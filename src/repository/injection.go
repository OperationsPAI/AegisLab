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

// TODO
func ListExistingDisplayConfigs(configs []string) ([]string, error) {
	query := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Select("display_config").
		Where("display_config in (?) AND status = ?", configs, consts.DatapackBuildSuccess)

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

func ListDatasetByExecutionIDs(executionIDs []int) ([]dto.DatasetItemWithID, error) {
	query := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Joins("JOIN execution_results ON execution_results.dataset = fault_injection_schedules.id")

	if len(executionIDs) > 0 {
		query = query.Where("execution_results.id IN (?)", executionIDs)
	}

	var injections []database.FaultInjectionSchedule
	if err := query.Find(&injections).Error; err != nil {
		return nil, err
	}

	items := make([]dto.DatasetItemWithID, 0, len(injections))
	for _, injection := range injections {
		var item dto.DatasetItemWithID
		if err := item.Convert(injection); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
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
		Where("status != -1").Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total injections: %v", err)
	}
	stats["total"] = total

	// Running injections (status = 1)
	var running int64
	if err := database.DB.Model(&database.FaultInjectionSchedule{}).
		Where("status = 1").Count(&running).Error; err != nil {
		return nil, fmt.Errorf("failed to count running injections: %v", err)
	}
	stats["running"] = running

	// Completed injections (status = 2)
	var completed int64
	if err := database.DB.Model(&database.FaultInjectionSchedule{}).
		Where("status = 2").Count(&completed).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed injections: %v", err)
	}
	stats["completed"] = completed

	// Failed injections (status = 3)
	var failed int64
	if err := database.DB.Model(&database.FaultInjectionSchedule{}).
		Where("status = 3").Count(&failed).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed injections: %v", err)
	}
	stats["failed"] = failed

	// Scheduled injections (status = 0)
	var scheduled int64
	if err := database.DB.Model(&database.FaultInjectionSchedule{}).
		Where("status = 0").Count(&scheduled).Error; err != nil {
		return nil, fmt.Errorf("failed to count scheduled injections: %v", err)
	}
	stats["scheduled"] = scheduled

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
func ListInjectionsV2(page, size int, taskID string, faultType, status *int, benchmark, search string) ([]database.FaultInjectionSchedule, int64, error) {
	query := database.DB.Model(&database.FaultInjectionSchedule{})

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

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * size
	query = query.Offset(offset).Limit(size)

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
		// Create injection record
		injection := database.FaultInjectionSchedule{
			TaskID:        utils.GetStringValue(item.TaskID, ""),
			FaultType:     item.FaultType,
			DisplayConfig: item.DisplayConfig,
			EngineConfig:  item.EngineConfig,
			PreDuration:   item.PreDuration,
			StartTime:     utils.GetTimeValue(item.StartTime, time.Now()),
			EndTime:       utils.GetTimeValue(item.EndTime, time.Now().Add(time.Hour)),
			Status:        utils.GetIntValue(&item.Status, 0),
			Description:   item.Description,
			Benchmark:     item.Benchmark,
			InjectionName: item.InjectionName,
		}

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
	label, err := CreateOrGetLabel(key, value, "system", "Injection source label")
	if err != nil {
		return fmt.Errorf("failed to create or get label: %v", err)
	}

	// Check if association already exists
	var count int64
	if err := tx.Model(&database.InjectionLabel{}).
		Where("injection_id = ? AND label_id = ?", injectionID, label.ID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing association: %v", err)
	}

	if count > 0 {
		return nil // Association already exists
	}

	// Create association
	injectionLabel := database.InjectionLabel{
		InjectionID: injectionID,
		LabelID:     label.ID,
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
		Joins("JOIN injection_labels ON injection_labels.label_id = labels.id").
		Where("injection_labels.injection_id = ?", injectionID).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get injection labels: %v", err)
	}
	return labels, nil
}

// RemoveInjectionLabel removes a specific label from an injection
func RemoveInjectionLabel(injectionID int, labelKey, labelValue string) error {
	return database.DB.
		Where("injection_id = ? AND label_id IN (SELECT id FROM labels WHERE key = ? AND value = ?)",
			injectionID, labelKey, labelValue).
		Delete(&database.InjectionLabel{}).Error
}

// Helper functions

// UpdateInjectionV2 updates injection by ID for V2 API
func UpdateInjectionV2(id int, updates map[string]interface{}) error {
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
	result := database.DB.Model(&database.FaultInjectionSchedule{}).Where("id = ?", id).Update("status", -1)
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
	if req.Page > 0 && req.Size > 0 {
		offset := (req.Page - 1) * req.Size
		query = query.Offset(offset).Limit(req.Size)
	}

	// Execute query
	var injections []database.FaultInjectionSchedule
	if err := query.Find(&injections).Error; err != nil {
		return nil, 0, err
	}

	return injections, total, nil
}
