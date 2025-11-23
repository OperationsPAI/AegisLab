package repository

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"aegis/consts"
	"aegis/database"
)

const BATCH_SIZE = 500

// =====================================================================
// Execution Repository Functions
// =====================================================================

// BatchDeleteExecutions marks multiple executions as deleted in batch
func BatchDeleteExecutions(db *gorm.DB, executions []int) error {
	if len(executions) == 0 {
		return nil
	}

	if err := db.Model(&database.Execution{}).
		Where("id IN (?) AND status != ?", executions, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to batch delete executions: %w", err)
	}

	return nil
}

// CreateExecution creates a new execution result record
func CreateExecution(db *gorm.DB, execution *database.Execution) error {
	if err := db.Create(execution).Error; err != nil {
		return fmt.Errorf("failed to create execution result: %w", err)
	}
	return nil
}

// GetExecutionByID retrieves an execution result by its ID with preloaded associations
func GetExecutionByID(db *gorm.DB, id int) (*database.Execution, error) {
	var result database.Execution
	if err := db.
		Preload("Algorithm.Container").
		Preload("Datapack.Benchmark.Container").
		Preload("Datapack.Pedestal.Container").
		Preload("Dataset").
		Preload("Task.Project").
		Preload("Labels").
		Where("id = ? AND status != ?", id, consts.CommonDeleted).
		First(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to find execution result with id %d: %w", id, err)
	}
	return &result, nil
}

// ListExecutions lists executions based on filters and pagination
func ListExecutions(db *gorm.DB, limit, offset int, event *consts.ExecutionState, status *consts.StatusType, labelConditons []map[string]string) ([]database.Execution, int64, error) {
	var executions []database.Execution
	var total int64

	query := db.Model(&database.Execution{}).
		Preload("Algorithm.Container").
		Preload("Datapack.Benchmark.Container").
		Preload("Datapack.Pedestal.Container").
		Preload("Dataset").
		Preload("Task.Project").
		Preload("Labels")
	if event != nil {
		query = query.Where("event = ?", *event)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if len(labelConditons) > 0 {
		for _, condition := range labelConditons {
			subQuery := db.Table("task_labels tl").
				Select("e.id").
				Joins("JOIN executions e ON e.task_id = tl.task_id").
				Joins("JOIN labels ON labels.id = tl.label_id").
				Where("labels.label_key = ? AND labels.label_value = ?", condition["key"], condition["value"])

			query = query.Where("execution.id IN (?)", subQuery)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count executions: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("updated_at DESC").Find(&executions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list executions: %w", err)
	}

	return executions, total, nil
}

func ListExecutionsByDatapackIDs(db *gorm.DB, datapackIDs []int) ([]database.Execution, error) {
	if len(datapackIDs) == 0 {
		return make([]database.Execution, 0), nil
	}

	var results []database.Execution

	query := db.
		Preload("Algorithm.Container").
		Preload("Datapack.Benchmark.Container").
		Preload("Datapack.Pedestal.Container").
		Preload("Dataset").
		Preload("Task.Project").
		Preload("Labels").
		Where("datapack_id IN (?) AND status != ?", datapackIDs, consts.CommonDeleted)
	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to list executions by datapack IDs: %w", err)
	}

	return results, nil
}

// UpdateExecution updates fields of an execution record
func UpdateExecution(db *gorm.DB, id int, updates map[string]any) error {
	result := db.Model(&database.Execution{}).
		Where("id = ? AND status != ?", id, consts.CommonDeleted).
		Updates(updates)
	if err := result.Error; err != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("execution not found or no changes made")
	}
	return nil
}

// =====================================================================
// ExecutionLabel Repository Functions
// =====================================================================

// Business layer: Execution labels are stored as TaskLabel in database
// Since Execution and Task are 1:1 relationship

// AddExecutionLabels adds multiple execution-label associations via TaskLabel
func AddExecutionLabels(db *gorm.DB, executionID int, labelIDs []int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	// Get the TaskID for this execution
	var execution database.Execution
	if err := db.Select("task_id").First(&execution, executionID).Error; err != nil {
		return fmt.Errorf("failed to get execution task_id: %w", err)
	}

	// Create TaskLabel associations
	taskLabels := make([]database.TaskLabel, 0, len(labelIDs))
	for _, labelID := range labelIDs {
		taskLabels = append(taskLabels, database.TaskLabel{
			TaskID:  execution.TaskID,
			LabelID: labelID,
		})
	}

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "task_id"}, {Name: "label_id"}},
		DoNothing: true,
	}).Create(&taskLabels).Error; err != nil {
		return fmt.Errorf("failed to add execution-label associatons: %w", err)
	}

	return nil
}

