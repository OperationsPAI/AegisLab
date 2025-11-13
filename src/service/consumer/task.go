package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"aegis/client"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"aegis/tracing"
	"aegis/utils"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type eventPublishOptions struct {
	callerLevel int
}

type eventPublishOption func(*eventPublishOptions)

func withCallerLevel(level int) eventPublishOption {
	return func(opts *eventPublishOptions) {
		opts.callerLevel = level
	}
}

// -----------------------------------------------------------------------------
// Constants and Global Variables
// -----------------------------------------------------------------------------

// Redis key constants for task queues and indexes
const (
	DelayedQueueKey    = "task:delayed"          // Sorted set for delayed tasks
	ReadyQueueKey      = "task:ready"            // List for ready-to-execute tasks
	DeadLetterKey      = "task:dead"             // Sorted set for failed tasks
	TaskIndexKey       = "task:index"            // Hash mapping task IDs to their queue
	ConcurrencyLockKey = "task:concurrency_lock" // Counter for concurrency control
	LastBatchInfoKey   = "last_batch_info"       // Key for batch processing information
	MaxConcurrency     = 20                      // Maximum concurrent tasks
)

// Prometheus metrics for task monitoring
var (
	// Counter for tracking processed tasks by type and status
	tasksProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "task_processed_total",
		Help: "Total number of processed tasks",
	}, []string{"type", "status"})

	// Histogram for measuring task duration by type
	taskDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "task_duration_seconds",
		Help:    "Task processing duration distribution",
		Buckets: []float64{0.1, 0.5, 1, 5, 10, 30},
	}, []string{"type"})
)

// Task cancellation registry
var (
	taskCancelFuncs      = make(map[string]context.CancelFunc) // Maps task IDs to their cancel functions
	taskCancelFuncsMutex sync.RWMutex                          // Mutex to protect the map
)

// StartScheduler starts the scheduler that moves tasks from delayed to ready queue
func StartScheduler(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			processDelayedTasks(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// processDelayedTasks moves tasks from delayed queue to ready queue when their time arrives
func processDelayedTasks(ctx context.Context) {
	result, err := repository.ProcessDelayedTasks(ctx)

	if err != nil && err != redis.Nil {
		logrus.Errorf("scheduler error: %v", err)
		return
	}

	for _, taskData := range result {
		var task dto.UnifiedTask
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			logrus.Warnf("failed to parse task: %v", err)
			continue
		}

		if task.CronExpr != "" {
			nextTime, err := common.CronNextTime(task.CronExpr)
			if err != nil {
				logrus.Warnf("invalid cron expr: %v", err)
				if err := repository.HandleCronRescheduleFailure(ctx, []byte(taskData)); err != nil {
					logrus.Errorf("failed to handle cron reschedule failure: %v", err)
				}
				continue
			}

			task.ExecuteTime = nextTime.Unix()
			taskData, err := json.Marshal(task)
			if err != nil {
				logrus.Errorf("failed to marshal cron task %s: %v", task.TaskID, err)
				return
			}

			if err := repository.SubmitDelayedTask(ctx, taskData, task.TaskID, task.ExecuteTime); err != nil {
				logrus.Errorf("failed to reschedule cron task %s: %v", task.TaskID, err)
				err := repository.HandleCronRescheduleFailure(ctx, []byte(taskData))
				if err != nil {
					logrus.Errorf("failed to handle cron reschedule failure: %v", err)
				}

			}
		}
	}
}

// -----------------------------------------------------------------------------
// Task Consumption and Processing
// -----------------------------------------------------------------------------

// ConsumeTasks starts a consumer that processes tasks from the ready queue
func ConsumeTasks() {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("consumer panic: %v", r)
		}
	}()
	logrus.Info("start consume tasks")
	ctx := context.Background()

	for {
		if !repository.AcquireConcurrencyLock(ctx) {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		taskData, err := repository.GetTask(ctx, 30*time.Second)
		if err != nil {
			repository.ReleaseConcurrencyLock(ctx)
			if err == redis.Nil {
				continue
			}
			logrus.Errorf("BRPop error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		go processTask(ctx, taskData)
	}
}

// processTask handles a task from the queue
func processTask(ctx context.Context, taskData string) {
	defer repository.ReleaseConcurrencyLock(ctx)
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("task panic: %v\n%s", r, debug.Stack())
		}
	}()

	var task dto.UnifiedTask
	if err := json.Unmarshal([]byte(taskData), &task); err != nil {
		logrus.Warnf("invalid task data: %v", err)
		return
	}

	// Previously, ctx is an empty context.
	// ExtractContext injects the context information into the context
	traceCtx, taskCtx := extractContext(&task)
	traceSpan := trace.SpanFromContext(traceCtx)
	defer traceSpan.End()

	taskSpan := trace.SpanFromContext(taskCtx)
	defer taskSpan.End()

	startTime := time.Now()

	tasksProcessed.WithLabelValues(consts.GetTaskTypeName(task.Type), "started").Inc()

	executeTaskWithRetry(taskCtx, &task)

	taskDuration.WithLabelValues(consts.GetTaskTypeName(task.Type)).Observe(time.Since(startTime).Seconds())
}

