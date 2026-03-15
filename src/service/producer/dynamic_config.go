package producer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"aegis/client"
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// =====================================================================
// Private Types (migrated from common for producer-only usage)
// =====================================================================

// configUpdateContext holds context information for a configuration update
type configUpdateContext struct {
	ChangeField consts.ConfigHistoryChangeField
	OldValue    string
	NewValue    string
	Reason      string
	OperatorID  int
	IpAddress   string
	UserAgent   string
}

// configHistoryParams encapsulates parameters for creating config history entries
type configHistoryParams struct {
	ConfigID       int
	ChangeType     consts.ConfigHistoryChangeType
	RollbackFromID *int

	ConfigUpdateContext configUpdateContext
}

// =====================================================================
// Configuration Service Layer
// =====================================================================

// etcdPrefixForScope returns the etcd key prefix for the given config scope.
func etcdPrefixForScope(scope consts.ConfigScope) string {
	switch scope {
	case consts.ConfigScopeProducer:
		return consts.ConfigEtcdProducerPrefix
	case consts.ConfigScopeConsumer:
		return consts.ConfigEtcdConsumerPrefix
	case consts.ConfigScopeGlobal:
		return consts.ConfigEtcdGlobalPrefix
	}
	return ""
}

// GetConfigDetail retrieves detailed information about a configuration by its key
func GetConfigDetail(containerID int) (*dto.ConfigDetailResp, error) {
	config, err := repository.GetConfigByID(database.DB, containerID, true)
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

	configs, total, err := repository.ListConfigs(database.DB, limit, offset, req.ValueType, req.Category, req.IsSecret, req.UpdatedBy)
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

// RollbackConfigValue rolls back a configuration value from history
func RollbackConfigValue(ctx context.Context, req *dto.RollbackConfigReq, configID, operatorID int, ipAddress, userAgent string) error {
	// Get the history entry to rollback to
	history, err := repository.GetConfigHistory(database.DB, req.HistoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: history entry with id %d not found", consts.ErrNotFound, req.HistoryID)
		}
		return fmt.Errorf("failed to get config history: %w", err)
	}

	// Validate this is a value change history
	if history.ChangeField != consts.ChangeFieldValue {
		return fmt.Errorf("history entry %d is not a value change (field: %v)", req.HistoryID, history.ChangeField)
	}

	// Get existing config
	existingConfig, err := repository.GetConfigByID(database.DB, configID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: configuration with id %d not found", consts.ErrNotFound, configID)
		}
		return fmt.Errorf("failed to get config: %w", err)
	}

	oldValue, err := client.EtcdGet(ctx, fmt.Sprintf("%s%s", etcdPrefixForScope(existingConfig.Scope), existingConfig.Key))
	if err != nil {
		return fmt.Errorf("failed to get current config value from etcd: %w", err)
	}

	newValue := history.OldValue

	if err := common.ValidateConfig(existingConfig, newValue); err != nil {
		return fmt.Errorf("invalid config after rollback: %w", err)
	}

	if err := setViperIfNeeded(existingConfig, newValue); err != nil {
		return fmt.Errorf("failed to set config value in viper: %w", err)
	}

	if _, err := createConfigRollback(existingConfig, utils.IntPtr(history.ID), configUpdateContext{
		ChangeField: consts.ChangeFieldValue,
		OldValue:    oldValue,
		NewValue:    newValue,
		Reason:      req.Reason,
		OperatorID:  operatorID,
		IpAddress:   ipAddress,
		UserAgent:   userAgent,
	}); err != nil {
		return err
	}

	return propagateValueChange(ctx, existingConfig, newValue, "rollback")
}

// RollbackConfigMetadata rolls back a configuration metadata field from history
func RollbackConfigMetadata(req *dto.RollbackConfigReq, configID, operatorID int, ipAddress, userAgent string) (*dto.ConfigResp, error) {
	// Get the history entry to rollback to
	history, err := repository.GetConfigHistory(database.DB, req.HistoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: history entry with id %d not found", consts.ErrNotFound, req.HistoryID)
		}
		return nil, fmt.Errorf("failed to get config history: %w", err)
	}

	// Validate this is a metadata change history
	if history.ChangeField == consts.ChangeFieldValue {
		return nil, fmt.Errorf("history entry %d is a value change, use RollbackConfigValue instead", req.HistoryID)
	}

	// Get existing config
	existingConfig, err := repository.GetConfigByID(database.DB, configID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: configuration with id %d not found", consts.ErrNotFound, configID)
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	// Rollback the metadata field
	oldValue, newValue, err := rollbackMetaFieldValue(existingConfig, history.ChangeField, history.OldValue)
	if err != nil {
		return nil, fmt.Errorf("failed to rollback metadata field: %w", err)
	}

	// Validate the configuration after metadata rollback
	if err := common.ValidateConfigMetadataConstraints(existingConfig); err != nil {
		return nil, fmt.Errorf("invalid config after metadata rollback: %w", err)
	}

	// Save to database with rollback history
	updatedConfig, err := createConfigRollback(existingConfig, utils.IntPtr(history.ID), configUpdateContext{
		ChangeField: history.ChangeField,
		OldValue:    oldValue,
		NewValue:    newValue,
		Reason:      req.Reason,
		OperatorID:  operatorID,
		IpAddress:   ipAddress,
		UserAgent:   userAgent,
	})
	if err != nil {
		return nil, err
	}

	return dto.NewConfigResp(updatedConfig), nil
}

