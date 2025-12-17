package dto

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"aegis/consts"
	"aegis/database"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

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
	Level        int                    `json:"level"`                        // Task level in the trace
	Sequence     int                    `json:"sequence"`                     // Task sequence in the trace
	ParentTaskID *string                `json:"parent_task_id,omitempty"`     // Parent task ID for sub-tasks
	TraceID      string                 `json:"trace_id"`                     // ID for tracing related tasks
	GroupID      string                 `json:"group_id"`                     // ID for grouping tasks
	ProjectID    int                    `json:"project_id"`                   // ID for the project (optional)
	UserID       int                    `json:"user_id"`                      // ID of the user who created the task (optional)
	State        consts.TaskState       `json:"state"`                        // Current state of the task
	TraceCarrier propagation.MapCarrier `json:"trace_carrier,omitempty"`      // Carrier for trace context
	GroupCarrier propagation.MapCarrier `json:"group_carrier,omitempty"`      // Carrier for group context
}

func (t *UnifiedTask) ConvertToTask() (*database.Task, error) {
	jsonPayload, err := json.Marshal(t.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := &database.Task{
		ID:           t.TaskID,
		Type:         t.Type,
		Immediate:    t.Immediate,
		ExecuteTime:  t.ExecuteTime,
		CronExpr:     t.CronExpr,
		Payload:      string(jsonPayload),
		Level:        t.Level,
		State:        t.State,
		Status:       consts.CommonEnabled,
		ParentTaskID: t.ParentTaskID,
		TraceID:      t.TraceID,
		GroupID:      t.GroupID,
		ProjectID:    t.ProjectID,
	}
	return task, nil
}

func (t *UnifiedTask) ConvertToTrace(withAlgorithms bool, leafNum int) (*database.Trace, error) {
	var traceType consts.TraceType
	switch t.Type {
	case consts.TaskTypeRestartPedestal:
		if withAlgorithms {
			traceType = consts.TraceTypeFullPipeline
		} else {
			traceType = consts.TraceTypeDatapackBuild
		}
	case consts.TaskTypeBuildDatapack:
		traceType = consts.TraceTypeDatapackBuild
	case consts.TaskTypeRunAlgorithm:
		traceType = consts.TraceTypeAlgorithmRun
	default:
		return nil, fmt.Errorf("unsupported task type for trace conversion: %s", consts.GetTaskTypeName(t.Type))
	}

	trace := &database.Trace{
		ID:        t.TraceID,
		Type:      traceType,
		StartTime: time.Now(),
		LeafNum:   leafNum,
		GroupID:   t.GroupID,
		ProjectID: t.ProjectID,
		State:     consts.TracePending,
		Status:    consts.CommonEnabled,
	}

	return trace, nil
}

// GetAnnotations generates the annotations for trace and group carriers
func (t *UnifiedTask) GetAnnotations(ctx context.Context) (map[string]string, error) {
	taskCarrier := make(propagation.MapCarrier)
	otel.GetTextMapPropagator().Inject(ctx, taskCarrier)

	taskCarrierBytes, err := json.Marshal(taskCarrier)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mapcarrier of task context: %w", err)
	}

	traceCarrierBytes, err := json.Marshal(t.TraceCarrier)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mapcarrier of trace context: %w", err)
	}

	return map[string]string{
		consts.TaskCarrier:  string(taskCarrierBytes),
		consts.TraceCarrier: string(traceCarrierBytes),
	}, nil
}

// GetLabels generates the labels for the task
func (t *UnifiedTask) GetLabels() map[string]string {
	return map[string]string{
		consts.JobLabelTaskID:    t.TaskID,
		consts.JobLabelTaskType:  consts.GetTaskTypeName(t.Type),
		consts.JobLabelTraceID:   t.TraceID,
		consts.JobLabelGroupID:   t.GroupID,
		consts.JobLabelProjectID: strconv.Itoa(t.ProjectID),
		consts.JobLabelUserID:    strconv.Itoa(t.UserID),
	}
}

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

func (t *UnifiedTask) Reschedule(executeTime time.Time) {
	t.ExecuteTime = executeTime.Unix()
	t.ReStartNum += 1
	t.State = consts.TaskRescheduled
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

// BatchDeleteTaskReq represents the request to batch delete tasks
type BatchDeleteTaskReq struct {
	IDs []string `json:"ids" binding:"required"` // List of task IDs for deletion
}

func (req *BatchDeleteTaskReq) Validate() error {
	for i, id := range req.IDs {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("empty id at index %d", i)
		}

		if !utils.IsValidUUID(id) {
			return fmt.Errorf("invalid UUID format for id at index %d: %s", i, id)
		}
	}
	return nil
}

