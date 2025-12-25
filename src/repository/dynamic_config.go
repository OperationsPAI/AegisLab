package repository

import (
	"fmt"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
)

const (
	configOmitFields = "active_key"
)

// =====================================================================
// DynamicConfig Repository Functions
// =====================================================================

// CreateConfig creates a new configuration item
func CreateConfig(db *gorm.DB, config *database.DynamicConfig) error {
	if err := db.Omit(configOmitFields).Create(config).Error; err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	return nil
}

// GetConfigByID retrieves a configuration by its ID
func GetConfigByID(db *gorm.DB, configID int) (*database.DynamicConfig, error) {
	var config database.DynamicConfig
	if err := db.
		Preload("UpdatedByUser").
		Where("id = ?", configID).
		First(&config).Error; err != nil {
		return nil, fmt.Errorf("failed to find config with id %d: %w", configID, err)
	}
	return &config, nil
}

// List ExistingConfigs lists all existing configurations
func ListExistingConfigs(db *gorm.DB) ([]database.DynamicConfig, error) {
	var configs []database.DynamicConfig
	if err := db.
		Order("config_key ASC").
		Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to list all existing configs: %w", err)
	}
	return configs, nil
}

// ListConfigs lists configs based on filter options
func ListConfigs(db *gorm.DB, limit, offset int, valueType *consts.ConfigValueType, isSecret *bool, updatedBy *int) ([]database.DynamicConfig, int64, error) {
	var configs []database.DynamicConfig
	var total int64

	query := db.Model(&database.DynamicConfig{})
	if valueType != nil {
		query = query.Where("value_type = ?", *valueType)
	}
	if isSecret != nil {
		query = query.Where("is_secret = ?", *isSecret)
	}
	if updatedBy != nil {
		query = query.Where("updated_by = ?", *updatedBy)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count configs: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("created DESC").Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list configs: %w", err)
	}

	return configs, total, nil
}

// UpdateConfig updates a configuration item
func UpdateConfig(db *gorm.DB, config *database.DynamicConfig) error {
	if err := db.Omit(configOmitFields).Save(config).Error; err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}
	return nil
}

// =====================================================================
// ConfigHistory Repository Functions
// =====================================================================

// CreateConfigHistory creates a new history record
func CreateConfigHistory(db *gorm.DB, history *database.ConfigHistory) error {
	if err := db.Create(history).Error; err != nil {
		return fmt.Errorf("failed to create config history: %w", err)
	}
	return nil
}

// GetConfigHistory retrieves a specific history entry by ID
func GetConfigHistory(db *gorm.DB, historyID int) (*database.ConfigHistory, error) {
	var history database.ConfigHistory
	if err := db.
		Preload("Operator").
		Preload("Config").
		First(&history, historyID).Error; err != nil {
		return nil, fmt.Errorf("failed to find config history with id %d: %w", historyID, err)
	}
	return &history, nil
}

// GetLatestConfigHistory retrieves the most recent configuration change
func GetLatestConfigHistory(db *gorm.DB) (*database.ConfigHistory, error) {
	var history database.ConfigHistory
	if err := db.
		Preload("Operator").
		Preload("Config").
		Order("created_at DESC").
		First(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest config history: %w", err)
	}
	return &history, nil
}

// ListConfigHistories lists configuration history entries with pagination and optional filters
func ListConfigHistories(db *gorm.DB, limit, offset int, configID int, changeType *consts.ConfigHistoryChangeType, operatorID *int) ([]database.ConfigHistory, int64, error) {
	var histories []database.ConfigHistory
	var total int64

	query := db.Model(&database.ConfigHistory{}).
		Where("config_id = ?", configID)

	if changeType != nil {
		query = query.Where("change_type = ?", *changeType)
	}
	if operatorID != nil {
		query = query.Where("operator_id = ?", *operatorID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count config histories: %w", err)
	}

	if err := query.
		Preload("Operator").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&histories).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list config histories: %w", err)
	}

	return histories, total, nil
}

// ListConfigHistoriesByConfigID lists all history entries for a specific configuration
func ListConfigHistoriesByConfigID(db *gorm.DB, configID int) ([]database.ConfigHistory, error) {
	var histories []database.ConfigHistory
	if err := db.
		Preload("Operator").
		Where("config_id = ?", configID).
		Order("created_at DESC").
		Find(&histories).Error; err != nil {
		return nil, fmt.Errorf("failed to list config histories for config %d: %w", configID, err)
	}
	return histories, nil
}
