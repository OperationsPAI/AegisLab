package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/LGU-SE-Internal/rcabench/client"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

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

func FindTaskItemByID(id string) (*dto.TaskItem, error) {
	var result database.Task
	if err := database.DB.Where("tasks.id = ?", id).First(&result).Error; err != nil {
		return nil, err
	}

	var item dto.TaskItem
	if err := item.Convert(result); err != nil {
		return nil, err
	}

	return &item, nil
}

// SubmitImmediateTask sends a task to the ready queue for immediate execution
func SubmitImmediateTask(ctx context.Context, taskData []byte, taskID string) error {
	redisCli := client.GetRedisClient()
	if err := redisCli.LPush(ctx, ReadyQueueKey, taskData).Err(); err != nil {
		return err
	}

	return redisCli.HSet(ctx, TaskIndexKey, taskID, ReadyQueueKey).Err()
}

// SubmitDelayedTask sends a task to the delayed queue for future execution
func SubmitDelayedTask(ctx context.Context, taskData []byte, taskID string, executeTime int64) error {
	redisCli := client.GetRedisClient()
	if err := redisCli.ZAdd(ctx, DelayedQueueKey, redis.Z{
		Score:  float64(executeTime),
		Member: taskData,
	}).Err(); err != nil {
		return err
	}

	return redisCli.HSet(ctx, TaskIndexKey, taskID, DelayedQueueKey).Err()
}

// ProcessDelayedTasks moves tasks from delayed queue to ready queue when their time arrives
func ProcessDelayedTasks(ctx context.Context) ([]string, error) {
	redisCli := client.GetRedisClient()
	now := time.Now().Unix()

	delayedTaskScript := redis.NewScript(`
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

	result, err := delayedTaskScript.Run(ctx, redisCli,
		[]string{DelayedQueueKey, ReadyQueueKey, TaskIndexKey},
		now,
	).StringSlice()

	if err != nil && err != redis.Nil {
		return nil, err
	}

	return result, nil
}

// HandleCronRescheduleFailure moves a failed cron task to the dead letter queue
func HandleCronRescheduleFailure(ctx context.Context, taskData []byte) error {
	return client.GetRedisClient().ZAdd(ctx, DeadLetterKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: taskData,
	}).Err()
}

// GetTask retrieves a task from the ready queue with blocking
func GetTask(ctx context.Context, timeout time.Duration) (string, error) {
	redisCli := client.GetRedisClient()
	result, err := redisCli.BRPop(ctx, timeout, ReadyQueueKey).Result()
	if err != nil {
		return "", err
	}

	return result[1], nil
}

// HandleFailedTask moves a failed task to the dead letter queue
func HandleFailedTask(ctx context.Context, taskData []byte, backoffSec int) error {
	deadLetterTime := time.Now().Add(time.Duration(backoffSec) * time.Second).Unix()
	redisCli := client.GetRedisClient()
	return redisCli.ZAdd(ctx, DeadLetterKey, redis.Z{
		Score:  float64(deadLetterTime),
		Member: taskData,
	}).Err()
}

// AcquireConcurrencyLock attempts to acquire a lock for task execution
func AcquireConcurrencyLock(ctx context.Context) bool {
	redisCli := client.GetRedisClient()
	currentCount, _ := redisCli.Get(ctx, ConcurrencyLockKey).Int64()
	if currentCount >= MaxConcurrency {
		return false
	}
	return redisCli.Incr(ctx, ConcurrencyLockKey).Err() == nil
}

// ReleaseConcurrencyLock releases a lock after task execution
func ReleaseConcurrencyLock(ctx context.Context) {
	redisCli := client.GetRedisClient()
	if err := redisCli.Decr(ctx, ConcurrencyLockKey).Err(); err != nil {
		logrus.Warnf("error releasing concurrency lock: %v", err)
	}
}

// InitConcurrencyLock initializes the concurrency lock counter
func InitConcurrencyLock(ctx context.Context) error {
	redisCli := client.GetRedisClient()
	return redisCli.Set(ctx, ConcurrencyLockKey, 0, 0).Err()
}

// GetTaskQueue retrieves the queue a task is in
func GetTaskQueue(ctx context.Context, taskID string) (string, error) {
	return client.GetRedisClient().HGet(ctx, TaskIndexKey, taskID).Result()
}

// RemoveFromList removes a task from a Redis list using Lua script
func RemoveFromList(ctx context.Context, key, taskID string) (bool, error) {
	cli := client.GetRedisClient()
	// Efficient list removal Lua script
	removeFromListScript := redis.NewScript(`
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

// RemoveFromZSet removes a task from a Redis sorted set
func RemoveFromZSet(ctx context.Context, key, taskID string) bool {
	cli := client.GetRedisClient()
	members, err := cli.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	}).Result()
	if err != nil {
		return false
	}

	for _, member := range members {
		var t dto.UnifiedTask
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

// DeleteTaskIndex removes a task from the task index
func DeleteTaskIndex(ctx context.Context, taskID string) error {
	return client.GetRedisClient().HDel(ctx, TaskIndexKey, taskID).Err()
}

// UpdateTaskStatus updates the task status in the database
func UpdateTaskStatus(ctx context.Context, taskID, status string) error {
	return database.DB.WithContext(ctx).Model(&database.Task{}).
		Where("id = ?", taskID).
		Update("status", status).Error
}

type EventConf struct {
	CallerLevel int
}

type EventConfOption func(*EventConf)

func WithCallerLevel(level int) func(*EventConf) {
	return func(c *EventConf) {
		c.CallerLevel = level
	}
}

func PublishEvent(ctx context.Context, stream string, event dto.StreamEvent, opts ...EventConfOption) {
	conf := &EventConf{
		CallerLevel: 2,
	}
	for _, opt := range opts {
		opt(conf)
	}

	file, line, fn := utils.GetCallerInfo(conf.CallerLevel)
	event.FileName = file
	event.Line = line
	event.FnName = fn

	_, err := client.GetRedisClient().XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: 10000,
		Approx: true,
		ID:     "*",
		Values: event.ToRedisStream(),
	}).Result()
	if err != nil {
		logrus.Errorf("Failed to publish event to Redis stream %s: %v", stream, err)
	}
}

