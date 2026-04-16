package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aegis/client"
	"aegis/dto"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Redis key constants for task queues and indexes.
const (
	DelayedQueueKey    = "task:delayed"
	ReadyQueueKey      = "task:ready"
	DeadLetterKey      = "task:dead"
	TaskIndexKey       = "task:index"
	ConcurrencyLockKey = "task:concurrency_lock"
	LastBatchInfoKey   = "last_batch_info"
	MaxConcurrency     = 20
)

type RedisClient interface {
	LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	BRPop(ctx context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd
	ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Decr(ctx context.Context, key string) *redis.IntCmd
	HGet(ctx context.Context, key, field string) *redis.StringCmd
	ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd
	ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd
}

var (
	getRedisClient = func() RedisClient {
		return client.GetRedisClient()
	}
	getRedisScriptClient = func() *redis.Client {
		return client.GetRedisClient()
	}
	getRedisListRange               = client.GetRedisListRange
	getRedisZRangeByScoreWithScores = client.GetRedisZRangeByScoreWithScores
	currentTime                     = time.Now
	runProcessDelayedTasksScript    = func(ctx context.Context, redisCli *redis.Client, now int64) ([]string, error) {
		delayedTaskScript := redis.NewScript(`
		local tasks = redis.call('ZRANGEBYSCORE', KEYS[1], 0, ARGV[1])
		if #tasks > 0 then
			redis.call('ZREMRANGEBYSCORE', KEYS[1], 0, ARGV[1])
			redis.call('LPUSH', KEYS[2], unpack(tasks))
			for _, task in ipairs(tasks) do
				local t = cjson.decode(task)
				redis.call('HSET', KEYS[3], t.task_id, KEYS[2])
			end
		end
		return tasks
	`)

		return delayedTaskScript.Run(ctx, redisCli,
			[]string{DelayedQueueKey, ReadyQueueKey, TaskIndexKey},
			now,
		).StringSlice()
	}
	runRemoveFromListScript = func(ctx context.Context, redisCli *redis.Client, key, taskID string) (int, error) {
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

		return removeFromListScript.Run(ctx, redisCli, []string{key}, taskID).Int()
	}
)

func SubmitImmediateTask(ctx context.Context, taskData []byte, taskID string) error {
	redisCli := getRedisClient()
	if err := redisCli.LPush(ctx, ReadyQueueKey, taskData).Err(); err != nil {
		return err
	}

	return redisCli.HSet(ctx, TaskIndexKey, taskID, ReadyQueueKey).Err()
}

func GetTask(ctx context.Context, timeout time.Duration) (string, error) {
	redisCli := getRedisClient()
	result, err := redisCli.BRPop(ctx, timeout, ReadyQueueKey).Result()
	if err != nil {
		return "", err
	}
	if len(result) < 2 {
		return "", fmt.Errorf("invalid BRPOP result")
	}

	return result[1], nil
}

func HandleFailedTask(ctx context.Context, taskData []byte, backoffSec int) error {
	deadLetterTime := currentTime().Add(time.Duration(backoffSec) * time.Second).Unix()
	redisCli := getRedisClient()
	return redisCli.ZAdd(ctx, DeadLetterKey, redis.Z{
		Score:  float64(deadLetterTime),
		Member: taskData,
	}).Err()
}

func SubmitDelayedTask(ctx context.Context, taskData []byte, taskID string, executeTime int64) error {
	redisCli := getRedisClient()
	if err := redisCli.ZAdd(ctx, DelayedQueueKey, redis.Z{
		Score:  float64(executeTime),
		Member: taskData,
	}).Err(); err != nil {
		return err
	}

	return redisCli.HSet(ctx, TaskIndexKey, taskID, DelayedQueueKey).Err()
}

func ProcessDelayedTasks(ctx context.Context) ([]string, error) {
	result, err := runProcessDelayedTasksScript(ctx, getRedisScriptClient(), currentTime().Unix())
	if err != nil && err != redis.Nil {
		return nil, err
	}

	return result, nil
}

func HandleCronRescheduleFailure(ctx context.Context, taskData []byte) error {
	return getRedisClient().ZAdd(ctx, DeadLetterKey, redis.Z{
		Score:  float64(currentTime().Unix()),
		Member: taskData,
	}).Err()
}

func AcquireConcurrencyLock(ctx context.Context) bool {
	redisCli := getRedisClient()
	currentCount, _ := redisCli.Get(ctx, ConcurrencyLockKey).Int64()
	if currentCount >= MaxConcurrency {
		return false
	}
	return redisCli.Incr(ctx, ConcurrencyLockKey).Err() == nil
}

func InitConcurrencyLock(ctx context.Context) error {
	return getRedisClient().Set(ctx, ConcurrencyLockKey, 0, 0).Err()
}

func ReleaseConcurrencyLock(ctx context.Context) {
	if err := getRedisClient().Decr(ctx, ConcurrencyLockKey).Err(); err != nil {
		logrus.Warnf("error releasing concurrency lock: %v", err)
	}
}

func GetTaskQueue(ctx context.Context, taskID string) (string, error) {
	return getRedisClient().HGet(ctx, TaskIndexKey, taskID).Result()
}

func ListDelayedTasks(ctx context.Context, limit int64) ([]string, error) {
	delayedTasksWithScore, err := getRedisZRangeByScoreWithScores(ctx, DelayedQueueKey, limit)
	if err != nil {
		return nil, err
	}

	taskDatas := make([]string, 0, len(delayedTasksWithScore))
	for _, z := range delayedTasksWithScore {
		taskData, ok := z.Member.(string)
		if !ok {
			return nil, fmt.Errorf("invalid delayed task data")
		}
		taskDatas = append(taskDatas, taskData)
	}

	return taskDatas, nil
}

func ListReadyTasks(ctx context.Context) ([]string, error) {
	return getRedisListRange(ctx, ReadyQueueKey)
}

func RemoveFromList(ctx context.Context, key, taskID string) (bool, error) {
	result, err := runRemoveFromListScript(ctx, getRedisScriptClient(), key, taskID)
	if err != nil {
		return false, fmt.Errorf("failed to remove from list: %w", err)
	}

	return result > 0, nil
}

func RemoveFromZSet(ctx context.Context, key, taskID string) bool {
	cli := getRedisClient()
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

func DeleteTaskIndex(ctx context.Context, taskID string) error {
	return getRedisClient().HDel(ctx, TaskIndexKey, taskID).Err()
}
