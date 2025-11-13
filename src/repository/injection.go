package repository

import (
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/database"
	"aegis/dto"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// =====================================================================
// Injection Repository Functions
// =====================================================================

// BatchDelteInjections marks multiple injections as deleted in batch
func BatchDeleteInjections(db *gorm.DB, injectionIDs []int) error {
	if len(injectionIDs) == 0 {
		return nil
	}

	if err := db.Model(&database.FaultInjection{}).
		Where("id IN (?) AND status != ?", injectionIDs, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to batch delete injections: %w", err)
	}

	return nil
}

// CreateInjection creates a fault injection record
func CreateInjection(db *gorm.DB, injection *database.FaultInjection) error {
	if err := db.Omit(commonOmitFields).Create(injection).Error; err != nil {
		return fmt.Errorf("failed to create injection: %w", err)
	}
	return nil
}

// GetInjectionByID gets injection by ID with preloaded associations
func GetInjectionByID(db *gorm.DB, id int) (*database.FaultInjection, error) {
	var injection database.FaultInjection
	if err := db.
		Preload("Task").
		Where("id = ?", id).First(&injection).Error; err != nil {
		return nil, fmt.Errorf("failed to find injection with id %d: %w", id, err)
	}
	return &injection, nil
}

// GetInjectionByName gets injection by name with preloaded associations
func GetInjectionByName(db *gorm.DB, name string) (*database.FaultInjection, error) {
	var injection database.FaultInjection
	if err := db.
		Where("name = ? AND status != ?", name, consts.CommonDeleted).First(&injection).Error; err != nil {
		return nil, fmt.Errorf("failed to find injection with name %s: %w", name, err)
	}
	return &injection, nil
}

// ListFaultInjectionsByID retrieves multiple fault injections by their IDs with preloaded associations
func ListFaultInjectionsByID(db *gorm.DB, injectionIDs []int) ([]database.FaultInjection, error) {
	if len(injectionIDs) == 0 {
		return []database.FaultInjection{}, nil
	}

	var injections []database.FaultInjection
	if err := db.
		Preload("Benchmark.Container").
		Preload("Pedestal.Container").
		Preload("Task.Project").
		Preload("Labels").
		Where("id IN (?) AND status != ?", injectionIDs, consts.CommonDeleted).
		Find(&injections).Error; err != nil {
		return nil, fmt.Errorf("failed to query fault injections: %w", err)
	}
	return injections, nil
}

// ListExistingEngineConfigs lists engine_config strings that already exist in DB and are considered completed builds.
// This is used to de-duplicate incoming injection requests by their engine configuration.
// Excludes records that have the "invalid" label.
func ListExistingEngineConfigs(configs []string) ([]string, error) {
	if len(configs) == 0 {
		return []string{}, nil
	}

	query := database.DB.
		Model(&database.FaultInjection{}).
		Select("engine_config").
		Where("engine_config in (?) AND status = ?", configs, consts.DatapackBuildSuccess)

	invalidLabelSubQuery := database.DB.Table("task_labels tl").
		Select("fis.id").
		Joins("JOIN fault_injection_schedules fis ON fis.task_id = tl.task_id").
		Joins("JOIN labels ON labels.id = tl.label_id").
		Where("labels.label_key = ? AND labels.label_value = ?", consts.LabelKeyTag, "invalid")

	query = query.Where("fault_injection_schedules.id NOT IN (?)", invalidLabelSubQuery)

	var existingEngineConfigs []string
	if err := query.Pluck("engine_config", &existingEngineConfigs).Error; err != nil {
		return nil, err
	}

	return existingEngineConfigs, nil
}

// ListEngineConfigByNames retrieves engine configurations by injection names
func ListEngineConfigByNames(db *gorm.DB, names []string) (map[string]string, error) {
	var records []struct {
		Name         string `gorm:"column:name"`
		EngineConfig string `gorm:"column:engine_config"`
	}

	if err := database.DB.
		Model(&database.FaultInjection{}).
		Select("name, engine_config").
		Where("name IN (?)", names).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to query engine configs: %v", err)
	}

	result := make(map[string]string, len(records))
	for _, record := range records {
		result[record.Name] = record.EngineConfig
	}

	return result, nil
}

