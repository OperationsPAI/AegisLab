package repository

import (
	"fmt"
	"time"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
)

// =====================================================================
// Project Repository Functions
// =====================================================================

// CreateProject creates a new project
func CreateProject(db *gorm.DB, project *database.Project) error {
	if err := db.Create(project).Error; err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

// DeleteProjct soft deletes a project by setting its status to deleted
func DeleteProject(db *gorm.DB, projectID int) (int64, error) {
	result := db.Model(&database.Project{}).
		Where("id = ? AND status != ?", projectID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to soft delete project %d: %w", projectID, result.Error)
	}
	return result.RowsAffected, nil
}

// GetProjectByID retrieves a project by its ID
func GetProjectByID(db *gorm.DB, id int) (*database.Project, error) {
	var project database.Project
	if err := db.Where("id = ?", id).First(&project).Error; err != nil {
		return nil, fmt.Errorf("failed to find project with id %d: %w", id, err)
	}
	return &project, nil
}

// GetProjectByName retrieves a project by its name
func GetProjectByName(db *gorm.DB, name string) (*database.Project, error) {
	var project database.Project
	if err := db.Where("name = ? AND status != ?", name, consts.CommonDeleted).First(&project).Error; err != nil {
		return nil, fmt.Errorf("failed to find project with name %s: %w", name, err)
	}
	return &project, nil
}

// GetProjectUserCount gets the count of users in a project
func GetProjectUserCount(db *gorm.DB, projectID int) (int, error) {
	var count int64
	if err := db.Model(&database.UserProject{}).
		Where("project_id = ? AND status = ?", projectID, consts.CommonEnabled).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count project users: %w", err)
	}
	return int(count), nil
}

// GetProjectStatistics returns project statistics
func GetProjectStatistics() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total projects (exclude deleted)
	var total int64
	if err := database.DB.Model(&database.Project{}).Where("status != -1").Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total projects: %w", err)
	}
	stats["total"] = total

	// Active projects
	var active int64
	if err := database.DB.Model(&database.Project{}).Where("status = 1").Count(&active).Error; err != nil {
		return nil, fmt.Errorf("failed to count active projects: %w", err)
	}
	stats["active"] = active

	// Inactive projects
	var inactive int64
	if err := database.DB.Model(&database.Project{}).Where("status = 0").Count(&inactive).Error; err != nil {
		return nil, fmt.Errorf("failed to count inactive projects: %w", err)
	}
	stats["inactive"] = inactive

	// New projects today
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	var newToday int64
	if err := database.DB.Model(&database.Project{}).
		Where("created_at >= ? AND created_at < ?", today, tomorrow).
		Count(&newToday).Error; err != nil {
		return nil, fmt.Errorf("failed to count new projects today: %w", err)
	}
	stats["new_today"] = newToday

	return stats, nil
}

// ListProjects lists projects based on filter options
func ListProjects(db *gorm.DB, limit, offset int, isPublic *bool, status *consts.StatusType) ([]database.Project, int64, error) {
	var projects []database.Project
	var total int64

	query := db.Model(&database.Project{})
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

	return projects, total, nil
}

// BatchGetProjectsByID retrieves multiple projects by their IDs
func ListProjectsByID(db *gorm.DB, projectIDs []int) ([]database.Project, error) {
	if len(projectIDs) == 0 {
		return []database.Project{}, nil
	}

	var projects []database.Project
	if err := db.
		Where("id IN (?) AND status != ?", projectIDs, consts.CommonDeleted).
		Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	return projects, nil
}

// UpdateProject updates a project
func UpdateProject(db *gorm.DB, project *database.Project) error {
	if err := db.Save(project).Error; err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

// ListInjectionsByProjectID gets all fault injections for a project
func ListInjectionsByProjectID(db *gorm.DB, projectID int) ([]database.FaultInjection, error) {
	var injections []database.FaultInjection
	if err := db.
		Joins("JOIN tasks ON fault_injections.task_id = tasks.id").
		Where("task.type = ? AND tasks.project_id = ?", consts.TaskTypeFaultInjection, projectID).
		Find(&injections).Error; err != nil {
		return nil, fmt.Errorf("failed to get fault injections of project id %d: %w", projectID, err)
	}

	return injections, nil
}

// =====================================================================
// ProjectLabel Repository Functions
// =====================================================================

// AddProjectLabels adds multiple project-label associations in a batch
func AddProjectLabels(db *gorm.DB, projectLabels []database.ProjectLabel) error {
	if len(projectLabels) == 0 {
		return nil
	}
	if err := db.Create(&projectLabels).Error; err != nil {
		return fmt.Errorf("failed to add project-label associations: %w", err)
	}
	return nil
}

// ClearProjectLabels removes label associations from specified projects
func ClearProjectLabels(db *gorm.DB, projectIDs []int, labelIDs []int) error {
	if len(projectIDs) == 0 {
		return nil
	}

	query := db.Table("project_labels").
		Where("project_id IN (?)", projectIDs)
	if len(labelIDs) > 0 {
		query = query.Where("label_id IN (?)", labelIDs)
	}

	if err := query.Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to clear project-label associations: %w", err)
	}
	return nil
}

