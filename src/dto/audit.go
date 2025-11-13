package dto

import (
	"aegis/consts"
	"aegis/database"
	"fmt"
	"time"
)

type ListAuditLogFilters struct {
	Action     string
	IpAddress  string
	UserID     int
	ResourceID int
	State      *consts.AuditLogState
	Status     *consts.StatusType
	StartTime  *time.Time
	EndTime    *time.Time
}

type ListAuditLogReq struct {
	PaginationReq

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
	if req.StartDate != "" {
		if err := validateTimeField(req.StartDate, time.DateOnly); err != nil {
			return fmt.Errorf("invalid start_time: %w", err)
		}
	}
	if req.EndDate != "" {
		if err := validateTimeField(req.EndDate, time.DateOnly); err != nil {
			return fmt.Errorf("invalid end_time: %w", err)
		}
	}

	if _, exists := consts.ValidAuditLogStates[*req.State]; !exists {
		return fmt.Errorf("invalid state: %d", *req.State)
	}

	return validateStatusField(req.Status, false)
}

func (req *ListAuditLogReq) ToFilterOptions() *ListAuditLogFilters {
	var startTimePtr, endTimePtr *time.Time

	if req.StartDate != "" {
		startTime, _ := time.Parse(time.DateTime, req.StartDate)
		startTimePtr = &startTime
	}

	if req.EndDate != "" {
		endTime, _ := time.Parse(time.DateTime, req.EndDate)
		endTimePtr = &endTime
	}

	return &ListAuditLogFilters{
		Action:     req.Action,
		IpAddress:  req.IPAddress,
		UserID:     req.UserID,
		ResourceID: req.ResourceID,
		State:      req.State,
		Status:     req.Status,
		StartTime:  startTimePtr,
		EndTime:    endTimePtr,
	}
}

// AuditLogResp represents a summarized view of an audit log
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

func NewAuditLogResp(log *database.AuditLog) *AuditLogResp {
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

// AuditLogDetailResp extends AuditLogResp with Details and ErrorMsg
type AuditLogDetailResp struct {
	AuditLogResp
	Details  string `json:"details"`
	ErrorMsg string `json:"error_msg,omitempty"`
}

func NewAuditLogDetailResp(log *database.AuditLog) *AuditLogDetailResp {
	return &AuditLogDetailResp{
		AuditLogResp: *NewAuditLogResp(log),
		Details:      log.Details,
		ErrorMsg:     log.ErrorMsg,
	}
}
