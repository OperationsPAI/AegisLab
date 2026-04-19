package execution

import (
	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) getProjectByName(name string) (*model.Project, error) {
	var project model.Project
	if err := r.db.Where("name = ? AND status != ?", name, consts.CommonDeleted).First(&project).Error; err != nil {
		return nil, fmt.Errorf("failed to find project with name %s: %w", name, err)
	}
	return &project, nil
}

func (r *Repository) listProjectExecutionsView(projectID, limit, offset int) ([]model.Execution, int64, error) {
	var (
		executions []model.Execution
		total      int64
	)

	baseQuery := r.db.Model(&model.Execution{}).
		Joins("JOIN tasks ON tasks.id = executions.task_id").
		Joins("JOIN traces on traces.id = tasks.trace_id").
		Where("traces.project_id = ? AND executions.status != ?", projectID, consts.CommonDeleted)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count executions for project %d: %w", projectID, err)
	}
	if err := baseQuery.
		Preload("AlgorithmVersion.Container").
		Preload("Datapack.Benchmark.Container").
		Preload("Datapack.Pedestal.Container").
		Preload("DatasetVersion").
		Limit(limit).
		Offset(offset).
		Order("executions.updated_at DESC").
		Find(&executions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list executions for project %d: %w", projectID, err)
	}
	return r.attachExecutionLabels(executions, total)
}

func (r *Repository) listExecutionsView(limit, offset int, req *ListExecutionReq) ([]model.Execution, int64, error) {
	labelConditions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		parts := strings.SplitN(item, ":", 2)
		condition := map[string]string{"key": parts[0], "value": ""}
		if len(parts) > 1 {
			condition["value"] = parts[1]
		}
		labelConditions = append(labelConditions, condition)
	}

	var (
		executions []model.Execution
		total      int64
	)

	query := r.db.Model(&model.Execution{}).
		Preload("AlgorithmVersion.Container").
		Preload("Datapack.Benchmark.Container").
		Preload("Datapack.Pedestal.Container").
		Preload("DatasetVersion").
		Preload("Task.Trace.Project")
	if req.State != nil {
		query = query.Where("event = ?", *req.State)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	for _, condition := range labelConditions {
		subQuery := r.db.Table("execution_injection_labels eil").
			Select("eil.execution_id").
			Joins("JOIN labels ON labels.id = eil.label_id").
			Where("labels.label_key = ? AND labels.label_value = ?", condition["key"], condition["value"])
		query = query.Where("executions.id IN (?)", subQuery)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count executions: %w", err)
	}
	if err := query.Limit(limit).Offset(offset).Order("updated_at DESC").Find(&executions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list executions: %w", err)
	}
	return r.attachExecutionLabels(executions, total)
}

func (r *Repository) getExecutionView(executionID int) (*model.Execution, []model.Label, error) {
	var execution model.Execution
	if err := r.db.
		Preload("AlgorithmVersion.Container").
		Preload("Datapack.Benchmark.Container").
		Preload("Datapack.Pedestal.Container").
		Preload("DatasetVersion").
		Preload("Task.Trace.Project").
		Where("id = ? AND status != ?", executionID, consts.CommonDeleted).
		First(&execution).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to find execution result with id %d: %w", executionID, err)
	}

	var labels []model.Label
	if err := r.db.Table("labels").
		Joins("JOIN execution_injection_labels eil ON labels.id = eil.label_id").
		Where("eil.execution_id = ?", execution.ID).
		Find(&labels).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get execution labels: %w", err)
	}
	return &execution, labels, nil
}

