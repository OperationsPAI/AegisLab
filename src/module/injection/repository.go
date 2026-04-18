package injectionmodule

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"aegis/searchx"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) loadInjection(id int) (*model.FaultInjection, error) {
	var injection model.FaultInjection
	if err := r.db.
		Preload("Task").
		Preload("Task.Trace").
		Preload("Benchmark.Container").
		Preload("Pedestal.Container").
		Where("id = ?", id).
		First(&injection).Error; err != nil {
		return nil, fmt.Errorf("failed to find injection with id %d: %w", id, err)
	}
	return &injection, nil
}

func (r *Repository) findInjectionByName(name string, preload bool) (*model.FaultInjection, error) {
	query := r.db
	if preload {
		query = query.Preload("Labels")
	}

	var injection model.FaultInjection
	if err := query.Where("name = ? AND status != ?", name, consts.CommonDeleted).
		First(&injection).Error; err != nil {
		return nil, fmt.Errorf("failed to find injection with name %s: %w", name, err)
	}
	return &injection, nil
}

func (r *Repository) createInjectionRecord(injection *model.FaultInjection) error {
	if err := r.db.Create(injection).Error; err != nil {
		return fmt.Errorf("failed to create injection: %w", err)
	}
	return nil
}

func (r *Repository) updateGroundtruth(id int, groundtruths []model.Groundtruth, source string) error {
	groundtruthJSON, err := json.Marshal(groundtruths)
	if err != nil {
		return fmt.Errorf("failed to marshal groundtruths: %w", err)
	}

	result := r.db.Model(&model.FaultInjection{}).
		Where("id = ? AND status != ?", id, consts.CommonDeleted).
		Updates(map[string]any{
			"groundtruths":       string(groundtruthJSON),
			"groundtruth_source": source,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update groundtruth for injection %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("injection with id %d: %w", id, consts.ErrNotFound)
	}
	return nil
}

func (r *Repository) updateInjectionFields(id int, fields map[string]any) error {
	result := r.db.Model(&model.FaultInjection{}).
		Where("id = ? AND status != ?", id, consts.CommonDeleted).
		Updates(fields)
	if result.Error != nil {
		return fmt.Errorf("failed to update injection %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: injection %d not found", consts.ErrNotFound, id)
	}
	return nil
}

func (r *Repository) addInjectionLabels(injectionID int, labelIDs []int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	links := make([]model.FaultInjectionLabel, 0, len(labelIDs))
	for _, labelID := range labelIDs {
		links = append(links, model.FaultInjectionLabel{
			FaultInjectionID: injectionID,
			LabelID:          labelID,
		})
	}
	if err := r.db.Create(&links).Error; err != nil {
		return fmt.Errorf("failed to add injection labels: %w", err)
	}
	return nil
}

func (r *Repository) resolveProject(name string) (*model.Project, error) {
	var project model.Project
	if err := r.db.Where("name = ? AND status != ?", name, consts.CommonDeleted).First(&project).Error; err != nil {
		return nil, fmt.Errorf("failed to find project with name %s: %w", name, err)
	}
	return &project, nil
}

func (r *Repository) loadTask(taskID string) (*model.Task, error) {
	var task model.Task
	if err := r.db.
		Preload("FaultInjection.Benchmark.Container").
		Preload("FaultInjection.Pedestal.Container").
		Preload("Execution.AlgorithmVersion.Container").
		Preload("Execution.Datapack").
		Preload("Execution.DatasetVersion").
		Where("id = ? AND status != ?", taskID, consts.CommonDeleted).
		First(&task).Error; err != nil {
		return nil, fmt.Errorf("failed to find task with id %s: %w", taskID, err)
	}
	return &task, nil
}

func (r *Repository) loadPedestalHelmConfig(versionID int) (*model.HelmConfig, error) {
	var helmConfig model.HelmConfig
	if err := r.db.Preload("ContainerVersion").
		Where("container_version_id = ?", versionID).
		First(&helmConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to find helm config for version id %d: %w", versionID, err)
	}
	return &helmConfig, nil
}

func (r *Repository) listExistingEngineConfigs(configs []string) ([]string, error) {
	if len(configs) == 0 {
		return []string{}, nil
	}

	invalidLabelSubQuery := r.db.Table("fault_injection_labels fil").
		Select("fil.fault_injection_id").
		Joins("JOIN labels ON labels.id = fil.label_id").
		Where("labels.label_key = ? AND labels.label_value = ?", consts.LabelKeyTag, "invalid")

	var existing []string
	if err := r.db.Model(&model.FaultInjection{}).
		Select("engine_config").
		Where("engine_config IN (?) AND state >= ? AND status = ?", configs, consts.DatapackInjectSuccess, consts.CommonEnabled).
		Where("fault_injections.id NOT IN (?)", invalidLabelSubQuery).
		Pluck("engine_config", &existing).Error; err != nil {
		return nil, err
	}
	return existing, nil
}

func (r *Repository) clearInjectionLabels(injectionIDs []int, labelIDs []int) error {
	if len(injectionIDs) == 0 {
		return nil
	}

	query := r.db.Table("fault_injection_labels").Where("fault_injection_id IN (?)", injectionIDs)
	if len(labelIDs) > 0 {
		query = query.Where("label_id IN (?)", labelIDs)
	}
	if err := query.Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to clear injection labels: %w", err)
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
		UpdateColumn("usage_count", expr).Error; err != nil {
		return fmt.Errorf("failed to batch decrease label usages: %w", err)
	}
	return nil
}

func (r *Repository) listExecutionsByDatapackIDs(datapackIDs []int) ([]model.Execution, error) {
	if len(datapackIDs) == 0 {
		return []model.Execution{}, nil
	}

	var executions []model.Execution
	if err := r.db.
		Preload("AlgorithmVersion.Container").
		Preload("Datapack.Benchmark.Container").
		Preload("Datapack.Pedestal.Container").
		Preload("DatasetVersion").
		Preload("Task.Trace.Project").
		Where("datapack_id IN (?) AND status != ?", datapackIDs, consts.CommonDeleted).
		Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("failed to list executions by datapack IDs: %w", err)
	}
	return executions, nil
}

