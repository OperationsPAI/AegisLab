package repository

import (
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"gorm.io/gorm"
)

// CreateLabel 创建标签
func CreateLabel(label *database.Label) error {
	if err := database.DB.Create(label).Error; err != nil {
		return fmt.Errorf("failed to create label: %v", err)
	}
	return nil
}

// GetLabelByID 根据ID获取标签
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

// GetLabelByKeyValue 根据键值对获取标签
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

// GetOrCreateLabel 获取或创建标签
func GetOrCreateLabel(key, value, category string) (*database.Label, error) {
	var label database.Label

	// 先尝试获取
	if err := database.DB.Where("key = ? AND value = ?", key, value).First(&label).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在则创建
			label = database.Label{
				Key:      key,
				Value:    value,
				Category: category,
			}
			if err := database.DB.Create(&label).Error; err != nil {
				return nil, fmt.Errorf("failed to create label: %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get label: %v", err)
		}
	}

	return &label, nil
}

// UpdateLabel 更新标签信息
func UpdateLabel(label *database.Label) error {
	if err := database.DB.Save(label).Error; err != nil {
		return fmt.Errorf("failed to update label: %v", err)
	}
	return nil
}

// DeleteLabel 删除标签（硬删除，因为标签被删除后关联关系也应该被清理）
func DeleteLabel(id int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 先删除所有关联关系
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

		// 最后删除标签本身
		if err := tx.Delete(&database.Label{}, id).Error; err != nil {
			return fmt.Errorf("failed to delete label: %v", err)
		}

		return nil
	})
}

// ListLabels 获取标签列表
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

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count labels: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("usage DESC, created_at DESC").Find(&labels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list labels: %v", err)
	}

	return labels, total, nil
}

// SearchLabels 搜索标签
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

// GetPopularLabels 获取热门标签
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

// GetUnusedLabels 获取未使用的标签
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

// CleanupUnusedLabels 清理未使用的标签
func CleanupUnusedLabels(olderThanDays int) (int64, error) {
	var count int64

	query := database.DB.Model(&database.Label{}).
		Where("usage = 0 AND is_system = false AND created_at < NOW() - INTERVAL ? DAY", olderThanDays)

	// 先获取要删除的数量
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count unused labels: %v", err)
	}

	// 执行删除
	if err := query.Delete(&database.Label{}).Error; err != nil {
		return 0, fmt.Errorf("failed to cleanup unused labels: %v", err)
	}

	return count, nil
}

// GetLabelsByCategory 根据分类获取标签
func GetLabelsByCategory(category string) ([]database.Label, error) {
	var labels []database.Label
	if err := database.DB.Where("category = ?", category).
		Order("usage DESC").Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get labels by category: %v", err)
	}
	return labels, nil
}

// GetSystemLabels 获取系统标签
func GetSystemLabels() ([]database.Label, error) {
	var labels []database.Label
	if err := database.DB.Where("is_system = true").
		Order("category, key, value").Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get system labels: %v", err)
	}
	return labels, nil
}
