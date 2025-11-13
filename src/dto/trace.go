package dto

import (
	"aegis/consts"
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

type DatapackResult struct {
	Datapack  *InjectionItem `json:"datapack"`
	Timestamp string         `json:"timestamp"`
}

type ExecutionResult struct {
	Algorithm   *ContainerVersionItem `json:"algorithm"`
	Datapack    *InjectionItem        `json:"datapack"`
	ExecutionID int                   `json:"execution_id"`
	Timestamp   string                `json:"timestamp"`
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
	LastID string `form:"last_id" binding:"required"`
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