func (r *Repository) removeLabelsFromExecutions(executionIDs []int) error {
	if len(executionIDs) == 0 {
		return nil
	}
	if err := r.db.Where("execution_id IN (?)", executionIDs).
		Delete(&model.ExecutionInjectionLabel{}).Error; err != nil {
		return fmt.Errorf("failed to remove all labels from executions %v: %w", executionIDs, err)
	}
	return nil
}

func (r *Repository) batchDeleteExecutions(executionIDs []int) error {
	if len(executionIDs) == 0 {
		return nil
	}
	if err := r.db.Model(&model.Execution{}).
		Where("id IN (?) AND status != ?", executionIDs, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to batch delete executions: %w", err)
	}
	return nil
}

func (r *Repository) batchDeleteInjections(injectionIDs []int) error {
	if len(injectionIDs) == 0 {
		return nil
	}
	if err := r.db.Model(&model.FaultInjection{}).
		Where("id IN (?) AND status != ?", injectionIDs, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to batch delete injections: %w", err)
	}
	return nil
}

func (r *Repository) deleteInjectionsCascade(injectionIDs []int) error {
	executions, err := r.listExecutionsByDatapackIDs(injectionIDs)
	if err != nil {
		return fmt.Errorf("failed to list executions by datapack ids: %w", err)
	}

	executionIDs := make([]int, 0, len(executions))
	for _, execution := range executions {
		executionIDs = append(executionIDs, execution.ID)
	}

	if len(executionIDs) > 0 {
		if err := r.removeLabelsFromExecutions(executionIDs); err != nil {
			return fmt.Errorf("failed to remove execution labels: %w", err)
		}
		if err := r.batchDeleteExecutions(executionIDs); err != nil {
			return fmt.Errorf("failed to delete executions: %w", err)
		}
	}

	if err := r.clearInjectionLabels(injectionIDs, nil); err != nil {
		return fmt.Errorf("failed to clear injection labels: %w", err)
	}
	if err := r.batchDeleteInjections(injectionIDs); err != nil {
		return fmt.Errorf("failed to delete injections: %w", err)
	}
	return nil
}

