package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func CheckInjectionExists(id int) (bool, error) {
	var injection database.FaultInjectionSchedule
	if err := database.DB.
		Where("id = ?", id).
		First(&injection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("failed to check injection: %v", err)
	}

	return true, nil
}

func DeleteDatasetByName(names []string) (int64, []string, error) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingNames []string
	if err := tx.Model(&database.FaultInjectionSchedule{}).
		Select("injection_name").
		Where("injection_name IN (?) AND status != ?", names, consts.DatapackDeleted).
		Pluck("injection_name", &existingNames).
		Error; err != nil {
		tx.Rollback()
		logrus.Errorf("failed to query existing records: %v", err)
		return 0, nil, fmt.Errorf("database operation query failed")
	}

	nonExisting := getMissingNames(names, existingNames)
	if len(nonExisting) > 0 {
		logrus.Warnf("Non-existing names: %v", nonExisting)
	}

	result := tx.Model(&database.FaultInjectionSchedule{}).
		Where("injection_name IN (?)", existingNames).
		Update("status", consts.DatapackDeleted)

	if result.Error != nil {
		tx.Rollback()
		logrus.Errorf("update failed: %v", result.Error)
		return 0, nil, fmt.Errorf("database operation update failed")
	}

	var failedUpdates []string
	if err := tx.Model(&database.FaultInjectionSchedule{}).
		Select("injection_name").
		Where("injection_name IN (?) AND status != ?", existingNames, consts.DatapackDeleted).
		Pluck("injection_name", &failedUpdates).
		Error; err != nil {
		tx.Rollback()
		logrus.Errorf("verification failed: %v", err)
		return 0, nil, fmt.Errorf("database operation query failed")
	}

	allFailed := utils.Union(nonExisting, failedUpdates)

	if err := tx.Commit().Error; err != nil {
		logrus.Errorf("commit failed: %v", err)
		return 0, nil, fmt.Errorf("database operation failed")
	}

	return result.RowsAffected, allFailed, nil
}

// Calculate difference set
func getMissingNames(requested []string, existing []string) []string {
	existingSet := make(map[string]struct{})
	for _, name := range existing {
		existingSet[name] = struct{}{}
	}

	var missing []string
	for _, name := range requested {
		if _, ok := existingSet[name]; !ok {
			missing = append(missing, name)
		}
	}

	return missing
}

func GetDatasetWithGroupIDs(groupIDs []string) ([]dto.DatasetJoinedResult, error) {
	var results []struct {
		GroupID string `gorm:"column:group_id"`
		Name    string `gorm:"column:injection_name"`
	}

	if err := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Joins("JOIN tasks ON tasks.id	 = fault_injection_schedules.task_id").
		Where("tasks.group_id IN ? AND fault_injection_schedules.status = ?", groupIDs, consts.DatapackBuildSuccess).
		Select("tasks.group_id, fault_injection_schedules.injection_name").
		Scan(&results).
		Error; err != nil {
		return nil, err
	}

	joinedResults := make([]dto.DatasetJoinedResult, 0, len(results))
	for _, r := range results {
		var joinedResult dto.DatasetJoinedResult
		joinedResult.Convert(r.GroupID, r.Name)
		joinedResults = append(joinedResults, joinedResult)
	}

	return joinedResults, nil
}

func GetDatasetByName(name string, status ...int) (*dto.DatasetItemWithID, error) {
	query := database.DB.Where("injection_name = ?", name)

	if len(status) == 0 {
		query = query.Where("status != ?", consts.DatapackDeleted)
	} else if len(status) == 1 {
		query = query.Where("status = ?", status[0])
	} else {
		query = query.Where("status IN ?", status)
	}

	var record database.FaultInjectionSchedule
	if err := query.First(&record).Error; err != nil {
		return nil, err
	}

	var item dto.DatasetItemWithID
	if err := item.Convert(record); err != nil {
		return nil, err
	}

	return &item, nil
}