// ListInjections lists fault injections based on filter options with preloaded associations
func ListInjections(db *gorm.DB, limit, offset int, filterOptions *dto.ListInjectionFilters) ([]database.FaultInjection, int64, error) {
	var injections []database.FaultInjection
	var total int64

	query := db.Model(&database.FaultInjection{}).
		Preload("Benchmark.Container").
		Preload("Pedestal.Container").
		Preload("Task.Project").
		Preload("Labels")
	if filterOptions.FaultType != nil {
		query = query.Where("fault_type = ?", *filterOptions.FaultType)
	}
	if filterOptions.Benchmark != "" {
		query = query.Where("benchmark = ?", filterOptions.Benchmark)
	}
	if filterOptions.State != nil {
		query = query.Where("event = ?", *filterOptions.State)
	}
	if filterOptions.Status != nil {
		query = query.Where("status = ?", *filterOptions.Status)
	}

	if len(filterOptions.LabelConditons) > 0 {
		for _, condition := range filterOptions.LabelConditons {
			subQuery := db.Table("task_labels tl").
				Select("fis.id").
				Joins("JOIN fault_injection_schedules fis ON fis.task_id = tl.task_id").
				Joins("JOIN labels ON labels.id = tl.label_id").
				Where("labels.label_key = ? AND labels.label_value = ?", condition["key"], condition["value"])

			query = query.Where("fault_injection_schedules.id IN (?)", subQuery)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count injections: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("updated_at DESC").Find(&injections).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list injections: %w", err)
	}

	return injections, total, nil
}

// SearchInjectionsV2 performs advanced search on injections
func SearchInjectionsV2(req *dto.SearchInjectionReq) ([]database.FaultInjection, int64, error) {
	query := database.DB.Model(&database.FaultInjection{}).
		Preload("Benchmark.Container").
		Preload("Pedestal.Container").
		Preload("Task.Project").
		Preload("Labels")

	// Apply filters
	if len(req.TaskIDs) > 0 {
		query = query.Where("task_id IN (?)", req.TaskIDs)
	}
	if len(req.FaultTypes) > 0 {
		query = query.Where("fault_type IN (?)", req.FaultTypes)
	}
	if len(req.Statuses) > 0 {
		query = query.Where("status IN (?)", req.Statuses)
	}
	if len(req.Benchmarks) > 0 {
		query = query.Where("benchmark IN (?)", req.Benchmarks)
	}
	if req.Search != "" {
		query = query.Where("injection_name LIKE ? OR description LIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}

	// Apply tags filter
	if len(req.Tags) > 0 {
		// Join with task_labels and labels tables to filter by tags
		query = query.Joins("JOIN task_labels tl ON tl.task_id = fault_injection_schedules.task_id").
			Joins("JOIN labels ON labels.id = tl.label_id").
			Where("labels.label_key = ? AND labels.label_value IN (?)", consts.LabelKeyTag, req.Tags).
			Group("fault_injection_schedules.id")
	}

	// Apply custom labels filter
	if len(req.Labels) > 0 {
		// Build subquery for custom labels
		for i, labelItem := range req.Labels {
			subQuery := database.DB.Table("task_labels tl"+fmt.Sprintf("%d", i)).
				Select("fis"+fmt.Sprintf("%d", i)+".id").
				Joins("JOIN fault_injection_schedules fis"+fmt.Sprintf("%d", i)+" ON fis"+fmt.Sprintf("%d", i)+".task_id = tl"+fmt.Sprintf("%d", i)+".task_id").
				Joins("JOIN labels l"+fmt.Sprintf("%d", i)+" ON l"+fmt.Sprintf("%d", i)+".id = tl"+fmt.Sprintf("%d", i)+".label_id").
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
	var injections []database.FaultInjection
	if err := query.Find(&injections).Error; err != nil {
		return nil, 0, err
	}

	return injections, total, nil
}

// UpdateInjection updates fields of a fault injection record
func UpdateInjection(db *gorm.DB, id int, updates map[string]any) error {
	result := db.Model(&database.FaultInjection{}).
		Where("id = ? AND status != ?", id, consts.CommonDeleted).
		Updates(updates)
	if err := result.Error; err != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("injection not found or no changes made")
	}
	return nil
}

// TODO 修改
// ListInjectionsNoIssues lists fault injections without issues based on label conditions and time range
func ListInjectionsNoIssues(db *gorm.DB, labelConditions []map[string]string, startTime, endTime *time.Time) ([]database.FaultInjectionNoIssues, error) {
	query := db.Model(&database.FaultInjectionNoIssues{}).Scopes(database.Sort("dataset_id desc"))
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	if len(labelConditions) > 0 {
		var whereConditions *gorm.DB
		for _, condition := range labelConditions {
			if whereConditions == nil {
				whereConditions = db.Where("label_key = ? AND label_value = ?", condition["key"], condition["value"])
			} else {
				whereConditions = whereConditions.Or("label_key = ? AND label_value = ?", condition["key"], condition["value"])
			}
		}

		if whereConditions != nil {
			query = query.Where(whereConditions)
		}

		query = query.
			Group("id").
			Having("COUNT(id) = ?", len(labelConditions))
	}

	var records []database.FaultInjectionNoIssues
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to query fault injections without issues: %v", err)
	}

	return records, nil
}

// ListInjectionsWithIssues lists fault injections with issues based on label conditions and time range
func ListInjectionsWithIssues(db *gorm.DB, labelConditions []map[string]string, startTime, endTime *time.Time) ([]database.FaultInjectionWithIssues, error) {
	query := db.Model(&database.FaultInjectionNoIssues{}).Scopes(database.Sort("dataset_id desc"))
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	if len(labelConditions) > 0 {
		var whereConditions *gorm.DB
		for _, condition := range labelConditions {
			if whereConditions == nil {
				whereConditions = db.Where("label_key = ? AND label_value = ?", condition["key"], condition["value"])
			} else {
				whereConditions = whereConditions.Or("label_key = ? AND label_value = ?", condition["key"], condition["value"])
			}
		}

		if whereConditions != nil {
			query = query.Where(whereConditions)
		}

		query = query.
			Group("id").
			Having("COUNT(id) = ?", len(labelConditions))
	}

	var records []database.FaultInjectionWithIssues
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to query fault injections without issues: %v", err)
	}

	return records, nil
}

// GetInjectionCountMapGroupByState gets a map of injection counts grouped by their state
func GetInjectionCountMapGroupByState(db *gorm.DB) (map[consts.DatapackState]int64, error) {
	type stateCount struct {
		state consts.DatapackState
		count int64
	}

	var results []stateCount
	if err := db.Model(&database.FaultInjection{}).
		Select("state, count(*) as count").
		Where("status != ?", consts.CommonDeleted).
		Group("state").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get detailed injection stats: %v", err)
	}

	stats := make(map[consts.DatapackState]int64)
	for _, res := range results {
		stats[res.state] = res.count
	}

	return stats, nil
}

// =====================================================================
// InjectionLabel Repository Functions
// =====================================================================

// Business layer: Injection labels are stored as TaskLabel in database
// Since Injection and Task are 1:1 relationship

// AddInjectionLabels adds multiple injection-label associations via TaskLabel
func AddInjectionLabels(db *gorm.DB, injectionID int, labelIDs []int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	// Get the TaskID for this injection
	var injection database.FaultInjection
	if err := db.Select("task_id").First(&injection, injectionID).Error; err != nil {
		return fmt.Errorf("failed to get injection task_id: %w", err)
	}

	// Create TaskLabel associations
	taskLabels := make([]database.TaskLabel, 0, len(labelIDs))
	for _, labelID := range labelIDs {
		taskLabels = append(taskLabels, database.TaskLabel{
			TaskID:  injection.TaskID,
			LabelID: labelID,
		})
	}

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "task_id"}, {Name: "label_id"}},
		DoNothing: true,
	}).Create(&taskLabels).Error; err != nil {
		return fmt.Errorf("failed to add injection-label associations: %w", err)
	}

	return nil
}

