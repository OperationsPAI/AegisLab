package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aegis/config"
	"aegis/consts"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type Gateway struct {
	client *redis.Client
}

func NewGateway(client *redis.Client) *Gateway {
	if client == nil {
		client = newClient()
	}
	return &Gateway{client: client}
}

func NewGatewayWithLifecycle(lc fx.Lifecycle) *Gateway {
	gateway := NewGateway(nil)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logrus.Info("Closing Redis client")
			return gateway.close()
		},
	})

	return gateway
}

func (g *Gateway) clientOrInit() *redis.Client {
	if g.client == nil {
		g.client = newClient()
	}
	return g.client
}

func (g *Gateway) close() error {
	if g.client == nil {
		return nil
	}
	return g.client.Close()
}

func (g *Gateway) CheckCachedField(ctx context.Context, key, field string) bool {
	exists, err := g.clientOrInit().HExists(ctx, key, field).Result()
	if err != nil {
		logrus.Errorf("failed to check if field %s exists in cache: %v", field, err)
		return false
	}
	return exists
}

func (g *Gateway) GetHashField(ctx context.Context, key, field string, target any) error {
	itemJSON, err := g.clientOrInit().HGet(ctx, key, field).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get hash field %s from key %s: %w", field, key, err)
	}
	if itemJSON == "" {
		logrus.Warnf("field %s not found in cache key %s", field, key)
		return nil
	}
	if err := json.Unmarshal([]byte(itemJSON), target); err != nil {
		return fmt.Errorf("failed to unmarshal cached items for field %s: %w", field, err)
	}
	return nil
}

func (g *Gateway) SetHashField(ctx context.Context, key, field string, item any) error {
	itemJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal items to JSON: %w", err)
	}
	if _, err := g.clientOrInit().Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, key, field, itemJSON)
		return nil
	}); err != nil {
		return fmt.Errorf("failed to set hash field %s in key %s: %w", field, key, err)
	}
	return nil
}

func (g *Gateway) ListRange(ctx context.Context, key string) ([]string, error) {
	result, err := g.clientOrInit().LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get list range for key '%s': %w", key, err)
	}
	return result, nil
}

func (g *Gateway) ListLength(ctx context.Context, key string) (int64, error) {
	result, err := g.clientOrInit().LLen(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get list length for key '%s': %w", key, err)
	}
	return result, nil
}

// ScanKeys iterates through keys matching the given pattern and returns all of them.
func (g *Gateway) ScanKeys(ctx context.Context, pattern string) ([]string, error) {
	iter := g.clientOrInit().Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan keys %q: %w", pattern, err)
	}
	return keys, nil
}

// DeleteKey removes a single key; returns 1 if it existed, 0 otherwise.
func (g *Gateway) DeleteKey(ctx context.Context, key string) (int64, error) {
	n, err := g.clientOrInit().Del(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to del key %q: %w", key, err)
	}
	return n, nil
}

func (g *Gateway) SetMembers(ctx context.Context, key string) ([]string, error) {
	result, err := g.clientOrInit().SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get set members for key '%s': %w", key, err)
	}
	return result, nil
}

func (g *Gateway) Exists(ctx context.Context, key string) (bool, error) {
	result, err := g.clientOrInit().Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key '%s': %w", key, err)
	}
	return result > 0, nil
}

func (g *Gateway) HashGetAll(ctx context.Context, key string) (map[string]string, error) {
	result, err := g.clientOrInit().HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash fields for key '%s': %w", key, err)
	}
	return result, nil
}

func (g *Gateway) HashGet(ctx context.Context, key, field string) (string, error) {
	result, err := g.clientOrInit().HGet(ctx, key, field).Result()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (g *Gateway) HashSet(ctx context.Context, key string, values map[string]any) error {
	if len(values) == 0 {
		return nil
	}
	if err := g.clientOrInit().HSet(ctx, key, values).Err(); err != nil {
		return fmt.Errorf("failed to set hash fields for key '%s': %w", key, err)
	}
	return nil
}

func (g *Gateway) SeedNamespaceState(ctx context.Context, namespaceKey, namespace string, endTime int64, status int) error {
	_, err := g.clientOrInit().Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.SAdd(ctx, consts.NamespacesKey, namespace)
		pipe.HSetNX(ctx, namespaceKey, "end_time", endTime)
		pipe.HSetNX(ctx, namespaceKey, "trace_id", "")
		pipe.HSetNX(ctx, namespaceKey, "status", status)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to seed namespace state for '%s': %w", namespace, err)
	}
	return nil
}

func (g *Gateway) ZRangeByScoreWithScores(ctx context.Context, key string, limit int64) ([]redis.Z, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be a positive number")
	}
	results, err := g.clientOrInit().ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  limit,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled tasks from key '%s': %w", key, err)
	}
	return results, nil
}

