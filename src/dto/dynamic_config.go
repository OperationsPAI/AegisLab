package dto

import (
	"encoding/json"
	"fmt"
	"time"

	"aegis/consts"
	"aegis/database"
	"aegis/utils"
)

// =====================================================================
// Configuration DTOs
// =====================================================================

// ListConfigReq represents config list query parameters
type ListConfigReq struct {
	PaginationReq
	ValueType *consts.ConfigValueType `form:"value_type" binding:"omitempty"`
	Category  *string                 `form:"category" binding:"omitempty"`
	IsSecret  *bool                   `form:"is_secret" binding:"omitempty"`
	UpdatedBy *int                    `form:"updated_by" binding:"omitempty,min_ptr=1"`
}

func (req *ListConfigReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if err := validateValuteType(req.ValueType); err != nil {
		return err
	}
	return nil
}

// RollbackConfigReq represents a request to rollback a configuration
type RollbackConfigReq struct {
	HistoryID int    `json:"history_id" binding:"required,min=1"`
	Reason    string `json:"reason" binding:"required"`
}

// UpdateConfigValueReq represents a request to update a configuration value (runtime config)
type UpdateConfigValueReq struct {
	Value  string `json:"value" binding:"required"`
	Reason string `json:"reason" binding:"required"`
}

// UpdateConfigMetadataReq represents a request to update configuration metadata
type UpdateConfigMetadataReq struct {
	// Metadata fields - only ONE field should be provided per request
	DefaultValue *string  `json:"default_value" binding:"omitempty"`
	Description  *string  `json:"description" binding:"omitempty"`
	MinValue     *float64 `json:"min_value" binding:"omitempty"`
	MaxValue     *float64 `json:"max_value" binding:"omitempty"`
	Pattern      *string  `json:"pattern" binding:"omitempty"`
	Options      *string  `json:"options" binding:"omitempty"`

	// Audit trail
	Reason string `json:"reason" binding:"required"`
}

func (req *UpdateConfigMetadataReq) Validate() error {
	// Count how many fields are being updated
	fieldCount := 0
	if req.DefaultValue != nil {
		fieldCount++
	}
	if req.Description != nil {
		fieldCount++
	}
	if req.MinValue != nil {
		fieldCount++
	}
	if req.MaxValue != nil {
		fieldCount++
	}
	if req.Pattern != nil {
		fieldCount++
	}
	if req.Options != nil {
		fieldCount++
	}

	if fieldCount == 0 {
		return fmt.Errorf("at least one metadata field must be provided for update")
	}
	if fieldCount > 1 {
		return fmt.Errorf("can only update one metadata field at a time")
	}

	return nil
}

func (req *UpdateConfigMetadataReq) PatchConfigModel(target *database.DynamicConfig) (string, string) {
	var oldValue string
	var newValue string

	if req.DefaultValue != nil {
		oldValue = target.DefaultValue
		newValue = *req.DefaultValue
		target.DefaultValue = *req.DefaultValue
	}
	if req.Description != nil {
		oldValue = target.Description
		newValue = *req.Description
		target.Description = *req.Description
	}
	if req.MinValue != nil {
		oldValue = fmt.Sprintf("%v", target.MinValue)
		newValue = fmt.Sprintf("%v", req.MinValue)
		target.MinValue = req.MinValue
	}
	if req.MaxValue != nil {
		oldValue = fmt.Sprintf("%v", target.MaxValue)
		newValue = fmt.Sprintf("%v", req.MaxValue)
		target.MaxValue = req.MaxValue
	}
	if req.Pattern != nil {
		oldValue = target.Pattern
		newValue = *req.Pattern
		target.Pattern = *req.Pattern
	}
	if req.Options != nil {
		oldValue = target.Options
		newValue = *req.Options
		target.Options = *req.Options
	}

	return oldValue, newValue
}

// GetChangeField returns the specific metadata field being changed
func (req *UpdateConfigMetadataReq) GetChangeField() consts.ConfigHistoryChangeField {
	if req.DefaultValue != nil {
		return consts.ChangeFieldDefaultValue
	}
	if req.Description != nil {
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
		ValueType: consts.GetDynamicConfigTypeName(config.ValueType),
		Category:  config.Category,
		UpdatedAt: config.UpdatedAt,
	}

	if config.UpdatedByUser != nil {
		resp.UpdatedByName = config.UpdatedByUser.Username
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
	LastUpdate     time.Time `json:"last_update"`
}

// ConfigUpdateResponse represents the response to a configuration update event
type ConfigUpdateResponse struct {
	ID          string    `json:"id"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
	ProcessedAt time.Time `json:"processed_at"`
	Payload     any       `json:"payload,omitempty"`
}

func NewConfigUpdateResponse() *ConfigUpdateResponse {
	return &ConfigUpdateResponse{
		ID:          utils.GenerateULID(nil),
		Success:     false,
		ProcessedAt: time.Now(),
	}
}

func (r *ConfigUpdateResponse) ToMap() (map[string]any, error) {
	m := map[string]any{
		"id":           r.ID,
		"success":      r.Success,
		"processed_at": r.ProcessedAt.Format(time.RFC3339),
	}

	if r.Error != "" {
		m["error"] = r.Error
	}
	if r.Payload != nil {
		payloadStr, err := json.Marshal(r.Payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload to JSON: %w", err)
		}
		m["payload"] = string(payloadStr)
	}

	return m, nil
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