// ClearInjectionLabels removes label associations from specified fault injections via TaskLabel
func ClearInjectionLabels(db *gorm.DB, injectionIDs []int, labelIDs []int) error {
	if len(injectionIDs) == 0 {
		return nil
	}

	// Use subquery to delete in a single operation
	subQuery := db.Model(&database.FaultInjection{}).
		Select("task_id").
		Where("id IN (?)", injectionIDs)

	query := db.Table("task_labels").
		Where("task_id IN (?)", subQuery)
	if len(labelIDs) > 0 {
		query = query.Where("label_id IN (?)", labelIDs)
	}

	if err := query.Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to clear injection labels: %w", err)
	}
	return nil
}

// RemoveInjectionsFromLabel removes all injection-label associations for a specific label
// This removes TaskLabel entries for all Injections that are associated with this label
func RemoveInjectionsFromLabel(db *gorm.DB, labelID int) (int64, error) {
	subQuery := db.Table("task_labels tl").
		Select("DISTINCT tl.task_id").
		Joins("JOIN fault_injection_schedules fis ON fis.task_id = tl.task_id").
		Where("tl.label_id = ?", labelID)

	result := db.Where("task_id IN (?) AND label_id = ?", subQuery, labelID).
		Delete(&database.TaskLabel{})
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to remove injection-label associations for label %d: %w", labelID, err)
	}

	return result.RowsAffected, nil
}

