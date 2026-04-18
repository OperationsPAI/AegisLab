package producer

import (
	"context"
	"fmt"
	"strings"

	"aegis/client"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const tokenBucketKeyPrefix = "token_bucket:"

func knownBuckets() map[string]int {
	return map[string]int{
		consts.RestartPedestalTokenBucket: consts.MaxConcurrentRestartPedestal,
		consts.BuildContainerTokenBucket:  consts.MaxConcurrentBuildContainer,
		consts.AlgoExecutionTokenBucket:   consts.MaxConcurrentAlgoExecution,
	}
}

// isTerminalState mirrors service/producer/task.go:isTaskTerminal.
// Issue #21 spells these "Success / Failed / -1"; codebase uses
// TaskCompleted (3), TaskError (-1), TaskCancelled (-2).
func isTerminalState(state consts.TaskState) bool {
	return state == consts.TaskCompleted || state == consts.TaskError || state == consts.TaskCancelled
}

// ListRateLimiters returns each token_bucket:* bucket with its holders.
func ListRateLimiters(ctx context.Context) (*dto.RateLimiterListResp, error) {
	redisCli := client.GetRedisClient()
	bucketCaps := knownBuckets()

	iter := redisCli.Scan(ctx, 0, tokenBucketKeyPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		if _, ok := bucketCaps[key]; !ok {
			bucketCaps[key] = 0
		}
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("scan token buckets: %w", err)
	}

	items := make([]dto.RateLimiterItem, 0, len(bucketCaps))
	for key, capacity := range bucketCaps {
		holders, err := redisCli.SMembers(ctx, key).Result()
		if err != nil && err != redis.Nil {
			return nil, fmt.Errorf("smembers %s: %w", key, err)
		}
		holderItems := make([]dto.RateLimiterHolder, 0, len(holders))
		for _, taskID := range holders {
			state, found, err := lookupTaskState(ctx, taskID)
			if err != nil {
				logrus.WithError(err).WithField("task_id", taskID).
					Warn("lookup task state for rate-limiter holder")
			}
			stateName := "Unknown"
			terminal := false
			if found {
				stateName = consts.GetTaskStateName(state)
				terminal = isTerminalState(state)
			} else {
				terminal = true
			}
			holderItems = append(holderItems, dto.RateLimiterHolder{
				TaskID: taskID, TaskState: stateName, IsTerminal: terminal,
			})
		}
		items = append(items, dto.RateLimiterItem{
			Bucket:   strings.TrimPrefix(key, tokenBucketKeyPrefix),
			Key:      key,
			Capacity: capacity,
			Held:     len(holders),
			Holders:  holderItems,
		})
	}
	return &dto.RateLimiterListResp{Items: items}, nil
}

// ResetRateLimiter deletes the given bucket key from Redis.
func ResetRateLimiter(ctx context.Context, bucket string) error {
	key := resolveBucketKey(bucket)
	if _, ok := knownBuckets()[key]; !ok {
		return fmt.Errorf("%w: unknown bucket %q", consts.ErrBadRequest, bucket)
	}
	redisCli := client.GetRedisClient()
	n, err := redisCli.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("del %s: %w", key, err)
	}
	if n == 0 {
		return fmt.Errorf("%w: bucket %q not present in redis", consts.ErrNotFound, bucket)
	}
	logrus.WithField("bucket", key).Warn("rate-limiter bucket reset")
	return nil
}

// GCRateLimiters releases tokens held by terminal-state tasks.
func GCRateLimiters(ctx context.Context) (released int, touchedBuckets int, err error) {
	return gcRateLimitersWith(ctx, client.GetRedisClient(), database.DB, knownBuckets())
}

// gcRateLimitersWith is the testable core.
func gcRateLimitersWith(ctx context.Context, redisCli *redis.Client, db *gorm.DB, buckets map[string]int) (released int, touchedBuckets int, err error) {
	for key := range buckets {
		holders, serr := redisCli.SMembers(ctx, key).Result()
		if serr != nil && serr != redis.Nil {
			return released, touchedBuckets, fmt.Errorf("smembers %s: %w", key, serr)
		}
		if len(holders) == 0 {
			continue
		}
		var toRelease []any
		for _, taskID := range holders {
			state, found, lerr := lookupTaskStateWith(ctx, db, taskID)
			if lerr != nil {
				logrus.WithError(lerr).WithField("task_id", taskID).
					Warn("gc: lookup task state failed; skipping")
				continue
			}
			if !found || isTerminalState(state) {
				toRelease = append(toRelease, taskID)
			}
		}
		if len(toRelease) == 0 {
			continue
		}
		n, rerr := redisCli.SRem(ctx, key, toRelease...).Result()
		if rerr != nil {
			return released, touchedBuckets, fmt.Errorf("srem %s: %w", key, rerr)
		}
		if n > 0 {
			released += int(n)
			touchedBuckets++
			logrus.WithFields(logrus.Fields{
				"bucket": key, "released": n, "holders": toRelease,
			}).Warn("rate-limiter gc: released leaked tokens")
		}
	}
	return released, touchedBuckets, nil
}

func lookupTaskState(ctx context.Context, taskID string) (consts.TaskState, bool, error) {
	return lookupTaskStateWith(ctx, database.DB, taskID)
}

func lookupTaskStateWith(ctx context.Context, db *gorm.DB, taskID string) (consts.TaskState, bool, error) {
	var task database.Task
	err := db.WithContext(ctx).Select("state").Where("id = ?", taskID).First(&task).Error
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return 0, false, nil
		}
		return 0, false, err
	}
	return task.State, true, nil
}

func resolveBucketKey(name string) string {
	if strings.HasPrefix(name, tokenBucketKeyPrefix) {
		return name
	}
	return tokenBucketKeyPrefix + name
}
