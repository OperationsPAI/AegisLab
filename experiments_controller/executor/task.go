package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/LGU-SE-Internal/rcabench/tracing"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

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

// LastBatchInfo stores information about the last batch execution
type LastBatchInfo struct {
	ExecutionTime time.Time // When the batch was executed
	Interval      int       // Interval between batches
	Num           int       // Number of tasks in the batch
}

// -----------------------------------------------------------------------------
// Task Submission Functions
// -----------------------------------------------------------------------------

// SubmitTask creates and submits a new task to the appropriate queue
//
// Task Context Hierarchy:
//  1. If GroupCarrier is not nil: task is an initial task that spawns several traces
//     1.2. If TraceCarrier is nil, create a new one
//  2. If TraceCarrier is not nil: task is within a task trace
//
// When calling SubmitTask:
// - For initial task: fill in the GroupCarrier (parent's parent)
// - For subsequent task: fill in the TraceCarrier (parent)
// - The context itself is the youngest span
//
// Hierarchy example:
//
//	Group -> Trace -> Task 1
//	               -> Task 2
//	               -> Task 3
//	               -> Task 4
//	               -> Task 5
func SubmitTask(ctx context.Context, task *dto.UnifiedTask) (string, string, error) {
	if task.TaskID == "" {
		task.TaskID = uuid.NewString()
	}
	if task.TraceID == "" {
		task.TraceID = uuid.NewString()
	}

	jsonPayload, err := json.Marshal(task.Payload)
	if err != nil {
		return "", "", err
	}

	t := database.Task{
		ID:          task.TaskID,
		Type:        string(task.Type),
		Payload:     string(jsonPayload),
		Immediate:   task.Immediate,
		ExecuteTime: task.ExecuteTime,
		CronExpr:    task.CronExpr,
		Status:      consts.TaskStatusPending,
		TraceID:     task.TraceID,
		GroupID:     task.GroupID,
	}
	if err := database.DB.Create(&t).Error; err != nil {
		logrus.Errorf("failed to save task to database, err: %v", err)
		return "", "", err
	}

	taskData, err := marshalTask(task)
	if err != nil {
		return "", "", err
	}

	if task.Immediate {
		err = repository.SubmitImmediateTask(ctx, taskData, task.TaskID)
	} else {
		executeTime, err := calculateExecuteTime(task)
		if err != nil {
			return "", "", err
		}
		err = repository.SubmitDelayedTask(ctx, taskData, task.TaskID, executeTime)
	}

	if err != nil {
		return "", "", err
	}

	return task.TaskID, task.TraceID, nil
}

// marshalTask serializes a task to JSON
func marshalTask(task *dto.UnifiedTask) ([]byte, error) {
	taskData, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("task marshaling failed: %w", err)
	}
	return taskData, nil
}

// -----------------------------------------------------------------------------
// Task Scheduling Functions
// -----------------------------------------------------------------------------

