package projectmodule

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *Repository) withDB(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) createProjectWithOwner(project *model.Project, userID int) error {
	var role model.Role
	if err := r.db.Where("name = ? AND status != ?", consts.RoleProjectAdmin.String(), consts.CommonDeleted).
		First(&role).Error; err != nil {
		return fmt.Errorf("failed to get project owner role: %w", err)
	}

	if err := r.db.Omit("ActiveName").Create(project).Error; err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	if err := r.db.Create(&model.UserProject{
		UserID:    userID,
		ProjectID: project.ID,
		RoleID:    role.ID,
		Status:    consts.CommonEnabled,
	}).Error; err != nil {
		return fmt.Errorf("failed to create user-project association: %w", err)
	}
	return nil
}

func (r *Repository) deleteProjectCascade(projectID int) (int64, error) {
	if err := r.db.Model(&model.UserProject{}).
		Where("project_id = ? AND status != ?", projectID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return 0, fmt.Errorf("failed to remove users from project: %w", err)
	}

	result := r.db.Model(&model.Project{}).
		Where("id = ? AND status != ?", projectID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to soft delete project %d: %w", projectID, result.Error)
	}
	return result.RowsAffected, nil
}

func (r *Repository) loadProjectDetail(projectID int) (*model.Project, *dto.ProjectStatistics, int, error) {
	project, err := r.loadProjectRecord(projectID)
	if err != nil {
		return nil, nil, 0, err
	}

	statsMap, err := r.listProjectStatistics([]int{project.ID})
	if err != nil {
		return nil, nil, 0, err
	}

	userCount, err := r.countProjectUsers(project.ID)
	if err != nil {
		return nil, nil, 0, err
	}

	return project, statsMap[project.ID], userCount, nil
}

func (r *Repository) listProjectViews(limit, offset int, isPublic *bool, status *consts.StatusType) ([]model.Project, map[int]*dto.ProjectStatistics, int64, error) {
	var (
		projects []model.Project
		total    int64
	)

	query := r.db.Model(&model.Project{})
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}
	if err := query.Limit(limit).Offset(offset).Find(&projects).Error; err != nil {
		return nil, nil, 0, fmt.Errorf("failed to list projects: %w", err)
	}

	projectIDs := make([]int, 0, len(projects))
	for _, project := range projects {
		projectIDs = append(projectIDs, project.ID)
	}

	labelsMap, err := r.listProjectLabels(projectIDs)
	if err != nil {
		return nil, nil, 0, err
	}

	statsMap, err := r.listProjectStatistics(projectIDs)
	if err != nil {
		return nil, nil, 0, err
	}

	for i := range projects {
		projects[i].Labels = labelsMap[projects[i].ID]
	}

	return projects, statsMap, total, nil
}

func (r *Repository) updateMutableProject(projectID int, patch func(*model.Project)) (*model.Project, error) {
	var project model.Project
	if err := r.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		return nil, fmt.Errorf("failed to find project with id %d: %w", projectID, err)
	}
	patch(&project)
	if err := r.db.Omit("ActiveName").Save(&project).Error; err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}
	return &project, nil
}

func (r *Repository) manageProjectLabels(projectID int, addLabelIDs []int, removeKeys []string) (*model.Project, error) {
	project, err := r.loadProjectRecord(projectID)
	if err != nil {
		return nil, err
	}
	if err := r.addProjectLabels(projectID, addLabelIDs); err != nil {
		return nil, err
	}
	if err := r.removeProjectLabelsByKeys(projectID, removeKeys); err != nil {
		return nil, err
	}
	labels, err := r.listLabelsByProjectID(project.ID)
	if err != nil {
		return nil, err
	}
	project.Labels = labels
	return project, nil
}

func (r *Repository) addProjectLabels(projectID int, labelIDs []int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	projectLabels := make([]model.ProjectLabel, 0, len(labelIDs))
	for _, labelID := range labelIDs {
		projectLabels = append(projectLabels, model.ProjectLabel{
			ProjectID: projectID,
			LabelID:   labelID,
		})
	}
	if err := r.db.Create(&projectLabels).Error; err != nil {
		return fmt.Errorf("failed to add project-label associations: %w", err)
	}
	return nil
}

func (r *Repository) removeProjectLabelsByKeys(projectID int, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	labelIDs, err := r.listProjectLabelIDsByKeys(projectID, keys)
	if err != nil {
		return fmt.Errorf("failed to find label ids by keys: %w", err)
	}
	if len(labelIDs) == 0 {
		return nil
	}

	if err := r.db.Table("project_labels").
		Where("project_id = ? AND label_id IN (?)", projectID, labelIDs).
		Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to clear project labels: %w", err)
	}
	if err := r.db.Model(&model.Label{}).
		Where("id IN (?)", labelIDs).
		UpdateColumn("usage_count", gorm.Expr("GREATEST(0, usage_count - ?)", 1)).Error; err != nil {
		return fmt.Errorf("failed to decrease label usage counts: %w", err)
	}
	return nil
}

