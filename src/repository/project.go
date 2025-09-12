package repository

import (
	"fmt"
	"time"

	"rcabench/consts"
	"rcabench/database"
)

func GetProject(column, param string) (*database.Project, error) {
	var record database.Project
	if err := database.DB.
		Where(fmt.Sprintf("%s = ?", column), param).
		First(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

func GetProjectByID(id int) (*database.Project, error) {
	var project database.Project
	if err := database.DB.
		Where("id = ? AND status != ?", id, consts.ProjectDeleted).
		First(&project).Error; err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return &project, nil
}

// GetProjectContainersMap gets all containers for multiple projects in batch (optimized)
func GetProjectContainersMap(projectIDs []int) (map[int][]database.Container, error) {
	if len(projectIDs) == 0 {
		return make(map[int][]database.Container), nil
	}

	var relations []database.ProjectContainer
	if err := database.DB.Preload("Container").
		Where("project_id IN ?", projectIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get project container relations: %v", err)
	}

	containersMap := make(map[int][]database.Container)
	for _, relation := range relations {
		if relation.Container != nil {
			containersMap[relation.ProjectID] = append(containersMap[relation.ProjectID], *relation.Container)
		}
	}

	for _, id := range projectIDs {
		if _, exists := containersMap[id]; !exists {
			containersMap[id] = []database.Container{}
		}
	}

	return containersMap, nil
}

// GetProjectDatasetsMap gets all datasets for multiple projects in batch (optimized)
func GetProjectDatasetsMap(projectIDs []int) (map[int][]database.Dataset, error) {
	if len(projectIDs) == 0 {
		return make(map[int][]database.Dataset), nil
	}

	var relations []database.ProjectDataset
	if err := database.DB.Preload("Dataset").
		Where("project_id IN ?", projectIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get project dataset relations: %v", err)
	}

	datasetsMap := make(map[int][]database.Dataset)
	for _, relation := range relations {
		if relation.Dataset != nil {
			datasetsMap[relation.ProjectID] = append(datasetsMap[relation.ProjectID], *relation.Dataset)
		}
	}

	for _, id := range projectIDs {
		if _, exists := datasetsMap[id]; !exists {
			datasetsMap[id] = []database.Dataset{}
		}
	}

	return datasetsMap, nil
}

// GetProjectInjectionsMap gets all labels for multiple projects in batch (optimized)
func GetProjectInjetionsMap(projectIDs []int) (map[int][]database.FaultInjectionSchedule, error) {
	if len(projectIDs) == 0 {
		return make(map[int][]database.FaultInjectionSchedule), nil
	}

	var injections []database.FaultInjectionSchedule
	if err := database.DB.
		Joins("JOIN tasks ON fault_injection_schedules.task_id = tasks.id").
		Where("tasks.project_id IN ?", projectIDs).
		Preload("Task").
		Find(&injections).Error; err != nil {
		return nil, fmt.Errorf("failed to get project fault injection relations: %v", err)
	}

	injectionsMap := make(map[int][]database.FaultInjectionSchedule)
	for _, injection := range injections {
		if injection.Task != nil {
			projectID := *injection.Task.ProjectID
			injectionsMap[projectID] = append(injectionsMap[projectID], injection)
		}
	}

	for _, id := range projectIDs {
		if _, exists := injectionsMap[id]; !exists {
			injectionsMap[id] = []database.FaultInjectionSchedule{}
		}
	}

	return injectionsMap, nil
}

// GetProjectLabelsMap gets all labels for multiple projects in batch (optimized)
func GetProjectLabelsMap(projectIDs []int) (map[int][]database.Label, error) {
	if len(projectIDs) == 0 {
		return make(map[int][]database.Label), nil
	}

	var relations []database.ProjectLabel
	if err := database.DB.Preload("Label").
		Where("project_id IN ?", projectIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get project label relations: %v", err)
	}

	labelsMap := make(map[int][]database.Label)
	for _, relation := range relations {
		if relation.Label != nil {
			labelsMap[relation.ProjectID] = append(labelsMap[relation.ProjectID], *relation.Label)
		}
	}

	for _, id := range projectIDs {
		if _, exists := labelsMap[id]; !exists {
			labelsMap[id] = []database.Label{}
		}
	}

	return labelsMap, nil
}

// GetProjectStatistics returns project statistics
func GetProjectStatistics() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total projects (exclude deleted)
	var total int64
	if err := database.DB.Model(&database.Project{}).Where("status != -1").Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total projects: %v", err)
	}
	stats["total"] = total

	// Active projects
	var active int64
	if err := database.DB.Model(&database.Project{}).Where("status = 1").Count(&active).Error; err != nil {
		return nil, fmt.Errorf("failed to count active projects: %v", err)
	}
	stats["active"] = active

	// Inactive projects
	var inactive int64
	if err := database.DB.Model(&database.Project{}).Where("status = 0").Count(&inactive).Error; err != nil {
		return nil, fmt.Errorf("failed to count inactive projects: %v", err)
	}
	stats["inactive"] = inactive

	// New projects today
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	var newToday int64
	if err := database.DB.Model(&database.Project{}).
		Where("created_at >= ? AND created_at < ?", today, tomorrow).
		Count(&newToday).Error; err != nil {
		return nil, fmt.Errorf("failed to count new projects today: %v", err)
	}
	stats["new_today"] = newToday

	return stats, nil
}
