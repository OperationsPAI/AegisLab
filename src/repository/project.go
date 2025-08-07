package repository

import (
	"fmt"
	"time"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
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
