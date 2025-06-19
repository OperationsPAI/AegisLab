package dto

import (
	"context"
	"encoding/json"
	"time"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
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
	TraceID      string                 `json:"trace_id,omitempty"`           // ID for tracing related tasks
	GroupID      string                 `json:"group_id,omitempty"`           // ID for grouping tasks
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

type StreamEvent struct {
	TimeStamp int              `json:"timestamp,omitempty"`
	TaskID    string           `json:"task_id"`
	TaskType  consts.TaskType  `json:"task_type"`
	FileName  string           `json:"file_name"`
	FnName    string           `json:"function_name"`
	Line      int              `json:"line"`
	EventName consts.EventType `json:"event_name"`
	// TODO 结构化部分payload
	Payload any `json:"payload"`
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

// TaskListReq defines the query parameters for listing tasks
type TaskListReq struct {
	// Filter parameters
	TaskID         string `form:"task_id"`
	TaskType       string `form:"task_type"`
	Status         string `form:"status"`
	TraceID        string `form:"trace_id"`
	GroupID        string `form:"group_id"`
	Immediate      *bool  `form:"immediate"`
	ExecuteTimeGT  *int64 `form:"execute_time_gt"`
	ExecuteTimeLT  *int64 `form:"execute_time_lt"`
	ExecuteTimeGTE *int64 `form:"execute_time_gte"`
	ExecuteTimeLTE *int64 `form:"execute_time_lte"`

	// Pagination parameters
	PaginationReq
	SortField string `form:"sort_field"` // Format: "field_name asc/desc"
}

type TaskDatabaseFilter struct {
	TaskID         *string
	TaskType       *string
	Immediate      *bool
	ExecuteTimeGT  *int64
	ExecuteTimeLT  *int64
	ExecuteTimeGTE *int64
	ExecuteTimeLTE *int64
	Status         *string
	TraceID        *string
	GroupID        *string
}

func (r *TaskListReq) Convert() TaskDatabaseFilter {
	filter := TaskDatabaseFilter{}

	if r.TaskID != "" {
		filter.TaskID = &r.TaskID
	}
	if r.TaskType != "" {
		filter.TaskType = &r.TaskType
	}
	if r.Status != "" {
		filter.Status = &r.Status
	}
	if r.TraceID != "" {
		filter.TraceID = &r.TraceID
	}
	if r.GroupID != "" {
		filter.GroupID = &r.GroupID
	}
	if r.Immediate != nil {
		filter.Immediate = r.Immediate
	}
	if r.ExecuteTimeGT != nil {
		filter.ExecuteTimeGT = r.ExecuteTimeGT
	}
	if r.ExecuteTimeLT != nil {
		filter.ExecuteTimeLT = r.ExecuteTimeLT
	}
	if r.ExecuteTimeGTE != nil {
		filter.ExecuteTimeGTE = r.ExecuteTimeGTE
	}
	if r.ExecuteTimeLTE != nil {
		filter.ExecuteTimeLTE = r.ExecuteTimeLTE
	}

	return filter
}