// ClearExecutionLabels removes label associations from specified executions via TaskLabel
func ClearExecutionLabels(db *gorm.DB, executionIDs []int, labelIDs []int) error {
	if len(executionIDs) == 0 {
		return nil
	}

	// Use subquery to delete in a single operation
	subQuery := db.Model(&database.Execution{}).
		Select("task_id").
		Where("id IN (?)", executionIDs)

	query := db.Table("task_labels").
		Where("task_id IN (?)", subQuery)
	if len(labelIDs) > 0 {
		query = query.Where("label_id IN (?)", labelIDs)
	}

	if err := query.Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to clear execution labels: %w", err)
	}
	return nil
}

// RemoveLabelsFromExecution removes all label associations from a specific execution via TaskLabel
func RemoveLabelsFromExecution(db *gorm.DB, executionID int) error {
	// Use subquery to delete in a single operation
	subQuery := db.Model(&database.Execution{}).
		Select("task_id").
		Where("id = ?", executionID)

	if err := db.Where("task_id IN (?)", subQuery).
		Delete(&database.TaskLabel{}).Error; err != nil {
		return fmt.Errorf("failed to remove all labels from execution %d: %w", executionID, err)
	}
	return nil
}

// RemoveLabelsFromExecutions removes all label associations from multiple executions via TaskLabel
func RemoveLabelsFromExecutions(db *gorm.DB, executionIDs []int) error {
	if len(executionIDs) == 0 {
		return nil
	}

	// Use subquery to delete in a single operation
	subQuery := db.Model(&database.Execution{}).
		Select("task_id").
		Where("id IN (?)", executionIDs)

	if err := db.Where("task_id IN (?)", subQuery).
		Delete(&database.TaskLabel{}).Error; err != nil {
		return fmt.Errorf("failed to remove all labels from executions %v: %w", executionIDs, err)
	}
	return nil
}

// RemoveExecutionsFromLabel deletes all execution-label associations for a specific label
// This removes TaskLabel entries for all Executions that are associated with this label
func RemoveExecutionsFromLabel(db *gorm.DB, labelID int) (int64, error) {
	// Use subquery to find TaskIDs and delete in a single operation
	subQuery := db.Table("task_labels tl").
		Select("DISTINCT tl.task_id").
		Joins("JOIN executions e ON e.task_id = tl.task_id").
		Where("tl.label_id = ?", labelID)

	result := db.Where("task_id IN (?) AND label_id = ?", subQuery, labelID).
		Delete(&database.TaskLabel{})
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete execution-label associations for label %d: %w", labelID, err)
	}

	return result.RowsAffected, nil
}

// RemoveExecutionsFromLabels removes all execution-label associations for multiple labels
func RemoveExecutionsFromLabels(db *gorm.DB, labelIDs []int) (int64, error) {
	if len(labelIDs) == 0 {
		return 0, nil
	}

	// Use subquery to find TaskIDs and delete in a single operation
	subQuery := db.Table("task_labels tl").
		Select("DISTINCT tl.task_id").
		Joins("JOIN executions e ON e.task_id = tl.task_id").
		Where("tl.label_id IN (?)", labelIDs)

	result := db.Where("task_id IN (?) AND label_id IN (?)", subQuery, labelIDs).
		Delete(&database.TaskLabel{})
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete execution-label associations for labels %v: %w", labelIDs, err)
	}

	return result.RowsAffected, nil
}

