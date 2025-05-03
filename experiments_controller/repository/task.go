package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
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