func (r *Repository) getInjectionWithLabels(injectionID int) (*model.FaultInjection, error) {
	injection, err := r.loadInjection(injectionID)
	if err != nil {
		return nil, err
	}

	var labels []model.Label
	if err := r.db.Table("labels").
		Joins("JOIN fault_injection_labels fil ON labels.id = fil.label_id").
		Where("fil.fault_injection_id = ?", injection.ID).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get injection labels: %w", err)
	}
	injection.Labels = labels
	return injection, nil
}

func (r *Repository) loadInjectionLabelIDsByItems(conditions []map[string]string, category consts.LabelCategory) (map[string]int, error) {
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

func (r *Repository) loadExistingInjectionsByID(injectionIDs []int) (map[int]*model.FaultInjection, error) {
	injections, err := r.listFaultInjectionsByIDWithLabels(injectionIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[int]*model.FaultInjection, len(injections))
	for i := range injections {
		injection := injections[i]
		result[injection.ID] = &injection
	}
	return result, nil
}

func (r *Repository) listInjectionsView(limit, offset int, filterOptions *ListInjectionFilters) ([]model.FaultInjection, int64, error) {
	query := r.db.Model(&model.FaultInjection{}).
		Preload("Benchmark.Container").
		Preload("Pedestal.Container").
		Preload("Task.Trace.Project").
		Preload("Labels")
	if filterOptions.FaultType != nil {
		query = query.Where("fault_type = ?", *filterOptions.FaultType)
	}
	if filterOptions.Category != nil {
		query = query.Where("category = ?", *filterOptions.Category)
	}
	if filterOptions.Benchmark != "" {
		query = query.Where("benchmark = ?", filterOptions.Benchmark)
	}
	if filterOptions.State != nil {
		query = query.Where("state = ?", *filterOptions.State)
	}
	if filterOptions.Status != nil {
		query = query.Where("status = ?", *filterOptions.Status)
	}
	if len(filterOptions.LabelConditions) > 0 {
		for _, condition := range filterOptions.LabelConditions {
			subQuery := r.db.Table("fault_injection_labels fil").
				Select("fil.fault_injection_id").
				Joins("JOIN labels ON labels.id = fil.label_id").
				Where("labels.label_key = ? AND labels.label_value = ?", condition["key"], condition["value"])
			query = query.Where("fault_injections.id IN (?)", subQuery)
		}
	}

	var (
		injections []model.FaultInjection
		total      int64
	)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count injections: %w", err)
	}
	if err := query.Limit(limit).Offset(offset).Order("updated_at DESC").Find(&injections).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list injections: %w", err)
	}

	injectionIDs := make([]int, 0, len(injections))
	for _, injection := range injections {
		injectionIDs = append(injectionIDs, injection.ID)
	}

	labelsMap, err := r.listInjectionLabels(injectionIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list injection labels: %w", err)
	}

	for i := range injections {
		if labels, exists := labelsMap[injections[i].ID]; exists {
			injections[i].Labels = labels
		}
	}

	return injections, total, nil
}

func (r *Repository) listProjectInjectionsView(projectID, limit, offset int) ([]model.FaultInjection, int64, error) {
	baseQuery := r.db.Model(&model.FaultInjection{}).
		Joins("JOIN tasks ON tasks.id = fault_injections.task_id").
		Joins("JOIN traces on traces.id = tasks.trace_id").
		Where("traces.project_id = ? AND fault_injections.status != ?", projectID, consts.CommonDeleted)

	var (
		injections []model.FaultInjection
		total      int64
	)
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count injections for project %d: %w", projectID, err)
	}

	if err := baseQuery.
		Preload("Benchmark.Container").
		Preload("Pedestal.Container").
		Limit(limit).
		Offset(offset).
		Order("fault_injections.updated_at DESC").
		Find(&injections).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list injections for project %d: %w", projectID, err)
	}

	injectionIDs := make([]int, 0, len(injections))
	for _, injection := range injections {
		injectionIDs = append(injectionIDs, injection.ID)
	}

	labelsMap, err := r.listInjectionLabels(injectionIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list injection labels: %w", err)
	}

	for i := range injections {
		injections[i].Labels = labelsMap[injections[i].ID]
	}
	return injections, total, nil
}

