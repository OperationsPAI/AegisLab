package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/tracing"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
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

// -----------------------------------------------------------------------------
// Data Structures
// -----------------------------------------------------------------------------

// UnifiedTask represents a task that can be scheduled and executed
type UnifiedTask struct {
	TaskID       string                 `json:"task_id"`                 // Unique identifier for the task
	Type         consts.TaskType        `json:"type"`                    // Task type (determines how it's processed)
	Immediate    bool                   `json:"immediate"`               // Whether to execute immediately
	ExecuteTime  int64                  `json:"execute_time"`            // Unix timestamp for delayed execution
	CronExpr     string                 `json:"cron_expr,omitempty"`     // Cron expression for recurring tasks
	ReStartNum   int                    `json:"restart_num"`             // Number of restarts for the task
	RetryPolicy  RetryPolicy            `json:"retry_policy"`            // Policy for retrying failed tasks
	Payload      map[string]any         `json:"payload"`                 // Task-specific data
	TraceID      string                 `json:"trace_id,omitempty"`      // ID for tracing related tasks
	GroupID      string                 `json:"group_id,omitempty"`      // ID for grouping tasks
	TraceCarrier propagation.MapCarrier `json:"trace_carrier,omitempty"` // Carrier for trace context
	GroupCarrier propagation.MapCarrier `json:"group_carrier,omitempty"` // Carrier for group context
}

// RetryPolicy defines how tasks should be retried on failure
type RetryPolicy struct {
	MaxAttempts int `json:"max_attempts"` // Maximum number of retry attempts
	BackoffSec  int `json:"backoff_sec"`  // Seconds to wait between retries
}

