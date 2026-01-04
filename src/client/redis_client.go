package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"aegis/config"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Singleton pattern Redis client
var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

// Get Redis client
func GetRedisClient() *redis.Client {
	redisOnce.Do(func() {
		logrus.Infof("Connecting to Redis %s", config.GetString("redis.host"))
		redisClient = redis.NewClient(&redis.Options{
			Addr:     config.GetString("redis.host"),
			Password: "",
			DB:       0,
		})

		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			logrus.Fatalf("Failed to connect to Redis: %v", err)
		}
	})
	return redisClient
}

// CheckCachedField checks if a field exists in Redis cache
func CheckCachedField(ctx context.Context, key, field string) bool {
	exists, err := GetRedisClient().HExists(ctx, key, field).Result()
	if err != nil {
		logrus.Errorf("failed to check if field %s exists in cache: %v", field, err)
		return false
	}

	return exists
}

// GetHashField retrieves a field from Redis hash and unmarshals it into the target
func GetHashField[T any](ctx context.Context, key, field string, target *T) error {
	itemJSON, err := GetRedisClient().HGet(ctx, key, field).Result()
	if err != nil {
		return fmt.Errorf("failed to get hash field %s from key %s: %w", field, key, err)
	}

	if err := json.Unmarshal([]byte(itemJSON), target); err != nil {
		return fmt.Errorf("failed to unmarshal cached items for field %s: %w", field, err)
	}

	return nil
}

// SetHashField sets a field in Redis hash with the provided item
func SetHashField[T any](ctx context.Context, key, field string, item T) error {
	itemJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal items to JSON: %w", err)
	}

	if _, err := GetRedisClient().Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, key, field, itemJSON)
		return nil
	}); err != nil {
		return fmt.Errorf("failed to set hash field %s in key %s: %w", field, key, err)
	}

	return nil
}

// GetRedisListRange retrieves all elements from a Redis list
func GetRedisListRange(ctx context.Context, key string) ([]string, error) {
	result, err := GetRedisClient().LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get list range for key '%s': %w", key, err)
	}
	return result, nil
}

// GetRedisZRangeByScoreWithScores retrieves elements from a Redis sorted set by score with a limit
func GetRedisZRangeByScoreWithScores(ctx context.Context, key string, limit int64) ([]redis.Z, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be a positive number")
	}
	options := &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  limit,
	}

	results, err := GetRedisClient().ZRangeByScoreWithScores(ctx, key, options).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled tasks from key '%s': %w", key, err)
	}

	return results, nil
}

// RedisXAdd adds an entry to a Redis stream
func RedisXAdd(ctx context.Context, stream string, values map[string]any) error {
	_, err := GetRedisClient().XAdd(ctx, &redis.XAddArgs{
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

// RedisXRead reads entries from Redis streams
func RedisXRead(ctx context.Context, streams []string, count int64, block time.Duration) ([]redis.XStream, error) {
	result, err := GetRedisClient().XRead(ctx, &redis.XReadArgs{
		Streams: streams,
		Count:   count,
		Block:   block,
	}).Result()

	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("redis XREAD failed: %w", err)
	}

	return result, nil
}