func (r *Repository) searchInjections(req *SearchInjectionReq, projectID *int) ([]model.FaultInjection, int64, error) {
	searchReq := req.ConvertToSearchReq()
	if projectID != nil {
		searchReq.AddFilter("project_id", dto.OpEqual, *projectID)
	}

	qb := searchx.NewQueryBuilder(r.db, consts.InjectionAllowedFields)
	qb.ApplySearchReq(searchReq.Filters, searchReq.Keyword, searchReq.Sort, searchReq.GroupBy, model.FaultInjection{})
	qb.ApplyIncludes(searchReq.Includes)
	qb.ApplyIncludeFields(searchReq.IncludeFields)
	qb.ApplyExcludeFields(searchReq.ExcludeFields, model.FaultInjection{})

	total, err := qb.GetCount()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count searched injections: %w", err)
	}

	query := qb.Query()
	if searchReq.Size != 0 && searchReq.Page != 0 {
		query = query.Offset(searchReq.GetOffset()).Limit(int(searchReq.Size))
	}

	var injections []model.FaultInjection
	if err := query.Find(&injections).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to execute injection search: %w", err)
	}

	if len(req.Labels) == 0 {
		return injections, total, nil
	}

	labelConditions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		labelConditions = append(labelConditions, map[string]string{"key": item.Key, "value": item.Value})
	}

	injectionIDs, err := r.listInjectionIDsByLabels(labelConditions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list injection ids by labels: %w", err)
	}

	injectionIDMap := make(map[int]struct{}, len(injectionIDs))
	for _, id := range injectionIDs {
		injectionIDMap[id] = struct{}{}
	}

	filtered := make([]model.FaultInjection, 0, len(injections))
	for _, injection := range injections {
		if _, exists := injectionIDMap[injection.ID]; exists {
			filtered = append(filtered, injection)
		}
	}

	return filtered, total, nil
}