// LastBatchInfo stores information about the last batch execution
type LastBatchInfo struct {
	ExecutionTime time.Time // When the batch was executed
	Interval      int       // Interval between batches
	Num           int       // Number of tasks in the batch
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
func SubmitTask(ctx context.Context, task *UnifiedTask) (string, string, error) {
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

	if task.Immediate {
		return task.TaskID, task.TraceID, submitImmediateTask(ctx, task)
	}

	return task.TaskID, task.TraceID, submitDelayedTask(ctx, task)
}

// submitImmediateTask sends a task to the ready queue for immediate execution
func submitImmediateTask(ctx context.Context, task *UnifiedTask) error {
	taskData, err := marshalTask(task)
	if err != nil {
		return err
	}

	redisCli := client.GetRedisClient()
	if err := redisCli.LPush(ctx, ReadyQueueKey, taskData).Err(); err != nil {
		return err
	}

	return redisCli.HSet(ctx, TaskIndexKey, task.TaskID, ReadyQueueKey).Err()
}

// submitDelayedTask sends a task to the delayed queue for future execution
func submitDelayedTask(ctx context.Context, task *UnifiedTask) error {
	executeTime, err := calculateExecuteTime(task)
	if err != nil {
		return err
	}

	taskData, err := marshalTask(task)
	if err != nil {
		return err
	}

	redisCli := client.GetRedisClient()
	if err := redisCli.ZAdd(ctx, DelayedQueueKey, redis.Z{
		Score:  float64(executeTime),
		Member: taskData,
	}).Err(); err != nil {
		return err
	}

	return redisCli.HSet(ctx, TaskIndexKey, task.TaskID, DelayedQueueKey).Err()
}

// marshalTask serializes a task to JSON
func marshalTask(task *UnifiedTask) ([]byte, error) {
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

// Lua script for processing delayed tasks efficiently
var delayedTaskScript = redis.NewScript(`
    local tasks = redis.call('ZRANGEBYSCORE', KEYS[1], 0, ARGV[1])
    if #tasks > 0 then
        redis.call('ZREMRANGEBYSCORE', KEYS[1], 0, ARGV[1])
        redis.call('LPUSH', KEYS[2], unpack(tasks))
        -- Update task index
        for _, task in ipairs(tasks) do
            local t = cjson.decode(task)
            redis.call('HSET', KEYS[3], t.task_id, KEYS[2])
        end
    end
    return tasks
`)

// ProcessDelayedTasks moves tasks from delayed queue to ready queue when their time arrives
func ProcessDelayedTasks(ctx context.Context) {
	redisCli := client.GetRedisClient()
	now := time.Now().Unix()

	result, err := delayedTaskScript.Run(ctx, redisCli,
		[]string{DelayedQueueKey, ReadyQueueKey, TaskIndexKey},
		now,
	).StringSlice()

	if err != nil && err != redis.Nil {
		logrus.Errorf("scheduler error: %v", err)
		return
	}

	for _, taskData := range result {
		var task UnifiedTask
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			logrus.Warnf("failed to parse task: %v", err)
			continue
		}

		if task.CronExpr != "" {
			nextTime, err := cronNextTime(task.CronExpr)
			if err != nil {
				logrus.Warnf("invalid cron expr: %v", err)
				handleCronRescheduleFailure(ctx, &task)
				continue
			}

			task.ExecuteTime = nextTime.Unix()
			if err := submitDelayedTask(ctx, &task); err != nil {
				logrus.Errorf("failed to reschedule cron task %s: %v", task.TaskID, err)
				handleCronRescheduleFailure(ctx, &task)
			}
		}
	}
}

// handleCronRescheduleFailure moves a failed cron task to the dead letter queue
func handleCronRescheduleFailure(ctx context.Context, task *UnifiedTask) {
	taskData, _ := json.Marshal(task)
	client.GetRedisClient().ZAdd(ctx, DeadLetterKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: taskData,
	})
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
	redisCli := client.GetRedisClient()

	for {
		if !acquireConcurrencyLock(ctx) {
			logrus.Info("no lock")
			time.Sleep(100 * time.Millisecond)
			continue
		}

		result, err := redisCli.BRPop(ctx, 30*time.Second, ReadyQueueKey).Result()
		if err != nil {
			releaseConcurrencyLock(ctx)
			if err == redis.Nil {
				continue
			}
			logrus.Errorf("BRPop error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		go processTask(ctx, result[1])
	}
}

// ExtractContext builds the trace and task contexts from a task
//
// Context hierarchy:
// 1. Always have group context
// 2.1 If there is no trace carrier, create a new trace span
// 2.2 If there is a trace carrier, extract the trace context
// 2.3 Always create a new task span
func ExtractContext(ctx context.Context, task *UnifiedTask) (context.Context, context.Context) {
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
	defer releaseConcurrencyLock(ctx)
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("task panic: %v\n%s", r, debug.Stack())
		}
	}()

	var task UnifiedTask
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
func executeTaskWithRetry(ctx context.Context, task *UnifiedTask) {
	retryCtx, retryCancel := context.WithCancel(ctx)
	registerCancelFunc(task.TaskID, retryCancel)
	defer retryCancel()
	defer unregisterCancelFunc(task.TaskID)

	span := trace.SpanFromContext(ctx)

	var err error
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

		err = dispatchTask(ctxWithCancel, task)
		if err == nil {
			tasksProcessed.WithLabelValues(string(task.Type), "success").Inc()
			span.SetStatus(codes.Ok, fmt.Sprintf("Task %s of type %s completed successfully after %d attempts",
				task.TaskID, task.Type, attempt+1))
			return
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			logrus.WithField("task_id", task.TaskID).Info("Task canceled")
			span.RecordError(err)
			return
		}

		message := fmt.Sprintf("Attempt %d failed: %v", attempt+1, err)
		span.AddEvent(message)
		logrus.WithField("task_id", task.TaskID).Warn(message)
	}

	tasksProcessed.WithLabelValues(string(task.Type), "failed").Inc()

	message := fmt.Sprintf("Task failed after %d attempts", task.RetryPolicy.MaxAttempts)
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
func handleFinalFailure(ctx context.Context, task *UnifiedTask, errMsg string) {
	deadLetterTime := time.Now().Add(time.Duration(task.RetryPolicy.BackoffSec) * time.Second).Unix()
	redisCli := client.GetRedisClient()
	taskData, _ := json.Marshal(task)
	redisCli.ZAdd(ctx, DeadLetterKey, redis.Z{
		Score:  float64(deadLetterTime),
		Member: taskData,
	})

	span := trace.SpanFromContext(ctx)
	span.AddEvent(errMsg)
	span.SetStatus(codes.Error, fmt.Sprintf(consts.SpanStatusDescription, task.TaskID, consts.TaskStatusError))
	span.End()

	logrus.WithField("task_id", task.TaskID).Errorf("failed after %d attempts", task.RetryPolicy.MaxAttempts)
}

// acquireConcurrencyLock attempts to acquire a lock for task execution
func acquireConcurrencyLock(ctx context.Context) bool {
	redisCli := client.GetRedisClient()
	currentCount, _ := redisCli.Get(ctx, ConcurrencyLockKey).Int64()
	if currentCount >= MaxConcurrency {
		return false
	}
	return redisCli.Incr(ctx, ConcurrencyLockKey).Err() == nil
}

// releaseConcurrencyLock releases a lock after task execution
func releaseConcurrencyLock(ctx context.Context) {
	redisCli := client.GetRedisClient()
	if err := redisCli.Decr(ctx, ConcurrencyLockKey).Err(); err != nil {
		logrus.Warnf("error releasing concurrency lock: %v", err)
	}
}

func InitConcurrencyLock(ctx context.Context) {
	redisCli := client.GetRedisClient()
	if err := redisCli.Set(ctx, ConcurrencyLockKey, 0, 0).Err(); err != nil {
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
	redisCli := client.GetRedisClient()

	// Locate queue using index
	queueType, err := redisCli.HGet(ctx, TaskIndexKey, taskID).Result()
	if err == nil {
		switch queueType {
		case ReadyQueueKey:
			if _, err := removeFromList(ctx, redisCli, ReadyQueueKey, taskID); err != nil {
				logrus.Warnf("failed to remove from list: %v", err)
			}
		case DelayedQueueKey:
			if s := removeFromZSet(ctx, redisCli, DelayedQueueKey, taskID); !s {
				logrus.Warnf("failed to remove from delayed queue: %v", err)
			}
		case DeadLetterKey:
			if s := removeFromZSet(ctx, redisCli, DeadLetterKey, taskID); !s {
				logrus.Warnf("failed to remove from dead letter queue: %v", err)
			}
		}
	}

	// Clean up index
	redisCli.HDel(ctx, TaskIndexKey, taskID)

	if exists || err == nil {
		return nil
	}

	return fmt.Errorf("task %s not found", taskID)
}

// -----------------------------------------------------------------------------
// Redis Utility Functions
// -----------------------------------------------------------------------------

// removeFromZSet removes a task from a Redis sorted set
func removeFromZSet(ctx context.Context, cli *redis.Client, key, taskID string) bool {
	members, err := cli.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	}).Result()
	if err != nil {
		return false
	}

	for _, member := range members {
		var t UnifiedTask
		if json.Unmarshal([]byte(member), &t) == nil && t.TaskID == taskID {
			if err := cli.ZRem(ctx, key, member).Err(); err != nil {
				logrus.Warnf("failed to remove from ZSet: %v", err)
				return false
			}
			return true
		}
	}

	return false
}