// UpdateConfigValue updates the value of a configuration and handles propagation based on its scope
func UpdateConfigValue(ctx context.Context, req *dto.UpdateConfigValueReq, configID, operatorID int, ipAddress, userAgent string) error {
	existingConfig, err := repository.GetConfigByID(database.DB, configID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: configuration with id %d not found", consts.ErrNotFound, configID)
		}
	}

	oldValue, err := client.EtcdGet(ctx, fmt.Sprintf("%s%s", etcdPrefixForScope(existingConfig.Scope), existingConfig.Key))
	if err != nil {
		return fmt.Errorf("failed to get current config value from etcd: %w", err)
	}

	newValue := req.Value

	if err := common.ValidateConfig(existingConfig, newValue); err != nil {
		return fmt.Errorf("invalid config value: %w", err)
	}

	if err := setViperIfNeeded(existingConfig, newValue); err != nil {
		return fmt.Errorf("failed to set config value in viper: %w", err)
	}

	if err := createConfigHistory(database.DB, configHistoryParams{
		ConfigID:   existingConfig.ID,
		ChangeType: consts.ChangeTypeUpdate,
		ConfigUpdateContext: configUpdateContext{
			ChangeField: consts.ChangeFieldValue,
			OldValue:    oldValue,
			NewValue:    newValue,
			Reason:      req.Reason,
			OperatorID:  operatorID,
			IpAddress:   ipAddress,
			UserAgent:   userAgent,
		},
	}); err != nil {
		return fmt.Errorf("failed to create config history: %w", err)
	}

	return propagateValueChange(ctx, existingConfig, newValue, "update")
}

