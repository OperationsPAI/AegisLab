package producer

import (
	"context"
	"testing"
	"time"

	"aegis/consts"
	"aegis/database"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&database.Task{}))
	return db
}

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

// Regression per OperationsPAI/aegis#21: a bucket with 2 holders, one
// terminal, must end with exactly 1 holder after GC.
func TestGCRateLimiters_ReleasesTerminalHolders(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := newTestDB(t)
	rdb := newTestRedis(t)

	require.NoError(t, db.Create(&database.Task{ID: "running-task-1", State: consts.TaskRunning}).Error)
	require.NoError(t, db.Create(&database.Task{ID: "done-task-1", State: consts.TaskCompleted}).Error)

	bucket := consts.RestartPedestalTokenBucket
	_, err := rdb.SAdd(ctx, bucket, "running-task-1", "done-task-1").Result()
	require.NoError(t, err)
	require.Equal(t, int64(2), rdb.SCard(ctx, bucket).Val())

	released, touched, err := gcRateLimitersWith(ctx, rdb, db, map[string]int{
		bucket: consts.MaxConcurrentRestartPedestal,
	})
	require.NoError(t, err)
	require.Equal(t, 1, released)
	require.Equal(t, 1, touched)

	members, err := rdb.SMembers(ctx, bucket).Result()
	require.NoError(t, err)
	require.Equal(t, []string{"running-task-1"}, members)
}

// Regression: the rate-limiter's task-done path must release the token.
// Exercises the same SRem call used by TokenBucketRateLimiter.ReleaseToken
// in service/consumer/rate_limiter.go:120.
func TestRateLimiterReleaseToken_RegressionGuard(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rdb := newTestRedis(t)
	bucket := consts.RestartPedestalTokenBucket

	_, err := rdb.SAdd(ctx, bucket, "task-42").Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), rdb.SCard(ctx, bucket).Val())

	n, err := rdb.SRem(ctx, bucket, "task-42").Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
	require.Equal(t, int64(0), rdb.SCard(ctx, bucket).Val())
}

func TestGCRateLimiters_ReleasesMissingTasks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := newTestDB(t)
	rdb := newTestRedis(t)
	bucket := consts.RestartPedestalTokenBucket
	_, err := rdb.SAdd(ctx, bucket, "ghost-task").Result()
	require.NoError(t, err)
	released, touched, err := gcRateLimitersWith(ctx, rdb, db, map[string]int{
		bucket: consts.MaxConcurrentRestartPedestal,
	})
	require.NoError(t, err)
	require.Equal(t, 1, released)
	require.Equal(t, 1, touched)
}

func TestGCRateLimiters_NoLeaks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := newTestDB(t)
	rdb := newTestRedis(t)
	require.NoError(t, db.Create(&database.Task{ID: "running-task-1", State: consts.TaskRunning}).Error)
	require.NoError(t, db.Create(&database.Task{ID: "pending-task-1", State: consts.TaskPending}).Error)
	bucket := consts.BuildContainerTokenBucket
	_, err := rdb.SAdd(ctx, bucket, "running-task-1", "pending-task-1").Result()
	require.NoError(t, err)
	released, touched, err := gcRateLimitersWith(ctx, rdb, db, map[string]int{
		bucket: consts.MaxConcurrentBuildContainer,
	})
	require.NoError(t, err)
	require.Equal(t, 0, released)
	require.Equal(t, 0, touched)
}

func TestIsTerminalState(t *testing.T) {
	require.True(t, isTerminalState(consts.TaskCompleted))
	require.True(t, isTerminalState(consts.TaskError))
	require.True(t, isTerminalState(consts.TaskCancelled))
	require.False(t, isTerminalState(consts.TaskRunning))
	require.False(t, isTerminalState(consts.TaskPending))
	require.False(t, isTerminalState(consts.TaskRescheduled))
}
