package repository

import (
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"gorm.io/gorm"
)

func CreateContainer(container *database.Container) error {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var existingContainer database.Container
		result := tx.Where("type = ? AND name = ? AND image = ? AND tag = ?",
			container.Type, container.Name, container.Image, container.Tag).
			FirstOrCreate(&existingContainer, container)

		if err := result.Error; err != nil {
			return err
		}

		if result.RowsAffected == 0 {
			return tx.Model(&existingContainer).Update("updated_at", tx.NowFunc()).Error
		}

		*container = existingContainer
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create or update container: %v", err)
	}

	return nil
}

func GetContaineInfo(opts *dto.GetContainerFilterOptions) (*database.Container, error) {
	query := database.DB.Where("name = ?", opts.Name)

	if opts != nil {
		if opts.Type != "" {
			query = query.Where("type = ?", opts.Type)
		}

		if opts.Image != "" {
			query = query.Where("image = ?", opts.Image)
		}

		if opts.Image != "" && opts.Tag != "" {
			query = query.Where("tag = ?", opts.Tag)
		}
	}

	var record database.Container
	if err := query.
		Order("updated_at DESC").
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("container info '%s' not found", opts.Name)
		}

		return nil, fmt.Errorf("failed to query container info: %v", err)
	}

	return &record, nil
}

func ListContainers(opts *dto.ListContainersFilterOptions) ([]database.Container, error) {
	query := database.DB.Order("created_at DESC")

	if opts != nil {
		if opts.Status != nil {
			query = query.Where("status = ?", *opts.Status)
		}

		if opts.Type != "" {
			query = query.Where("type = ?", opts.Type)
		}

		if len(opts.Names) > 0 {
			query = query.Where("name IN ?", opts.Names)
		}
	}

	var containers []database.Container
	if err := query.Find(&containers).Error; err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	return containers, nil
}

// GetContainerLabelsMap gets all labels for multiple containers in batch (optimized)
func GetContainerLabelsMap(containerIDs []int) (map[int][]database.Label, error) {
	if len(containerIDs) == 0 {
		return make(map[int][]database.Label), nil
	}

	var relations []database.ContainerLabel
	if err := database.DB.Preload("Label").
		Where("container_id IN ?", containerIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get container label relations: %v", err)
	}

	labelsMap := make(map[int][]database.Label)
	for _, relation := range relations {
		if relation.Label != nil {
			labelsMap[relation.ContainerID] = append(labelsMap[relation.ContainerID], *relation.Label)
		}
	}

	for _, id := range containerIDs {
		if _, exists := labelsMap[id]; !exists {
			labelsMap[id] = []database.Label{}
		}
	}

	return labelsMap, nil
}

// GetContainerStatistics returns statistics about containers
func GetContainerStatistics() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total containers
	var total int64
	if err := database.DB.Model(&database.Container{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total containers: %v", err)
	}
	stats["total"] = total

	// Active containers
	var active int64
	if err := database.DB.Model(&database.Container{}).Where("status = 1").Count(&active).Error; err != nil {
		return nil, fmt.Errorf("failed to count active containers: %v", err)
	}
	stats["active"] = active

	// Disabled containers
	var disabled int64
	if err := database.DB.Model(&database.Container{}).Where("status = 0").Count(&disabled).Error; err != nil {
		return nil, fmt.Errorf("failed to count disabled containers: %v", err)
	}
	stats["disabled"] = disabled

	// Deleted containers
	var deleted int64
	if err := database.DB.Model(&database.Container{}).Where("status = -1").Count(&deleted).Error; err != nil {
		return nil, fmt.Errorf("failed to count deleted containers: %v", err)
	}
	stats["deleted"] = deleted

	return stats, nil
}

// GetContainerCountByType returns count of containers grouped by type
func GetContainerCountByType() (map[string]int64, error) {
	type TypeCount struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	var results []TypeCount
	err := database.DB.Model(&database.Container{}).
		Select("type, COUNT(*) as count").
		Where("status != -1").
		Group("type").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count containers by type: %v", err)
	}

	typeCounts := make(map[string]int64)
	for _, result := range results {
		typeCounts[result.Type] = result.Count
	}

	return typeCounts, nil
}