func (g *Gateway) ZRangeByScore(ctx context.Context, key, min, max string) ([]string, error) {
	result, err := g.clientOrInit().ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get sorted set range for key '%s': %w", key, err)
	}
	return result, nil
}

func (g *Gateway) SortedSetCard(ctx context.Context, key string) (int64, error) {
	result, err := g.clientOrInit().ZCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get sorted set size for key '%s': %w", key, err)
	}
	return result, nil
}

func (g *Gateway) ZAdd(ctx context.Context, key string, member redis.Z) error {
	if err := g.clientOrInit().ZAdd(ctx, key, member).Err(); err != nil {
		return fmt.Errorf("failed to add sorted set member for key '%s': %w", key, err)
	}
	return nil
}

func (g *Gateway) ZRemRangeByScore(ctx context.Context, key, min, max string) error {
	if err := g.clientOrInit().ZRemRangeByScore(ctx, key, min, max).Err(); err != nil {
		return fmt.Errorf("failed to trim sorted set for key '%s': %w", key, err)
	}
	return nil
}

func (g *Gateway) SetRemove(ctx context.Context, key string, members ...any) (int64, error) {
	result, err := g.clientOrInit().SRem(ctx, key, members...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to remove set members for key '%s': %w", key, err)
	}
	return result, nil
}

func (g *Gateway) SetCard(ctx context.Context, key string) (int64, error) {
	result, err := g.clientOrInit().SCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get set size for key '%s': %w", key, err)
	}
	return result, nil
}

func (g *Gateway) XAdd(ctx context.Context, stream string, values map[string]any) error {
	_, err := g.clientOrInit().XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: 1000,
		Approx: true,
		ID:     "*",
		Values: values,
	}).Result()
	if err != nil {
		return fmt.Errorf("redis XADD failed for stream '%s': %w", stream, err)
	}
	return nil
}

func (g *Gateway) RunScript(ctx context.Context, script *redis.Script, keys []string, args ...any) (any, error) {
	result, err := script.Run(ctx, g.clientOrInit(), keys, args...).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (g *Gateway) Ping(ctx context.Context) error {
	if err := g.clientOrInit().Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis PING failed: %w", err)
	}
	return nil
}

func (g *Gateway) HashLength(ctx context.Context, key string) (int64, error) {
	result, err := g.clientOrInit().HLen(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get hash length for key '%s': %w", key, err)
	}
	return result, nil
}

func (g *Gateway) GetInt64(ctx context.Context, key string) (int64, error) {
	result, err := g.clientOrInit().Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get int64 value for key '%s': %w", key, err)
	}
	return result, nil
}

func (g *Gateway) Watch(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error {
	return g.clientOrInit().Watch(ctx, fn, keys...)
}

func (g *Gateway) XRead(ctx context.Context, streams []string, count int64, block time.Duration) ([]redis.XStream, error) {
	result, err := g.clientOrInit().XRead(ctx, &redis.XReadArgs{
		Streams: streams,
		Count:   count,
		Block:   block,
	}).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("redis XREAD failed: %w", err)
	}
	return result, nil
}

func (g *Gateway) Publish(ctx context.Context, channel string, message any) error {
	var payload string
	switch v := message.(type) {
	case string:
		payload = v
	default:
		data, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		payload = string(data)
	}

	if err := g.clientOrInit().Publish(ctx, channel, payload).Err(); err != nil {
		return fmt.Errorf("redis PUBLISH failed for channel '%s': %w", channel, err)
	}
	return nil
}

func (g *Gateway) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := g.clientOrInit().Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("redis SET failed for key '%s': %w", key, err)
	}
	return nil
}

func (g *Gateway) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	result, err := g.clientOrInit().SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis SETNX failed for key '%s': %w", key, err)
	}
	return result, nil
}

func (g *Gateway) Subscribe(ctx context.Context, channel string) (*redis.PubSub, error) {
	pubsub := g.clientOrInit().Subscribe(ctx, channel)
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, fmt.Errorf("redis SUBSCRIBE failed for channel '%s': %w", channel, err)
	}
	return pubsub, nil
}

func (g *Gateway) InitConcurrencyLock(ctx context.Context) error {
	return g.clientOrInit().Set(ctx, ConcurrencyLockKey, 0, 0).Err()
}

func newClient() *redis.Client {
	logrus.Infof("Connecting to Redis %s", config.GetString("redis.host"))
	client := redis.NewClient(&redis.Options{
		Addr:     config.GetString("redis.host"),
		Password: "",
		DB:       0,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		logrus.Fatalf("Failed to connect to Redis: %v", err)
	}
	return client
}