func (r *Repository) listEvaluationExecutionsByDatapack(algorithmVersionID int, datapackName string, filterLabels []dto.LabelItem) ([]model.Execution, error) {
	var executions []model.Execution

	query := r.db.Model(&model.Execution{}).
		Preload("DetectorResults").
		Preload("GranularityResults").
		Preload("AlgorithmVersion.Container").
		Preload("Datapack.Groundtruths").
		Joins("JOIN fault_injections fi ON executions.datapack_id = fi.id").
		Where(
			"executions.algorithm_version_id = ? AND fi.name = ? AND executions.status != ?",
			algorithmVersionID, datapackName, consts.CommonDeleted,
		)

	if len(filterLabels) > 0 {
		query = query.
			Joins("JOIN execution_injection_labels eil ON eil.execution_id = executions.id").
			Joins("JOIN labels l ON l.id = eil.label_id")

		var whereConditions *gorm.DB
		for _, label := range filterLabels {
			if whereConditions == nil {
				whereConditions = r.db.Where("l.label_key = ? AND l.label_value = ?", label.Key, label.Value)
			} else {
				whereConditions = whereConditions.Or("l.label_key = ? AND l.label_value = ?", label.Key, label.Value)
			}
		}

		if whereConditions != nil {
			query = query.Where(whereConditions)
		}
		query = query.Group("executions.id").Having("COUNT(executions.id) = ?", len(filterLabels))
	}

	if err := query.Order("executions.updated_at DESC").Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("failed to list evaluation executions for algorithm %d and datapack %s: %w", algorithmVersionID, datapackName, err)
	}
	return executions, nil
}

func (r *Repository) listEvaluationExecutionsByDataset(algorithmVersionID, datasetVersionID int, filterLabels []dto.LabelItem) ([]model.Execution, error) {
	var executions []model.Execution

	query := r.db.Model(&model.Execution{}).
		Preload("DetectorResults").
		Preload("GranularityResults").
		Preload("AlgorithmVersion.Container").
		Preload("Datapack.Groundtruths").
		Preload("DatasetVersion").
		Preload("DatasetVersion.Injections").
		Where(
			"executions.algorithm_version_id = ? AND executions.dataset_version_id = ? AND executions.status != ?",
			algorithmVersionID, datasetVersionID, consts.CommonDeleted,
		)

	if len(filterLabels) > 0 {
		query = query.
			Joins("JOIN execution_injection_labels eil ON eil.execution_id = executions.id").
			Joins("JOIN labels l ON l.id = eil.label_id")

		var whereConditions *gorm.DB
		for _, label := range filterLabels {
			if whereConditions == nil {
				whereConditions = r.db.Where("l.label_key = ? AND l.label_value = ?", label.Key, label.Value)
			} else {
				whereConditions = whereConditions.Or("l.label_key = ? AND l.label_value = ?", label.Key, label.Value)
			}
		}

		if whereConditions != nil {
			query = query.Where(whereConditions)
		}
		query = query.Group("executions.id").Having("COUNT(executions.id) = ?", len(filterLabels))
	}

	if err := query.Order("executions.updated_at DESC").Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("failed to list evaluation executions for algorithm %d and dataset version %d: %w", algorithmVersionID, datasetVersionID, err)
	}
	return executions, nil
}

func (r *Repository) getExecutionResultView(executionID int) (*model.Execution, []model.Label, []model.DetectorResult, []model.GranularityResult, error) {
	execution, labels, err := r.getExecutionView(executionID)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	if execution.AlgorithmVersion.Container.Name == config.GetDetectorName() {
		var detectorResults []model.DetectorResult
		if err := r.db.Where("execution_id = ?", execution.ID).Find(&detectorResults).Error; err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to get detector results: %w", err)
		}
		return execution, labels, detectorResults, nil, nil
	}

	var granularityResults []model.GranularityResult
	if err := r.db.Where("execution_id = ?", execution.ID).Find(&granularityResults).Error; err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get granularity results: %w", err)
	}
	return execution, labels, nil, granularityResults, nil
}

func (r *Repository) listAvailableExecutionLabels() ([]model.Label, error) {
	var labels []model.Label
	if err := r.db.
		Where("status != ?", consts.CommonDeleted).
		Order("usage_count DESC, created_at DESC").
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
	}

	executionLabels := make([]model.Label, 0)
	for _, label := range labels {
		if label.Category == consts.ExecutionCategory {
			executionLabels = append(executionLabels, label)
		}
	}
	return executionLabels, nil
}

