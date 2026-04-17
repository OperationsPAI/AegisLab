package dto

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"aegis/consts"
	"aegis/model"

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
	TaskID       string                   `json:"task_id"`                      // Unique identifier for the task
	Type         consts.TaskType          `json:"type"`                         // Task type (determines how it's processed)
	Immediate    bool                     `json:"immediate"`                    // Whether to execute immediately
	ExecuteTime  int64                    `json:"execute_time"`                 // Unix timestamp for delayed execution
	CronExpr     string                   `json:"cron_expr,omitempty"`          // Cron expression for recurring tasks
	ReStartNum   int                      `json:"restart_num"`                  // Number of restarts for the task
	RetryPolicy  RetryPolicy              `json:"retry_policy"`                 // Policy for retrying failed tasks
	Payload      map[string]any           `json:"payload" swaggertype:"object"` // Task-specific data
	Level        int                      `json:"level"`                        // Task level in the trace
	Sequence     int                      `json:"sequence"`                     // Task sequence in the trace
	ParentTaskID *string                  `json:"parent_task_id,omitempty"`     // Parent task ID for sub-tasks
	TraceID      string                   `json:"trace_id"`                     // ID for tracing related tasks
	GroupID      string                   `json:"group_id"`                     // ID for grouping tasks
	ProjectID    int                      `json:"project_id"`                   // ID for the project (optional)
	UserID       int                      `json:"user_id"`                      // ID of the user who created the task (optional)
	State        consts.TaskState         `json:"state"`                        // Current state of the task
	TraceCarrier propagation.MapCarrier   `json:"trace_carrier,omitempty"`      // Carrier for trace context
	GroupCarrier propagation.MapCarrier   `json:"group_carrier,omitempty"`      // Carrier for group context
	Extra        map[consts.TaskExtra]any `json:"extra,omitempty"`              // Additional metadata
}

func (t *UnifiedTask) ConvertToTask() (*model.Task, error) {
	jsonPayload, err := json.Marshal(t.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := &model.Task{
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
	}
	return task, nil
}

func (t *UnifiedTask) ConvertToTrace(withAlgorithms bool, leafNum int) (*model.Trace, error) {
	var traceType consts.TraceType
	switch t.Type {
	case consts.TaskTypeRestartPedestal:
		if withAlgorithms {
			traceType = consts.TraceTypeFullPipeline
		} else {
			traceType = consts.TraceTypeFaultInjection
		}
	case consts.TaskTypeBuildDatapack:
		traceType = consts.TraceTypeDatapackBuild
	case consts.TaskTypeRunAlgorithm:
		traceType = consts.TraceTypeAlgorithmRun
	default:
		return nil, fmt.Errorf("unsupported task type for trace conversion: %s", consts.GetTaskTypeName(t.Type))
	}

	trace := &model.Trace{
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
		logrus.WithFields(
			logrus.Fields{
				"task_id":   t.TaskID,
				"task_type": consts.GetTaskTypeName(t.Type),
			},
		).Warn("No trace context, create a new one")
		return context.Background()
	}

	traceCtx := otel.GetTextMapPropagator().Extract(context.Background(), t.TraceCarrier)
	return traceCtx
}

// GetGroupCtx extracts the group context from the carrier
func (t *UnifiedTask) GetGroupCtx() context.Context {
	if t.GroupCarrier == nil {
		logrus.WithFields(
			logrus.Fields{
				"task_id":   t.TaskID,
				"task_type": consts.GetTaskTypeName(t.Type),
			},
		).Warn("No group context, create a new one")
		return context.Background()
	}

	groupCtx := otel.GetTextMapPropagator().Extract(context.Background(), t.GroupCarrier)
	return groupCtx
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
