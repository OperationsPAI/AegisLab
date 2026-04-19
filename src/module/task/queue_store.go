package task

import (
	"context"
	"fmt"

	redisinfra "aegis/infra/redis"
	goredis "github.com/redis/go-redis/v9"
)

const jobLogsChannelPrefix = "joblogs"

type TaskQueueStore struct {
	redis *redisinfra.Gateway
}

func NewTaskQueueStore(redis *redisinfra.Gateway) *TaskQueueStore {
	return &TaskQueueStore{redis: redis}
}

func (s *TaskQueueStore) SubscribeJobLogs(ctx context.Context, taskID string) (*goredis.PubSub, error) {
	channel := fmt.Sprintf("%s:%s", jobLogsChannelPrefix, taskID)
	return s.redis.Subscribe(ctx, channel)
}