// ListExecutionsByDatapackFilter lists executions for a specific algorithm version and datapack name, with optional label filtering
func ListExecutionsByDatapackFilter(db *gorm.DB, algorithmVersionID int, datapackName string, labelConditions []map[string]string) ([]database.Execution, error) {
	var executions []database.Execution

	query := db.Model(&database.Execution{}).
		Preload("DetectorResults").
		Preload("GranularityResults").
		Preload("Algorithm.Container").
		Preload("Datapack").
		Joins("JOIN fault_injections fi ON executions.datapack_id = fi.id").
		Where("executions.algorithm_version_id = ? AND fi.name = ? AND executions.status != ?",
			algorithmVersionID, datapackName, consts.CommonDeleted)

	if len(labelConditions) > 0 {
		query = query.
			Joins("JOIN task_labels tl ON tl.task_id = executions.task_id").
			Joins("JOIN labels l ON l.id = tl.label_id")

		var whereConditions *gorm.DB
		for _, condition := range labelConditions {
			if whereConditions == nil {
				whereConditions = db.Where("l.label_key = ? AND l.label_value = ?", condition["key"], condition["value"])
			} else {
				whereConditions = whereConditions.Or("l.label_key = ? AND l.label_value = ?", condition["key"], condition["value"])
			}
		}

		if whereConditions != nil {
			query = query.Where(whereConditions)
		}

		query = query.
			Group("executions.id").
			Having("COUNT(executions.id) = ?", len(labelConditions))
	}

	if err := query.Order("executions.updated_at DESC").Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("failed to list executions for algorithm %d and datapack %s: %w",
			algorithmVersionID, datapackName, err)
	}

	return executions, nil
}

// ListExecutionsByDatasetFilter lists executions for a specific algorithm version and dataset version, with optional label filtering
func ListExecutionsByDatasetFilter(db *gorm.DB, algorithmVersionID, datasetVersionID int, labelConditions []map[string]string) ([]database.Execution, error) {
	var executions []database.Execution

	query := db.Model(&database.Execution{}).
		Preload("DetectorResults").
		Preload("GranularityResults").
		Preload("Algorithm.Container").
		Preload("Datapack").
		Preload("Dataset").
		Preload("Dataset.Injections").
		Where("executions.algorithm_version_id = ? AND executions.dataset_version_id = ? AND executions.status != ?",
			algorithmVersionID, datasetVersionID, consts.CommonDeleted)

	if len(labelConditions) > 0 {
		query = query.
			Joins("JOIN task_labels tl ON tl.task_id = executions.task_id").
			Joins("JOIN labels l ON l.id = tl.label_id")

		var whereConditions *gorm.DB
		for _, condition := range labelConditions {
			if whereConditions == nil {
				whereConditions = db.Where("l.label_key = ? AND l.label_value = ?", condition["key"], condition["value"])
			} else {
				whereConditions = whereConditions.Or("l.label_key = ? AND l.label_value = ?", condition["key"], condition["value"])
			}
		}

		if whereConditions != nil {
			query = query.Where(whereConditions)
		}

		query = query.
			Group("executions.id").
			Having("COUNT(executions.id) = ?", len(labelConditions))
	}

	if err := query.Order("executions.updated_at DESC").Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("failed to list executions for algorithm %d and dataset version %d: %w",
			algorithmVersionID, datasetVersionID, err)
	}

	return executions, nil
}

// ListExecutionIDsByLabels gets execution IDs associated with all specified labels via TaskLabel
func ListExecutionIDsByLabels(db *gorm.DB, labelConditions []map[string]string) ([]int, error) {
	var executionIDs []int
	query := db.Model(&database.Execution{}).
		Select("DISTINCT executions.id").
		Joins("JOIN task_labels tl ON tl.task_id = executions.task_id").
		Joins("JOIN labels ON labels.id = tl.label_id").
		Where("executions.status != ?", consts.CommonDeleted)

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

	if err := query.Pluck("executions.id", &executionIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to list execution IDs by labels: %w", err)
	}

	return executionIDs, nil
}