func (r *Repository) listExecutionLabelIDsByKeys(executionID int, keys []string) ([]int, error) {
	var labelIDs []int
	if err := r.db.Table("labels l").
		Select("l.id").
		Joins("JOIN execution_injection_labels eil ON eil.label_id = l.id").
		Where("eil.execution_id = ? AND l.label_key IN (?)", executionID, keys).
		Pluck("l.id", &labelIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to find label IDs by key '%s': %w", keys, err)
	}
	return labelIDs, nil
}

func (r *Repository) addExecutionLabels(executionID int, labelIDs []int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	executionLabels := make([]model.ExecutionInjectionLabel, 0, len(labelIDs))
	for _, labelID := range labelIDs {
		executionLabels = append(executionLabels, model.ExecutionInjectionLabel{
			ExecutionID: executionID,
			LabelID:     labelID,
		})
	}
	if err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "execution_id"}, {Name: "label_id"}},
		DoNothing: true,
	}).Create(&executionLabels).Error; err != nil {
		return fmt.Errorf("failed to add execution-label associatons: %w", err)
	}
	return nil
}

func (r *Repository) clearExecutionLabels(executionIDs []int, labelIDs []int) error {
	if len(executionIDs) == 0 {
		return nil
	}

	query := r.db.Table("execution_injection_labels").Where("execution_id IN (?)", executionIDs)
	if len(labelIDs) > 0 {
		query = query.Where("label_id IN (?)", labelIDs)
	}
	if err := query.Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to clear execution labels: %w", err)
	}
	return nil
}

func (r *Repository) batchDecreaseLabelUsages(labelIDs []int, decrement int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	expr := gorm.Expr("GREATEST(0, usage_count - ?)", decrement)
	if err := r.db.Model(&model.Label{}).
		Where("id IN (?)", labelIDs).
		Clauses(clause.Returning{}).
		UpdateColumn("usage_count", expr).Error; err != nil {
		return fmt.Errorf("failed to batch decrease label usages: %w", err)
	}
	return nil
}

func (r *Repository) listExecutionIDsByLabelItems(labelItems []dto.LabelItem) ([]int, error) {
	labelConditions := make([]map[string]string, 0, len(labelItems))
	for _, item := range labelItems {
		labelConditions = append(labelConditions, map[string]string{"key": item.Key, "value": item.Value})
	}

	var executionIDs []int
	query := r.db.Model(&model.Execution{}).
		Select("DISTINCT executions.id").
		Joins("JOIN execution_injection_labels eil ON eil.execution_id = executions.id").
		Joins("JOIN labels ON labels.id = eil.label_id").
		Where("executions.status != ?", consts.CommonDeleted)

	var whereClauses []string
	var whereArgs []any
	for _, condition := range labelConditions {
		whereClauses = append(whereClauses, "(labels.label_key = ? AND labels.label_value = ?)")
		whereArgs = append(whereArgs, condition["key"], condition["value"])
	}
	if len(whereClauses) > 0 {
		query = query.Where(strings.Join(whereClauses, " OR "), whereArgs...)
	}

	if err := query.Pluck("executions.id", &executionIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to list execution IDs by labels: %w", err)
	}
	return executionIDs, nil
}

func (r *Repository) batchDeleteExecutions(executionIDs []int) error {
	if len(executionIDs) == 0 {
		return nil
	}
	if err := r.db.Where("execution_id IN (?)", executionIDs).
		Delete(&model.ExecutionInjectionLabel{}).Error; err != nil {
		return fmt.Errorf("failed to delete execution labels: %w", err)
	}
	if err := r.db.Model(&model.Execution{}).
		Where("id IN (?) AND status != ?", executionIDs, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to batch delete executions: %w", err)
	}
	return nil
}

