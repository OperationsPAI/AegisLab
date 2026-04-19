package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aegis/dto"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	DelayedQueueKey    = "task:delayed"
	ReadyQueueKey      = "task:ready"
	DeadLetterKey      = "task:dead"
	TaskIndexKey       = "task:index"
	ConcurrencyLockKey = "task:concurrency_lock"
	LastBatchInfoKey   = "last_batch_info"
	MaxConcurrency     = 20
)

type TaskQueueStats struct {
	ReadyCount       int64
	DelayedCount     int64
	DeadCount        int64
	IndexedCount     int64
	ConcurrencyCount int64
}

func (g *Gateway) SubmitImmediateTask(ctx context.Context, taskData []byte, taskID string) error {
	redisCli := g.clientOrInit()
	if err := redisCli.LPush(ctx, ReadyQueueKey, taskData).Err(); err != nil {
		return err
	}
	return redisCli.HSet(ctx, TaskIndexKey, taskID, ReadyQueueKey).Err()
}

func (g *Gateway) GetTask(ctx context.Context, timeout time.Duration) (string, error) {
	redisCli := g.clientOrInit()
	result, err := redisCli.BRPop(ctx, timeout, ReadyQueueKey).Result()
	if err != nil {
		return "", err
	}
	return result[1], nil
}

func (g *Gateway) HandleFailedTask(ctx context.Context, taskData []byte, backoffSec int) error {
	deadLetterTime := time.Now().Add(time.Duration(backoffSec) * time.Second).Unix()
	redisCli := g.clientOrInit()
	if err := redisCli.ZAdd(ctx, DeadLetterKey, redis.Z{
		Score:  float64(deadLetterTime),
		Member: taskData,
	}).Err(); err != nil {
		return err
	}

	var task dto.UnifiedTask
	if err := json.Unmarshal(taskData, &task); err == nil && task.TaskID != "" {
		return redisCli.HSet(ctx, TaskIndexKey, task.TaskID, DeadLetterKey).Err()
	}
	return nil
}

func (g *Gateway) SubmitDelayedTask(ctx context.Context, taskData []byte, taskID string, executeTime int64) error {
	redisCli := g.clientOrInit()
	if err := redisCli.ZAdd(ctx, DelayedQueueKey, redis.Z{
		Score:  float64(executeTime),
		Member: taskData,
	}).Err(); err != nil {
		return err
	}
	return redisCli.HSet(ctx, TaskIndexKey, taskID, DelayedQueueKey).Err()
}

func (g *Gateway) ProcessDelayedTasks(ctx context.Context) ([]string, error) {
	redisCli := g.clientOrInit()
	now := time.Now().Unix()

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

	result, err := delayedTaskScript.Run(ctx, redisCli,
		[]string{DelayedQueueKey, ReadyQueueKey, TaskIndexKey},
		now,
	).StringSlice()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	return result, nil
}

func (g *Gateway) HandleCronRescheduleFailure(ctx context.Context, taskData []byte) error {
	redisCli := g.clientOrInit()
	if err := redisCli.ZAdd(ctx, DeadLetterKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: taskData,
	}).Err(); err != nil {
		return err
	}

	var task dto.UnifiedTask
	if err := json.Unmarshal(taskData, &task); err == nil && task.TaskID != "" {
		return redisCli.HSet(ctx, TaskIndexKey, task.TaskID, DeadLetterKey).Err()
	}
	return nil
}

func (g *Gateway) AcquireConcurrencyLock(ctx context.Context) bool {
	redisCli := g.clientOrInit()
	currentCount, _ := redisCli.Get(ctx, ConcurrencyLockKey).Int64()
	if currentCount >= MaxConcurrency {
		return false
	}
	return redisCli.Incr(ctx, ConcurrencyLockKey).Err() == nil
}

func (g *Gateway) ReleaseConcurrencyLock(ctx context.Context) {
	if err := g.clientOrInit().Decr(ctx, ConcurrencyLockKey).Err(); err != nil {
		logrus.Warnf("error releasing concurrency lock: %v", err)
	}
}

func (g *Gateway) GetTaskQueue(ctx context.Context, taskID string) (string, error) {
	return g.clientOrInit().HGet(ctx, TaskIndexKey, taskID).Result()
}

func (g *Gateway) ListDelayedTasks(ctx context.Context, limit int64) ([]string, error) {
	delayedTasksWithScore, err := g.ZRangeByScoreWithScores(ctx, DelayedQueueKey, limit)
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

func (g *Gateway) ListDeadLetterTasks(ctx context.Context, limit int64) ([]string, error) {
	deadTasksWithScore, err := g.ZRangeByScoreWithScores(ctx, DeadLetterKey, limit)
	if err != nil {
		return nil, err
	}

	taskDatas := make([]string, 0, len(deadTasksWithScore))
	for _, z := range deadTasksWithScore {
		taskData, ok := z.Member.(string)
		if !ok {
			return nil, fmt.Errorf("invalid dead letter task data")
		}
		taskDatas = append(taskDatas, taskData)
	}

	return taskDatas, nil
}

func (g *Gateway) ListReadyTasks(ctx context.Context) ([]string, error) {
	return g.ListRange(ctx, ReadyQueueKey)
}

func (g *Gateway) RemoveFromList(ctx context.Context, key, taskID string) (bool, error) {
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

	result, err := removeFromListScript.Run(ctx, g.clientOrInit(), []string{key}, taskID).Int()
	if err != nil {
		return false, fmt.Errorf("failed to remove from list: %w", err)
	}
	return result > 0, nil
}

func (g *Gateway) RemoveFromZSet(ctx context.Context, key, taskID string) bool {
	cli := g.clientOrInit()
	members, err := cli.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	}).Result()
	if err != nil {
		return false
	}

	for _, member := range members {
		var task dto.UnifiedTask
		if json.Unmarshal([]byte(member), &task) == nil && task.TaskID == taskID {
			if err := cli.ZRem(ctx, key, member).Err(); err != nil {
				logrus.Warnf("failed to remove from ZSet: %v", err)
				return false
			}
			return true
		}
	}

	return false
}

func (g *Gateway) DeleteTaskIndex(ctx context.Context, taskID string) error {
	return g.clientOrInit().HDel(ctx, TaskIndexKey, taskID).Err()
}

func (g *Gateway) GetTaskQueueStats(ctx context.Context) (TaskQueueStats, error) {
	readyCount, err := g.ListLength(ctx, ReadyQueueKey)
	if err != nil {
		return TaskQueueStats{}, err
	}

	delayedCount, err := g.SortedSetCard(ctx, DelayedQueueKey)
	if err != nil {
		return TaskQueueStats{}, err
	}

	deadCount, err := g.SortedSetCard(ctx, DeadLetterKey)
	if err != nil {
		return TaskQueueStats{}, err
	}

	indexedCount, err := g.HashLength(ctx, TaskIndexKey)
	if err != nil {
		return TaskQueueStats{}, err
	}

	concurrencyCount, err := g.GetInt64(ctx, ConcurrencyLockKey)
	if err != nil {
		return TaskQueueStats{}, err
	}

	return TaskQueueStats{
		ReadyCount:       readyCount,
		DelayedCount:     delayedCount,
		DeadCount:        deadCount,
		IndexedCount:     indexedCount,
		ConcurrencyCount: concurrencyCount,
	}, nil
}