// StartScheduler starts the scheduler that moves tasks from delayed to ready queue
func StartScheduler(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ProcessDelayedTasks(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// ProcessDelayedTasks moves tasks from delayed queue to ready queue when their time arrives
func ProcessDelayedTasks(ctx context.Context) {
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
			nextTime, err := cronNextTime(task.CronExpr)
			if err != nil {
				logrus.Warnf("invalid cron expr: %v", err)
				repository.HandleCronRescheduleFailure(ctx, []byte(taskData))
				continue
			}

			task.ExecuteTime = nextTime.Unix()
			taskData, _ := marshalTask(&task)
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
			logrus.Info("no lock")
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

// ExtractContext builds the trace and task contexts from a task
//
// Context hierarchy:
// 1. Always have group context
// 2.1 If there is no trace carrier, create a new trace span
// 2.2 If there is a trace carrier, extract the trace context
// 2.3 Always create a new task span
func ExtractContext(ctx context.Context, task *dto.UnifiedTask) (context.Context, context.Context) {
	var traceCtx context.Context
	var traceSpan trace.Span

	if task.TraceCarrier != nil {
		// Means it is a father span
		traceCtx = task.GetTraceCtx()
		logrus.WithField("task_id", task.TaskID).WithField("task_type", task.Type).Infof("Initial task group")
	} else {
		// Means it is a grand father span
		groupCtx := task.GetGroupCtx()

		// Create father first
		traceCtx, traceSpan = otel.Tracer("rcabench/trace").Start(groupCtx, fmt.Sprintf("start_task/%s", task.Type), trace.WithAttributes(
			attribute.String("trace_id", task.TraceID),
		))

		// Inject father into the carrier
		task.SetTraceCtx(traceCtx)

		traceSpan.SetStatus(codes.Ok, fmt.Sprintf("Started processing task trace %s", task.TraceID))
		logrus.WithField("task_id", task.TaskID).WithField("task_type", task.Type).Infof("Subsequent task")
	}

	taskCtx, _ := otel.Tracer("rcabench/task").Start(traceCtx,
		fmt.Sprintf("consume %s task", task.Type),
		trace.WithAttributes(
			attribute.String("task_id", task.TaskID),
			attribute.String("task_type", string(task.Type)),
		))

	return traceCtx, taskCtx
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
	traceCtx, taskCtx := ExtractContext(ctx, &task)
	traceSpan := trace.SpanFromContext(traceCtx)
	defer traceSpan.End()

	taskSpan := trace.SpanFromContext(taskCtx)
	defer taskSpan.End()

	startTime := time.Now()

	tasksProcessed.WithLabelValues(string(task.Type), "started").Inc()

	executeTaskWithRetry(taskCtx, &task)

	taskDuration.WithLabelValues(string(task.Type)).Observe(time.Since(startTime).Seconds())
}

// executeTaskWithRetry attempts to execute a task with retry logic
func executeTaskWithRetry(ctx context.Context, task *dto.UnifiedTask) {
	retryCtx, retryCancel := context.WithCancel(ctx)
	registerCancelFunc(task.TaskID, retryCancel)
	defer retryCancel()
	defer unregisterCancelFunc(task.TaskID)

	span := trace.SpanFromContext(ctx)

	errs := make([]error, 0)
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
			tasksProcessed.WithLabelValues(string(task.Type), "success").Inc()
			span.SetStatus(codes.Ok, fmt.Sprintf("Task %s of type %s completed successfully after %d attempts",
				task.TaskID, task.Type, attempt+1))
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
		repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  task.Type,
			EventName: consts.EventTaskRetryStatus,
			Payload: dto.InfoPayloadTemplate{
				Status: consts.TaskStatusError,
				Msg:    err.Error(),
			},
		})
	}

	tasksProcessed.WithLabelValues(string(task.Type), "failed").Inc()

	message := fmt.Sprintf("Task failed after %d attempts, errors: [%v]", task.RetryPolicy.MaxAttempts, errs)
	handleFinalFailure(ctx, task, message)
	updateTaskStatus(
		ctx,
		task.TraceID,
		task.TaskID,
		message,
		consts.TaskStatusError,
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
	taskData, _ := marshalTask(task)
	repository.HandleFailedTask(ctx, taskData, task.RetryPolicy.BackoffSec)

	span := trace.SpanFromContext(ctx)
	span.AddEvent(errMsg)
	span.SetStatus(codes.Error, fmt.Sprintf(consts.SpanStatusDescription, task.TaskID, consts.TaskStatusError))
	span.End()

	logrus.WithField("task_id", task.TaskID).Errorf("failed after %d attempts", task.RetryPolicy.MaxAttempts)
}

// InitConcurrencyLock initializes the concurrency lock counter
func InitConcurrencyLock(ctx context.Context) {
	if err := repository.InitConcurrencyLock(ctx); err != nil {
		logrus.Fatalf("error setting concurrency lock to 0: %v", err)
	}
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
	repository.DeleteTaskIndex(ctx, taskID)

	if exists || err == nil {
		return nil
	}

	return fmt.Errorf("task %s not found", taskID)
}

// -----------------------------------------------------------------------------
// Task Status Update Functions
// -----------------------------------------------------------------------------

// updateTaskStatus updates the task status and publishes the update
func updateTaskStatus(ctx context.Context, traceID, taskID, message, taskStatus string, taskType consts.TaskType) {
	tracing.WithSpan(ctx, func(ctx context.Context) error {
		span := trace.SpanFromContext(ctx)
		logEntry := logrus.WithField("trace_id", traceID).WithField("task_id", taskID)
		span.AddEvent(message)

		description := fmt.Sprintf(consts.SpanStatusDescription, taskID, taskStatus)
		if taskStatus == consts.TaskStatusCompleted {
			span.SetStatus(codes.Ok, description)
		}

		if taskStatus == consts.TaskStatusError {
			span.SetStatus(codes.Error, description)
		}

		repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, traceID), dto.StreamEvent{
			TaskID:    taskID,
			TaskType:  taskType,
			EventName: consts.EventTaskStatusUpdate,
			Payload: dto.InfoPayloadTemplate{
				Status: taskStatus,
				Msg:    message,
			},
		}, repository.WithCallerLevel(5))

		err := repository.UpdateTaskStatus(ctx, taskID, taskStatus)
		if err != nil {
			logEntry.Errorf("failed to update database: %v", err)
		}
		return err
	})
}

// -----------------------------------------------------------------------------
// Utility Functions
// -----------------------------------------------------------------------------

// calculateExecuteTime determines when a task should be executed
func calculateExecuteTime(task *dto.UnifiedTask) (int64, error) {
	if task.Type == "cron" {
		next, err := cronNextTime(task.CronExpr)
		if err != nil {
			return 0, err
		}
		return next.Unix(), nil
	}
	return task.ExecuteTime, nil
}

// cronNextTime calculates the next execution time from a cron expression
func cronNextTime(expr string) (time.Time, error) {
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(expr)
	if err != nil {
		return time.Time{}, err
	}

	return schedule.Next(time.Now()), nil
}
