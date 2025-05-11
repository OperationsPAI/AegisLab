package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/utils"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
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

	res, err := client.GetRedisClient().XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: 10000,
		Approx: true,
		ID:     "*",
		Values: event.ToRedisStream(),
	}).Result()
	if err != nil {
		logrus.Errorf("Failed to publish event to Redis stream %s: %v", stream, err)
	}
	logrus.Debugf("Published event to Redis stream %s: %s", stream, res)
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

// CreateConsumerGroup 创建 Redis Stream 消费者组
func CreateConsumerGroup(ctx context.Context, stream, group, startID string) error {
	err := client.GetRedisClient().XGroupCreate(ctx, stream, group, startID).Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}
	return nil
}

// ConsumeStreamEvents 使用消费者组消费 Redis Stream 事件
func ConsumeStreamEvents(ctx context.Context, stream, group, consumer string, count int64, block time.Duration) ([]redis.XStream, error) {
	return client.GetRedisClient().XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    block,
	}).Result()
}

func AcknowledgeMessage(ctx context.Context, stream, group, id string) error {
	return client.GetRedisClient().XAck(ctx, stream, group, id).Err()
}

// ParseEventFromValues 从 Redis Stream 消息解析事件
func ParseEventFromValues(values map[string]any) (*dto.StreamEvent, error) {
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

// ProcessStreamMessagesForSSE processes Redis stream messages and prepares them for SSE events
// Returns the last message ID and a slice of prepared SSE messages
func ProcessStreamMessagesForSSE(messages []redis.XStream) (string, []dto.SSEMessageData, error) {
	var lastID string
	var sseMessages []dto.SSEMessageData

	for _, stream := range messages {
		for _, msg := range stream.Messages {
			lastID = msg.ID

			streamEvent, err := ParseEventFromValues(msg.Values)
			if err != nil {
				return "", nil, fmt.Errorf("failed to parse stream message value: %v", err)
			}

			sseMessage, err := streamEvent.ToSSE()
			if err != nil {
				return "", nil, fmt.Errorf("failed to parse streamEvent to sse message: %v", err)
			}

			// Check if this is a completion event
			isCompleted := false
			if streamEvent.TaskType == consts.TaskTypeCollectResult && streamEvent.EventName == consts.EventTaskStatusUpdate {
				if payloadStr, ok := streamEvent.Payload.(string); ok {
					var payload dto.InfoPayloadTemplate
					if err := json.Unmarshal([]byte(payloadStr), &payload); err == nil {
						if payload.Status == consts.TaskStatusCompleted || payload.Status == consts.TaskStatusError || payload.Status == consts.TaskStatusCanceled {
							isCompleted = true
						}
					}
				}
			}

			sseMessages = append(sseMessages, dto.SSEMessageData{
				ID:          msg.ID,
				Data:        sseMessage,
				IsCompleted: isCompleted,
			})
		}
	}

	return lastID, sseMessages, nil
}

// TaskFilter defines filtering options for task queries
type TaskFilter struct {
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

// FindTasks searches for tasks with pagination and filtering
func FindTasks(filter TaskFilter, pageNum, pageSize int, sortField string) (int64, []database.Task, error) {
	// Build the WHERE condition dynamically
	whereConditions := []string{}
	whereArgs := []interface{}{}

	if filter.TaskID != nil {
		whereConditions = append(whereConditions, "id = ?")
		whereArgs = append(whereArgs, *filter.TaskID)
	}

	if filter.TaskType != nil {
		whereConditions = append(whereConditions, "type = ?")
		whereArgs = append(whereArgs, *filter.TaskType)
	}

	if filter.Immediate != nil {
		whereConditions = append(whereConditions, "immediate = ?")
		whereArgs = append(whereArgs, *filter.Immediate)
	}

	if filter.ExecuteTimeGT != nil {
		whereConditions = append(whereConditions, "execute_time > ?")
		whereArgs = append(whereArgs, *filter.ExecuteTimeGT)
	}

	if filter.ExecuteTimeLT != nil {
		whereConditions = append(whereConditions, "execute_time < ?")
		whereArgs = append(whereArgs, *filter.ExecuteTimeLT)
	}

	if filter.ExecuteTimeGTE != nil {
		whereConditions = append(whereConditions, "execute_time >= ?")
		whereArgs = append(whereArgs, *filter.ExecuteTimeGTE)
	}

	if filter.ExecuteTimeLTE != nil {
		whereConditions = append(whereConditions, "execute_time <= ?")
		whereArgs = append(whereArgs, *filter.ExecuteTimeLTE)
	}

	if filter.Status != nil {
		whereConditions = append(whereConditions, "status = ?")
		whereArgs = append(whereArgs, *filter.Status)
	}

	if filter.TraceID != nil {
		whereConditions = append(whereConditions, "trace_id = ?")
		whereArgs = append(whereArgs, *filter.TraceID)
	}

	if filter.GroupID != nil {
		whereConditions = append(whereConditions, "group_id = ?")
		whereArgs = append(whereArgs, *filter.GroupID)
	}

	// Combine all conditions with AND
	whereClause := "1=1" // Default condition that's always true
	if len(whereConditions) > 0 {
		whereClause = strings.Join(whereConditions, " AND ")
	}

	// If sortField is empty, default to created_at desc
	if sortField == "" {
		sortField = "created_at desc"
	}

	// Use the generic pagination function
	return paginateQuery[database.Task](
		whereClause,
		whereArgs,
		sortField,
		pageNum,
		pageSize,
		nil,
	)
}
