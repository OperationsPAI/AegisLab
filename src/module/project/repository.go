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

func (r *Repository) loadProjectDetailBase(projectID int) (*model.Project, int, error) {
	project, err := r.loadProjectRecord(projectID)
	if err != nil {
		return nil, 0, err
	}

	var userCount int64
	if err := r.db.Model(&model.UserProject{}).
		Where("project_id = ? AND status = ?", project.ID, consts.CommonEnabled).
		Count(&userCount).Error; err != nil {
		return nil, 0, err
	}

	return project, int(userCount), nil
}

func (r *Repository) listProjectViews(limit, offset int, isPublic *bool, status *consts.StatusType, teamID *int) ([]model.Project, int64, error) {
	var (
		projects []model.Project
		total    int64
	)

	query := r.db.Model(&model.Project{})
	if teamID != nil {
		query = query.Where("team_id = ?", *teamID)
	}
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}
	if err := query.Limit(limit).Offset(offset).Find(&projects).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list projects: %w", err)
	}

	projectIDs := make([]int, 0, len(projects))
	for _, project := range projects {
		projectIDs = append(projectIDs, project.ID)
	}

	type projectLabelResult struct {
		model.Label
		ProjectID int `gorm:"column:project_id"`
	}

	labelsMap := make(map[int][]model.Label, len(projectIDs))
	for _, projectID := range projectIDs {
		labelsMap[projectID] = []model.Label{}
	}
	if len(projectIDs) > 0 {
		var flatResults []projectLabelResult
		if err := r.db.Model(&model.Label{}).
			Joins("JOIN project_labels pl ON pl.label_id = labels.id").
			Where("pl.project_id IN (?)", projectIDs).
			Select("labels.*, pl.project_id").
			Find(&flatResults).Error; err != nil {
			return nil, 0, fmt.Errorf("failed to batch query project labels: %w", err)
		}

		for _, result := range flatResults {
			labelsMap[result.ProjectID] = append(labelsMap[result.ProjectID], result.Label)
		}
	}

	for i := range projects {
		projects[i].Labels = labelsMap[projects[i].ID]
	}

	return projects, total, nil
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

	if len(addLabelIDs) > 0 {
		projectLabels := make([]model.ProjectLabel, 0, len(addLabelIDs))
		for _, labelID := range addLabelIDs {
			projectLabels = append(projectLabels, model.ProjectLabel{
				ProjectID: projectID,
				LabelID:   labelID,
			})
		}
		if err := r.db.Create(&projectLabels).Error; err != nil {
			return nil, fmt.Errorf("failed to add project-label associations: %w", err)
		}
	}

	if len(removeKeys) > 0 {
		var labelIDs []int
		if err := r.db.Table("labels l").
			Select("l.id").
			Joins("JOIN project_labels pl ON pl.label_id = l.id").
			Where("pl.project_id = ? AND l.label_key IN (?)", projectID, removeKeys).
			Pluck("l.id", &labelIDs).Error; err != nil {
			return nil, fmt.Errorf("failed to find label IDs by key '%v': %w", removeKeys, err)
		}
		if len(labelIDs) > 0 {
			if err := r.db.Table("project_labels").
				Where("project_id = ? AND label_id IN (?)", projectID, labelIDs).
				Delete(nil).Error; err != nil {
				return nil, fmt.Errorf("failed to clear project labels: %w", err)
			}
			if err := r.db.Model(&model.Label{}).
				Where("id IN (?)", labelIDs).
				UpdateColumn("usage_count", gorm.Expr("GREATEST(0, usage_count - ?)", 1)).Error; err != nil {
				return nil, fmt.Errorf("failed to decrease label usage counts: %w", err)
			}
		}
	}

	var labels []model.Label
	if err := r.db.Model(&model.Label{}).
		Joins("JOIN project_labels pl ON pl.label_id = labels.id").
		Where("pl.project_id = ?", project.ID).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to list labels for project %d: %w", project.ID, err)
	}
	project.Labels = labels
	return project, nil
}

func (r *Repository) loadProjectRecord(projectID int) (*model.Project, error) {
	var project model.Project
	if err := r.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		return nil, fmt.Errorf("failed to find project with id %d: %w", projectID, err)
	}
	return &project, nil
}

func (r *Repository) ListProjectStatistics(projectIDs []int) (map[int]*dto.ProjectStatistics, error) {
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
		Where("tr.project_id IN (?) AND fi.status != ?", projectIDs, consts.CommonDeleted).
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
		Where("tr.project_id IN (?) AND e.status != ?", projectIDs, consts.CommonDeleted).
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
