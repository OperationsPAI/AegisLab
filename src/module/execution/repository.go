package executionmodule

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

func (r *Repository) withDB(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
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

func (r *Repository) loadExecutionLabelIDsByItems(conditions []map[string]string, category consts.LabelCategory) (map[string]int, error) {
	if len(conditions) == 0 {
		return map[string]int{}, nil
	}

	query := r.db.Model(&model.Label{}).
		Where("status != ? AND category = ?", consts.CommonDeleted, category)
	orBuilder := r.db.Where("1 = 0")
	for _, condition := range conditions {
		andBuilder := r.db.Where("1 = 1")
		if key, ok := condition["key"]; ok {
			andBuilder = andBuilder.Where("label_key = ?", key)
		}
		if value, ok := condition["value"]; ok {
			andBuilder = andBuilder.Where("label_value = ?", value)
		}
		orBuilder = orBuilder.Or(andBuilder)
	}

	var labels []model.Label
	if err := query.Where(orBuilder).Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to list label IDs by conditions: %w", err)
	}

	result := make(map[string]int, len(labels))
	for _, label := range labels {
		result[label.Key+":"+label.Value] = label.ID
	}
	return result, nil
}

func (r *Repository) AddExecutionLabels(executionID int, labelIDs []int) error {
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

func (r *Repository) ClearExecutionLabels(executionIDs []int, labelIDs []int) error {
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

func (r *Repository) BatchDecreaseLabelUsages(labelIDs []int, decrement int) error {
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

func (r *Repository) ListExecutionIDsByLabelItems(labelItems []dto.LabelItem) ([]int, error) {
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

func (r *Repository) BatchDeleteExecutions(executionIDs []int) error {
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

func (r *Repository) UpdateExecutionDuration(executionID int, duration float64) error {
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

func (r *Repository) SaveDetectorResults(results []model.DetectorResult) error {
	if len(results) == 0 {
		return fmt.Errorf("no detector results to save")
	}
	if err := r.db.Create(&results).Error; err != nil {
		return fmt.Errorf("failed to save detector results: %w", err)
	}
	return nil
}

func (r *Repository) SaveGranularityResults(results []model.GranularityResult) error {
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