func ReadStreamEvents(ctx context.Context, stream string, lastID string, count int64, block time.Duration) ([]redis.XStream, error) {
	if lastID == "" {
		lastID = "0"
	}

	return client.GetRedisClient().XRead(ctx, &redis.XReadArgs{
		Streams: []string{stream, lastID},
		Count:   count,
		Block:   block,
	}).Result()
}

type StreamProcessor struct {
	hasIssues      bool
	isCompleted    bool
	detectorTaskID string
	algorithms     map[string]struct{}
	finishedCount  int
}

func NewStreamProcessor(algorithmItems []dto.AlgorithmItem) *StreamProcessor {
	algorithms := make(map[string]struct{}, len(algorithmItems))
	for _, item := range algorithmItems {
		algorithms[item.Name] = struct{}{}
	}

	return &StreamProcessor{
		hasIssues:      false,
		isCompleted:    false,
		detectorTaskID: "",
		algorithms:     algorithms,
		finishedCount:  0,
	}
}

func (sp *StreamProcessor) IsCompleted() bool {
	return sp.isCompleted
}

func (sp *StreamProcessor) ProcessMessageForSSE(msg redis.XMessage) (string, string, error) {
	streamEvent, err := parseEventFromValues(msg.Values)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse stream message value: %v", err)
	}

	sseMessage, err := streamEvent.ToSSE()
	if err != nil {
		return "", "", fmt.Errorf("failed to parse streamEvent to sse message: %v", err)
	}

	switch streamEvent.EventName {
	case consts.EventImageBuildSucceed:
		sp.isCompleted = true

	case consts.EventDatasetNoAnomaly, consts.EventDatasetNoConclusionFile:
		sp.detectorTaskID = streamEvent.TaskID
		sp.hasIssues = false

	case consts.EventDatasetResultCollection:
		sp.detectorTaskID = streamEvent.TaskID
		sp.hasIssues = true

	case consts.EventAlgoRunSucceed, consts.EventAlgoRunFailed:
		payloadStr, ok := streamEvent.Payload.(string)
		if !ok {
			return "", "", fmt.Errorf("invalid payload type for task status update event: %T", streamEvent.Payload)
		}

		var payload dto.ExecutionOptions
		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			return "", "", fmt.Errorf("failed to unmarshal payload: %v", err)
		}

		if payload.Algorithm.Name != config.GetString("algo.detector") {
			if len(sp.algorithms) != 0 {
				if _, exists := sp.algorithms[payload.Algorithm.Name]; exists {
					sp.finishedCount++
				}
			} else {
				sp.finishedCount++
			}
		}

	case consts.EventTaskStatusUpdate:
		payloadStr, ok := streamEvent.Payload.(string)
		if !ok {
			return "", "", fmt.Errorf("invalid payload type for task status update event: %T", streamEvent.Payload)
		}

		var payload dto.InfoPayloadTemplate
		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			return "", "", fmt.Errorf("failed to unmarshal payload: %v", err)
		}

		switch payload.Status {
		case consts.TaskStatusError:
			sp.isCompleted = true
		case consts.TaskStatusCompleted:
			if streamEvent.TaskType == consts.TaskTypeCollectResult {
				if sp.detectorTaskID != "" && streamEvent.TaskID == sp.detectorTaskID {
					sp.isCompleted = !sp.hasIssues || len(sp.algorithms) == 0
				} else {
					sp.isCompleted = len(sp.algorithms) == 0 || sp.finishedCount == len(sp.algorithms)
				}
			}
		}
	}

	return msg.ID, sseMessage, nil
}