// removeFromList removes a task from a Redis list using Lua script
func removeFromList(ctx context.Context, cli *redis.Client, key, taskID string) (bool, error) {
	// Efficient list removal Lua script
	var removeFromListScript = redis.NewScript(`
		local key = KEYS[1]
		local taskID = ARGV[1]
		local count = 0

		for i=0, redis.call('LLEN', key)-1 do
			local item = redis.call('LINDEX', key, i)
			if item then
				local task = cjson.decode(item)
				if task.task_id == taskID then
					redis.call('LSET', key, i, "__DELETED__")
					count = count + 1
				end
			end
		end

		if count > 0 then
			redis.call('LREM', key, count, "__DELETED__")
		end
		
		return count
	`)
	result, err := removeFromListScript.Run(ctx, cli, []string{key}, taskID).Int()
	if err != nil {
		return false, fmt.Errorf("failed to remove from list: %w", err)
	}

	return result > 0, nil
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

		client.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, traceID), client.StreamEvent{
			TaskID:    taskID,
			TaskType:  taskType,
			EventName: consts.EventTaskStatusUpdate,
			Payload: client.InfoPayloadTemplate{
				Status: taskStatus,
				Msg:    message,
			},
		}, client.WithCallerLevel(5))

		tx := database.DB.WithContext(ctx).Begin()
		if err := tx.Model(&database.Task{}).
			Where("id = ?", taskID).
			Update("status", taskStatus).Error; err != nil {
			tx.Rollback()
			logEntry.Errorf("failed to update database: %v", err)
			return err
		}
		tx.Commit()
		return nil
	})
}

// -----------------------------------------------------------------------------
// Utility Functions
// -----------------------------------------------------------------------------

// calculateExecuteTime determines when a task should be executed
func calculateExecuteTime(task *UnifiedTask) (int64, error) {
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
