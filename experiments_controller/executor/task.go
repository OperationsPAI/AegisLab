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
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// 常量定义
const (
	DelayedQueueKey     = "task:delayed"
	ReadyQueueKey       = "task:ready"
	DeadLetterKey       = "task:dead"
	TaskIndexKey        = "task:index"
	TaskDatasetIndexKey = "task:dataset_index"
	ConcurrencyLockKey  = "task:concurrency_lock"
	MaxConcurrency      = 20
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
	TaskID      string         `json:"task_id"`
	Type        TaskType       `json:"type"`
	Immediate   bool           `json:"immediate"`
	ExecuteTime int64          `json:"execute_time"`
	CronExpr    string         `json:"cron_expr,omitempty"`
	RetryPolicy RetryPolicy    `json:"retry_policy"`
	Payload     map[string]any `json:"payload"`
	TraceID     string         `json:"trace_id,omitempty"`
	GroupID     string         `json:"group_id,omitempty"`
}

type RetryPolicy struct {
	MaxAttempts int `json:"max_attempts"`
	BackoffSec  int `json:"backoff_sec"`
}

type RdbMsg struct {
	Status string   `json:"status"`
	Error  *string  `json:"error"`
	Type   TaskType `json:"task_type"`
}

type TaskMeta struct {
	Benchmark   string `redis:"benchmark" mapstructure:"benchmark"`
	PreDuration int    `redis:"pre_duration" mapstructure:"pre_duration"`
	TraceID     string `redis:"trace_id" mapstructure:"trace_id"`
	GroupID     string `redis:"group_id" mapstructure:"group_id"`
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
		Status:      TaskStatusPending,
		TraceID:     task.TraceID,
		GroupID:     task.GroupID,
	}
	if err := database.DB.Create(&t).Error; err != nil {
		logrus.Errorf("Failed to save task to database, err: %v", err)
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
		logrus.Errorf("Scheduler error: %v", err)
		return
	}

	for _, taskData := range result {
		var task UnifiedTask
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			logrus.WithError(err).Warn("Failed to parse task")
			continue
		}

		if task.CronExpr != "" {
			nextTime, err := cronNextTime(task.CronExpr)
			if err != nil {
				logrus.WithError(err).Warn("Invalid cron expr")
				handleCronRescheduleFailure(ctx, &task)
				continue
			}
			task.ExecuteTime = nextTime.Unix()
			if err := submitDelayedTask(ctx, &task); err != nil {
				logrus.Errorf("Failed to reschedule cron task %s: %v", task.TaskID, err)
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
			logrus.Errorf("Consumer panic: %v", r)
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
			logrus.Errorf("Task panic: %v\n%s", r, debug.Stack())
		}
	}()

	var task UnifiedTask
	if err := json.Unmarshal([]byte(taskData), &task); err != nil {
		logrus.WithError(err).Warn("Invalid task data")
		return
	}
	logrus.Infof("Dealing with task %s, type: %s, groupID: %s", task.TaskID, task.Type, task.GroupID)

	startTime := time.Now()
	tasksProcessed.WithLabelValues(string(task.Type), "started").Inc()

	executeTaskWithRetry(ctx, &task)

	duration := time.Since(startTime).Seconds()
	taskDuration.WithLabelValues(string(task.Type)).Observe(duration)
}

func executeTaskWithRetry(ctx context.Context, task *UnifiedTask) {
	retryCtx, retryCancel := context.WithCancel(ctx)
	defer retryCancel()
	registerCancelFunc(task.TaskID, retryCancel)
	defer unregisterCancelFunc(task.TaskID)

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

		taskCtx, _ := context.WithCancel(retryCtx)
		err = dispatchTask(taskCtx, task)
		if err == nil {
			tasksProcessed.WithLabelValues(string(task.Type), "success").Inc()
			return
		}

		if errors.Is(err, context.Canceled) {
			logrus.WithField("task_id", task.TaskID).Info("Task canceled")
			return
		}

		logrus.WithField("task_id", task.TaskID).Warnf("Attempt %d failed: %v", attempt+1, err)
	}

	tasksProcessed.WithLabelValues(string(task.Type), "failed").Inc()
	handleFinalFailure(ctx, task)

	message := err.Error()
	updateTaskStatus(task.TaskID, task.TraceID,
		message,
		map[string]any{
			RdbMsgStatus: TaskStatusError,
			RdbMsgError:  message,
		})
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
	logrus.WithField("task_id", task.TaskID).Errorf("Failed after %d attempts", task.RetryPolicy.MaxAttempts)
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
		logrus.WithError(err).Warn("Error releasing concurrency lock")
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
			removeFromList(ctx, redisCli, ReadyQueueKey, taskID)
		case DelayedQueueKey:
			removeFromZSet(ctx, redisCli, DelayedQueueKey, taskID)
		case DeadLetterKey:
			removeFromZSet(ctx, redisCli, DeadLetterKey, taskID)
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
				logrus.WithError(err).Warn("Failed to remove from ZSet")
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