// UpdateConfigMetadata updates the metadata of a configuration
func UpdateConfigMetadata(req *dto.UpdateConfigMetadataReq, configID, operatorID int, ipAddress, userAgent string) (*dto.ConfigResp, error) {
	existingConfig, err := repository.GetConfigByID(database.DB, configID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: configuration with id %d not found", consts.ErrNotFound, configID)
		}
	}

	oldValue, newValue := req.PatchConfigModel(existingConfig)

	// Validate the configuration after metadata update
	if err := common.ValidateConfigMetadataConstraints(existingConfig); err != nil {
		return nil, fmt.Errorf("invalid config after metadata update: %w", err)
	}

	var updatedConfig *database.DynamicConfig
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		existingConfig.UpdatedBy = utils.IntPtr(operatorID)

		if err := repository.UpdateConfig(tx, existingConfig); err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}

		updatedConfig = existingConfig

		if err := createConfigHistory(tx, configHistoryParams{
			ConfigID:   updatedConfig.ID,
			ChangeType: consts.ChangeTypeUpdate,
			ConfigUpdateContext: configUpdateContext{
				ChangeField: req.GetChangeField(),
				OldValue:    oldValue,
				NewValue:    newValue,
				Reason:      req.Reason,
				OperatorID:  operatorID,
				IpAddress:   ipAddress,
				UserAgent:   userAgent,
			},
		}); err != nil {
			return fmt.Errorf("failed to create config history: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewConfigResp(updatedConfig), nil
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

// ===================== Helper Functions =====================

// createConfigHistory creates a ConfigHistory entry from a config update (private version)
func createConfigHistory(db *gorm.DB, params configHistoryParams) error {
	entry := &database.ConfigHistory{
		ChangeType:       params.ChangeType,
		OldValue:         params.ConfigUpdateContext.OldValue,
		NewValue:         params.ConfigUpdateContext.NewValue,
		Reason:           params.ConfigUpdateContext.Reason,
		ConfigID:         params.ConfigID,
		OperatorID:       utils.IntPtr(params.ConfigUpdateContext.OperatorID),
		IPAddress:        params.ConfigUpdateContext.IpAddress,
		UserAgent:        params.ConfigUpdateContext.UserAgent,
		RolledBackFromID: params.RollbackFromID,
		ChangeField:      params.ConfigUpdateContext.ChangeField,
	}
	if err := repository.CreateConfigHistory(db, entry); err != nil {
		return fmt.Errorf("failed to create config history: %w", err)
	}
	return nil
}

// createConfigRollback updates config and creates a rollback history entry
// This wraps the common history creation logic but with rollback-specific parameters
func createConfigRollback(config *database.DynamicConfig, historyID *int, updateContext configUpdateContext) (*database.DynamicConfig, error) {
	var updatedConfig *database.DynamicConfig

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Update the config in database
		if err := repository.UpdateConfig(tx, config); err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}

		updatedConfig = config

		// Create rollback history entry using common function
		if err := createConfigHistory(tx, configHistoryParams{
			ConfigID:            config.ID,
			ChangeType:          consts.ChangeTypeRollback,
			ConfigUpdateContext: updateContext,
			RollbackFromID:      historyID,
		}); err != nil {
			return fmt.Errorf("failed to create rollback history: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return updatedConfig, nil
}

// rollbackMetaFieldValue rolls back a specific field in the config based on the change field type
// Returns the old value (before rollback) and new value (after rollback)
func rollbackMetaFieldValue(config *database.DynamicConfig, changeField consts.ConfigHistoryChangeField, targetValue string) (oldValue string, newValue string, err error) {
	newValue = targetValue

	switch changeField {
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

// setViperIfNeeded updates the local Viper cache for scopes that need immediate local reflection
// (producer and global). Consumer configs live only in etcd and are applied remotely.
func setViperIfNeeded(cfg *database.DynamicConfig, newValue string) error {
	if cfg.Scope == consts.ConfigScopeConsumer {
		return nil
	}
	return config.SetViperValue(cfg.Key, newValue, cfg.ValueType)
}

// propagateValueChange publishes the new value to etcd and, for consumer scope, waits for ack.
// Producer scope requires no network propagation, so this is a no-op for that scope.
func propagateValueChange(ctx context.Context, cfg *database.DynamicConfig, newValue, opDesc string) error {
	if cfg.Scope != consts.ConfigScopeGlobal && cfg.Scope != consts.ConfigScopeConsumer {
		return nil
	}

	etcdKey := fmt.Sprintf("%s%s", etcdPrefixForScope(cfg.Scope), cfg.Key)
	if err := publishConfigToEtcdWithRetry(etcdKey, newValue, 3); err != nil {
		return fmt.Errorf("config saved to database but failed to publish to etcd: %w", err)
	}

	if cfg.Scope == consts.ConfigScopeConsumer {
		logrus.Infof("Waiting for consumer config %s response...", opDesc)
		resp, err := waitForConfigUpdateResponse(10 * time.Second)
		if err != nil {
			return fmt.Errorf("config %s but consumer did not respond: %w", opDesc, err)
		}
		if !resp.Success {
			return fmt.Errorf("consumer failed to process config %s: %s", opDesc, resp.Error)
		}
		logrus.Infof("Config %s successfully processed by consumer", opDesc)
	}

	return nil
}

// publishConfigToEtcdWithRetry publishes configuration to etcd with exponential backoff retry
func publishConfigToEtcdWithRetry(key, value string, maxRetries int) error {
	var lastErr error
	baseDelay := 500 * time.Millisecond

	for attempt := range maxRetries {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			logrus.Warnf("Retrying etcd publish after %v (attempt %d/%d)", delay, attempt+1, maxRetries)
			time.Sleep(delay)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := client.EtcdPut(ctx, key, value, 0)
		cancel()

		if err == nil {
			if attempt > 0 {
				logrus.Infof("Successfully published config to etcd after %d retries", attempt)
			}
			return nil
		}

		lastErr = err
		logrus.Warnf("Failed to publish config to etcd (attempt %d/%d): %v", attempt+1, maxRetries, err)
	}

	return fmt.Errorf("failed to publish config to etcd after %d attempts: %w", maxRetries, lastErr)
}

// waitForConfigUpdateResponse uses Redis Pub/Sub to synchronously wait for a response to a configuration update with timeout
func waitForConfigUpdateResponse(timeout time.Duration) (*dto.ConfigUpdateResponse, error) {
	redisClient := client.GetRedisClient()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pubsub := redisClient.Subscribe(ctx, consts.ConfigUpdateResponseChannel)
	defer func() { _ = pubsub.Close() }()

	if _, err := pubsub.Receive(ctx); err != nil {
		return nil, fmt.Errorf("failed to confirm subscription: %w", err)
	}

	msgChan := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for config update response after %v", timeout)

		case msg, ok := <-msgChan:
			if !ok {
				return nil, fmt.Errorf("subscription channel closed unexpectedly")
			}

			var response dto.ConfigUpdateResponse
			if err := json.Unmarshal([]byte(msg.Payload), &response); err != nil {
				logrus.Warnf("failed to parse response message: %v", err)
				continue
			}

			logrus.WithFields(logrus.Fields{
				"response_id": response.ID,
				"success":     response.Success,
			}).Info("Received matching config update response")
			return &response, nil
		}
	}
}