func (r *Repository) updateExecutionDuration(executionID int, duration float64) error {
	var execution model.Execution
	if err := r.db.
		Preload("AlgorithmVersion.Container").
		Preload("Datapack.Benchmark.Container").
		Preload("Datapack.Pedestal.Container").
		Preload("DatasetVersion").
		Preload("Task.Trace.Project").
		Where("id = ? AND status != ?", executionID, consts.CommonDeleted).
		First(&execution).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: execution %d not found", consts.ErrNotFound, executionID)
		}
		return fmt.Errorf("execution %d not found: %w", executionID, err)
	}

	if execution.Status != consts.CommonEnabled {
		return fmt.Errorf("must upload results for an active execution %d", executionID)
	}
	if execution.State == consts.ExecutionSuccess {
		return fmt.Errorf("cannot upload results for a successful execution %d", executionID)
	}

	result := r.db.Model(&model.Execution{}).
		Where("id = ? AND status != ?", executionID, consts.CommonDeleted).
		Updates(map[string]any{"duration": duration})
	if err := result.Error; err != nil {
		return fmt.Errorf("failed to update execution %d duration: %w", executionID, err)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("execution not found or no changes made")
	}
	return nil
}

func (r *Repository) loadExecution(executionID int) (*model.Execution, error) {
	var execution model.Execution
	if err := r.db.Where("id = ? AND status != ?", executionID, consts.CommonDeleted).First(&execution).Error; err != nil {
		return nil, fmt.Errorf("failed to find execution %d: %w", executionID, err)
	}
	return &execution, nil
}

func (r *Repository) createExecutionRecord(execution *model.Execution) error {
	if err := r.db.Create(execution).Error; err != nil {
		return fmt.Errorf("failed to create execution: %w", err)
	}
	return nil
}

func (r *Repository) updateExecutionFields(executionID int, fields map[string]any) error {
	result := r.db.Model(&model.Execution{}).
		Where("id = ? AND status != ?", executionID, consts.CommonDeleted).
		Updates(fields)
	if err := result.Error; err != nil {
		return fmt.Errorf("failed to update execution %d: %w", executionID, err)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: execution %d not found", consts.ErrNotFound, executionID)
	}
	return nil
}

func (r *Repository) saveDetectorResults(results []model.DetectorResult) error {
	if len(results) == 0 {
		return fmt.Errorf("no detector results to save")
	}
	if err := r.db.Create(&results).Error; err != nil {
		return fmt.Errorf("failed to save detector results: %w", err)
	}
	return nil
}

func (r *Repository) saveGranularityResults(results []model.GranularityResult) error {
	if len(results) == 0 {
		return fmt.Errorf("no granularity results to create")
	}
	for i := range results {
		resultPtr := &results[i]
		err := r.db.Omit("active_name").Create(resultPtr).Error
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: index %d", consts.ErrAlreadyExists, i)
			}
			return fmt.Errorf("failed to create record index %d: %w", i, err)
		}
	}
	return nil
}

func (r *Repository) attachExecutionLabels(executions []model.Execution, total int64) ([]model.Execution, int64, error) {
	executionIDs := make([]int, 0, len(executions))
	for _, execution := range executions {
		executionIDs = append(executionIDs, execution.ID)
	}

	if len(executionIDs) == 0 {
		return executions, total, nil
	}

	type executionLabelResult struct {
		model.Label
		executionID int `gorm:"column:execution_id"`
	}

	var flatResults []executionLabelResult
	if err := r.db.Model(&model.Label{}).
		Joins("JOIN execution_injection_labels eil ON eil.label_id = labels.id").
		Where("eil.execution_id IN (?)", executionIDs).
		Select("labels.*, eil.execution_id").
		Find(&flatResults).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to batch query execution labels: %w", err)
	}

	labelsMap := make(map[int][]model.Label, len(executionIDs))
	for _, id := range executionIDs {
		labelsMap[id] = []model.Label{}
	}
	for _, res := range flatResults {
		labelsMap[res.executionID] = append(labelsMap[res.executionID], res.Label)
	}

	for i := range executions {
		executions[i].Labels = labelsMap[executions[i].ID]
	}
	return executions, total, nil
}