func (r *Repository) loadProjectRecord(projectID int) (*model.Project, error) {
	var project model.Project
	if err := r.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		return nil, fmt.Errorf("failed to find project with id %d: %w", projectID, err)
	}
	return &project, nil
}

func (r *Repository) countProjectUsers(projectID int) (int, error) {
	var userCount int64
	if err := r.db.Model(&model.UserProject{}).
		Where("project_id = ? AND status = ?", projectID, consts.CommonEnabled).
		Count(&userCount).Error; err != nil {
		return 0, fmt.Errorf("failed to count project users: %w", err)
	}
	return int(userCount), nil
}

func (r *Repository) listProjectStatistics(projectIDs []int) (map[int]*dto.ProjectStatistics, error) {
	statsMap := make(map[int]*dto.ProjectStatistics, len(projectIDs))
	for _, projectID := range projectIDs {
		statsMap[projectID] = &dto.ProjectStatistics{}
	}
	if len(projectIDs) == 0 {
		return statsMap, nil
	}

	var injectionStats []struct {
		ProjectID int
		Count     int64
		LastAt    *time.Time
	}
	if err := r.db.Table("fault_injections fi").
		Select("tr.project_id, COUNT(*) as count, MAX(fi.updated_at) as last_at").
		Joins("JOIN tasks t ON fi.task_id = t.id").
		Joins("JOIN traces tr ON t.trace_id = tr.id").
		Where("tr.project_id IN (?)", projectIDs).
		Group("tr.project_id").
		Scan(&injectionStats).Error; err != nil {
		return nil, fmt.Errorf("failed to batch get injection statistics: %w", err)
	}
	for _, stat := range injectionStats {
		statsMap[stat.ProjectID].InjectionCount = int(stat.Count)
		statsMap[stat.ProjectID].LastInjectionAt = stat.LastAt
	}

	var executionStats []struct {
		ProjectID int
		Count     int64
		LastAt    *time.Time
	}
	if err := r.db.Table("executions e").
		Select("tr.project_id, COUNT(*) as count, MAX(e.updated_at) as last_at").
		Joins("JOIN tasks t ON e.task_id = t.id").
		Joins("JOIN traces tr ON t.trace_id = tr.id").
		Where("tr.project_id IN (?)", projectIDs).
		Group("tr.project_id").
		Scan(&executionStats).Error; err != nil {
		return nil, fmt.Errorf("failed to batch get execution statistics: %w", err)
	}
	for _, stat := range executionStats {
		statsMap[stat.ProjectID].ExecutionCount = int(stat.Count)
		statsMap[stat.ProjectID].LastExecutionAt = stat.LastAt
	}

	return statsMap, nil
}

func (r *Repository) listProjectLabels(projectIDs []int) (map[int][]model.Label, error) {
	labelsMap := make(map[int][]model.Label, len(projectIDs))
	for _, projectID := range projectIDs {
		labelsMap[projectID] = []model.Label{}
	}
	if len(projectIDs) == 0 {
		return labelsMap, nil
	}

	type projectLabelResult struct {
		model.Label
		ProjectID int `gorm:"column:project_id"`
	}

	var flatResults []projectLabelResult
	if err := r.db.Model(&model.Label{}).
		Joins("JOIN project_labels pl ON pl.label_id = labels.id").
		Where("pl.project_id IN (?)", projectIDs).
		Select("labels.*, pl.project_id").
		Find(&flatResults).Error; err != nil {
		return nil, fmt.Errorf("failed to batch query project labels: %w", err)
	}

	for _, result := range flatResults {
		labelsMap[result.ProjectID] = append(labelsMap[result.ProjectID], result.Label)
	}
	return labelsMap, nil
}

func (r *Repository) listLabelsByProjectID(projectID int) ([]model.Label, error) {
	var labels []model.Label
	if err := r.db.Model(&model.Label{}).
		Joins("JOIN project_labels pl ON pl.label_id = labels.id").
		Where("pl.project_id = ?", projectID).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to list labels for project %d: %w", projectID, err)
	}
	return labels, nil
}

func (r *Repository) listProjectLabelIDsByKeys(projectID int, keys []string) ([]int, error) {
	var labelIDs []int
	if err := r.db.Table("labels l").
		Select("l.id").
		Joins("JOIN project_labels pl ON pl.label_id = l.id").
		Where("pl.project_id = ? AND l.label_key IN (?)", projectID, keys).
		Pluck("l.id", &labelIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to find label IDs by key '%v': %w", keys, err)
	}
	return labelIDs, nil
}
