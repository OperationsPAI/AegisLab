package systemmodule

import (
	"aegis/consts"
	"aegis/model"
	"fmt"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) withDB(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *Repository) GetAuditLogByID(id int) (*model.AuditLog, error) {
	var auditLog model.AuditLog
	if err := r.db.Where("id = ?", id).First(&auditLog).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}
	return &auditLog, nil
}

func (r *Repository) ListAuditLogs(limit, offset int, filters *ListAuditLogFilters) ([]model.AuditLog, int64, error) {
	var (
		logs  []model.AuditLog
		total int64
	)

	query := r.db.Model(&model.AuditLog{}).Preload("User").Preload("Resource")
	if filters != nil {
		if filters.Action != "" {
			query = query.Where("action = ?", filters.Action)
		}
		if filters.IPAddress != "" {
			query = query.Where("ip_address = ?", filters.IPAddress)
		}
		if filters.UserID != 0 {
			query = query.Where("user_id = ?", filters.UserID)
		}
		if filters.ResourceID != 0 {
			query = query.Where("resource_id = ?", filters.ResourceID)
		}
		if filters.State != nil {
			query = query.Where("state = ?", *filters.State)
		}
		if filters.Status != nil {
			query = query.Where("status = ?", *filters.Status)
		}
		if filters.StartTime != nil {
			query = query.Where("created_at >= ?", *filters.StartTime)
		}
		if filters.EndTime != nil {
			query = query.Where("created_at <= ?", *filters.EndTime)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}
	return logs, total, nil
}

func (r *Repository) GetConfigByID(configID int, includeUser bool) (*model.DynamicConfig, error) {
	query := r.db
	if includeUser {
		query = query.Preload("UpdatedByUser")
	}

	var cfg model.DynamicConfig
	if err := query.Where("id = ?", configID).First(&cfg).Error; err != nil {
		return nil, fmt.Errorf("failed to find config with id %d: %w", configID, err)
	}
	return &cfg, nil
}

func (r *Repository) ListConfigs(limit, offset int, valueType *consts.ConfigValueType, category *string, isSecret *bool, updatedBy *int) ([]model.DynamicConfig, int64, error) {
	var (
		configs []model.DynamicConfig
		total   int64
	)

	query := r.db.Model(&model.DynamicConfig{})
	if valueType != nil {
		query = query.Where("value_type = ?", *valueType)
	}
	if category != nil {
		query = query.Where("category = ?", *category)
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
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list configs: %w", err)
	}
	return configs, total, nil
}

func (r *Repository) UpdateConfig(config *model.DynamicConfig) error {
	if err := r.db.Save(config).Error; err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}
	return nil
}

func (r *Repository) GetConfigHistory(historyID int) (*model.ConfigHistory, error) {
	var history model.ConfigHistory
	if err := r.db.Preload("Operator").Preload("Config").First(&history, historyID).Error; err != nil {
		return nil, fmt.Errorf("failed to find config history with id %d: %w", historyID, err)
	}
	return &history, nil
}

func (r *Repository) CreateConfigHistory(history *model.ConfigHistory) error {
	if err := r.db.Create(history).Error; err != nil {
		return fmt.Errorf("failed to create config history: %w", err)
	}
	return nil
}

func (r *Repository) ListConfigHistories(limit, offset int, configID int, changeType *consts.ConfigHistoryChangeType, operatorID *int) ([]model.ConfigHistory, int64, error) {
	var (
		histories []model.ConfigHistory
		total     int64
	)

	query := r.db.Model(&model.ConfigHistory{}).Where("config_id = ?", configID)
	if changeType != nil {
		query = query.Where("change_type = ?", *changeType)
	}
	if operatorID != nil {
		query = query.Where("operator_id = ?", *operatorID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count config histories: %w", err)
	}
	if err := query.Preload("Operator").Limit(limit).Offset(offset).Order("created_at DESC").Find(&histories).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list config histories: %w", err)
	}
	return histories, total, nil
}

func (r *Repository) ListConfigHistoriesByConfigID(configID int) ([]model.ConfigHistory, error) {
	var histories []model.ConfigHistory
	if err := r.db.Preload("Operator").Where("config_id = ?", configID).Order("created_at DESC").Find(&histories).Error; err != nil {
		return nil, fmt.Errorf("failed to list config histories for config %d: %w", configID, err)
	}
	return histories, nil
}
