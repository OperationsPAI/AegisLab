package dto

import (
	"aegis/consts"
	"aegis/database"
	"aegis/utils"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type StreamEvent struct {
	TimeStamp int              `json:"timestamp,omitempty" swaggerignore:"true"`
	TaskID    string           `json:"task_id"`
	TaskType  consts.TaskType  `json:"task_type"`
	FileName  string           `json:"file_name" swaggerignore:"true"`
	FnName    string           `json:"function_name" swaggerignore:"true"`
	Line      int              `json:"line" swaggerignore:"true"`
	EventName consts.EventType `json:"event_name"`
	Payload   any              `json:"payload,omitempty" swaggertype:"object"`
}

func (s *StreamEvent) ToRedisStream() map[string]any {
	payload, err := json.Marshal(s.Payload)
	if err != nil {
		return nil
	}

	return map[string]any{
		consts.RdbEventTaskID:   s.TaskID,
		consts.RdbEventTaskType: consts.GetTaskTypeName(s.TaskType),
		consts.RdbEventFileName: s.FileName,
		consts.RdbEventFn:       s.FnName,
		consts.RdbEventLine:     s.Line,
		consts.RdbEventName:     string(s.EventName),
		consts.RdbEventPayload:  payload,
	}
}

func (s *StreamEvent) ToSSE() (string, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

type DatapackInfo struct {
	Datapack *InjectionItem `json:"datapack"`
	JobName  string         `json:"job_name"`
}

type DatapackResult struct {
	Datapack string `json:"datapack"`
	JobName  string `json:"job_name"`
}

type ExecutionInfo struct {
	Algorithm   *ContainerVersionItem `json:"algorithm"`
	Datapack    *InjectionItem        `json:"datapack"`
	ExecutionID int                   `json:"execution_id"`
	JobName     string                `json:"job_name"`
}

type ExecutionResult struct {
	Algorithm string `json:"algorithm"`
	JobName   string `json:"job_name"`
}

type InfoPayloadTemplate struct {
	State string `json:"task_state"`
	Msg   string `json:"msg"`
}

type JobMessage struct {
	JobName   string `json:"job_name"`
	Namespace string `json:"namespace"`
	LogFile   string `json:"log_file,omitempty"`
}

type TraceQuery struct {
	TraceID       string          `json:"trace_id"`
	FirstTaskType consts.TaskType `json:"first_task_type"`
	StartTime     time.Time       `json:"start_time"`
	EndTime       time.Time       `json:"end_time"`
}

type GetTraceStreamReq struct {
	LastID string `form:"last_id" binding:"omitempty"`
}

func (req *GetTraceStreamReq) Validate() error {
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

// GetGroupStatsReq represents the request to get group stats
type GetGroupStatsReq struct {
	GroupID string `form:"group_id" binding:"required"` // Group ID to query
}

func (req *GetGroupStatsReq) Validate() error {
	if !utils.IsValidUUID(req.GroupID) {
		return fmt.Errorf("invalid group_id: must be a valid UUID")
	}
	return nil
}

// TraceStatsItem represents the stat of a trace
type TraceStatsItem struct {
	TraceID           string             `json:"trace_id"`
	Type              string             `json:"type"`
	State             string             `json:"state"`
	StartTime         time.Time          `json:"start_time"`
	EndTime           *time.Time         `json:"end_time,omitempty"`
	CurrentEvent      string             `json:"current_event"`
	CurrentTask       string             `json:"current_task"`
	TaskTypeDurations map[string]float64 `json:"task_type_durations,omitempty" swaggertype:"object"` // Average durations per task type in seconds
}

func NewTraceStats(trace *database.Trace) *TraceStatsItem {
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

		taskTypeMap := make(map[string][]database.Task)
		for _, task := range trace.Tasks {
			if task.State == consts.TaskCompleted || task.State == consts.TaskError {
				taskTypeName := consts.GetTaskTypeName(task.Type)
				if _, exists := taskTypeMap[taskTypeName]; !exists {
					taskTypeMap[taskTypeName] = []database.Task{}
				}
				taskTypeMap[taskTypeName] = append(taskTypeMap[taskTypeName], task)
			}
		}

		detail.TaskTypeDurations = make(map[string]float64)
		for taskTypeName, tasks := range taskTypeMap {
			totalDuration := 0.0
			for _, task := range tasks {
				duration := task.UpdatedAt.Sub(task.CreatedAt).Seconds()
				totalDuration += duration
			}
			detail.TaskTypeDurations[taskTypeName] = totalDuration / float64(len(tasks))
		}
	}

	return detail
}

// GroupStats represents the response for group stats
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
