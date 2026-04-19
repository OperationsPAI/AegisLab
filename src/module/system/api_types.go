package system

import (
	"fmt"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	systemmetric "aegis/module/systemmetric"
	task "aegis/module/task"
)

// HealthCheckResp represents system health check response.
type HealthCheckResp struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Services  map[string]ServiceInfo `json:"services" swaggertype:"object"`
}

// ServiceInfo represents individual service health information.
type ServiceInfo struct {
	Status       string    `json:"status"`
	LastChecked  time.Time `json:"last_checked"`
	ResponseTime string    `json:"response_time"`
	Error        string    `json:"error,omitempty"`
	Details      any       `json:"details,omitempty"`
}

// SystemInfo represents system information.
type SystemInfo struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	LoadAverage string  `json:"load_average"`
}

// MonitoringQueryReq represents monitoring query request.
type MonitoringQueryReq struct {
	Query     string    `json:"query" binding:"required"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Step      string    `json:"step,omitempty"`
}

// MonitoringMetricsResp represents monitoring metrics response.
type MonitoringMetricsResp struct {
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]MetricValue `json:"metrics"`
	Labels    map[string]string      `json:"labels,omitempty"`
}

type MetricValue = systemmetric.MetricValue
type ListNamespaceLockResp = systemmetric.ListNamespaceLockResp
type QueuedTasksResp = task.QueuedTasksResp

type ListAuditLogFilters struct {
	Action     string
	IPAddress  string
	UserID     int
	ResourceID int
	State      *consts.AuditLogState
	Status     *consts.StatusType
	StartTime  *time.Time
	EndTime    *time.Time
}

type ListAuditLogReq struct {
	dto.PaginationReq

	Action     string                `form:"action" binding:"omitempty"`
	IPAddress  string                `form:"ip_address" binding:"omitempty"`
	UserID     int                   `form:"user_id" binding:"omitempty"`
	ResourceID int                   `form:"resource_id" binding:"omitempty"`
	State      *consts.AuditLogState `form:"state" binding:"omitempty"`
	Status     *consts.StatusType    `form:"status" binding:"omitempty"`
	StartDate  string                `form:"start_date" binding:"omitempty"`
	EndDate    string                `form:"end_date" binding:"omitempty"`
}

func (req *ListAuditLogReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if err := validateDateField(req.StartDate); err != nil {
		return fmt.Errorf("invalid start_time: %w", err)
	}
	if err := validateDateField(req.EndDate); err != nil {
		return fmt.Errorf("invalid end_time: %w", err)
	}
	if req.State != nil {
		if _, exists := consts.ValidAuditLogStates[*req.State]; !exists {
			return fmt.Errorf("invalid state: %d", *req.State)
		}
	}
	return validateStatusValue(req.Status, false)
}

func (req *ListAuditLogReq) ToFilterOptions() *ListAuditLogFilters {
	var startTimePtr, endTimePtr *time.Time
	if req.StartDate != "" {
		startTime, _ := time.Parse(time.DateOnly, req.StartDate)
		startTimePtr = &startTime
	}
	if req.EndDate != "" {
		endTime, _ := time.Parse(time.DateOnly, req.EndDate)
		endTimePtr = &endTime
	}

	return &ListAuditLogFilters{
		Action:     req.Action,
		IPAddress:  req.IPAddress,
		UserID:     req.UserID,
		ResourceID: req.ResourceID,
		State:      req.State,
		Status:     req.Status,
		StartTime:  startTimePtr,
		EndTime:    endTimePtr,
	}
}

type AuditLogResp struct {
	ID         int                 `json:"id"`
	Action     string              `json:"action"`
	IPAddress  string              `json:"ip_address"`
	Duration   int                 `json:"duration"`
	UserAgent  string              `json:"user_agent"`
	UserID     int                 `json:"user_id,omitempty"`
	Username   string              `json:"username,omitempty"`
	ResourceID int                 `json:"resource_id,omitempty"`
	Resource   consts.ResourceName `json:"resource,omitempty"`
	State      string              `json:"state"`
	Status     string              `json:"status"`
	CreatedAt  time.Time           `json:"created_at"`
}

func NewAuditLogResp(log *model.AuditLog) *AuditLogResp {
	resp := &AuditLogResp{
		ID:         log.ID,
		Action:     log.Action,
		IPAddress:  log.IPAddress,
		Duration:   log.Duration,
		UserAgent:  log.UserAgent,
		UserID:     log.UserID,
		ResourceID: log.ResourceID,
		State:      consts.GetAuditLogStateName(log.State),
		Status:     consts.GetStatusTypeName(log.Status),
		CreatedAt:  log.CreatedAt,
	}
	if log.User != nil {
		resp.Username = log.User.Username
	}
	if log.Resource != nil {
		resp.Resource = log.Resource.Name
	}
	return resp
}

type AuditLogDetailResp struct {
	AuditLogResp
	Details  string `json:"details"`
	ErrorMsg string `json:"error_msg,omitempty"`
}

func NewAuditLogDetailResp(log *model.AuditLog) *AuditLogDetailResp {
	return &AuditLogDetailResp{
		AuditLogResp: *NewAuditLogResp(log),
		Details:      log.Details,
		ErrorMsg:     log.ErrorMsg,
	}
}

type ListConfigReq struct {
	dto.PaginationReq
	ValueType *consts.ConfigValueType `form:"value_type" binding:"omitempty"`
	Category  *string                 `form:"category" binding:"omitempty"`
	IsSecret  *bool                   `form:"is_secret" binding:"omitempty"`
	UpdatedBy *int                    `form:"updated_by" binding:"omitempty,min_ptr=1"`
}

func (req *ListConfigReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	return validateConfigValueType(req.ValueType)
}

type RollbackConfigReq struct {
	HistoryID int    `json:"history_id" binding:"required,min=1"`
	Reason    string `json:"reason" binding:"required"`
}

type UpdateConfigValueReq struct {
	Value  string `json:"value" binding:"required"`
	Reason string `json:"reason" binding:"required"`
}

type UpdateConfigMetadataReq struct {
	DefaultValue *string  `json:"default_value" binding:"omitempty"`
	Description  *string  `json:"description" binding:"omitempty"`
	MinValue     *float64 `json:"min_value" binding:"omitempty"`
	MaxValue     *float64 `json:"max_value" binding:"omitempty"`
	Pattern      *string  `json:"pattern" binding:"omitempty"`
	Options      *string  `json:"options" binding:"omitempty"`
	Reason       string   `json:"reason" binding:"required"`
}

func (req *UpdateConfigMetadataReq) Validate() error {
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

func (req *UpdateConfigMetadataReq) PatchConfigModel(target *model.DynamicConfig) (string, string) {
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
	dto.PaginationReq
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

type ConfigResp struct {
	ID            int       `json:"id"`
	Key           string    `json:"key"`
	ValueType     string    `json:"value_type"`
	Category      string    `json:"category"`
	UpdatedAt     time.Time `json:"updated_at"`
	UpdatedByID   int       `json:"updated_by_id"`
	UpdatedByName string    `json:"updated_by_name"`
}

func NewConfigResp(config *model.DynamicConfig) *ConfigResp {
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

func NewConfigDetailResp(config *model.DynamicConfig) *ConfigDetailResp {
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

func NewConfigHistoryResp(history *model.ConfigHistory) *ConfigHistoryResp {
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

func validateDateField(value string) error {
	if value == "" {
		return nil
	}
	if _, err := time.Parse(time.DateOnly, value); err != nil {
		return fmt.Errorf("invalid time format: %s", value)
	}
	return nil
}

func validateStatusValue(statusPtr *consts.StatusType, isMutation bool) error {
	if statusPtr == nil {
		return nil
	}
	status := *statusPtr
	if _, exists := consts.ValidStatuses[status]; !exists {
		return fmt.Errorf("invalid status value: %d", status)
	}
	if isMutation && status == consts.CommonDeleted {
		return fmt.Errorf("status value cannot be set to deleted (%d) directly through this update/create operation", consts.CommonDeleted)
	}
	return nil
}

func validateConfigValueType(valueType *consts.ConfigValueType) error {
	if valueType != nil {
		if _, ok := consts.ValidDynamicConfigTypes[*valueType]; !ok {
			return fmt.Errorf("invalid value type: %v", valueType)
		}
	}
	return nil
}