func GetInjection(column, param string) (*database.FaultInjectionSchedule, error) {
	var record database.FaultInjectionSchedule
	if err := database.DB.
		Where(fmt.Sprintf("%s = ?", column), param).
		First(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

func ListDisplayConfigsByTraceIDs(traceIDs []string) (map[string]any, error) {
	query := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Select("tasks.trace_id, fault_injection_schedules.display_config").
		Joins("JOIN tasks ON tasks.id = fault_injection_schedules.task_id")

	if len(traceIDs) > 0 {
		query = query.Where("tasks.trace_id IN (?)", traceIDs)
	}

	var records []struct {
		TraceID       string `gorm:"column:trace_id"`
		DisplayConfig string `gorm:"column:display_config"`
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to query display configs: %v", err)
	}

	result := make(map[string]any, len(records))
	if len(traceIDs) > 0 {
		for _, traceID := range traceIDs {
			result[traceID] = nil
		}
	}

	for _, record := range records {
		var config map[string]any
		if err := json.Unmarshal([]byte(record.DisplayConfig), &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal display config for trace_id %s: %v", record.TraceID, err)
		}

		result[record.TraceID] = config
	}

	return result, nil
}

// ListExistingEngineConfigs lists engine_config strings that already exist in DB and are considered completed builds.
// This is used to de-duplicate incoming injection requests by their engine configuration.
// Excludes records that have the "invalid" label.
func ListExistingEngineConfigs(configs []string) ([]string, error) {
	if len(configs) == 0 {
		return []string{}, nil
	}

	query := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Select("engine_config").
		Where("engine_config in (?) AND status = ?", configs, consts.DatapackBuildSuccess)

	// Exclude records that have the "invalid" label
	invalidLabelSubQuery := database.DB.Table("fault_injection_labels").
		Select("fault_injection_id").
		Joins("JOIN labels ON labels.id = fault_injection_labels.label_id").
		Where("labels.label_key = ? AND labels.label_value = ?", consts.LabelKeyTag, "invalid")

	query = query.Where("fault_injection_schedules.id NOT IN (?)", invalidLabelSubQuery)

	var existingEngineConfigs []string
	if err := query.Pluck("engine_config", &existingEngineConfigs).Error; err != nil {
		return nil, err
	}

	return existingEngineConfigs, nil
}

func ListEngineConfigsByNames(names []string) (map[string]string, error) {
	var records []struct {
		InjectionName string `gorm:"column:injection_name"`
		EngineConfig  string `gorm:"column:engine_config"`
	}

	if err := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Select("injection_name, engine_config").
		Where("injection_name IN (?)", names).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to query engine configs: %v", err)
	}

	result := make(map[string]string, len(records))
	for _, record := range records {
		result[record.InjectionName] = record.EngineConfig
	}

	return result, nil
}

func ListInjections(params *dto.ListInjectionsReq) (int64, []database.FaultInjectionProject, error) {
	opts, err := params.TimeRangeQuery.Convert()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to convert time range query: %v", err)
	}

	builder := func(db *gorm.DB) *gorm.DB {
		query := db

		if params.ProjectName != "" {
			query = query.Where("project_name = ?", params.ProjectName)
		}

		if params.Env != "" {
			query = query.Where("env = ?", params.Env)
		}

		if params.Batch != "" {
			query = query.Where("batch = ?", params.Batch)
		}

		if params.Tag != "" {
			query = query.Where("tag=?", params.Tag)
		}

		if params.Benchmark != "" {
			query = query.Where("benchmark = ?", params.Benchmark)
		}

		if params.Status != nil {
			query = query.Where("status = ?", *params.Status)
		}

		if params.FaultType != nil {
			query = query.Where("fault_type = ?", *params.FaultType)
		}

		query = opts.AddTimeFilter(query, "created_at")
		return query
	}

	sortField := ""
	if params.SortField != "" && params.SortOrder != "" {
		sortField = fmt.Sprintf("%s %s", params.SortField, params.SortOrder)
	}

	genericQueryParams := &GenericQueryParams{
		Builder:   builder,
		SortField: sortField,
		Limit:     params.Limit,
		PageNum:   params.PageNum,
		PageSize:  params.PageSize,
	}
	return GenericQueryWithBuilder[database.FaultInjectionProject](genericQueryParams)
}

func UpdateStatusByDataset(name string, status int) error {
	return updateRecord(name, map[string]any{
		"status": status,
	})
}

func UpdateTimeByInjectionName(name string, startTime, endTime time.Time) error {
	return updateRecord(name, map[string]any{
		"start_time": startTime,
		"end_time":   endTime,
		"status":     consts.DatapackInjectSuccess,
	})
}

func updateRecord(name string, updates map[string]any) error {
	if len(updates) == 0 {
		return fmt.Errorf("empty update fields")
	}

	var record database.FaultInjectionSchedule
	err := database.DB.
		Where("injection_name = ?", name).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("record with name %q not found", name)
		}
		return fmt.Errorf("failed to query record: %v", err)
	}

	result := database.DB.
		Model(&record).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update record: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("record found but no fields were updated, possibly because values are unchanged")
	}

	return nil
}