// RemoveInjectionsFromLabels removes all injection-label associations for multiple labels
func RemoveInjectionsFromLabels(db *gorm.DB, labelIDs []int) (int64, error) {
	if len(labelIDs) == 0 {
		return 0, nil
	}

	subQuery := db.Table("task_labels tl").
		Select("DISTINCT tl.task_id").
		Joins("JOIN fault_injection_schedules fis ON fis.task_id = tl.task_id").
		Where("tl.label_id IN (?)", labelIDs)

	result := db.Where("task_id IN (?) AND label_id IN (?)", subQuery, labelIDs).
		Delete(&database.TaskLabel{})
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to remove injection-label associations for labels %v: %w", labelIDs, err)
	}

	return result.RowsAffected, nil
}

// RemoveLabelsFromInjection removes all label associations from a specific injection via TaskLabel
func RemoveLabelsFromInjection(db *gorm.DB, injectionID int) error {
	subQuery := db.Model(&database.FaultInjection{}).
		Select("task_id").
		Where("id = ?", injectionID)

	if err := db.Where("task_id IN (?)", subQuery).
		Delete(&database.TaskLabel{}).Error; err != nil {
		return fmt.Errorf("failed to remove all labels from injection %d: %w", injectionID, err)
	}
	return nil
}

// RemoveLabelsFromInjections removes all label associations from multiple injections via TaskLabel
func RemoveLabelsFromInjections(db *gorm.DB, injectionIDs []int) error {
	if len(injectionIDs) == 0 {
		return nil
	}

	// Use subquery to delete in a single operation
	subQuery := db.Model(&database.FaultInjection{}).
		Select("task_id").
		Where("id IN (?)", injectionIDs)

	if err := db.Where("task_id IN (?)", subQuery).
		Delete(&database.TaskLabel{}).Error; err != nil {
		return fmt.Errorf("failed to remove all labels from injections %v: %w", injectionIDs, err)
	}
	return nil
}

