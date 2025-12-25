package producer

import (
	"errors"
	"fmt"

	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"aegis/utils"

	"gorm.io/gorm"
)

// configHistoryParams encapsulates parameters for creating config history entries
type configHistoryParams struct {
	changeType     consts.ConfigHistoryChangeType
	changeField    consts.ConfigHistoryChangeField
	oldValue       string
	newValue       string
	reason         string
	configID       int
	operatorID     *int
	ipAddress      string
	userAgent      string
	rollbackFromID *int
}

// =====================================================================
// Configuration Service Layer
// =====================================================================

// GetConfigDetail retrieves detailed information about a configuration by its key
func GetConfigDetail(containerID int) (*dto.ConfigDetailResp, error) {
	config, err := repository.GetConfigByID(database.DB, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get config detail: %w", err)
	}

	histories, err := repository.ListConfigHistoriesByConfigID(database.DB, config.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get config histories: %w", err)
	}

	resp := dto.NewConfigDetailResp(config)
	for _, history := range histories {
		resp.Histories = append(resp.Histories, *dto.NewConfigHistoryResp(&history))
	}

	return resp, nil
}

// ListConfigs lists configurations based on the provided filters
func ListConfigs(req *dto.ListConfigReq) (*dto.ListResp[dto.ConfigResp], error) {
	limit, offset := req.ToGormParams()

	configs, total, err := repository.ListConfigs(database.DB, limit, offset, req.ValueType, req.IsSecret, req.UpdatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}

	configResps := make([]dto.ConfigResp, 0, len(configs))
	for _, config := range configs {
		configResps = append(configResps, *dto.NewConfigResp(&config))
	}

	resp := dto.ListResp[dto.ConfigResp]{
		Items:      configResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// UpdateConfig updates an existing configuration with validation and history tracking
func UpdateConfig(req *dto.UpdateConfigReq, configID, operatorID int, ipAddress, userAgent string) (*dto.ConfigResp, error) {
	var updatedConfig *database.DynamicConfig

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		existingConfig, err := repository.GetConfigByID(tx, configID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: configuration with id %d not found", consts.ErrNotFound, configID)
			}
		}

		oldValue := existingConfig.Value

		req.PatchConfigModel(existingConfig)
		existingConfig.UpdatedBy = utils.IntPtr(operatorID)

		if err := common.ValidateConfig(existingConfig); err != nil {
			return fmt.Errorf("invalid config value: %w", err)
		}

		if err := repository.UpdateConfig(tx, existingConfig); err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}

		updatedConfig = existingConfig

		if err := createConfigHistory(tx, configHistoryParams{
			changeType:  consts.ChangeTypeUpdate,
			changeField: req.GetChangeField(),
			oldValue:    oldValue,
			newValue:    updatedConfig.Value,
			reason:      req.Reason,
			configID:    updatedConfig.ID,
			operatorID:  updatedConfig.UpdatedBy,
			ipAddress:   ipAddress,
			userAgent:   userAgent,
		}); err != nil {
			return fmt.Errorf("failed to create config history: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Trigger config reload and callbacks for injection category changes
	if updatedConfig.Category == "injection" {
		if err := config.GetChaosSystemConfigManager().Reload(); err != nil {
			return nil, fmt.Errorf("failed to reload system config after updating injection config: %w", err)
		}
	}

	return dto.NewConfigResp(updatedConfig), nil
}

// RollbackConfig rolls back a configuration to a previous value from history
func RollbackConfig(req *dto.RollbackConfigReq, configID, operatorID int, ipAddress, userAgent string) (*dto.ConfigResp, error) {
	var rollbackedConfig *database.DynamicConfig

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		existingConfig, err := repository.GetConfigByID(tx, configID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: configuration with id %d not found", consts.ErrNotFound, configID)
			}
		}

		existingConfig.UpdatedBy = utils.IntPtr(operatorID)

		history, err := repository.GetConfigHistory(tx, req.HistoryID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: history entry with id %d not found", consts.ErrNotFound, req.HistoryID)
			}
		}

		// Rollback the specific field to its old value
		oldValue, newValue, err := rollbackFieldValue(existingConfig, history.ChangeField, history.OldValue)
		if err != nil {
			return fmt.Errorf("failed to rollback field: %w", err)
		}

		if err := common.ValidateConfig(existingConfig); err != nil {
			return fmt.Errorf("invalid config value: %w", err)
		}

		if err := repository.UpdateConfig(tx, existingConfig); err != nil {
			return fmt.Errorf("failed to rollback config: %w", err)
		}

		rollbackedConfig = existingConfig

		if err := createConfigHistory(tx, configHistoryParams{
			changeType:     consts.ChangeTypeRollback,
			changeField:    history.ChangeField,
			oldValue:       oldValue,
			newValue:       newValue,
			reason:         req.Reason,
			configID:       configID,
			operatorID:     utils.IntPtr(operatorID),
			ipAddress:      ipAddress,
			userAgent:      userAgent,
			rollbackFromID: utils.IntPtr(req.HistoryID),
		}); err != nil {
			return fmt.Errorf("failed to create config history: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewConfigResp(rollbackedConfig), nil
}

// ===================== ConfigHistory =====================

func ListConfigHistories(req *dto.ListConfigHistoryReq, configID int) (*dto.ListResp[dto.ConfigHistoryResp], error) {
	limit, offset := req.ToGormParams()

	histories, total, err := repository.ListConfigHistories(database.DB, limit, offset, configID, req.ChangeType, req.OperatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to list config histories: %w", err)
	}

	historyResps := make([]dto.ConfigHistoryResp, 0, len(histories))
	for _, history := range histories {
		historyResps = append(historyResps, *dto.NewConfigHistoryResp(&history))
	}

	resp := dto.ListResp[dto.ConfigHistoryResp]{
		Items:      historyResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// createConfigHistory creates a ConfigHistory entry from a config update
func createConfigHistory(db *gorm.DB, params configHistoryParams) error {
	entry := &database.ConfigHistory{
		ChangeType:       params.changeType,
		OldValue:         params.oldValue,
		NewValue:         params.newValue,
		Reason:           params.reason,
		ConfigID:         params.configID,
		OperatorID:       params.operatorID,
		IPAddress:        params.ipAddress,
		UserAgent:        params.userAgent,
		RolledBackFromID: params.rollbackFromID,
	}
	if err := repository.CreateConfigHistory(db, entry); err != nil {
		return fmt.Errorf("failed to create config history: %w", err)
	}
	return nil
}

// rollbackFieldValue rolls back a specific field in the config based on the change field type
// Returns the old value (before rollback) and new value (after rollback)
func rollbackFieldValue(config *database.DynamicConfig, changeField consts.ConfigHistoryChangeField, targetValue string) (oldValue string, newValue string, err error) {
	newValue = targetValue

	switch changeField {
	case consts.ChangeFieldValue:
		oldValue = config.Value
		config.Value = newValue

	case consts.ChangeFieldDefaultValue:
		oldValue = config.DefaultValue
		config.DefaultValue = newValue

	case consts.ChangeFieldDescription:
		oldValue = config.Description
		config.Description = newValue

	case consts.ChangeFieldMinValue:
		if config.MinValue != nil {
			oldValue = fmt.Sprintf("%f", *config.MinValue)
		}
		if newValue == "" {
			config.MinValue = nil
		} else {
			var minVal float64
			if _, err := fmt.Sscanf(newValue, "%f", &minVal); err != nil {
				return "", "", fmt.Errorf("failed to parse min value: %w", err)
			}
			config.MinValue = &minVal
		}

	case consts.ChangeFieldMaxValue:
		if config.MaxValue != nil {
			oldValue = fmt.Sprintf("%f", *config.MaxValue)
		}
		if newValue == "" {
			config.MaxValue = nil
		} else {
			var maxVal float64
			if _, err := fmt.Sscanf(newValue, "%f", &maxVal); err != nil {
				return "", "", fmt.Errorf("failed to parse max value: %w", err)
			}
			config.MaxValue = &maxVal
		}

	case consts.ChangeFieldPattern:
		oldValue = config.Pattern
		config.Pattern = newValue

	case consts.ChangeFieldOptions:
		oldValue = config.Options
		config.Options = newValue

	default:
		return "", "", fmt.Errorf("unknown change field: %d", changeField)
	}

	return oldValue, newValue, nil
}