func GetAllFaultInjectionNoIssues(params *dto.FaultInjectionNoIssuesReq) (int64, []database.FaultInjectionNoIssues, error) {
	opts, err := params.TimeRangeQuery.Convert()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to convert time range query: %v", err)
	}

	builder := func(db *gorm.DB) *gorm.DB {
		query := db

		// Directly use fields from view for query
		if params.Env != "" {
			query = query.Where("env = ?", params.Env)
		}

		if params.Batch != "" {
			query = query.Where("batch = ?", params.Batch)
		}

		query = opts.AddTimeFilter(query, "created_at")
		return query
	}

	genericQueryParams := &GenericQueryParams{
		Builder:       builder,
		SortField:     "dataset_id desc",
		SelectColumns: []string{},
	}
	return GenericQueryWithBuilder[database.FaultInjectionNoIssues](genericQueryParams)
}

func GetAllFaultInjectionWithIssues(params *dto.FaultInjectionWithIssuesReq) (int64, []database.FaultInjectionWithIssues, error) {
	opts, err := params.TimeRangeQuery.Convert()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to convert time range query: %v", err)
	}

	builder := func(db *gorm.DB) *gorm.DB {
		query := db

		// Directly use fields from view for query
		if params.Env != "" {
			query = query.Where("env = ?", params.Env)
		}

		if params.Batch != "" {
			query = query.Where("batch = ?", params.Batch)
		}

		query = opts.AddTimeFilter(query, "created_at")
		return query
	}

	genericQueryParams := &GenericQueryParams{
		Builder:       builder,
		SortField:     "dataset_id desc",
		SelectColumns: []string{},
	}
	return GenericQueryWithBuilder[database.FaultInjectionWithIssues](genericQueryParams)
}

func GetFLByDatasetName(datasetName string) (*database.FaultInjectionSchedule, error) {
	var record database.FaultInjectionSchedule
	if err := database.DB.Where("injection_name = ?", datasetName).First(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

func GetInjectionStats(req *dto.TimeRangeQuery) (map[string]int64, error) {
	opts, err := req.Convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert time range query: %v", err)
	}

	startTime, endTime := opts.GetTimeRange()

	var args []any
	args = append(args, startTime)
	args = append(args, endTime)

	timeCondition := "WHERE created_at >= ? AND created_at <= ?"
	sql := fmt.Sprintf(`
        SELECT 
            'no_issues' as type,
            COUNT(*) as record_count,
            COUNT(DISTINCT injection_name) as name_count
        FROM fault_injection_no_issues %s
        UNION ALL
        SELECT 
            'with_issues' as type,
            COUNT(*) as record_count,
            COUNT(DISTINCT injection_name) as name_count
        FROM fault_injection_with_issues %s
    `, timeCondition, timeCondition)

	var results []struct {
		Type        string `gorm:"column:type"`
		RecordCount int64  `gorm:"column:record_count"`
		NameCount   int64  `gorm:"column:name_count"`
	}

	allArgs := append(args, args...)
	if err := database.DB.Raw(sql, allArgs...).Scan(&results).Error; err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, result := range results {
		prefix := result.Type
		stats[prefix+"_records"] = result.RecordCount
		stats[prefix+"_injections"] = result.NameCount
	}

	return stats, nil
}

// GetInjectionDetailedStats gets detailed fault injection status statistics
func GetInjectionDetailedStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total injections (exclude deleted)
	var total int64
	if err := database.DB.Model(&database.FaultInjectionSchedule{}).
		Where("status != ?", consts.DatapackDeleted).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total injections: %v", err)
	}
	stats["total"] = total

	// Running injections (status = 1)
	var running int64
	if err := database.DB.Model(&database.FaultInjectionSchedule{}).
		Where("status = ?", consts.DatapackInitial).Count(&running).Error; err != nil {
		return nil, fmt.Errorf("failed to count running injections: %v", err)
	}
	stats["running"] = running

	// Completed injections (status = 2)
	var completed int64
	if err := database.DB.Model(&database.FaultInjectionSchedule{}).
		Where("status = ?", consts.DatapackInjectSuccess).Count(&completed).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed injections: %v", err)
	}
	stats["completed"] = completed

	// Failed injections (status = 3)
	var failed int64
	if err := database.DB.Model(&database.FaultInjectionSchedule{}).
		Where("status = ?", consts.DatapackInjectFailed).Count(&failed).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed injections: %v", err)
	}
	stats["failed"] = failed

	return stats, nil
}

// V2 API Repository Methods

// GetInjectionByIDV2 gets injection by ID for V2 API
func GetInjectionByIDV2(id int) (*database.FaultInjectionSchedule, error) {
	var injection database.FaultInjectionSchedule
	if err := database.DB.First(&injection, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("injection not found")
		}
		return nil, err
	}
	return &injection, nil
}

// GetInjectionByNameV2 gets injection by name for V2 API
func GetInjectionByNameV2(name string) (*database.FaultInjectionSchedule, error) {
	var injection database.FaultInjectionSchedule
	if err := database.DB.Where("injection_name = ? AND status != ?", name, consts.DatapackDeleted).First(&injection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("injection not found")
		}
		return nil, err
	}
	return &injection, nil
}