// ListInjectionIDsByLabels gets injection IDs associated with all specified labels via TaskLabel
func ListInjectionIDsByLabels(db *gorm.DB, labelConditions []map[string]string) ([]int, error) {
	var injectionIDs []int
	query := db.Model(&database.FaultInjection{}).
		Select("DISTINCT fault_injection_schedules.id").
		Joins("JOIN tasks ON tasks.id = fault_injection_schedules.task_id").
		Joins("JOIN task_labels tl ON tl.task_id = tasks.id").
		Joins("JOIN labels ON labels.id = tl.label_id").
		Where("fault_injection_schedules.status != ?", consts.CommonDeleted)

	var whereClauses []string
	var whereArgs []any

	for _, condition := range labelConditions {
		whereClauses = append(whereClauses, "(labels.label_key = ? AND labels.label_value = ?)")
		whereArgs = append(whereArgs, condition["key"], condition["value"])
	}

	if len(whereClauses) > 0 {
		whereClause := strings.Join(whereClauses, " OR ")
		query = query.Where(whereClause, whereArgs...)
	}

	if err := query.Pluck("fault_injection_schedules.id", &injectionIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to list injection IDs by labels: %v", err)
	}

	return injectionIDs, nil
}

// ListInjectionLabels gets labels for multiple injections in batch via TaskLabel
func ListInjectionLabels(db *gorm.DB, injectionIDs []int) (map[int][]database.Label, error) {
	if len(injectionIDs) == 0 {
		return nil, nil
	}

	type injectionLabelResult struct {
		database.Label
		InjectionID int `gorm:"column:injection_id"`
	}

	var flatResults []injectionLabelResult
	if err := db.Model(&database.Label{}).
		Joins("JOIN task_labels tl ON tl.label_id = labels.id").
		Joins("JOIN fault_injection_schedules fis ON fis.task_id = tl.task_id").
		Where("fis.id IN (?)", injectionIDs).
		Select("labels.*, fis.id as injection_id").
		Find(&flatResults).Error; err != nil {
		return nil, fmt.Errorf("failed to batch query fault injection labels: %w", err)
	}

	labelsMap := make(map[int][]database.Label)
	for _, id := range injectionIDs {
		labelsMap[id] = []database.Label{}
	}

	for _, res := range flatResults {
		label := res.Label
		labelsMap[res.InjectionID] = append(labelsMap[res.InjectionID], label)
	}

	return labelsMap, nil
}

// ListInjectionLabelCounts retrieves the count of injections associated with each label ID
func ListInjectionLabelCounts(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	if len(labelIDs) == 0 {
		return make(map[int]int64), nil
	}

	type injectionLabelResult struct {
		labelID int `gorm:"column:label_id"`
		count   int64
	}

	var results []injectionLabelResult
	// Count via task_labels joined with fault_injection_schedules
	if err := db.Table("task_labels tl").
		Select("tl.label_id, count(DISTINCT fis.id) as count").
		Joins("JOIN fault_injection_schedules fis ON fis.task_id = tl.task_id").
		Where("tl.label_id IN (?)", labelIDs).
		Group("tl.label_id").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to count injection-label associations: %w", err)
	}

	countMap := make(map[int]int64, len(results))
	for _, result := range results {
		countMap[result.labelID] = result.count
	}

	return countMap, nil
}

// ListInjectionLabelsByInjectionID gets labels for a specific injection via TaskLabel
func ListLabelsByInjectionID(db *gorm.DB, injectionID int) ([]database.Label, error) {
	var labels []database.Label
	if err := db.Table("labels").
		Joins("JOIN task_labels tl ON labels.id = tl.label_id").
		Joins("JOIN fault_injection_schedules fis ON fis.task_id = tl.task_id").
		Where("fis.id = ?", injectionID).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get injection labels: %v", err)
	}
	return labels, nil
}

// ListLabelIDsByKeyAndInjectionID finds label IDs by keys associated with a specific injection via TaskLabel
func ListLabelIDsByKeyAndInjectionID(db *gorm.DB, injectionID int, keys []string) ([]int, error) {
	var labelIDs []int

	err := db.Table("labels l").
		Select("l.id").
		Joins("JOIN task_labels tl ON tl.label_id = l.id").
		Joins("JOIN fault_injection_schedules fis ON fis.task_id = tl.task_id").
		Where("fis.id = ? AND l.label_key IN (?)", injectionID, keys).
		Pluck("l.id", &labelIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find label IDs by key '%s': %w", keys, err)
	}

	return labelIDs, nil
}
