package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/client"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const InjectAlgorithmsKey = "injection:algorithms"

func CheckCachedTraceID(ctx context.Context, traceID string) bool {
	exists, err := client.GetRedisClient().HExists(ctx, InjectAlgorithmsKey, traceID).Result()
	if err != nil {
		logrus.Errorf("failed to check if trace ID %s exists in cache: %v", traceID, err)
		return false
	}

	return exists
}

func GetCachedAlgorithmsFromRedis(ctx context.Context, traceID string) ([]string, error) {
	namesJSON, err := client.GetRedisClient().HGet(ctx, InjectAlgorithmsKey, traceID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("no cached algorithms found for trace id %s", traceID)
		}
		return nil, fmt.Errorf("failed to get cached algorithms for trace id %s: %v", traceID, err)
	}

	var names []string
	if err := json.Unmarshal([]byte(namesJSON), &names); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached algorithms for trace id %s: %v", traceID, err)
	}

	return names, nil
}

func SetAlgorithmsToRedis(ctx context.Context, traceID string, names []string) error {
	namesJSON, err := json.Marshal(names)
	if err != nil {
		return fmt.Errorf("failed to marshal algorithm names to JSON: %v", err)
	}

	if _, err := client.GetRedisClient().Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, InjectAlgorithmsKey, traceID, namesJSON)
		return nil
	}); err != nil {
		return fmt.Errorf("failed to cache algorithms for trace id %s: %v", traceID, err)
	}

	return nil
}
