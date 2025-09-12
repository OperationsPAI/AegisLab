package dto

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"rcabench/consts"
	"rcabench/database"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// -----------------------------------------------------------------------------
// Data Structures
// -----------------------------------------------------------------------------
// RetryPolicy defines how tasks should be retried on failure
type RetryPolicy struct {
	MaxAttempts int `json:"max_attempts"` // Maximum number of retry attempts
	BackoffSec  int `json:"backoff_sec"`  // Seconds to wait between retries
}

// UnifiedTask represents a task that can be scheduled and executed
type UnifiedTask struct {
	TaskID       string                 `json:"task_id"`                      // Unique identifier for the task
	Type         consts.TaskType        `json:"type"`                         // Task type (determines how it's processed)
	Immediate    bool                   `json:"immediate"`                    // Whether to execute immediately
	ExecuteTime  int64                  `json:"execute_time"`                 // Unix timestamp for delayed execution
	CronExpr     string                 `json:"cron_expr,omitempty"`          // Cron expression for recurring tasks
	ReStartNum   int                    `json:"restart_num"`                  // Number of restarts for the task
	RetryPolicy  RetryPolicy            `json:"retry_policy"`                 // Policy for retrying failed tasks
	Payload      map[string]any         `json:"payload" swaggertype:"object"` // Task-specific data
	Status       string                 `json:"status"`                       // Status of the task
	TraceID      string                 `json:"trace_id"`                     // ID for tracing related tasks
	GroupID      string                 `json:"group_id"`                     // ID for grouping tasks
	ProjectID    *int                   `json:"project_id,omitempty"`         // ID for the project (optional)
	TraceCarrier propagation.MapCarrier `json:"trace_carrier,omitempty"`      // Carrier for trace context
	GroupCarrier propagation.MapCarrier `json:"group_carrier,omitempty"`      // Carrier for group context
}

// -----------------------------------------------------------------------------
// Context Management Methods
// -----------------------------------------------------------------------------

// GetTraceCtx extracts the trace context from the carrier
func (t *UnifiedTask) GetTraceCtx() context.Context {
	if t.TraceCarrier == nil {
		logrus.WithField("task_id", t.TaskID).WithField("task_type", t.Type).Error("No group context, create a new one")
		return context.Background()
	}

	traceCtx := otel.GetTextMapPropagator().Extract(context.Background(), t.TraceCarrier)
	return traceCtx
}

// GetGroupCtx extracts the group context from the carrier
func (t *UnifiedTask) GetGroupCtx() context.Context {
	if t.GroupCarrier == nil {
		logrus.WithField("task_id", t.TaskID).WithField("task_type", t.Type).Error("No group context, create a new one")
		return context.Background()
	}

	traceCtx := otel.GetTextMapPropagator().Extract(context.Background(), t.GroupCarrier)
	return traceCtx
}

// SetTraceCtx injects the trace context into the carrier
func (t *UnifiedTask) SetTraceCtx(ctx context.Context) {
	if t.TraceCarrier == nil {
		t.TraceCarrier = make(propagation.MapCarrier)
	}

	otel.GetTextMapPropagator().Inject(ctx, t.TraceCarrier)
}

// SetGroupCtx injects the group context into the carrier
func (t *UnifiedTask) SetGroupCtx(ctx context.Context) {
	if t.GroupCarrier == nil {
		t.GroupCarrier = make(propagation.MapCarrier)
	}

	otel.GetTextMapPropagator().Inject(ctx, t.GroupCarrier)
}

type TaskDetailResp struct {
	Task TaskItem `json:"task"`
	Logs []string `json:"logs"`
}

type TaskItem struct {
	ID        string         `json:"id"`
	TraceID   string         `json:"trace_id"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload" swaggertype:"object"`
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
}

func (t *TaskItem) Convert(task database.Task) error {
	var payload map[string]any
	if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
		return err
	}

	t.ID = task.ID
	t.TraceID = task.TraceID
	t.Type = task.Type
	t.Payload = payload
	t.Status = task.Status
	t.CreatedAt = task.CreatedAt

	return nil
}

type TaskReq struct {
	TaskID string `uri:"task_id" binding:"required"`
}

type TaskStreamItem struct {
	Type    string
	TraceID string
}

func (t *TaskStreamItem) Convert(task database.Task) {
	t.Type = task.Type
	t.TraceID = task.TraceID
}

type DatasetOptions struct {
	Dataset string `json:"dataset"`
}

type ExecutionOptions struct {
	Algorithm   AlgorithmItem `json:"algorithm"`
	Dataset     string        `json:"dataset"`
	ExecutionID int           `json:"execution_id"`
	Timestamp   string        `json:"timestamp"`
}

type JobMessage struct {
	JobName   string            `json:"job_name"`
	Namespace string            `json:"namespace"`
	Status    string            `json:"status"`
	Logs      map[string]string `json:"logs"`
}

type StreamEvent struct {
	TimeStamp int              `json:"timestamp,omitempty"`
	TaskID    string           `json:"task_id"`
	TaskType  consts.TaskType  `json:"task_type"`
	FileName  string           `json:"file_name"`
	FnName    string           `json:"function_name"`
	Line      int              `json:"line"`
	EventName consts.EventType `json:"event_name"`
	Payload   any              `json:"payload"`
}

type InfoPayloadTemplate struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

func (s *StreamEvent) ToRedisStream() map[string]any {
	payload, err := json.Marshal(s.Payload)
	if err != nil {
		logrus.Errorf("Failed to marshal payload: %v", err)
		return nil
	}

	return map[string]any{
		consts.RdbEventTaskID:   s.TaskID,
		consts.RdbEventTaskType: string(s.TaskType),
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

type ListTasksReq struct {
	TraceID string `form:"trace_id"`
	GroupID string `form:"group_id"`

	TaskType  string `form:"task_type"`
	Status    string `form:"status"`
	Immediate *bool  `form:"immediate"`

	ListOptionsQuery
	TimeRangeQuery
}

func (req *ListTasksReq) Validate() error {
	idFieldsUsed := 0
	if req.TraceID != "" {
		idFieldsUsed++
	}
	if req.GroupID != "" {
		idFieldsUsed++
	}

	if idFieldsUsed > 1 {
		return fmt.Errorf("only one of task_id, trace_id, or group_id can be specified")
	}

	if err := req.ListOptionsQuery.Validate(); err != nil {
		return err
	}

	if err := req.TimeRangeQuery.Validate(); err != nil {
		return err
	}

	return nil
}

type ListTasksResp []database.Task
