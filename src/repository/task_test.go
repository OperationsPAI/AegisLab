package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"aegis/dto"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTaskRedisClient struct {
	mock.Mock
}

func (m *mockTaskRedisClient) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	args := []interface{}{ctx, key}
	args = append(args, values...)
	return m.Called(args...).Get(0).(*redis.IntCmd)
}

func (m *mockTaskRedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	args := []interface{}{ctx, key}
	args = append(args, values...)
	return m.Called(args...).Get(0).(*redis.IntCmd)
}

func (m *mockTaskRedisClient) BRPop(ctx context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	args := []interface{}{ctx, timeout}
	for _, key := range keys {
		args = append(args, key)
	}
	return m.Called(args...).Get(0).(*redis.StringSliceCmd)
}

func (m *mockTaskRedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	args := []interface{}{ctx, key}
	for _, member := range members {
		args = append(args, member)
	}
	return m.Called(args...).Get(0).(*redis.IntCmd)
}

func (m *mockTaskRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return m.Called(ctx, key).Get(0).(*redis.StringCmd)
}

func (m *mockTaskRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	return m.Called(ctx, key).Get(0).(*redis.IntCmd)
}

func (m *mockTaskRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return m.Called(ctx, key, value, expiration).Get(0).(*redis.StatusCmd)
}

func (m *mockTaskRedisClient) Decr(ctx context.Context, key string) *redis.IntCmd {
	return m.Called(ctx, key).Get(0).(*redis.IntCmd)
}

func (m *mockTaskRedisClient) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return m.Called(ctx, key, field).Get(0).(*redis.StringCmd)
}

func (m *mockTaskRedisClient) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	return m.Called(ctx, key, opt).Get(0).(*redis.StringSliceCmd)
}

func (m *mockTaskRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	args := []interface{}{ctx, key}
	args = append(args, members...)
	return m.Called(args...).Get(0).(*redis.IntCmd)
}

func (m *mockTaskRedisClient) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	args := []interface{}{ctx, key}
	for _, field := range fields {
		args = append(args, field)
	}
	return m.Called(args...).Get(0).(*redis.IntCmd)
}

func useMockTaskRedis(t *testing.T, cli *mockTaskRedisClient) {
	t.Helper()

	origGetClient := getTaskRedisClient
	origGetScriptClient := getTaskRedisScriptClient
	origListRange := getRedisListRange
	origZRange := getRedisZRangeByScoreWithScores
	origNow := currentTaskTime
	origProcessDelayed := runProcessDelayedTasksScript
	origRemoveFromList := runRemoveFromListScript

	getTaskRedisClient = func() taskRedisClient { return cli }
	getTaskRedisScriptClient = func() *redis.Client { return nil }
	getRedisListRange = origListRange
	getRedisZRangeByScoreWithScores = origZRange
	currentTaskTime = origNow
	runProcessDelayedTasksScript = origProcessDelayed
	runRemoveFromListScript = origRemoveFromList

	t.Cleanup(func() {
		getTaskRedisClient = origGetClient
		getTaskRedisScriptClient = origGetScriptClient
		getRedisListRange = origListRange
		getRedisZRangeByScoreWithScores = origZRange
		currentTaskTime = origNow
		runProcessDelayedTasksScript = origProcessDelayed
		runRemoveFromListScript = origRemoveFromList
		cli.AssertExpectations(t)
	})
}

func TestSubmitImmediateTask(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	taskData := []byte(`{"task_id":"task-1"}`)
	cli.On("LPush", ctx, ReadyQueueKey, taskData).Return(redis.NewIntResult(1, nil)).Once()
	cli.On("HSet", ctx, TaskIndexKey, "task-1", ReadyQueueKey).Return(redis.NewIntResult(1, nil)).Once()

	err := SubmitImmediateTask(ctx, taskData, "task-1")
	require.NoError(t, err)
}

func TestSubmitImmediateTaskReturnsPushError(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	pushErr := errors.New("push failed")
	taskData := []byte(`{"task_id":"task-1"}`)
	cli.On("LPush", ctx, ReadyQueueKey, taskData).Return(redis.NewIntResult(0, pushErr)).Once()

	err := SubmitImmediateTask(ctx, taskData, "task-1")
	require.ErrorIs(t, err, pushErr)
}