func parseRdbMsg(payload map[string]any) (*RdbMsg, error) {
	status, ok := payload[RdbMsgStatus].(string)
	if !ok || status == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", RdbMsgStatus)
	}

	return &RdbMsg{
		Status: status,
	}, nil
}

func addDatasetIndex(taskID, name string) {
	ctx := context.Background()
	redisCli := client.GetRedisClient()

	pipe := redisCli.TxPipeline()
	pipe.HSet(ctx, TaskDatasetIndexKey, name, taskID)

	if _, err := pipe.Exec(ctx); err != nil {
		logrus.WithField("dataset", name).WithField("task_id", taskID).Error("Failed to build index")
		return
	}
}

func checkDatasetIndex(name string) string {
	ctx := context.Background()
	redisCli := client.GetRedisClient()

	taskID, err := redisCli.HGet(ctx, TaskDatasetIndexKey, name).Result()
	if err != nil {
		logrus.WithField("dataset", name).Errorf("The name is not in dataset index: %v", err)
	}

	return taskID
}

func addTaskMeta(taskID string, values ...any) {
	ctx := context.Background()
	redisCli := client.GetRedisClient()

	pipe := redisCli.TxPipeline()
	pipe.HSet(ctx, fmt.Sprintf(MetaKey, taskID), values...)

	if _, err := pipe.Exec(ctx); err != nil {
		logrus.WithField("task_id", taskID).Errorf("Failed to store task meta: %v", err)
		return
	}
}

func getTaskMeta(taskID string) (*TaskMeta, error) {
	if taskID == "" {
		message := "The task_id can not be blank"
		logrus.Error(message)
		return nil, fmt.Errorf(message)
	}

	ctx := context.Background()
	redisCli := client.GetRedisClient()

	result, err := redisCli.HGetAll(ctx, fmt.Sprintf(MetaKey, taskID)).Result()
	if err != nil {
		logrus.WithField("task_id", taskID).Errorf("Failed to read metadata: %v", err)
		return nil, err
	}

	if len(result) == 0 {
		message := "Task metadata does not exist"
		logrus.WithField("task_id", taskID).Warn(message)
		return nil, fmt.Errorf(message)
	}

	var meta TaskMeta
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		Result:           &meta,
		WeaklyTypedInput: true, // 启用弱类型转换
	})
	if err != nil {
		logrus.WithField("task_id", taskID).Errorf("Failed to create decoder: %v", err)
		return nil, err
	}

	if err := decoder.Decode(result); err != nil {
		logrus.WithField("task_id", taskID).Errorf("Failed to parse metadata: %v", err)
		return nil, err
	}

	return &meta, nil
}

// 事务型状态更新
func updateTaskStatus(taskID, traceID, message string, payload map[string]any) {
	ctx := context.Background()
	redisCli := client.GetRedisClient()

	rdbMsg, err := parseRdbMsg(payload)
	if err != nil {
		logrus.WithField("task_id", taskID).Error(err)
		return
	}

	// Redis事务
	pipe := redisCli.TxPipeline()
	pipe.HSet(ctx, fmt.Sprintf(StatusKey, taskID),
		"status", rdbMsg.Status,
		"updated_at", time.Now().Unix(),
	)

	pipe.RPush(ctx, fmt.Sprintf(LogKey, taskID), fmt.Sprintf(LogFormat, rdbMsg.Status, message))
	if _, err := pipe.Exec(ctx); err != nil {
		logrus.WithField("task_id", taskID).WithError(err).Error("Failed to update task status")
		return
	}

	// 数据库事务
	tx := database.DB.Begin()
	if err := tx.Model(&database.Task{}).
		Where("id = ?", taskID).
		Update("status", rdbMsg.Status).Error; err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()

	msg, err := json.Marshal(payload)
	if err != nil {
		logrus.WithField("task_id", taskID).WithError(err).Error("Failed to marshal payload")
		return
	}

	redisCli.Publish(ctx, fmt.Sprintf(SubChannel, traceID), msg)
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
