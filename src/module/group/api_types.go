package groupmodule

import (
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"aegis/utils"
)

// GroupStreamEvent represents a lightweight event pushed to group-level Redis stream.
type GroupStreamEvent struct {
	TraceID   string            `json:"trace_id"`
	State     consts.TraceState `json:"state"`
	LastEvent consts.EventType  `json:"last_event"`
}

func (e *GroupStreamEvent) ToRedisStream() map[string]any {
	return map[string]any{
		consts.RdbEventTraceID:        e.TraceID,
		consts.RdbEventTraceState:     e.State,
		consts.RdbEventTraceLastEvent: e.LastEvent,
	}
}

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

type GetGroupStatsReq struct {
	GroupID string `form:"group_id" binding:"required"`
}

func (req *GetGroupStatsReq) Validate() error {
	if !utils.IsValidUUID(req.GroupID) {
		return fmt.Errorf("invalid group_id: must be a valid UUID")
	}
	return nil
}

type TraceStatsItem struct {
	TraceID           string             `json:"trace_id"`
	Type              string             `json:"type"`
	State             string             `json:"state"`
	StartTime         time.Time          `json:"start_time"`
	EndTime           *time.Time         `json:"end_time,omitempty"`
	CurrentEvent      string             `json:"current_event"`
	CurrentTask       string             `json:"current_task"`
	TaskTypeDurations map[string]float64 `json:"task_type_durations,omitempty" swaggertype:"object"`
}

func NewTraceStats(trace *model.Trace) *TraceStatsItem {
	detail := &TraceStatsItem{
		TraceID:      trace.ID,
		Type:         consts.GetTraceTypeName(trace.Type),
		State:        consts.GetTraceStateName(trace.State),
		StartTime:    trace.StartTime,
		EndTime:      trace.EndTime,
		CurrentEvent: trace.LastEvent.String(),
	}

	if len(trace.Tasks) > 0 {
		detail.CurrentTask = trace.Tasks[0].ID

		taskTypeMap := make(map[string][]model.Task)
		for _, task := range trace.Tasks {
			if task.State == consts.TaskCompleted || task.State == consts.TaskError {
				taskTypeName := consts.GetTaskTypeName(task.Type)
				taskTypeMap[taskTypeName] = append(taskTypeMap[taskTypeName], task)
			}
		}

		detail.TaskTypeDurations = make(map[string]float64)
		for taskTypeName, tasks := range taskTypeMap {
			totalDuration := 0.0
			for _, task := range tasks {
				totalDuration += task.UpdatedAt.Sub(task.CreatedAt).Seconds()
			}
			detail.TaskTypeDurations[taskTypeName] = totalDuration / float64(len(tasks))
		}
	}

	return detail
}

type GroupStats struct {
	TotalTraces   int                         `json:"total_traces"`
	AvgDuration   float64                     `json:"avg_duration"`
	MinDuration   float64                     `json:"min_duration"`
	MaxDuration   float64                     `json:"max_duration"`
	TraceStateMap map[string][]TraceStatsItem `json:"trace_state_map"`
}

func NewDefaultGroupStats() *GroupStats {
	return &GroupStats{
		TotalTraces: 0,
		AvgDuration: 0.0,
		MinDuration: 0.0,
		MaxDuration: 0.0,
	}
}

type GroupTraceListResp = dto.ListResp[TraceStatsItem]