// RemoveLabelsFromProject removes all label associations from a specific project
func RemoveLabelsFromProject(db *gorm.DB, projectID int) error {
	if err := db.Where("project_id = ?", projectID).
		Delete(&database.ProjectLabel{}).Error; err != nil {
		return fmt.Errorf("failed to delete all labels from project %d: %w", projectID, err)
	}
	return nil
}

// RemoveProjectsFromLabel removes all project associations from a specific label
func RemoveProjectsFromLabel(db *gorm.DB, labelID int) (int64, error) {
	result := db.Where("label_id = ?", labelID).
		Delete(&database.ProjectLabel{})
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete all projects from label %d: %w", labelID, err)
	}
	return result.RowsAffected, nil
}

// RemoveProjectsFromLabels removes all project associations from multiple labels
func RemoveProjectsFromLabels(db *gorm.DB, labelIDs []int) (int64, error) {
	if len(labelIDs) == 0 {
		return 0, nil
	}

	result := db.Where("label_id IN (?)", labelIDs).
		Delete(&database.ProjectLabel{})
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete all projects from labels %v: %w", labelIDs, err)
	}
	return result.RowsAffected, nil
}

// ListProjectLabels gets labels for multiple projects in batch
func ListProjectLabels(db *gorm.DB, projectIDs []int) (map[int][]database.Label, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}

	type projectLabelResult struct {
		database.Label
		projectID int `gorm:"column:project_id"`
	}

	var flatResults []projectLabelResult
	if err := db.Model(&database.Label{}).
		Joins("JOIN project_labels pl ON pl.label_id = labels.id").
		Where("pl.project_id IN (?)", projectIDs).
		Select("labels.*, pl.project_id").
		Find(&flatResults).Error; err != nil {
		return nil, fmt.Errorf("failed to batch query project labels: %w", err)
	}

	labelsMap := make(map[int][]database.Label)
	for _, id := range projectIDs {
		labelsMap[id] = []database.Label{}
	}

	for _, res := range flatResults {
		label := res.Label
		labelsMap[res.projectID] = append(labelsMap[res.projectID], label)
	}

	return labelsMap, nil
}

// ListProjectLabelCounts retrieves the count of projects associated with each label ID
func ListProjectLabelCounts(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	if len(labelIDs) == 0 {
		return make(map[int]int64), nil
	}

	type projectLabelResult struct {
		labelID int `gorm:"column:label_id"`
		count   int64
	}

	var results []projectLabelResult
	if err := db.Model(&database.ProjectLabel{}).
		Select("label_id, count(label_id) as count").
		Where("label_id IN (?)", labelIDs).
		Group("label_id").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to count project-label associations: %w", err)
	}

	countMap := make(map[int]int64, len(results))
	for _, result := range results {
		countMap[result.labelID] = result.count
	}

	return countMap, nil
}

// ListLabelsByProjectID lists all labels associated with a specific project
func ListLabelsByProjectID(db *gorm.DB, projectID int) ([]database.Label, error) {
	var labels []database.Label
	if err := db.Model(&database.Label{}).
		Joins("JOIN project_labels pl ON pl.label_id = labels.id").
		Where("pl.project_id = ?", projectID).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to list labels for project %d: %w", projectID, err)
	}
	return labels, nil
}

// ListLabelIDsByKeyAndProjectID finds label IDs by keys associated with a specific project
func ListLabelIDsByKeyAndProjectID(db *gorm.DB, projectID int, keys []string) ([]int, error) {
	var labelIDs []int

	err := db.Table("labels l").
		Select("l.id").
		Joins("JOIN project_labels pl ON pl.label_id = l.id").
		Where("pl.project_id = ? AND l.label_key IN (?)", projectID, keys).
		Pluck("l.id", &labelIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find label IDs by key '%s': %w", keys, err)
	}

	return labelIDs, nil
}
