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
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// 常量定义
const (
	DelayedQueueKey    = "task:delayed"
	ReadyQueueKey      = "task:ready"
	DeadLetterKey      = "task:dead"
	TaskIndexKey       = "task:index"
	GroupIndexKey      = "group:index"
	ConcurrencyLockKey = "task:concurrency_lock"
	MaxConcurrency     = 20
)

// 监控指标
var (
	tasksProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "task_processed_total",
		Help: "Total number of processed tasks",
	}, []string{"type", "status"})

	taskDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "task_duration_seconds",
		Help:    "Task processing duration distribution",
		Buckets: []float64{0.1, 0.5, 1, 5, 10, 30},
	}, []string{"type"})
)

// UnifiedTask 统一任务结构
type UnifiedTask struct {
	TaskID       string                 `json:"task_id"`
	Type         consts.TaskType        `json:"type"`
	Immediate    bool                   `json:"immediate"`
	ExecuteTime  int64                  `json:"execute_time"`
	CronExpr     string                 `json:"cron_expr,omitempty"`
	RetryPolicy  RetryPolicy            `json:"retry_policy"`
	Payload      map[string]any         `json:"payload"`
	TraceID      string                 `json:"trace_id,omitempty"`
	GroupID      string                 `json:"group_id,omitempty"`
	TraceCarrier propagation.MapCarrier `json:"trace_carrier,omitempty"`
	GroupCarrier propagation.MapCarrier `json:"group_carrier,omitempty"`
}

type RetryPolicy struct {
	MaxAttempts int `json:"max_attempts"`
	BackoffSec  int `json:"backoff_sec"`
}

var (
	taskCancelFuncs      = make(map[string]context.CancelFunc)
	taskCancelFuncsMutex sync.RWMutex
)

// SubmitTask 提交任务入口
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

func submitImmediateTask(ctx context.Context, task *UnifiedTask) error {
	taskData, err := marshalTask(task)
	if err != nil {
		return err
	}

	redisCli := client.GetRedisClient()
	if err := redisCli.LPush(ctx, ReadyQueueKey, taskData).Err(); err != nil {
		return err
	}

	// 创建任务索引
	return redisCli.HSet(ctx, TaskIndexKey, task.TaskID, ReadyQueueKey).Err()
}

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
	if err := redisCli.ZAdd(ctx, DelayedQueueKey, &redis.Z{
		Score:  float64(executeTime),
		Member: taskData,
	}).Err(); err != nil {
		return err
	}

	// 创建任务索引
	return redisCli.HSet(ctx, TaskIndexKey, task.TaskID, DelayedQueueKey).Err()
}

func marshalTask(task *UnifiedTask) ([]byte, error) {
	taskData, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("task marshaling failed: %w", err)
	}
	return taskData, nil
}

// StartScheduler 启动调度器
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

// 优化后的Lua脚本
var delayedTaskScript = redis.NewScript(`
    local tasks = redis.call('ZRANGEBYSCORE', KEYS[1], 0, ARGV[1])
    if #tasks > 0 then
        redis.call('ZREMRANGEBYSCORE', KEYS[1], 0, ARGV[1])
        redis.call('LPUSH', KEYS[2], unpack(tasks))
        -- 更新任务索引
        for _, task in ipairs(tasks) do
            local t = cjson.decode(task)
            redis.call('HSET', KEYS[3], t.task_id, KEYS[2])
        end
    end
    return tasks
`)

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

func handleCronRescheduleFailure(ctx context.Context, task *UnifiedTask) {
	taskData, _ := json.Marshal(task)
	client.GetRedisClient().ZAdd(ctx, DeadLetterKey, &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: taskData,
	})
}