// GetInjectionsByIDsAndNames gets injections by IDs and names in batch
func GetInjectionsByIDsAndNames(ids []int, names []string) ([]database.FaultInjectionSchedule, error) {
	var injections []database.FaultInjectionSchedule

	query := database.DB.Model(&database.FaultInjectionSchedule{})

	// Build OR conditions for IDs and names
	var conditions []string
	var args []interface{}

	if len(ids) > 0 {
		conditions = append(conditions, "id IN ?")
		args = append(args, ids)
	}

	if len(names) > 0 {
		conditions = append(conditions, "injection_name IN ?")
		args = append(args, names)
	}

	if len(conditions) == 0 {
		return injections, nil
	}

	// Combine conditions with OR
	whereClause := strings.Join(conditions, " OR ")
	query = query.Where(whereClause, args...)

	if err := query.Find(&injections).Error; err != nil {
		return nil, err
	}

	return injections, nil
}

// ListInjectionsV2 lists injections with pagination and filtering
func ListInjectionsV2(page, size int, taskID string, faultType, status *int, benchmark, search string, tags []string) ([]database.FaultInjectionSchedule, int64, error) {
	query := database.DB.Model(&database.FaultInjectionSchedule{})
	tags = append(tags, "valid")

	// Apply filters
	if taskID != "" {
		query = query.Where("task_id = ?", taskID)
	}
	if faultType != nil {
		query = query.Where("fault_type = ?", *faultType)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if benchmark != "" {
		query = query.Where("benchmark = ?", benchmark)
	}
	if search != "" {
		query = query.Where("injection_name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Apply tags filter
	if len(tags) > 0 {
		// For each tag, add a condition that the injection must have that tag
		// This creates an AND condition - injection must have ALL specified tags
		for _, tag := range tags {
			subQuery := database.DB.Table("fault_injection_labels").
				Select("fault_injection_id").
				Joins("JOIN labels ON labels.id = fault_injection_labels.label_id").
				Where("labels.label_key = ? AND labels.label_value = ?", consts.LabelKeyTag, tag)

			query = query.Where("fault_injection_schedules.id IN (?)", subQuery)
		}
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * size
	query = query.Offset(offset).Limit(size).Order("id desc")

	// Execute query
	var injections []database.FaultInjectionSchedule
	if err := query.Find(&injections).Error; err != nil {
		return nil, 0, err
	}

	return injections, total, nil
}

// CreateInjectionsV2 creates multiple injections with label support
func CreateInjectionsV2(injections []dto.InjectionV2CreateItem) ([]database.FaultInjectionSchedule, []dto.InjectionCreateError, error) {
	var createdInjections []database.FaultInjectionSchedule
	var failedItems []dto.InjectionCreateError

	// Use transaction for batch creation
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i, item := range injections {
		injection := item.ToEntity()
		if err := tx.Create(&injection).Error; err != nil {
			failedItems = append(failedItems, dto.InjectionCreateError{
				Index: i,
				Error: fmt.Sprintf("Failed to create injection: %v", err),
				Item:  item,
			})
			continue
		}

		// Add label based on TaskID
		var labelValue string
		if item.TaskID == nil || *item.TaskID == "" {
			labelValue = consts.ExecutionSourceManual
		} else {
			labelValue = consts.ExecutionSourceSystem
		}

		// Add injection label (non-blocking)
		if err := AddInjectionLabelWithTx(tx, injection.ID, consts.ExecutionLabelSource, labelValue); err != nil {
			logrus.Warnf("Failed to add label to injection %d: %v", injection.ID, err)
		}

		createdInjections = append(createdInjections, injection)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return createdInjections, failedItems, nil
}

// AddInjectionLabelWithTx adds a label to an injection within a transaction
func AddInjectionLabelWithTx(tx *gorm.DB, injectionID int, key, value string) error {
	// Create or get label
	label, err := CreateOrGetLabel(key, value, consts.LabelSystem, "Injection source label")
	if err != nil {
		return fmt.Errorf("failed to create or get label: %v", err)
	}

	// Check if association already exists
	var count int64
	if err := tx.Model(&database.FaultInjectionLabel{}).
		Where("fault_injection_id = ? AND label_id = ?", injectionID, label.ID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing association: %v", err)
	}

	if count > 0 {
		return nil // Association already exists
	}

	// Create association
	injectionLabel := database.FaultInjectionLabel{
		FaultInjectionID: injectionID,
		LabelID:          label.ID,
	}

	if err := tx.Create(&injectionLabel).Error; err != nil {
		return fmt.Errorf("failed to create injection label association: %v", err)
	}

	return nil
}

// AddInjectionLabel adds a label to an injection (without transaction)
func AddInjectionLabel(injectionID int, key, value string) error {
	return AddInjectionLabelWithTx(database.DB, injectionID, key, value)
}

// GetInjectionLabels gets all labels for an injection
func GetInjectionLabels(injectionID int) ([]database.Label, error) {
	var labels []database.Label
	if err := database.DB.
		Joins("JOIN fault_injection_labels ON fault_injection_labels.label_id = labels.id").
		Where("fault_injection_labels.fault_injection_id = ?", injectionID).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get injection labels: %v", err)
	}
	return labels, nil
}

// AddLabelToInjection adds a label to injection by label ID
func AddLabelToInjection(injectionID, labelID int) error {
	// Check if association already exists
	var count int64
	if err := database.DB.Model(&database.FaultInjectionLabel{}).
		Where("fault_injection_id = ? AND label_id = ?", injectionID, labelID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing association: %v", err)
	}

	if count > 0 {
		return nil
	}

	// Use transaction to ensure atomicity
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	injectionLabel := &database.FaultInjectionLabel{
		FaultInjectionID: injectionID,
		LabelID:          labelID,
	}

	if err := tx.Create(injectionLabel).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to add label to injection: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// RemoveLabelFromInjection removes a label from injection by label ID
func RemoveLabelFromInjection(injectionID, labelID int) error {
	if err := database.DB.Where("fault_injection_id = ? AND label_id = ?", injectionID, labelID).
		Delete(&database.FaultInjectionLabel{}).Error; err != nil {
		return fmt.Errorf("failed to remove label from injection: %v", err)
	}

	if err := database.DB.Model(&database.Label{}).Where("id = ?", labelID).
		UpdateColumn("usage_count", gorm.Expr("GREATEST(0, usage_count - 1)")).Error; err != nil {
		return fmt.Errorf("failed to update label usage: %v", err)
	}

	return nil
}

// AddTagToInjection adds a tag to injection
func AddTagToInjection(injectionID int, tagValue string) error {
	// Create or get label with key "tag"
	label, err := CreateOrGetLabel(consts.LabelKeyTag, tagValue, consts.LabelInjection, "Injection tag")
	if err != nil {
		return fmt.Errorf("failed to create or get tag: %v", err)
	}

	return AddLabelToInjection(injectionID, label.ID)
}

// RemoveTagFromInjection removes a tag from injection
func RemoveTagFromInjection(injectionID int, tagValue string) error {
	label, err := GetLabelByKeyandValue(consts.LabelKeyTag, tagValue)
	if err != nil {
		return fmt.Errorf("failed to get tag '%s': %v", tagValue, err)
	}

	return RemoveLabelFromInjection(injectionID, label.ID)
}

// Helper functions

// UpdateInjectionV2 updates injection by ID for V2 API
func UpdateInjectionV2(id int, updates map[string]any) error {
	result := database.DB.Model(&database.FaultInjectionSchedule{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("injection not found or no changes made")
	}
	return nil
}

// DeleteInjectionV2 soft deletes injection by ID for V2 API
func DeleteInjectionV2(id int) error {
	result := database.DB.Model(&database.FaultInjectionSchedule{}).Where("id = ?", id).Update("status", consts.DatapackDeleted)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("injection not found")
	}
	return nil
}

// SearchInjectionsV2 performs advanced search on injections
func SearchInjectionsV2(req *dto.InjectionV2SearchReq) ([]database.FaultInjectionSchedule, int64, error) {
	query := database.DB.Model(&database.FaultInjectionSchedule{})

	// Apply filters
	if len(req.TaskIDs) > 0 {
		query = query.Where("task_id IN ?", req.TaskIDs)
	}
	if len(req.FaultTypes) > 0 {
		query = query.Where("fault_type IN ?", req.FaultTypes)
	}
	if len(req.Statuses) > 0 {
		query = query.Where("status IN ?", req.Statuses)
	}
	if len(req.Benchmarks) > 0 {
		query = query.Where("benchmark IN ?", req.Benchmarks)
	}
	if req.Search != "" {
		query = query.Where("injection_name LIKE ? OR description LIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}

	// Apply tags filter
	if len(req.Tags) > 0 {
		// Join with fault_injection_labels and labels tables to filter by tags
		query = query.Joins("JOIN fault_injection_labels ON fault_injection_labels.fault_injection_id = fault_injection_schedules.id").
			Joins("JOIN labels ON labels.id = fault_injection_labels.label_id").
			Where("labels.label_key = ? AND labels.label_value IN ?", consts.LabelKeyTag, req.Tags).
			Group("fault_injection_schedules.id")
	}

	// Apply custom labels filter
	if len(req.Labels) > 0 {
		// Build subquery for custom labels
		for i, labelItem := range req.Labels {
			subQuery := database.DB.Table("fault_injection_labels fil"+fmt.Sprintf("%d", i)).
				Select("fil"+fmt.Sprintf("%d", i)+".fault_injection_id").
				Joins("JOIN labels l"+fmt.Sprintf("%d", i)+" ON l"+fmt.Sprintf("%d", i)+".id = fil"+fmt.Sprintf("%d", i)+".label_id").
				Where("l"+fmt.Sprintf("%d", i)+".label_key = ?", labelItem.Key)

			if labelItem.Value != "" {
				subQuery = subQuery.Where("l"+fmt.Sprintf("%d", i)+".label_value = ?", labelItem.Value)
			}

			query = query.Where("fault_injection_schedules.id IN (?)", subQuery)
		}
	}

	// Time range filters
	if req.StartTimeGte != nil {
		query = query.Where("start_time >= ?", *req.StartTimeGte)
	}
	if req.StartTimeLte != nil {
		query = query.Where("start_time <= ?", *req.StartTimeLte)
	}
	if req.EndTimeGte != nil {
		query = query.Where("end_time >= ?", *req.EndTimeGte)
	}
	if req.EndTimeLte != nil {
		query = query.Where("end_time <= ?", *req.EndTimeLte)
	}
	if req.CreatedAtGte != nil {
		query = query.Where("created_at >= ?", *req.CreatedAtGte)
	}
	if req.CreatedAtLte != nil {
		query = query.Where("created_at <= ?", *req.CreatedAtLte)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	if req.SortBy != "" && req.SortOrder != "" {
		query = query.Order(fmt.Sprintf("%s %s", req.SortBy, req.SortOrder))
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if req.Page != nil && req.Size != nil {
		page := *req.Page
		size := *req.Size
		if page > 0 && size > 0 {
			offset := (page - 1) * size
			query = query.Offset(offset).Limit(size)
		}
	}

	// Execute query
	var injections []database.FaultInjectionSchedule
	if err := query.Find(&injections).Error; err != nil {
		return nil, 0, err
	}

	return injections, total, nil
}

// AddCustomLabelToInjection adds a custom label to injection
func AddCustomLabelToInjection(injectionID int, key, value string) error {
	// Generate color based on key
	color := utils.GenerateColorFromKey(key)

	// Generate description with creation info
	description := fmt.Sprintf(consts.CustomLabelDescriptionTemplate, key)

	// Create or get custom label with injection category
	label, err := CreateOrGetLabel(key, value, consts.LabelInjection, description)
	if err != nil {
		return fmt.Errorf("failed to create or get custom label: %v", err)
	}

	// Update color for the label
	if err := database.DB.Model(&database.Label{}).Where("id = ?", label.ID).Update("color", color).Error; err != nil {
		return fmt.Errorf("failed to update label color: %v", err)
	}

	return AddLabelToInjection(injectionID, label.ID)
}

// RemoveCustomLabelFromInjection removes a custom label from injection by key
func RemoveCustomLabelFromInjection(injectionID int, key string) error {
	// Find all labels with the given key for this injection
	var labels []database.Label
	if err := database.DB.
		Joins("JOIN fault_injection_labels ON fault_injection_labels.label_id = labels.id").
		Where("fault_injection_labels.fault_injection_id = ? AND labels.label_key = ?", injectionID, key).
		Find(&labels).Error; err != nil {
		return fmt.Errorf("failed to find labels with key '%s': %v", key, err)
	}

	if len(labels) == 0 {
		return fmt.Errorf("no label found with key '%s' for injection %d", key, injectionID)
	}

	// Remove all labels with this key (in case there are multiple with same key but different values)
	for _, label := range labels {
		if err := RemoveLabelFromInjection(injectionID, label.ID); err != nil {
			return fmt.Errorf("failed to remove label %d: %v", label.ID, err)
		}
	}

	return nil
}

// GetInjectionLabelsMap gets all labels for multiple injections in batch (optimized)
func GetInjectionLabelsMap(injectionIDs []int) (map[int][]database.Label, error) {
	if len(injectionIDs) == 0 {
		return make(map[int][]database.Label), nil
	}

	// Method 1: Use single query with direct model association
	var relations []database.FaultInjectionLabel
	if err := database.DB.Preload("Label").
		Where("fault_injection_id IN ?", injectionIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get injection label relations: %v", err)
	}

	// Group labels by injection ID
	labelsMap := make(map[int][]database.Label)
	for _, relation := range relations {
		if relation.Label != nil {
			labelsMap[relation.FaultInjectionID] = append(labelsMap[relation.FaultInjectionID], *relation.Label)
		}
	}

	// Initialize empty slices for injections with no labels
	for _, id := range injectionIDs {
		if _, exists := labelsMap[id]; !exists {
			labelsMap[id] = []database.Label{}
		}
	}

	return labelsMap, nil
}

// AddCustomLabelToInjectionWithOverride adds a custom label to injection with override behavior
// If a label with the same key already exists, it will be removed first, then the new label will be added
func AddCustomLabelToInjectionWithOverride(injectionID int, key, value string) error {
	// First, remove any existing labels with the same key
	if err := RemoveCustomLabelFromInjection(injectionID, key); err != nil {
		// If no label found, that's fine - we can proceed
		if !strings.Contains(err.Error(), "no label found") {
			return fmt.Errorf("failed to remove existing label with key '%s': %v", key, err)
		}
	}

	// Generate color based on key
	color := utils.GenerateColorFromKey(key)

	// Generate description with creation info
	description := fmt.Sprintf(consts.CustomLabelDescriptionTemplate, key)

	// Create or get custom label with injection category
	label, err := CreateOrGetLabel(key, value, consts.LabelInjection, description)
	if err != nil {
		return fmt.Errorf("failed to create or get custom label: %v", err)
	}

	// Update color for the label
	if err := database.DB.Model(&database.Label{}).Where("id = ?", label.ID).Update("color", color).Error; err != nil {
		return fmt.Errorf("failed to update label color: %v", err)
	}

	return AddLabelToInjection(injectionID, label.ID)
}

// BatchDeleteInjectionsV2 performs batch deletion of injections with cascading deletes
func BatchDeleteInjectionsV2(injectionIDs []int) (dto.InjectionV2BatchDeleteResponse, error) {
	var response dto.InjectionV2BatchDeleteResponse
	var successItems []dto.InjectionV2DeletedItem
	var failedItems []dto.InjectionV2DeleteError

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Track cascade delete statistics
	var cascadeStats dto.InjectionV2CascadeDeleteStats

	for _, injectionID := range injectionIDs {
		// Get injection details before deletion
		var injection database.FaultInjectionSchedule
		if err := tx.First(&injection, injectionID).Error; err != nil {
			failedItems = append(failedItems, dto.InjectionV2DeleteError{
				ID:    injectionID,
				Error: fmt.Sprintf("injection not found: %v", err),
			})
			continue
		}

		// Perform cascading deletes
		if err := performCascadeDelete(tx, injectionID, &cascadeStats); err != nil {
			failedItems = append(failedItems, dto.InjectionV2DeleteError{
				ID:            injectionID,
				InjectionName: injection.InjectionName,
				Error:         fmt.Sprintf("cascade delete failed: %v", err),
			})
			continue
		}

		// Soft delete the injection itself
		if err := tx.Model(&database.FaultInjectionSchedule{}).
			Where("id = ?", injectionID).
			Update("status", consts.DatapackDeleted).Error; err != nil {
			failedItems = append(failedItems, dto.InjectionV2DeleteError{
				ID:            injectionID,
				InjectionName: injection.InjectionName,
				Error:         fmt.Sprintf("failed to delete injection: %v", err),
			})
			continue
		}

		successItems = append(successItems, dto.InjectionV2DeletedItem{
			ID:            injectionID,
			InjectionName: injection.InjectionName,
			Benchmark:     injection.Benchmark,
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return response, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Build response
	message := fmt.Sprintf("Successfully deleted %d injection(s)", len(successItems))
	if len(failedItems) > 0 {
		message += fmt.Sprintf(", %d failed", len(failedItems))
	}

	response = dto.InjectionV2BatchDeleteResponse{
		SuccessCount:   len(successItems),
		SuccessItems:   successItems,
		FailedCount:    len(failedItems),
		FailedItems:    failedItems,
		CascadeDeleted: cascadeStats,
		Message:        message,
	}

	return response, nil
}

// BatchDeleteInjectionsByLabelsV2 performs batch deletion of injections by labels with cascading deletes
func BatchDeleteInjectionsByLabelsV2(labelFilters []string) (dto.InjectionV2BatchDeleteResponse, error) {
	var response dto.InjectionV2BatchDeleteResponse

	// Parse label filters
	var labelConditions []map[string]string
	for _, labelFilter := range labelFilters {
		parts := strings.SplitN(labelFilter, ":", 2)
		if len(parts) != 2 {
			return response, fmt.Errorf("invalid label format: %s, expected 'key:value'", labelFilter)
		}
		labelConditions = append(labelConditions, map[string]string{
			"key":   parts[0],
			"value": parts[1],
		})
	}

	// Find injections matching the labels
	var injectionIDs []int
	query := database.DB.Model(&database.FaultInjectionSchedule{}).
		Select("DISTINCT fault_injection_schedules.id").
		Joins("JOIN fault_injection_labels ON fault_injection_labels.fault_injection_id = fault_injection_schedules.id").
		Joins("JOIN labels ON labels.id = fault_injection_labels.label_id").
		Where("fault_injection_schedules.status != ?", consts.DatapackDeleted) // Exclude already deleted

	// Build WHERE condition for labels
	var whereClauses []string
	var whereArgs []interface{}

	for _, condition := range labelConditions {
		whereClauses = append(whereClauses, "(labels.label_key = ? AND labels.label_value = ?)")
		whereArgs = append(whereArgs, condition["key"], condition["value"])
	}

	if len(whereClauses) > 0 {
		whereClause := strings.Join(whereClauses, " OR ")
		query = query.Where(whereClause, whereArgs...)
	}

	if err := query.Pluck("fault_injection_schedules.id", &injectionIDs).Error; err != nil {
		return response, fmt.Errorf("failed to find injections by labels: %v", err)
	}

	if len(injectionIDs) == 0 {
		response.Message = "No injections found matching the specified labels"
		return response, nil
	}

	// Perform batch deletion
	return BatchDeleteInjectionsV2(injectionIDs)
}

// performCascadeDelete handles cascade deletion of related records
func performCascadeDelete(tx *gorm.DB, injectionID int, stats *dto.InjectionV2CascadeDeleteStats) error {
	// First, get all execution_result IDs that belong to this injection
	var executionResultIDs []int
	if err := tx.Model(&database.ExecutionResult{}).
		Where("datapack_id = ?", injectionID).
		Pluck("id", &executionResultIDs).Error; err != nil {
		return fmt.Errorf("failed to get execution result IDs: %v", err)
	}

	if len(executionResultIDs) == 0 {
		result := tx.Where("fault_injection_id = ?", injectionID).Delete(&database.FaultInjectionLabel{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete fault_injection_labels: %v", result.Error)
		}
		stats.FaultInjectionLabels += int(result.RowsAffected)

		// Delete dataset_fault_injections
		result = tx.Where("fault_injection_id = ?", injectionID).Delete(&database.DatasetFaultInjection{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete dataset_fault_injections: %v", result.Error)
		}
		stats.DatasetFaultInjections += int(result.RowsAffected)

		return nil
	}

	// Count what will be deleted before actual deletion for accurate counting
	var granularityCount int64
	tx.Model(&database.GranularityResult{}).
		Where("execution_id IN ?", executionResultIDs).
		Count(&granularityCount)
	stats.GranularityResults += int(granularityCount)

	var executionResultLabelsCount int64
	tx.Table("execution_result_labels").
		Where("execution_id IN ?", executionResultIDs).
		Count(&executionResultLabelsCount)
	stats.ExecutionResultLabels += int(executionResultLabelsCount)

	var detectorsCount int64
	tx.Model(&database.Detector{}).
		Where("execution_id IN ?", executionResultIDs).
		Count(&detectorsCount)
	stats.Detectors += int(detectorsCount)

	// Step 1: Delete granularity_results (child table)
	if err := tx.Where("execution_id IN ?", executionResultIDs).Delete(&database.GranularityResult{}).Error; err != nil {
		return fmt.Errorf("failed to delete granularity_results: %v", err)
	}

	// Step 2: Delete detectors (child table)
	if err := tx.Where("execution_id IN ?", executionResultIDs).Delete(&database.Detector{}).Error; err != nil {
		return fmt.Errorf("failed to delete detectors: %v", err)
	}

	// Step 3: Delete execution_result_labels (many-to-many relationship table)
	if err := tx.Table("execution_result_labels").
		Where("execution_id IN ?", executionResultIDs).
		Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to delete execution_result_labels: %v", err)
	}

	// Step 4: Now we can safely delete execution_results
	result := tx.Where("datapack_id = ?", injectionID).Delete(&database.ExecutionResult{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete execution_results: %v", result.Error)
	}
	stats.ExecutionResults += int(result.RowsAffected)

	// Step 5: Delete fault_injection_labels
	result = tx.Where("fault_injection_id = ?", injectionID).Delete(&database.FaultInjectionLabel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete fault_injection_labels: %v", result.Error)
	}
	stats.FaultInjectionLabels += int(result.RowsAffected)

	// Step 6: Delete dataset_fault_injections
	result = tx.Where("fault_injection_id = ?", injectionID).Delete(&database.DatasetFaultInjection{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete dataset_fault_injections: %v", result.Error)
	}
	stats.DatasetFaultInjections += int(result.RowsAffected)

	return nil
}