func TestGetTask(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	cli.On("BRPop", ctx, 5*time.Second, ReadyQueueKey).
		Return(redis.NewStringSliceResult([]string{ReadyQueueKey, `{"task_id":"task-1"}`}, nil)).
		Once()

	task, err := GetTask(ctx, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, `{"task_id":"task-1"}`, task)
}

func TestSubmitDelayedTask(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	taskData := []byte(`{"task_id":"task-2"}`)
	executeTime := int64(1700000000)

	cli.On("ZAdd", ctx, DelayedQueueKey, mock.MatchedBy(func(member redis.Z) bool {
		return member.Score == float64(executeTime) && string(member.Member.([]byte)) == string(taskData)
	})).Return(redis.NewIntResult(1, nil)).Once()
	cli.On("HSet", ctx, TaskIndexKey, "task-2", DelayedQueueKey).Return(redis.NewIntResult(1, nil)).Once()

	err := SubmitDelayedTask(ctx, taskData, "task-2", executeTime)
	require.NoError(t, err)
}

func TestProcessDelayedTasks(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	frozen := time.Unix(1700000000, 0)
	currentTaskTime = func() time.Time { return frozen }

	runProcessDelayedTasksScript = func(callCtx context.Context, redisCli *redis.Client, now int64) ([]string, error) {
		assert.Equal(t, ctx, callCtx)
		assert.Nil(t, redisCli)
		assert.Equal(t, frozen.Unix(), now)
		return []string{`{"task_id":"task-3"}`}, nil
	}

	tasks, err := ProcessDelayedTasks(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{`{"task_id":"task-3"}`}, tasks)
}

func TestHandleFailedTask(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	frozen := time.Unix(1700000000, 0)
	currentTaskTime = func() time.Time { return frozen }
	taskData := []byte(`{"task_id":"task-4"}`)

	cli.On("ZAdd", ctx, DeadLetterKey, mock.MatchedBy(func(member redis.Z) bool {
		return member.Score == float64(frozen.Add(30*time.Second).Unix()) &&
			string(member.Member.([]byte)) == string(taskData)
	})).Return(redis.NewIntResult(1, nil)).Once()

	err := HandleFailedTask(ctx, taskData, 30)
	require.NoError(t, err)
}

func TestHandleCronRescheduleFailure(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	frozen := time.Unix(1700000000, 0)
	currentTaskTime = func() time.Time { return frozen }
	taskData := []byte(`{"task_id":"task-5"}`)

	cli.On("ZAdd", ctx, DeadLetterKey, mock.MatchedBy(func(member redis.Z) bool {
		return member.Score == float64(frozen.Unix()) &&
			string(member.Member.([]byte)) == string(taskData)
	})).Return(redis.NewIntResult(1, nil)).Once()

	err := HandleCronRescheduleFailure(ctx, taskData)
	require.NoError(t, err)
}

func TestAcquireConcurrencyLock(t *testing.T) {
	ctx := context.Background()

	t.Run("acquires when below limit", func(t *testing.T) {
		cli := &mockTaskRedisClient{}
		useMockTaskRedis(t, cli)

		cli.On("Get", ctx, ConcurrencyLockKey).Return(redis.NewStringResult("3", nil)).Once()
		cli.On("Incr", ctx, ConcurrencyLockKey).Return(redis.NewIntResult(4, nil)).Once()

		assert.True(t, AcquireConcurrencyLock(ctx))
	})

	t.Run("refuses when at limit", func(t *testing.T) {
		cli := &mockTaskRedisClient{}
		useMockTaskRedis(t, cli)

		cli.On("Get", ctx, ConcurrencyLockKey).Return(redis.NewStringResult("20", nil)).Once()

		assert.False(t, AcquireConcurrencyLock(ctx))
	})
}

func TestInitReleaseAndQueueLookupHelpers(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	cli.On("Set", ctx, ConcurrencyLockKey, 0, time.Duration(0)).Return(redis.NewStatusResult("OK", nil)).Once()
	cli.On("Decr", ctx, ConcurrencyLockKey).Return(redis.NewIntResult(0, nil)).Once()
	cli.On("HGet", ctx, TaskIndexKey, "task-6").Return(redis.NewStringResult(ReadyQueueKey, nil)).Once()
	cli.On("HDel", ctx, TaskIndexKey, "task-6").Return(redis.NewIntResult(1, nil)).Once()

	require.NoError(t, InitConcurrencyLock(ctx))
	ReleaseConcurrencyLock(ctx)

	queue, err := GetTaskQueue(ctx, "task-6")
	require.NoError(t, err)
	assert.Equal(t, ReadyQueueKey, queue)

	require.NoError(t, DeleteTaskIndex(ctx, "task-6"))
}