// ExtractContext builds the trace and task contexts from a task
//
// Context hierarchy:
// 1. Always have group context
// 2.1 If there is no trace carrier, create a new trace span
// 2.2 If there is a trace carrier, extract the trace context
// 2.3 Always create a new task span
func extractContext(task *dto.UnifiedTask) (context.Context, context.Context) {
	var traceCtx context.Context
	var traceSpan trace.Span

	if task.TraceCarrier != nil {
		// Means it is a father span
		traceCtx = task.GetTraceCtx()
		logrus.WithField("task_id", task.TaskID).WithField("task_type", consts.GetTaskTypeName(task.Type)).Infof("Initial task group")
	} else {
		// Means it is a grand father span
		groupCtx := task.GetGroupCtx()

		// Create father first
		traceCtx, traceSpan = otel.Tracer("rcabench/trace").Start(groupCtx, fmt.Sprintf("start_task/%s", consts.GetTaskTypeName(task.Type)), trace.WithAttributes(
			attribute.String("trace_id", task.TraceID),
		))

		// Inject father into the carrier
		task.SetTraceCtx(traceCtx)

		traceSpan.SetStatus(codes.Ok, fmt.Sprintf("Started processing task trace %s", task.TraceID))
		logrus.WithField("task_id", task.TaskID).WithField("task_type", consts.GetTaskTypeName(task.Type)).Infof("Subsequent task")
	}

	taskCtx, _ := otel.Tracer("rcabench/task").Start(traceCtx,
		fmt.Sprintf("consume %s task", consts.GetTaskTypeName(task.Type)),
		trace.WithAttributes(
			attribute.String("task_id", task.TaskID),
			attribute.String("task_type", consts.GetTaskTypeName(task.Type)),
		))

	return traceCtx, taskCtx
}

// executeTaskWithRetry attempts to execute a task with retry logic
func executeTaskWithRetry(ctx context.Context, task *dto.UnifiedTask) {
	retryCtx, retryCancel := context.WithCancel(ctx)
	registerCancelFunc(task.TaskID, retryCancel)
	defer retryCancel()
	defer unregisterCancelFunc(task.TaskID)

	span := trace.SpanFromContext(ctx)

	errs := make([]error, 0)
	// TODO Task backoff
	for attempt := 0; attempt <= task.RetryPolicy.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-retryCtx.Done():
				logrus.Infof("Task %s canceled during retry", task.TaskID)
				return
			case <-time.After(time.Duration(task.RetryPolicy.BackoffSec) * time.Second):
			}
		}

		ctxWithCancel, cancel := context.WithCancel(ctx)
		_ = cancel

		err := dispatchTask(ctxWithCancel, task)
		if err == nil {
			tasksProcessed.WithLabelValues(consts.GetTaskTypeName(task.Type), "success").Inc()
			span.SetStatus(codes.Ok, fmt.Sprintf("Task %s of type %s completed successfully after %d attempts",
				task.TaskID, consts.GetTaskTypeName(task.Type), attempt+1))
			return
		}

		errs = append(errs, err)

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			logrus.WithField("task_id", task.TaskID).Info("Task canceled")
			span.RecordError(err)
			return
		}

		message := fmt.Sprintf("Attempt %d failed: %v", attempt+1, err)
		span.AddEvent(message)
		logrus.WithField("task_id", task.TaskID).Warn(message)
		publishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  task.Type,
			EventName: consts.EventTaskRetryStatus,
			Payload: dto.InfoPayloadTemplate{
				State: consts.GetTaskStateName(consts.TaskError),
				Msg:   err.Error(),
			},
		})
	}

	tasksProcessed.WithLabelValues(consts.GetTaskTypeName(task.Type), "failed").Inc()

	message := fmt.Sprintf("Task failed after %d attempts, errors: [%v]", task.RetryPolicy.MaxAttempts, errs)
	handleFinalFailure(ctx, task, message)
	updateTaskState(
		ctx,
		task.TraceID,
		task.TaskID,
		message,
		consts.TaskError,
		task.Type,
	)
}

// -----------------------------------------------------------------------------
// Task Cancellation and Control Functions
// -----------------------------------------------------------------------------

// registerCancelFunc stores a task's cancel function
func registerCancelFunc(taskID string, cancel context.CancelFunc) {
	taskCancelFuncsMutex.Lock()
	defer taskCancelFuncsMutex.Unlock()
	taskCancelFuncs[taskID] = cancel
}

