// Package ratelimiter provides admin/operator APIs over the token-bucket
// rate limiters that back the restart_pedestal, build_container and
// algo_execution concurrency gates. Operators can inspect bucket state,
// reset a bucket, or garbage-collect tokens still held by terminal-state
// tasks (OperationsPAI/aegis#21).
package ratelimiter

import (
	"context"
	"fmt"
	"strings"

	"aegis/consts"
	redisinfra "aegis/infra/redis"
	"aegis/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const tokenBucketKeyPrefix = "token_bucket:"

// Service exposes rate-limiter admin operations. It is deliberately thin —
// all the state lives in Redis/MySQL and we just wrap scans + set ops.
type Service struct {
	redis *redisinfra.Gateway
	db    *gorm.DB
}

func NewService(redis *redisinfra.Gateway, db *gorm.DB) *Service {
	return &Service{redis: redis, db: db}
}

func knownBuckets() map[string]int {
	return map[string]int{
		consts.RestartPedestalTokenBucket: consts.MaxConcurrentRestartPedestal,
		consts.BuildContainerTokenBucket:  consts.MaxConcurrentBuildContainer,
		consts.AlgoExecutionTokenBucket:   consts.MaxConcurrentAlgoExecution,
	}
}

// isTerminalState mirrors the codebase's TaskCompleted (3), TaskError (-1),
// TaskCancelled (-2) states.
func isTerminalState(state consts.TaskState) bool {
	return state == consts.TaskCompleted || state == consts.TaskError || state == consts.TaskCancelled
}

// List returns each token_bucket:* bucket with its holders.
func (s *Service) List(ctx context.Context) (*RateLimiterListResp, error) {
	bucketCaps := knownBuckets()

	extra, err := s.redis.ScanKeys(ctx, tokenBucketKeyPrefix+"*")
	if err != nil {
		return nil, fmt.Errorf("scan token buckets: %w", err)
	}
	for _, key := range extra {
		if _, ok := bucketCaps[key]; !ok {
			bucketCaps[key] = 0
		}
	}

	items := make([]RateLimiterItem, 0, len(bucketCaps))
	for key, capacity := range bucketCaps {
		holders, err := s.redis.SetMembers(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("smembers %s: %w", key, err)
		}
		holderItems := make([]RateLimiterHolder, 0, len(holders))
		for _, taskID := range holders {
			state, found, err := s.lookupTaskState(ctx, taskID)
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
			holderItems = append(holderItems, RateLimiterHolder{
				TaskID: taskID, TaskState: stateName, IsTerminal: terminal,
			})
		}
		items = append(items, RateLimiterItem{
			Bucket:   strings.TrimPrefix(key, tokenBucketKeyPrefix),
			Key:      key,
			Capacity: capacity,
			Held:     len(holders),
			Holders:  holderItems,
		})
	}
	return &RateLimiterListResp{Items: items}, nil
}

// Reset deletes the given bucket key from Redis.
func (s *Service) Reset(ctx context.Context, bucket string) error {
	key := resolveBucketKey(bucket)
	if _, ok := knownBuckets()[key]; !ok {
		return fmt.Errorf("%w: unknown bucket %q", consts.ErrBadRequest, bucket)
	}
	n, err := s.redis.DeleteKey(ctx, key)
	if err != nil {
		return fmt.Errorf("del %s: %w", key, err)
	}
	if n == 0 {
		return fmt.Errorf("%w: bucket %q not present in redis", consts.ErrNotFound, bucket)
	}
	logrus.WithField("bucket", key).Warn("rate-limiter bucket reset")
	return nil
}

// GC releases tokens held by terminal-state (or unknown) tasks across all
// known buckets. Returns (released, touchedBuckets, err).
func (s *Service) GC(ctx context.Context) (int, int, error) {
	return gcWith(ctx, s.redis, s.db, knownBuckets())
}

// gcWith is the testable core.
func gcWith(ctx context.Context, r *redisinfra.Gateway, db *gorm.DB, buckets map[string]int) (int, int, error) {
	var released, touched int
	for key := range buckets {
		holders, err := r.SetMembers(ctx, key)
		if err != nil {
			return released, touched, fmt.Errorf("smembers %s: %w", key, err)
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
		n, rerr := r.SetRemove(ctx, key, toRelease...)
		if rerr != nil {
			return released, touched, fmt.Errorf("srem %s: %w", key, rerr)
		}
		if n > 0 {
			released += int(n)
			touched++
			logrus.WithFields(logrus.Fields{
				"bucket": key, "released": n, "holders": toRelease,
			}).Warn("rate-limiter gc: released leaked tokens")
		}
	}
	return released, touched, nil
}

func (s *Service) lookupTaskState(ctx context.Context, taskID string) (consts.TaskState, bool, error) {
	return lookupTaskStateWith(ctx, s.db, taskID)
}

func lookupTaskStateWith(ctx context.Context, db *gorm.DB, taskID string) (consts.TaskState, bool, error) {
	var task model.Task
	err := db.WithContext(ctx).Select("state").Where("id = ?", taskID).First(&task).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
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