func TestListDelayedTasks(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	getRedisZRangeByScoreWithScores = func(callCtx context.Context, key string, limit int64) ([]redis.Z, error) {
		assert.Equal(t, ctx, callCtx)
		assert.Equal(t, DelayedQueueKey, key)
		assert.EqualValues(t, 2, limit)
		return []redis.Z{
			{Member: `{"task_id":"task-7"}`},
			{Member: `{"task_id":"task-8"}`},
		}, nil
	}

	tasks, err := ListDelayedTasks(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, []string{`{"task_id":"task-7"}`, `{"task_id":"task-8"}`}, tasks)
}

func TestListDelayedTasksRejectsNonStringMembers(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	getRedisZRangeByScoreWithScores = func(context.Context, string, int64) ([]redis.Z, error) {
		return []redis.Z{{Member: []byte("bad")}}, nil
	}

	_, err := ListDelayedTasks(ctx, 1)
	require.EqualError(t, err, "invalid delayed task data")
}

func TestListReadyTasks(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	getRedisListRange = func(callCtx context.Context, key string) ([]string, error) {
		assert.Equal(t, ctx, callCtx)
		assert.Equal(t, ReadyQueueKey, key)
		return []string{"task-a", "task-b"}, nil
	}

	tasks, err := ListReadyTasks(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"task-a", "task-b"}, tasks)
}

func TestRemoveFromList(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	runRemoveFromListScript = func(callCtx context.Context, redisCli *redis.Client, key, taskID string) (int, error) {
		assert.Equal(t, ctx, callCtx)
		assert.Nil(t, redisCli)
		assert.Equal(t, ReadyQueueKey, key)
		assert.Equal(t, "task-9", taskID)
		return 1, nil
	}

	removed, err := RemoveFromList(ctx, ReadyQueueKey, "task-9")
	require.NoError(t, err)
	assert.True(t, removed)
}

func TestRemoveFromListReturnsScriptError(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	runRemoveFromListScript = func(context.Context, *redis.Client, string, string) (int, error) {
		return 0, errors.New("script failed")
	}

	removed, err := RemoveFromList(ctx, ReadyQueueKey, "task-9")
	require.EqualError(t, err, "failed to remove from list: script failed")
	assert.False(t, removed)
}

func TestRemoveFromZSet(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	taskJSON := `{"task_id":"task-10"}`
	cli.On("ZRangeByScore", ctx, DelayedQueueKey, mock.MatchedBy(func(opt *redis.ZRangeBy) bool {
		return opt.Min == "-inf" && opt.Max == "+inf"
	})).Return(redis.NewStringSliceResult([]string{taskJSON}, nil)).Once()
	cli.On("ZRem", ctx, DelayedQueueKey, taskJSON).Return(redis.NewIntResult(1, nil)).Once()

	assert.True(t, RemoveFromZSet(ctx, DelayedQueueKey, "task-10"))
}

func TestRemoveFromZSetReturnsFalseForNoMatch(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	taskJSON := `{"task_id":"other-task"}`
	cli.On("ZRangeByScore", ctx, DelayedQueueKey, mock.AnythingOfType("*redis.ZRangeBy")).
		Return(redis.NewStringSliceResult([]string{taskJSON}, nil)).
		Once()

	assert.False(t, RemoveFromZSet(ctx, DelayedQueueKey, "task-10"))
}

func TestQueueHelpersOperateOnSerializedTaskPayloads(t *testing.T) {
	ctx := context.Background()
	cli := &mockTaskRedisClient{}
	useMockTaskRedis(t, cli)

	task := dto.UnifiedTask{TaskID: "task-11"}
	taskJSON := `{"task_id":"task-11"}`

	cli.On("ZRangeByScore", ctx, DeadLetterKey, mock.AnythingOfType("*redis.ZRangeBy")).
		Return(redis.NewStringSliceResult([]string{taskJSON}, nil)).
		Once()
	cli.On("ZRem", ctx, DeadLetterKey, taskJSON).Return(redis.NewIntResult(1, nil)).Once()

	assert.NotEmpty(t, task.TaskID)
	assert.True(t, RemoveFromZSet(ctx, DeadLetterKey, task.TaskID))
}
