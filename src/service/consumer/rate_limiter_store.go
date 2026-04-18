package consumer

import (
	"context"
	"fmt"

	redisinfra "aegis/infra/redis"
	"github.com/redis/go-redis/v9"
)

type tokenBucketStore struct {
	bucketKey string
	client    *redisinfra.Gateway
}

func newTokenBucketStore(client *redisinfra.Gateway, bucketKey string) tokenBucketStore {
	return tokenBucketStore{bucketKey: bucketKey, client: client}
}

func (s tokenBucketStore) acquire(ctx context.Context, maxTokens int, taskID, traceID string) (bool, error) {
	script := redis.NewScript(`
		local bucket_key = KEYS[1]
		local max_tokens = tonumber(ARGV[1])
		local task_id = ARGV[2]
		local trace_id = ARGV[3]
		local expire_time = tonumber(ARGV[4])
		
		local current_tokens = redis.call('SCARD', bucket_key)
		
		if current_tokens < max_tokens then
			redis.call('SADD', bucket_key, task_id)
			redis.call('EXPIRE', bucket_key, expire_time)
			return 1
		else
			return 0
		end
	`)

	const expireTime = 10 * 60
	result, err := s.client.RunScript(ctx, script, []string{s.bucketKey},
		maxTokens, taskID, traceID, expireTime)
	if err != nil {
		return false, fmt.Errorf("failed to acquire token: %v", err)
	}
	return result.(int64) == 1, nil
}

func (s tokenBucketStore) release(ctx context.Context, taskID string) (int64, error) {
	result, err := s.client.SetRemove(ctx, s.bucketKey, taskID)
	if err != nil {
		return 0, fmt.Errorf("failed to release token: %v", err)
	}
	return result, nil
}

func (s tokenBucketStore) inUse(ctx context.Context) (int64, error) {
	result, err := s.client.SetCard(ctx, s.bucketKey)
	if err != nil {
		return 0, fmt.Errorf("failed to get token usage: %v", err)
	}
	return result, nil
}