func (r *Repository) listIssuesFreeInjections(labelConditions []map[string]string, startTime, endTime *time.Time, projectID *int) ([]model.FaultInjectionNoIssues, error) {
	var injections []model.FaultInjectionNoIssues
	query := r.db.Model(&model.FaultInjectionNoIssues{}).
		Joins("JOIN fault_injections fi ON fi.id = fault_injection_no_issues.datapack_id").
		Joins("JOIN tasks t ON t.id = fi.task_id").
		Joins("JOIN traces tr ON tr.id = t.trace_id").
		Where("fi.status != ?", consts.CommonDeleted)
	if projectID != nil {
		query = query.Where("tr.project_id = ?", *projectID)
	}
	if startTime != nil {
		query = query.Where("fi.created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("fi.created_at <= ?", *endTime)
	}
	for _, condition := range labelConditions {
		subQuery := r.db.Table("fault_injection_labels fil").
			Select("fil.fault_injection_id").
			Joins("JOIN labels l ON l.id = fil.label_id").
			Where("l.label_key = ? AND l.label_value = ?", condition["key"], condition["value"])
		query = query.Where("fi.id IN (?)", subQuery)
	}
	anomalySubQuery := r.db.Table("executions e").
		Select("DISTINCT fi2.id").
		Joins("JOIN fault_injections fi2 ON fi2.id = e.datapack_id").
		Where("e.status != ? AND e.has_anomaly = ?", consts.CommonDeleted, true)
	query = query.Where("fi.id NOT IN (?)", anomalySubQuery)
	if err := query.Scan(&injections).Error; err != nil {
		return nil, fmt.Errorf("failed to list injections without issues: %w", err)
	}
	return injections, nil
}

func (r *Repository) listIssueInjections(labelConditions []map[string]string, startTime, endTime *time.Time, projectID *int) ([]model.FaultInjectionWithIssues, error) {
	var injections []model.FaultInjectionWithIssues
	query := r.db.Model(&model.FaultInjectionWithIssues{}).
		Joins("JOIN fault_injections fi ON fi.id = fault_injection_with_issues.datapack_id").
		Joins("JOIN tasks t ON t.id = fi.task_id").
		Joins("JOIN traces tr ON tr.id = t.trace_id").
		Where("fi.status != ?", consts.CommonDeleted)
	if projectID != nil {
		query = query.Where("tr.project_id = ?", *projectID)
	}
	if startTime != nil {
		query = query.Where("fi.created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("fi.created_at <= ?", *endTime)
	}
	for _, condition := range labelConditions {
		subQuery := r.db.Table("fault_injection_labels fil").
			Select("fil.fault_injection_id").
			Joins("JOIN labels l ON l.id = fil.label_id").
			Where("l.label_key = ? AND l.label_value = ?", condition["key"], condition["value"])
		query = query.Where("fi.id IN (?)", subQuery)
	}
	if err := query.Scan(&injections).Error; err != nil {
		return nil, fmt.Errorf("failed to list injections with issues: %w", err)
	}
	return injections, nil
}

func (r *Repository) listInjectionLabelIDsByKeys(injectionID int, keys []string) ([]int, error) {
	var labelIDs []int
	if err := r.db.Table("labels l").
		Select("l.id").
		Joins("JOIN fault_injection_labels fil ON fil.label_id = l.id").
		Where("fil.fault_injection_id = ? AND l.label_key IN (?)", injectionID, keys).
		Pluck("l.id", &labelIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to find label IDs by key '%v': %w", keys, err)
	}
	return labelIDs, nil
}

func (r *Repository) listFaultInjectionsByIDWithLabels(injectionIDs []int) ([]model.FaultInjection, error) {
	if len(injectionIDs) == 0 {
		return []model.FaultInjection{}, nil
	}

	var injections []model.FaultInjection
	if err := r.db.
		Preload("Benchmark.Container").
		Preload("Pedestal.Container").
		Preload("Task.Trace.Project").
		Preload("Labels").
		Where("id IN (?) AND status != ?", injectionIDs, consts.CommonDeleted).
		Find(&injections).Error; err != nil {
		return nil, fmt.Errorf("failed to query fault injections: %w", err)
	}

	labelsMap, err := r.listInjectionLabels(injectionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list injection labels: %w", err)
	}

	for i := range injections {
		if labels, exists := labelsMap[injections[i].ID]; exists {
			injections[i].Labels = labels
		}
	}

	return injections, nil
}

func (r *Repository) listInjectionIDsByLabelConditions(labelConditions []map[string]string) ([]int, error) {
	return r.listInjectionIDsByLabels(labelConditions)
}

func (r *Repository) listInjectionIDsByLabels(labelConditions []map[string]string) ([]int, error) {
	var injectionIDs []int
	query := r.db.Model(&model.FaultInjection{}).
		Select("DISTINCT fault_injections.id").
		Joins("JOIN fault_injection_labels fil ON fil.fault_injection_id = fault_injections.id").
		Joins("JOIN labels ON labels.id = fil.label_id").
		Where("fault_injections.status != ?", consts.CommonDeleted)

	var whereClauses []string
	var whereArgs []any
	for _, condition := range labelConditions {
		whereClauses = append(whereClauses, "(labels.label_key = ? AND labels.label_value = ?)")
		whereArgs = append(whereArgs, condition["key"], condition["value"])
	}
	if len(whereClauses) > 0 {
		query = query.Where(strings.Join(whereClauses, " OR "), whereArgs...)
	}

	if err := query.Pluck("fault_injections.id", &injectionIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to list injection IDs by labels: %v", err)
	}
	return injectionIDs, nil
}

func (r *Repository) listInjectionLabels(injectionIDs []int) (map[int][]model.Label, error) {
	labelsMap := make(map[int][]model.Label, len(injectionIDs))
	for _, id := range injectionIDs {
		labelsMap[id] = []model.Label{}
	}
	if len(injectionIDs) == 0 {
		return labelsMap, nil
	}

	type injectionLabelResult struct {
		model.Label
		InjectionID int `gorm:"column:injection_id"`
	}

	var flatResults []injectionLabelResult
	if err := r.db.Model(&model.Label{}).
		Joins("JOIN fault_injection_labels fil ON fil.label_id = labels.id").
		Where("fil.fault_injection_id IN (?)", injectionIDs).
		Select("labels.*, fil.fault_injection_id as injection_id").
		Find(&flatResults).Error; err != nil {
		return nil, fmt.Errorf("failed to batch query fault injection labels: %w", err)
	}

	for _, result := range flatResults {
		labelsMap[result.InjectionID] = append(labelsMap[result.InjectionID], result.Label)
	}
	return labelsMap, nil
}
