package repository

import (
	"errors"
	"fmt"

	"aegis/consts"
	"aegis/database"
	"aegis/dto"

	"gorm.io/gorm"
)

func CheckContainerExists(id int) (bool, error) {
	var container database.Container
	if err := database.DB.
		Where("id = ?", id).
		First(&container).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("failed to check container: %v", err)
	}

	return true, nil
}

func CreateContainer(container *database.Container, tag string) error {
	return CreateContainerWithTx(nil, container, tag)
}

// CreateContainerWithTx creates a container with explicit transaction control
// If tx is nil, it creates its own transaction; otherwise uses the provided transaction
func CreateContainerWithTx(tx *gorm.DB, container *database.Container, tag string) error {
	if tx == nil {
		return database.DB.Transaction(func(newTx *gorm.DB) error {
			return createContainerInTx(newTx, container, tag)
		})
	}

	// Use provided transaction
	return createContainerInTx(tx, container, tag)
}

// Internal function that performs the actual creation within a transaction
func createContainerInTx(tx *gorm.DB, container *database.Container, tag string) error {
	if err := tx.Create(container).Error; err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	if tag != "" {
		label, err := CreateOrGetLabelWithTx(tx, consts.ContainerTag, tag, consts.ContainerCategory, "")
		if err != nil {
			return fmt.Errorf("failed to create or get label: %v", err)
		}

		relation := &database.ContainerLabel{
			ContainerID: container.ID,
			LabelID:     label.ID,
		}
		if err := tx.Create(relation).Error; err != nil {
			return fmt.Errorf("failed to create container-label relation: %v", err)
		}
	}

	return nil
}

func GetContainerInfo(opts *dto.GetContainerFilterOptions, userID int) (*dto.ContainerInfo, error) {
	query := database.DB.
		Preload("User").
		Where("name = ? AND type = ? AND user_id = ?", opts.Name, opts.Type, userID)

	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	var container database.Container
	if err := query.
		Preload("HelmConfig").
		First(&container).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("container '%s' not found for user %d", opts.Name, userID)
		}

		return nil, fmt.Errorf("failed to query container: %v", err)
	}

	var tags []database.Label
	if err := database.DB.
		Joins("JOIN container_labels ON container_labels.label_id = labels.id").
		Where("container_labels.container_id = ?", container.ID).
		Where("labels.label_key = ?", "container_tag").
		Order("labels.created_at DESC").
		Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to get container tags: %v", err)
	}

	return &dto.ContainerInfo{
		Container: container,
		Tags:      tags,
	}, nil
}

func GetContainerWithTag(containerType consts.ContainerType, name string, requestedTag string, userID int) (*database.Container, string, error) {
	enabledStatus := consts.ContainerEnabled
	containerInfo, err := GetContainerInfo(&dto.GetContainerFilterOptions{
		Type:   containerType,
		Name:   name,
		Status: &enabledStatus,
	}, userID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get container info: %w", err)
	}

	var selectedTag string
	if requestedTag != "" && containerInfo.IsTagExists(requestedTag) {
		selectedTag = requestedTag
	} else {
		selectedTag = containerInfo.Container.DefaultTag
	}

	return &containerInfo.Container, selectedTag, nil
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

func GetContainerLabel(containerID, labelID int) (*database.ContainerLabel, error) {
	var relation database.ContainerLabel
	if err := database.DB.
		Where("container_id = ? AND label_id = ?", containerID, labelID).
		First(&relation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to query container-label relation: %v", err)
	}

	return &relation, nil
}
