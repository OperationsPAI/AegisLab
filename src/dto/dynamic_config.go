package dto

import (
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/database"
)

// =====================================================================
// Configuration DTOs
// =====================================================================

// ListConfigReq represents config list query parameters
type ListConfigReq struct {
	PaginationReq
	ValueType *consts.ConfigValueType `form:"value_type" binding:"omitempty"`
	IsSecret  *bool                   `form:"is_secret" binding:"omitempty"`
	Status    *consts.StatusType      `form:"status" binding:"omitempty"`
	UpdatedBy *int                    `form:"updated_by" binding:"omitempty,min_ptr=1"`
}

func (req *ListConfigReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if err := validateValuteType(req.ValueType); err != nil {
		return err
	}
	return validateStatusField(req.Status, false)
}

// RollbackConfigReq represents a request to rollback a configuration
type RollbackConfigReq struct {
	HistoryID int    `json:"history_id" binding:"required,min=1"`
	Reason    string `json:"reason" binding:"required"`
}

// UpdateConfigReq represents a request to update a configuration item
// Supports partial updates - only provided fields will be updated
type UpdateConfigReq struct {
	// Runtime value
	Value *string `json:"value" binding:"omitempty"`

	// Metadata fields
	DefaultValue *string  `json:"default_value" binding:"omitempty"`
	Description  string   `json:"description" binding:"omitempty"`
	MinValue     *float64 `json:"min_value" binding:"omitempty"`
	MaxValue     *float64 `json:"max_value" binding:"omitempty"`
	Pattern      *string  `json:"pattern" binding:"omitempty"`
	Options      *string  `json:"options" binding:"omitempty"`

	// Audit trail
	Reason string `json:"reason" binding:"required"`
}

func (req *UpdateConfigReq) Validate() error {
	// Check if at least one field is being updated
	hasValueUpdate := req.Value != nil
	if !hasValueUpdate && !req.HasMetadataUpdate() {
		return fmt.Errorf("at least one field must be provided for update")
	}
	if hasValueUpdate && req.HasMetadataUpdate() {
		return fmt.Errorf("cannot update value and metadata fields in the same request")
	}

	// If updating metadata, only allow ONE metadata field at a time
	if req.HasMetadataUpdate() {
		metadataFieldCount := 0
		if req.DefaultValue != nil {
			metadataFieldCount++
		}
		if req.Description != "" {
			metadataFieldCount++
		}
		if req.MinValue != nil {
			metadataFieldCount++
		}
		if req.MaxValue != nil {
			metadataFieldCount++
		}
		if req.Pattern != nil {
			metadataFieldCount++
		}
		if req.Options != nil {
			metadataFieldCount++
		}
		if metadataFieldCount > 1 {
			return fmt.Errorf("can only update one metadata field at a time")
		}
	}

	return nil
}

func (req *UpdateConfigReq) PatchConfigModel(target *database.DynamicConfig) {
	if req.Value != nil {
		target.Value = *req.Value
	}
	if req.DefaultValue != nil {
		target.DefaultValue = *req.DefaultValue
	}
	if req.Description != "" {
		target.Description = req.Description
	}
	if req.MinValue != nil {
		target.MinValue = req.MinValue
	}
	if req.MaxValue != nil {
		target.MaxValue = req.MaxValue
	}
	if req.Pattern != nil {
		target.Pattern = *req.Pattern
	}
	if req.Options != nil {
		target.Options = *req.Options
	}
}

// HasMetadataUpdate returns true if the request updates any metadata fields
func (req *UpdateConfigReq) HasMetadataUpdate() bool {
	return req.DefaultValue != nil || req.Description != "" || req.MinValue != nil || req.MaxValue != nil ||
		req.Pattern != nil || req.Options != nil
}

// GetChangeField returns the specific metadata field being changed
func (req *UpdateConfigReq) GetChangeField() consts.ConfigHistoryChangeField {
	if req.DefaultValue != nil {
		return consts.ChangeFieldDefaultValue
	}
	if req.Description != "" {
		return consts.ChangeFieldDescription
	}
	if req.MinValue != nil {
		return consts.ChangeFieldMinValue
	}
	if req.MaxValue != nil {
		return consts.ChangeFieldMaxValue
	}
	if req.Pattern != nil {
		return consts.ChangeFieldPattern
	}
	if req.Options != nil {
		return consts.ChangeFieldOptions
	}
	return consts.ChangeFieldValue
}

type ListConfigHistoryReq struct {
	PaginationReq
	ChangeType *consts.ConfigHistoryChangeType `form:"change_type" binding:"omitempty"`
	OperatorID *int                            `form:"operator_id" binding:"omitempty,min_ptr=1"`
}

