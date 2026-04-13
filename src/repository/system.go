package repository

import (
	"fmt"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
)

func ListSystems(db *gorm.DB, limit, offset int) ([]database.System, int64, error) {
	var systems []database.System
	var total int64

	query := db.Model(&database.System{}).
		Where("status != ?", consts.CommonDeleted)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count systems: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("updated_at DESC").Find(&systems).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list systems: %w", err)
	}

	return systems, total, nil
}

func GetSystemByID(db *gorm.DB, id int) (*database.System, error) {
	var system database.System
	if err := db.
		Where("id = ? AND status != ?", id, consts.CommonDeleted).
		First(&system).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("system with id %d: %w", id, consts.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to find system with id %d: %w", id, err)
	}
	return &system, nil
}

func GetSystemByName(db *gorm.DB, name string) (*database.System, error) {
	var system database.System
	if err := db.
		Where("name = ? AND status != ?", name, consts.CommonDeleted).
		First(&system).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("system with name %s: %w", name, consts.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to find system with name %s: %w", name, err)
	}
	return &system, nil
}

func CreateSystem(db *gorm.DB, system *database.System) error {
	if err := db.Create(system).Error; err != nil {
		return fmt.Errorf("failed to create system: %w", err)
	}
	return nil
}

func UpdateSystem(db *gorm.DB, id int, updates map[string]interface{}) error {
	result := db.Model(&database.System{}).
		Where("id = ? AND status != ?", id, consts.CommonDeleted).
		Updates(updates)
	if err := result.Error; err != nil {
		return fmt.Errorf("failed to update system with id %d: %w", id, err)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("system with id %d: %w", id, consts.ErrNotFound)
	}
	return nil
}

func DeleteSystem(db *gorm.DB, id int) error {
	result := db.Model(&database.System{}).
		Where("id = ? AND status != ?", id, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if err := result.Error; err != nil {
		return fmt.Errorf("failed to delete system with id %d: %w", id, err)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("system with id %d: %w", id, consts.ErrNotFound)
	}
	return nil
}

func ListEnabledSystems(db *gorm.DB) ([]database.System, error) {
	var systems []database.System
	if err := db.Where("status = ?", consts.CommonEnabled).Find(&systems).Error; err != nil {
		return nil, fmt.Errorf("failed to list enabled systems: %w", err)
	}
	return systems, nil
}