// ParseEventFromValues 从 Redis Stream 消息解析事件
func parseEventFromValues(values map[string]any) (*dto.StreamEvent, error) {
	message := "missing or invalid key %s in redis stream message values"

	taskID, ok := values[consts.RdbEventTaskID].(string)
	if !ok || taskID == "" {
		return nil, fmt.Errorf(message, consts.RdbEventTaskID)
	}
	event := &dto.StreamEvent{
		TaskID: taskID,
	}

	if _, exists := values[consts.RdbEventTaskType]; exists {
		taskType, ok := values[consts.RdbEventTaskType].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventTaskType)
		}
		event.TaskType = consts.TaskType(taskType)
	}

	if _, exists := values[consts.RdbEventFn]; exists {
		fnName, ok := values[consts.RdbEventFn].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventFn)
		}
		event.FnName = fnName
	}

	if _, exists := values[consts.RdbEventPayload]; exists {
		payload, ok := values[consts.RdbEventPayload]
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventPayload)
		}

		event.Payload = payload
	}

	if _, exists := values[consts.RdbEventFileName]; exists {
		fileName, ok := values[consts.RdbEventFileName].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventTaskID)
		}

		event.FileName = fileName
	}

	if _, exists := values[consts.RdbEventLine]; exists {
		lineInt64, ok := values[consts.RdbEventLine].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventLine)
		}

		line, err := strconv.Atoi(lineInt64)
		if err != nil {
			return nil, fmt.Errorf("invalid line number: %w", err)
		}
		event.Line = line
	}

	if _, exists := values[consts.RdbEventName]; exists {
		eventName, ok := values[consts.RdbEventName].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventName)
		}
		event.EventName = consts.EventType(eventName)
	}

	return event, nil
}

func ListTasks(params *dto.ListTasksReq) (int64, []database.Task, error) {
	opts, err := params.TimeRangeQuery.Convert()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to convert time range query: %v", err)
	}

	builder := func(db *gorm.DB) *gorm.DB {
		query := db

		if params.TaskID != "" {
			query = query.Where("id = ?", params.TaskID)
		}

		if params.TaskType != "" {
			query = query.Where("type = ?", params.TaskType)
		}

		if params.Immediate != nil {
			query = query.Where("immediate = ?", *params.Immediate)
		}

		if params.Status != "" {
			query = query.Where("status = ?", params.Status)
		}

		if params.TraceID != "" {
			query = query.Where("trace_id = ?", params.TraceID)
		}

		if params.GroupID != "" {
			query = query.Where("group_id = ?", params.GroupID)
		}

		query = opts.AddTimeFilter(query, "created_at")
		return query
	}

	genericQueryParams := &genericQueryParams{
		builder:   builder,
		sortField: fmt.Sprintf("%s %s", params.SortField, params.SortOrder),
		limit:     params.Limit,
	}
	return genericQueryWithBuilder[database.Task](genericQueryParams)
}

// GetTaskStatistics returns statistics about tasks
func GetTaskStatistics() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total tasks
	var total int64
	if err := database.DB.Model(&database.Task{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total tasks: %v", err)
	}
	stats["total"] = total

	// Tasks by status
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	var statusCounts []StatusCount
	err := database.DB.Model(&database.Task{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count tasks by status: %v", err)
	}

	for _, sc := range statusCounts {
		stats[sc.Status] = sc.Count
	}

	// Tasks by type
	type TypeCount struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	var typeCounts []TypeCount
	err = database.DB.Model(&database.Task{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Find(&typeCounts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count tasks by type: %v", err)
	}

	for _, tc := range typeCounts {
		stats[tc.Type+"_tasks"] = tc.Count
	}

	return stats, nil
}

// GetRecentTaskActivity returns task activity for the last N days
func GetRecentTaskActivity(days int) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Last N days activity
	startDate := time.Now().AddDate(0, 0, -days)
	var recentCount int64
	if err := database.DB.Model(&database.Task{}).Where("created_at >= ?", startDate).Count(&recentCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count recent tasks: %v", err)
	}
	stats[fmt.Sprintf("last_%d_days", days)] = recentCount

	// Today's tasks
	today := time.Now().Truncate(24 * time.Hour)
	var todayCount int64
	if err := database.DB.Model(&database.Task{}).Where("created_at >= ?", today).Count(&todayCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count today's tasks: %v", err)
	}
	stats["today"] = todayCount

	return stats, nil
}