func (req *ListConfigHistoryReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if req.ChangeType != nil {
		if _, ok := consts.ValidConfigHistoryChanteTypes[*req.ChangeType]; !ok {
			return fmt.Errorf("invalid change type: %v", req.ChangeType)
		}
	}
	return nil
}

// ConfigResp represents a configuration item response
type ConfigResp struct {
	ID            int       `json:"id"`
	Key           string    `json:"key"`
	Value         string    `json:"value"`
	ValueType     string    `json:"value_type"`
	Category      string    `json:"category"`
	UpdatedAt     time.Time `json:"updated_at"`
	UpdatedByID   int       `json:"updated_by_id"`
	UpdatedByName string    `json:"updated_by_name"`
}

// NewConfigResp converts a DynamicConfig entity to ConfigResp DTO
func NewConfigResp(config *database.DynamicConfig) *ConfigResp {
	resp := &ConfigResp{
		ID:        config.ID,
		Key:       config.Key,
		Value:     config.Value,
		ValueType: consts.GetDynamicConfigTypeName(config.ValueType),
		Category:  config.Category,
		UpdatedAt: config.UpdatedAt,
	}

	if config.UpdatedByUser != nil {
		resp.UpdatedByName = config.UpdatedByUser.Username
	}

	// Mask secret values
	if config.IsSecret {
		resp.Value = maskSecretValue(config.Value)
	}

	return resp
}

type ConfigDetailResp struct {
	ConfigResp

	DefaultValue string              `json:"default_value"`
	Description  string              `json:"description"`
	MinValue     *float64            `json:"min_value,omitempty"`
	MaxValue     *float64            `json:"max_value,omitempty"`
	Pattern      string              `json:"pattern,omitempty"`
	Options      string              `json:"options,omitempty"`
	Histories    []ConfigHistoryResp `json:"histories,omitempty"`
}

func NewConfigDetailResp(config *database.DynamicConfig) *ConfigDetailResp {
	return &ConfigDetailResp{
		ConfigResp:   *NewConfigResp(config),
		DefaultValue: config.DefaultValue,
		Description:  config.Description,
		MinValue:     config.MinValue,
		MaxValue:     config.MaxValue,
		Pattern:      config.Pattern,
		Options:      config.Options,
	}
}

// ConfigHistoryResp represents a configuration change history entry response
type ConfigHistoryResp struct {
	ID               int       `json:"id"`
	ChangeType       string    `json:"change_type"`
	OldValue         string    `json:"old_value"`
	NewValue         string    `json:"new_value"`
	Reason           string    `json:"reason"`
	ConfigID         int       `json:"config_id"`
	OperatorID       *int      `json:"operator_id"`
	OperatorName     string    `json:"operator_name,omitempty"`
	IPAddress        string    `json:"ip_address,omitempty"`
	UserAgent        string    `json:"user_agent,omitempty"`
	RolledBackFromID *int      `json:"rolled_back_from_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

func NewConfigHistoryResp(history *database.ConfigHistory) *ConfigHistoryResp {
	resp := &ConfigHistoryResp{
		ID:               history.ID,
		ChangeType:       consts.GetConfigHistoryChangeTypeName(history.ChangeType),
		ConfigID:         history.ConfigID,
		OldValue:         history.OldValue,
		NewValue:         history.NewValue,
		Reason:           history.Reason,
		OperatorID:       history.OperatorID,
		IPAddress:        history.IPAddress,
		UserAgent:        history.UserAgent,
		RolledBackFromID: history.RolledBackFromID,
		CreatedAt:        history.CreatedAt,
	}

	if history.Operator != nil {
		resp.OperatorName = history.Operator.Username
	}
	return resp
}

// ConfigStatsResp represents statistics about the configuration system
type ConfigStatsResp struct {
	TotalConfigs   int       `json:"total_configs"`
	DynamicConfigs int       `json:"dynamic_configs"`
	StaticConfigs  int       `json:"static_configs"`
	TotalChanges   int       `json:"total_changes"`
	ChangesLast24h int       `json:"changes_last_24h"`
	Categories     []string  `json:"categories"`
	LastUpdate     time.Time `json:"last_update,omitempty"`
}

// maskSecretValue masks sensitive configuration values
func maskSecretValue(value string) string {
	maskLen := min(len(value), 8)
	if maskLen == 0 {
		return ""
	}
	return strings.Repeat("*", maskLen)
}

// validateValuteType checks if the provided config value type is valid
func validateValuteType(valueType *consts.ConfigValueType) error {
	if valueType != nil {
		if _, ok := consts.ValidDynamicConfigTypes[*valueType]; !ok {
			return fmt.Errorf("invalid value type: %v", valueType)
		}
	}
	return nil
}
