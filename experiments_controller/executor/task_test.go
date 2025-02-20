package executor_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask(t *testing.T) {
	task := &executor.UnifiedTask{
		TaskID:    "immediate-1",
		Type:      executor.TaskTypeFaultInjection,
		Immediate: true,
	}
	if _, err := executor.SubmitTask(context.Background(), task); err != nil {
		t.Error(err)
	}

	task1 := &executor.UnifiedTask{
		TaskID:      "delayed-1",
		Type:        executor.TaskTypeRunAlgorithm,
		ExecuteTime: time.Now().Add(10 * time.Second).Unix(),
	}
	if _, err := executor.SubmitTask(context.Background(), task1); err != nil {
		t.Error(err)
	}

	task2 := &executor.UnifiedTask{
		Type:        executor.TaskTypeBuildImages,
		ExecuteTime: time.Now().Add(10 * time.Second).Unix(),
	}
	if _, err := executor.SubmitTask(context.Background(), task2); err != nil {
		t.Error(err)
	}
	time.Sleep(5 * time.Second)
}

func TestSubmitTask(t *testing.T) {
	ctx := context.Background()
	redisCli := client.GetRedisClient()
	defer redisCli.FlushDB(ctx)

	t.Run("SubmitValidImmediateTask", func(t *testing.T) {
		task := &executor.UnifiedTask{
			TaskID:    "immediate-1",
			Type:      executor.TaskTypeFaultInjection,
			Immediate: true,
		}

		_, err := executor.SubmitTask(ctx, task)
		require.NoError(t, err)

		// 验证就绪队列
		taskData, err := redisCli.LPop(ctx, executor.ReadyQueueKey).Result()
		require.NoError(t, err)
		var resultTask executor.UnifiedTask
		require.NoError(t, json.Unmarshal([]byte(taskData), &resultTask))
		assert.Equal(t, task.TaskID, resultTask.TaskID)

		// 验证任务索引
		queueType, err := redisCli.HGet(ctx, executor.TaskIndexKey, task.TaskID).Result()
		require.NoError(t, err)
		assert.Equal(t, executor.ReadyQueueKey, queueType)
	})

	t.Run("SubmitValidDelayedTask", func(t *testing.T) {
		task := &executor.UnifiedTask{
			TaskID:      "delayed-1",
			Type:        executor.TaskTypeRunAlgorithm,
			ExecuteTime: time.Now().Add(1 * time.Hour).Unix(),
		}

		_, err := executor.SubmitTask(ctx, task)
		require.NoError(t, err)

		// 验证延迟队列
		zset, err := redisCli.ZRangeWithScores(ctx, executor.DelayedQueueKey, 0, -1).Result()
		require.NoError(t, err)
		require.Len(t, zset, 1)
		assert.Equal(t, float64(task.ExecuteTime), zset[0].Score)

		// 验证任务索引
		queueType, err := redisCli.HGet(ctx, executor.TaskIndexKey, task.TaskID).Result()
		require.NoError(t, err)
		assert.Equal(t, executor.DelayedQueueKey, queueType)
	})

	t.Run("SubmitTaskWithoutID", func(t *testing.T) {
		task := &executor.UnifiedTask{Immediate: true}
		_, err := executor.SubmitTask(ctx, task)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "task ID cannot be empty")
	})
}

func TestScheduler(t *testing.T) {
	ctx := context.Background()
	redisCli := client.GetRedisClient()
	defer redisCli.FlushDB(ctx)

	t.Run("ProcessDelayedTasks", func(t *testing.T) {
		// 准备测试任务
		task := &executor.UnifiedTask{
			TaskID:      "delayed-2",
			ExecuteTime: time.Now().Unix(),
		}
		taskData, _ := json.Marshal(task)
		redisCli.ZAdd(ctx, executor.DelayedQueueKey, &redis.Z{
			Score:  float64(task.ExecuteTime),
			Member: taskData,
		})

		executor.ProcessDelayedTasks(ctx)

		// 验证任务已转移
		readyLen, err := redisCli.LLen(ctx, executor.ReadyQueueKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), readyLen)

		delayedLen, err := redisCli.ZCard(ctx, executor.DelayedQueueKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), delayedLen)
	})

	// t.Run("HandleCronReschedule", func(t *testing.T) {
	// 	cronTask := &executor.UnifiedTask{
	// 		TaskID:      "cron-1",
	// 		Type:        executor.TaskTypeCron,
	// 		CronExpr:    "* * * * * *",
	// 		ExecuteTime: time.Now().Unix(),
	// 	}
	// 	// 提交并处理任务
	// 	// ...（类似逻辑验证cron任务重新调度）
	// })
}

func TestTaskProcessing(t *testing.T) {
	ctx := context.Background()
	redisCli := client.GetRedisClient()
	defer redisCli.FlushDB(ctx)

	t.Run("SuccessfulProcessing", func(t *testing.T) {
		// 准备测试任务
		task := &executor.UnifiedTask{
			TaskID:    "success-1",
			Immediate: true,
			RetryPolicy: executor.RetryPolicy{
				MaxAttempts: 3,
				BackoffSec:  1,
			},
		}
		_, err := executor.SubmitTask(ctx, task)
		require.NoError(t, err)

		// 模拟消费者处理
		go executor.ConsumeTasks()
		defer redisCli.FlushDB(ctx)

		// 验证任务处理结果
		assert.Eventually(t, func() bool {
			return redisCli.Exists(ctx, "task:success-1:status").Val() == 1
		}, 5*time.Second, 100*time.Millisecond)
	})

	t.Run("RetryMechanism", func(t *testing.T) {
		// 创建会失败的任务（需要根据实际实现模拟）
		// 验证重试次数和最终状态
	})
}

func TestCancelTask(t *testing.T) {
	ctx := context.Background()
	redisCli := client.GetRedisClient()
	defer redisCli.FlushDB(ctx)

	t.Run("CancelPendingTask", func(t *testing.T) {
		task := &executor.UnifiedTask{
			TaskID:    "cancel-1",
			Immediate: true,
		}
		_, err := executor.SubmitTask(ctx, task)
		require.NoError(t, err)

		err = executor.CancelTask(task.TaskID)
		require.NoError(t, err)

		// 验证任务已删除
		exists, err := redisCli.HExists(ctx, executor.TaskIndexKey, task.TaskID).Result()
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("CancelRunningTask", func(t *testing.T) {
		// 创建长时间运行的任务
		// 调用取消并验证上下文取消
	})
}

func TestConcurrencyControl(t *testing.T) {
	ctx := context.Background()
	redisCli := client.GetRedisClient()
	defer redisCli.FlushDB(ctx)

	t.Run("MaxConcurrencyLimit", func(t *testing.T) {
		// 用MaxConcurrency+1个任务测试并发控制
		// 验证同时运行的任务不超过限制
	})
}

func TestErrorHandling(t *testing.T) {
	ctx := context.Background()
	redisCli := client.GetRedisClient()
	defer redisCli.FlushDB(ctx)

	t.Run("InvalidTaskData", func(t *testing.T) {
		redisCli.LPush(ctx, executor.ReadyQueueKey, "invalid-json")
		go executor.ConsumeTasks()

		// 验证错误处理日志
	})

	t.Run("DeadLetterHandling", func(t *testing.T) {
		// 创建会最终失败的任务
		// 验证是否进入死信队列
	})
}
