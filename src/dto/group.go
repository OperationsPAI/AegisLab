package dto

import (
	"aegis/consts"
	"fmt"
	"strings"
)

// ===================== Group Stream DTO =====================

// GroupStreamEvent represents a lightweight event pushed to group-level Redis stream
// when a trace reaches a terminal state (Completed/Failed).
type GroupStreamEvent struct {
	TraceID   string            `json:"trace_id"`
	State     consts.TraceState `json:"state"`
	LastEvent consts.EventType  `json:"last_event"`
}

// ToRedisStream converts GroupStreamEvent to Redis stream field-value pairs
func (e *GroupStreamEvent) ToRedisStream() map[string]any {
	return map[string]any{
		consts.RdbEventTraceID:        e.TraceID,
		consts.RdbEventTraceState:     e.State,
		consts.RdbEventTraceLastEvent: e.LastEvent,
	}
}

// GetGroupStreamReq represents the request to subscribe to a group stream
type GetGroupStreamReq struct {
	LastID string `form:"last_id" binding:"omitempty"`
}

func (req *GetGroupStreamReq) Validate() error {
	if req.LastID == "" {
		req.LastID = "0"
	}

	if req.LastID == "0" {
		return nil
	}

	if strings.Count(req.LastID, "-") != 1 {
		return fmt.Errorf("invalid last_id format: must be '0' or a valid stream ID (e.g., 1678886400000-0)")
	}

	return nil
}