// ListTaskFilters represents the filters for listing tasks
type ListTaskFilters struct {
	TaskType  *consts.TaskType
	Immediate *bool
	TraceID   string
	GroupID   string
	ProjectID int
	State     *consts.TaskState
	Status    *consts.StatusType
}

// ListTaskReq represents the request to list tasks
type ListTaskReq struct {
	PaginationReq
	TaskType  *consts.TaskType   `form:"task_type" binding:"omitempty"`
	Immediate *bool              `form:"immediate" binding:"omitempty"`
	TraceID   string             `form:"trace_id" binding:"omitempty"`
	GroupID   string             `form:"group_id" binding:"omitempty"`
	ProjectID int                `form:"project_id" binding:"omitempty"`
	State     *consts.TaskState  `form:"state" binding:"omitempty"`
	Status    *consts.StatusType `form:"status" binding:"omitempty"`
}

func (req *ListTaskReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if err := validateTaskType(req.TaskType); err != nil {
		return err
	}
	if err := validateUUID(req.TraceID); err != nil {
		return err
	}
	if err := validateUUID(req.GroupID); err != nil {
		return err
	}

	if req.ProjectID <= 0 {
		return fmt.Errorf("invalid project ID: %d", req.ProjectID)
	}

	if err := validateState(req.State); err != nil {
		return err
	}
	return validateStatusField(req.Status, true)
}

func (req *ListTaskReq) ToFilterOptions() *ListTaskFilters {
	return &ListTaskFilters{
		Immediate: req.Immediate,
		TaskType:  req.TaskType,
		TraceID:   req.TraceID,
		GroupID:   req.GroupID,
		ProjectID: req.ProjectID,
		State:     req.State,
		Status:    req.Status,
	}
}

// TaskResp represents the response for a task
type TaskResp struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Immediate   bool   `json:"immediate"`
	ExecuteTime int64  `json:"execute_time"`
	CronExpr    string `json:"cron_expr,omitempty"`
	TraceID     string `json:"trace_id"`
	GroupID     string `json:"group_id"`

	State       string    `json:"state"`
	Status      string    `json:"status"`
	ProjectID   int       `json:"project_id,omitempty"`
	ProjectName string    `json:"project_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewTaskResp(task *database.Task) *TaskResp {
	resp := &TaskResp{
		ID:          task.ID,
		Type:        consts.GetTaskTypeName(task.Type),
		Immediate:   task.Immediate,
		ExecuteTime: task.ExecuteTime,
		CronExpr:    task.CronExpr,
		TraceID:     task.TraceID,
		GroupID:     task.GroupID,
		State:       consts.GetTaskStateName(task.State),
		Status:      consts.GetStatusTypeName(task.Status),
		ProjectID:   task.Project.ID,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}

	if task.Project != nil {
		resp.ProjectName = task.Project.Name
	}
	return resp
}

type TaskDetailResp struct {
	TaskResp

	Payload map[string]any `json:"payload,omitempty" swaggertype:"object"`
	Logs    []string       `json:"logs"`
}

func NewTaskDetailResp(task *database.Task, logs []string) *TaskDetailResp {
	resp := &TaskDetailResp{
		TaskResp: *NewTaskResp(task),
		Logs:     logs,
	}

	if task.Payload != "" {
		var payload map[string]any
		if err := json.Unmarshal([]byte(task.Payload), &payload); err == nil {
			resp.Payload = payload
		}
	}
	return resp
}

// QueuedTasksResp represents the response for queued tasks
type QueuedTasksResp struct {
	ReadyTasks   []TaskResp `json:"ready_tasks"`
	DelayedTasks []TaskResp `json:"delayed_tasks"`
}

func validateState(state *consts.TaskState) error {
	if state != nil {
		if _, exists := consts.ValidTaskStates[*state]; !exists {
			return fmt.Errorf("invalid task state: %d", *state)
		}
	}
	return nil
}

func validateTaskType(taskType *consts.TaskType) error {
	if taskType != nil {
		if _, exists := consts.ValidTaskTypes[*taskType]; !exists {
			return fmt.Errorf("invalid task type: %d", *taskType)
		}
	}
	return nil
}

func validateUUID(id string) error {
	if !utils.IsValidUUID(id) {
		return fmt.Errorf("invalid UUID format: %s", id)
	}
	return nil
}
