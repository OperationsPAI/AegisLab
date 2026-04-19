package dto

import (
	"aegis/consts"
	"encoding/json"
)

type TraceStreamEvent struct {
	TimeStamp int              `json:"timestamp"`
	TaskID    string           `json:"task_id"`
	TaskType  consts.TaskType  `json:"task_type"`
	FileName  string           `json:"file_name" swaggerignore:"true"`
	FnName    string           `json:"function_name" swaggerignore:"true"`
	Line      int              `json:"line" swaggerignore:"true"`
	EventName consts.EventType `json:"event_name"`
	Payload   any              `json:"payload,omitempty" swaggertype:"object"`
}

func (s *TraceStreamEvent) ToRedisStream() map[string]any {
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

func (s *TraceStreamEvent) ToSSE() (string, error) {
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

// TaskScheduledPayload is the payload for the task.scheduled trace event
// emitted whenever a task is enqueued into the delayed queue.
type TaskScheduledPayload struct {
	ExecuteTime int64  `json:"execute_time"`
	Reason      string `json:"reason"`
}

// Reasons for task.scheduled events.
const (
	TaskScheduledReasonPreDurationWait  = "pre_duration wait"
	TaskScheduledReasonTokenUnavailable = "token unavailable"
	TaskScheduledReasonRetryBackoff     = "retry_backoff"
	TaskScheduledReasonManual           = "manual"
	TaskScheduledReasonCronNext         = "cron_next"
	TaskScheduledReasonExpedite         = "expedite"
)

type JobMessage struct {
	JobName   string `json:"job_name"`
	Namespace string `json:"namespace"`
	LogFile   string `json:"log_file,omitempty"`
}