// ListExecutionLabels gets labels for multiple executions in batch
func ListExecutionLabels(db *gorm.DB, executionIDs []int) (map[int][]database.Label, error) {
	if len(executionIDs) == 0 {
		return nil, nil
	}

	type executionLabelResult struct {
		database.Label
		executionID int `gorm:"column:execution_id"`
	}

	var flatResults []executionLabelResult
	if err := db.Model(&database.Label{}).
		Joins("JOIN execution_result_labels erl ON erl.label_id = labels.id").
		Where("erl.execution_id IN (?)", executionIDs).
		Select("labels.*, erl.execution_id").
		Find(&flatResults).Error; err != nil {
		return nil, fmt.Errorf("failed to batch query execution labels: %w", err)
	}

	labelsMap := make(map[int][]database.Label)
	for _, id := range executionIDs {
		labelsMap[id] = []database.Label{}
	}

	for _, res := range flatResults {
		label := res.Label
		labelsMap[res.executionID] = append(labelsMap[res.executionID], label)
	}

	return labelsMap, nil
}

// ListExecutionLabelCounts retrieves the count of executions associated with each label ID
func ListExecutionLabelCounts(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	if len(labelIDs) == 0 {
		return make(map[int]int64), nil
	}

	type executionLabelResult struct {
		labelID int `gorm:"column:label_id"`
		count   int64
	}

	var results []executionLabelResult
	// Count via task_labels joined with executions
	if err := db.Table("task_labels tl").
		Select("tl.label_id, count(DISTINCT e.id) as count").
		Joins("JOIN executions e ON e.task_id = tl.task_id").
		Where("tl.label_id IN (?)", labelIDs).
		Group("tl.label_id").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to count execution-label associations: %w", err)
	}

	countMap := make(map[int]int64, len(results))
	for _, result := range results {
		countMap[result.labelID] = result.count
	}

	return countMap, nil
}

// ListLabelsByExecutionID retrieves all labels associated with a specific execution via TaskLabel
func ListLabelsByExecutionID(db *gorm.DB, executionID int) ([]database.Label, error) {
	var labels []database.Label
	if err := db.Table("labels").
		Joins("JOIN task_labels tl ON labels.id = tl.label_id").
		Joins("JOIN executions e ON e.task_id = tl.task_id").
		Where("e.id = ?", executionID).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get execution labels: %v", err)
	}
	return labels, nil
}

// ListLabelIDsByKeyAndExecutionID retrieves label IDs for a specific execution based on label keys via TaskLabel
func ListLabelIDsByKeyAndExecutionID(db *gorm.DB, executionID int, keys []string) ([]int, error) {
	var labelIDs []int

	err := db.Table("labels l").
		Select("l.id").
		Joins("JOIN task_labels tl ON tl.label_id = l.id").
		Joins("JOIN executions e ON e.task_id = tl.task_id").
		Where("e.id = ? AND l.label_key IN (?)", executionID, keys).
		Pluck("l.id", &labelIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find label IDs by key '%s': %w", keys, err)
	}

	return labelIDs, nil
}

// GetExecutionStatistics returns statistics about executions
func GetExecutionStatistics() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total executions
	var total int64
	if err := database.DB.Model(&database.Execution{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total executions: %w", err)
	}
	stats["total"] = total

	// Executions by status
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	var statusCounts []StatusCount
	err := database.DB.Model(&database.Execution{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count executions by status: %w", err)
	}

	// Set status counts
	for _, sc := range statusCounts {
		switch sc.Status {
		case "pending":
			stats["pending"] = sc.Count
		case "running":
			stats["running"] = sc.Count
		case "completed":
			stats["completed"] = sc.Count
		case "failed":
			stats["failed"] = sc.Count
		case "cancelled":
			stats["cancelled"] = sc.Count
		default:
			stats[sc.Status] = sc.Count
		}
	}

	// Initialize missing statuses with 0
	statuses := []string{"pending", "running", "completed", "failed", "cancelled"}
	for _, status := range statuses {
		if _, exists := stats[status]; !exists {
			stats[status] = 0
		}
	}

	return stats, nil
}
