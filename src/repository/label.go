package repository

import (
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"gorm.io/gorm"
)

// CreateOrGetLabel creates a new label or gets existing one
func CreateOrGetLabel(key, value, category, description string) (*database.Label, error) {
	var label database.Label

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("key = ? AND value = ? AND category = ?", key, value, category).First(&label).Error
		if err == nil {
			return tx.Model(&label).UpdateColumn("usage", gorm.Expr("usage + ?", 1)).Error
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to query label: %v", err)
		}

		label = database.Label{
			Key:         key,
			Value:       value,
			Category:    category,
			Description: description,
			Color:       "#1890ff",
			IsSystem:    true,
			Usage:       1,
		}
		return tx.Create(&label).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create or get label: %v", err)
	}

	return &label, nil
}

// GetLabelByID gets label by ID
func GetLabelByID(id int) (*database.Label, error) {
	var label database.Label
	if err := database.DB.First(&label, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("label with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get label: %v", err)
	}
	return &label, nil
}

// GetLabelByKeyValue gets label by key-value pair
func GetLabelByKeyValue(key, value string) (*database.Label, error) {
	var label database.Label
	if err := database.DB.Where("key = ? AND value = ?", key, value).First(&label).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("label '%s:%s' not found", key, value)
		}
		return nil, fmt.Errorf("failed to get label: %v", err)
	}
	return &label, nil
}

// UpdateLabel updates label information
func UpdateLabel(label *database.Label) error {
	if err := database.DB.Save(label).Error; err != nil {
		return fmt.Errorf("failed to update label: %v", err)
	}
	return nil
}

// UpdateLabelUsage updates the usage of a label (any increase or decrease)
func UpdateLabelUsage(id int, increment int) error {
	var operation string
	if increment >= 0 {
		operation = fmt.Sprintf("usage + %d", increment)
	} else {
		operation = fmt.Sprintf("GREATEST(0, usage + %d)", increment)
	}

	result := database.DB.Model(&database.Label{}).Where("id = ?", id).
		UpdateColumn("usage", gorm.Expr(operation))

	if result.Error != nil {
		return fmt.Errorf("failed to update label usage: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("label with id %d not found", id)
	}

	return nil
}

// DeleteLabel deletes label (hard delete, because relationships should also be cleaned up when label is deleted)
func DeleteLabel(id int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Delete all related relationships first
		if err := tx.Where("label_id = ?", id).Delete(&database.DatasetLabel{}).Error; err != nil {
			return fmt.Errorf("failed to delete dataset label relations: %v", err)
		}

		if err := tx.Where("label_id = ?", id).Delete(&database.FaultInjectionLabel{}).Error; err != nil {
			return fmt.Errorf("failed to delete fault injection label relations: %v", err)
		}

		if err := tx.Where("label_id = ?", id).Delete(&database.ContainerLabel{}).Error; err != nil {
			return fmt.Errorf("failed to delete container label relations: %v", err)
		}

		if err := tx.Where("label_id = ?", id).Delete(&database.ProjectLabel{}).Error; err != nil {
			return fmt.Errorf("failed to delete project label relations: %v", err)
		}

		// Finally delete the label itself
		if err := tx.Delete(&database.Label{}, id).Error; err != nil {
			return fmt.Errorf("failed to delete label: %v", err)
		}

		return nil
	})
}

// ListLabels gets the label list
func ListLabels(page, pageSize int, category string, isSystem *bool) ([]database.Label, int64, error) {
	var labels []database.Label
	var total int64

	query := database.DB.Model(&database.Label{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if isSystem != nil {
		query = query.Where("is_system = ?", *isSystem)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count labels: %v", err)
	}

	// Pagination query
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("usage DESC, created_at DESC").Find(&labels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list labels: %v", err)
	}

	return labels, total, nil
}

// SearchLabels searches for labels
func SearchLabels(keyword string, category string, limit int) ([]database.Label, error) {
	var labels []database.Label

	query := database.DB.Model(&database.Label{})

	if keyword != "" {
		query = query.Where("key ILIKE ? OR value ILIKE ? OR description ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("usage DESC, created_at DESC").Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to search labels: %v", err)
	}

	return labels, nil
}

// GetPopularLabels gets popular labels
func GetPopularLabels(category string, limit int) ([]database.Label, error) {
	var labels []database.Label

	query := database.DB.Model(&database.Label{}).Where("usage > 0")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("usage DESC").Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get popular labels: %v", err)
	}

	return labels, nil
}

// GetUnusedLabels gets unused labels
func GetUnusedLabels(category string) ([]database.Label, error) {
	var labels []database.Label

	query := database.DB.Model(&database.Label{}).Where("usage = 0")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Order("created_at ASC").Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get unused labels: %v", err)
	}

	return labels, nil
}

// CleanupUnusedLabels cleans up unused labels
func CleanupUnusedLabels(olderThanDays int) (int64, error) {
	var count int64

	query := database.DB.Model(&database.Label{}).
		Where("usage = 0 AND is_system = false AND created_at < NOW() - INTERVAL ? DAY", olderThanDays)

	// First get the count to be deleted
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count unused labels: %v", err)
	}

	// Execute deletion
	if err := query.Delete(&database.Label{}).Error; err != nil {
		return 0, fmt.Errorf("failed to cleanup unused labels: %v", err)
	}

	return count, nil
}

// GetLabelsByCategory gets labels by category
func GetLabelsByCategory(category string) ([]database.Label, error) {
	var labels []database.Label
	if err := database.DB.Where("category = ?", category).
		Order("usage DESC").Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get labels by category: %v", err)
	}
	return labels, nil
}

// GetSystemLabels gets system labels
func GetSystemLabels() ([]database.Label, error) {
	var labels []database.Label
	if err := database.DB.Where("is_system = true").
		Order("category, key, value").Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get system labels: %v", err)
	}
	return labels, nil
}