// ConsumeTasks 消费任务
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

	var traceCtx context.Context
	var span trace.Span
	if task.TraceCarrier != nil {
		traceCtx = otel.GetTextMapPropagator().Extract(context.Background(), task.TraceCarrier)
	} else {
		groupCtx := otel.GetTextMapPropagator().Extract(context.Background(), task.GroupCarrier)
		traceCtx, span = otel.Tracer("rcabench/trace").Start(groupCtx, "consume trace", trace.WithAttributes(
			attribute.String("trace_id", task.TraceID),
		))
		defer span.End()

		task.TraceCarrier = make(propagation.MapCarrier)
		otel.GetTextMapPropagator().Inject(traceCtx, task.TraceCarrier)

		span.SetStatus(codes.Ok, fmt.Sprintf("Started processing task trace %s", task.TraceID))
	}

	taskCtx, _ := otel.Tracer("rcabench/task").Start(traceCtx,
		fmt.Sprintf("consume %s task", task.Type),
		trace.WithAttributes(
			attribute.String("task_id", task.TaskID),
			attribute.String("task_type", string(task.Type)),
		))

	logrus.Infof("dealing with task %s, type: %s, groupID: %s", task.TaskID, task.Type, task.GroupID)

	startTime := time.Now()
	tasksProcessed.WithLabelValues(string(task.Type), "started").Inc()

	executeTaskWithRetry(taskCtx, &task)

	duration := time.Since(startTime).Seconds()
	taskDuration.WithLabelValues(string(task.Type)).Observe(duration)
}

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
			span.End()
			return
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			logrus.WithField("task_id", task.TaskID).Info("Task canceled")
			return
		}

		message := fmt.Sprintf("Attempt %d failed: %v", attempt+1, err)
		span.AddEvent(message)
		logrus.WithField("task_id", task.TaskID).Warn(message)
	}

	tasksProcessed.WithLabelValues(string(task.Type), "failed").Inc()
	handleFinalFailure(ctx, task)

	updateTaskError(
		nil,
		task.TaskID,
		task.TraceID,
		task.Type,
		fmt.Sprintf("Task failed after %d attempts: %v", task.RetryPolicy.MaxAttempts, err),
	)
}

// 注册取消函数
func registerCancelFunc(taskID string, cancel context.CancelFunc) {
	taskCancelFuncsMutex.Lock()
	defer taskCancelFuncsMutex.Unlock()
	taskCancelFuncs[taskID] = cancel
}

// 注销取消函数
func unregisterCancelFunc(taskID string) {
	taskCancelFuncsMutex.Lock()
	defer taskCancelFuncsMutex.Unlock()
	delete(taskCancelFuncs, taskID)
}

func handleFinalFailure(ctx context.Context, task *UnifiedTask) {
	deadLetterTime := time.Now().Add(time.Duration(task.RetryPolicy.BackoffSec) * time.Second).Unix()
	redisCli := client.GetRedisClient()
	taskData, _ := json.Marshal(task)
	redisCli.ZAdd(ctx, DeadLetterKey, &redis.Z{
		Score:  float64(deadLetterTime),
		Member: taskData,
	})

	span := trace.SpanFromContext(ctx)
	span.SetStatus(codes.Error, fmt.Sprintf("failed to execute task %s", task.TaskID))
	span.End()

	logrus.WithField("task_id", task.TaskID).Errorf("failed after %d attempts", task.RetryPolicy.MaxAttempts)
}

// 分布式并发控制
func acquireConcurrencyLock(ctx context.Context) bool {
	redisCli := client.GetRedisClient()
	currentCount, _ := redisCli.Get(ctx, ConcurrencyLockKey).Int64()
	if currentCount >= MaxConcurrency {
		return false
	}
	return redisCli.Incr(ctx, ConcurrencyLockKey).Err() == nil
}

func releaseConcurrencyLock(ctx context.Context) {
	redisCli := client.GetRedisClient()
	if err := redisCli.Decr(ctx, ConcurrencyLockKey).Err(); err != nil {
		logrus.Warnf("error releasing concurrency lock: %v", err)
	}
}

// 改进的任务取消机制
func CancelTask(taskID string) error {
	// 取消执行上下文
	taskCancelFuncsMutex.RLock()
	cancelFunc, exists := taskCancelFuncs[taskID]
	taskCancelFuncsMutex.RUnlock()

	if exists {
		cancelFunc()
	}

	// 从Redis删除任务
	ctx := context.Background()
	redisCli := client.GetRedisClient()

	// 通过索引快速定位队列
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

	// 清理索引
	redisCli.HDel(ctx, TaskIndexKey, taskID)

	if exists || err == nil {
		return nil
	}

	return fmt.Errorf("task %s not found", taskID)
}

// 高效任务删除实现
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