// unregisterCancelFunc removes a task's cancel function
func unregisterCancelFunc(taskID string) {
	taskCancelFuncsMutex.Lock()
	defer taskCancelFuncsMutex.Unlock()
	delete(taskCancelFuncs, taskID)
}

// handleFinalFailure moves a failed task to the dead letter queue
func handleFinalFailure(ctx context.Context, task *dto.UnifiedTask, errMsg string) {
	taskData, err := json.Marshal(task)
	if err != nil {
		logrus.Errorf("failed to marshal failed task %s: %v", task.TaskID, err)
		return
	}

	if err := repository.HandleFailedTask(ctx, taskData, task.RetryPolicy.BackoffSec); err != nil {
		logrus.Errorf("failed to handle failed task %s: %v", task.TaskID, err)
	}

	span := trace.SpanFromContext(ctx)
	span.AddEvent(errMsg)
	span.SetStatus(codes.Error, fmt.Sprintf(consts.SpanStatusDescription, task.TaskID, consts.GetTaskStateName(consts.TaskError)))
	span.End()

	logrus.WithField("task_id", task.TaskID).Errorf("failed after %d attempts", task.RetryPolicy.MaxAttempts)
}

// CancelTask cancels a task and removes it from the queues
func CancelTask(taskID string) error {
	// Cancel execution context
	taskCancelFuncsMutex.RLock()
	cancelFunc, exists := taskCancelFuncs[taskID]
	taskCancelFuncsMutex.RUnlock()

	if exists {
		cancelFunc()
	}

	// Remove task from Redis
	ctx := context.Background()

	// Locate queue using index
	queueType, err := repository.GetTaskQueue(ctx, taskID)
	if err == nil {
		switch queueType {
		case repository.ReadyQueueKey:
			if _, err := repository.RemoveFromList(ctx, repository.ReadyQueueKey, taskID); err != nil {
				logrus.Warnf("failed to remove from list: %v", err)
			}
		case repository.DelayedQueueKey:
			if s := repository.RemoveFromZSet(ctx, repository.DelayedQueueKey, taskID); !s {
				logrus.Warnf("failed to remove from delayed queue: %v", err)
			}
		case repository.DeadLetterKey:
			if s := repository.RemoveFromZSet(ctx, repository.DeadLetterKey, taskID); !s {
				logrus.Warnf("failed to remove from dead letter queue: %v", err)
			}
		}
	}

	// Clean up index
	if err := repository.DeleteTaskIndex(ctx, taskID); err != nil {
		logrus.Warnf("failed to delete task index: %v", err)
	}

	if exists || err == nil {
		return nil
	}

	return fmt.Errorf("task %s not found", taskID)
}

// ===================== Redis Event Publishing =====================

// publishEvent publishes a StreamEvent to the specified Redis stream
// This adds caller information and handles error logging
func publishEvent(ctx context.Context, stream string, event dto.StreamEvent, opts ...eventPublishOption) {
	options := &eventPublishOptions{
		callerLevel: 2,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Business logic: Enhance event with caller information
	file, line, fn := utils.GetCallerInfo(options.callerLevel)
	event.FileName = file
	event.Line = line
	event.FnName = fn

	// Call repository layer for data access
	if err := client.RedisXAdd(ctx, stream, event.ToRedisStream()); err != nil {
		if err == redis.Nil {
			logrus.Warnf("No new messages to publish to Redis stream %s", stream)
			return
		}
		logrus.Errorf("Failed to publish event to Redis stream %s: %v", stream, err)
	}
}

// updateTaskState updates the task states and publishes the update
func updateTaskState(ctx context.Context, traceID, taskID, message string, taskState consts.TaskState, taskType consts.TaskType) {
	err := tracing.WithSpan(ctx, func(ctx context.Context) error {
		span := trace.SpanFromContext(ctx)
		logEntry := logrus.WithField("trace_id", traceID).WithField("task_id", taskID)
		span.AddEvent(message)

		description := fmt.Sprintf(consts.SpanStatusDescription, taskID, consts.GetTaskStateName(taskState))
		if taskState == consts.TaskCompleted {
			span.SetStatus(codes.Ok, description)
		}

		if taskState == consts.TaskError {
			span.SetStatus(codes.Error, description)
		}

		publishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, traceID), dto.StreamEvent{
			TaskID:    taskID,
			TaskType:  taskType,
			EventName: consts.EventTaskStateUpdate,
			Payload: dto.InfoPayloadTemplate{
				State: consts.GetTaskStateName(taskState),
				Msg:   message,
			},
		}, withCallerLevel(5))

		err := repository.UpdateTaskState(database.DB, ctx, taskID, taskState)
		if err != nil {
			logEntry.Errorf("failed to update database: %v", err)
		}
		return err
	})

	if err != nil {
		logrus.WithField("task_id", taskID).Errorf("failed to update task state: %v", err)
	}
}