func removeFromList(ctx context.Context, cli *redis.Client, key, taskID string) (bool, error) {
	// 高效列表删除Lua脚本
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

func parseRdbMsgFromPayload(payload map[string]any) (*dto.RdbMsg, error) {
	message := "missing or invalid '%s' key in payload"

	status, ok := payload[consts.RdbMsgStatus].(string)
	if !ok || status == "" {
		return nil, fmt.Errorf(message, consts.RdbMsgStatus)
	}

	taskType, ok := payload[consts.RdbMsgTaskType].(consts.TaskType)
	if !ok {
		return nil, fmt.Errorf(message, consts.RdbMsgTaskType)
	}

	return &dto.RdbMsg{
		Status: status,
		Type:   taskType,
	}, nil
}

func updateTaskError(taskCarrier propagation.MapCarrier, taskID, traceID string, taskType consts.TaskType, errMsg string) {
	updateTaskStatus(
		taskCarrier,
		taskID,
		traceID,
		fmt.Sprintf(consts.TaskMsgFailed, taskID),
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusError,
			consts.RdbMsgTaskID:   taskID,
			consts.RdbMsgTaskType: taskType,
			consts.RdbMsgError:    errMsg,
		})
}

// 事务型状态更新
func updateTaskStatus(taskCarrier propagation.MapCarrier, taskID, traceID, message string, payload map[string]any) {
	rdbMsg, err := parseRdbMsgFromPayload(payload)
	if err != nil {
		logrus.WithField("task_id", taskID).Error(err)
		return
	}

	ctx := context.Background()

	// Redis事务
	redisCli := client.GetRedisClient()
	pipe := redisCli.TxPipeline()
	pipe.HSet(ctx, fmt.Sprintf(consts.StatusKey, taskID),
		"status", rdbMsg.Status,
		"updated_at", time.Now().Unix(),
	)
	if _, err := pipe.Exec(ctx); err != nil {
		logrus.WithField("task_id", taskID).Errorf("failed to update task status in redis: %v", err)
		return
	}

	// 数据库事务
	tx := database.DB.WithContext(ctx).Begin()
	if err := tx.Model(&database.Task{}).
		Where("id = ?", taskID).
		Update("status", rdbMsg.Status).Error; err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()

	msg, err := json.Marshal(payload)
	if err != nil {
		logrus.WithField("task_id", taskID).Errorf("failed to marshal payload: %v", err)
		return
	}

	redisCli.Publish(ctx, fmt.Sprintf(consts.SubChannel, traceID), msg)

	if taskCarrier != nil {
		taskCtx := otel.GetTextMapPropagator().Extract(context.Background(), taskCarrier)
		// 处理TaskCtx
		_, span := otel.Tracer("rcabench/task/feedback").Start(taskCtx, "process feedback", trace.WithAttributes(
			attribute.String("task_id", taskID),
			attribute.String("task_type", string(rdbMsg.Type)),
		))
		defer span.End()

		span.AddEvent(message)
		description := fmt.Sprintf("task %s %s", taskID, rdbMsg.Status)
		if rdbMsg.Status == consts.TaskStatusCompleted {
			span.SetStatus(codes.Ok, description)
		}

		if rdbMsg.Status == consts.TaskStatusError {
			span.SetStatus(codes.Error, description)
		}
	}
}

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

func cronNextTime(expr string) (time.Time, error) {
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(expr)
	if err != nil {
		return time.Time{}, err
	}
	return schedule.Next(time.Now()), nil
}

func AddGroupIndex(ctx context.Context, groupID, traceID string) {
	redisCli := client.GetRedisClient()

	pipe := redisCli.TxPipeline()
	pipe.HSet(ctx, GroupIndexKey, groupID, traceID)

	if _, err := pipe.Exec(ctx); err != nil {
		logrus.WithFields(logrus.Fields{
			"group_id": groupID,
			"trace_id": traceID,
		}).Error("failed to build index")
	}
}

func getFinalTraceIndex(ctx context.Context, groupID string) string {
	redisCli := client.GetRedisClient()

	taskID, err := redisCli.HGet(ctx, GroupIndexKey, groupID).Result()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"group_id": groupID,
		}).Errorf("the group ID %s is not in dataset index: %v", groupID, err)
		return ""
	}

	return taskID
}
