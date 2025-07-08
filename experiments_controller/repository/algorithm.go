package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/client"
	"github.com/LGU-SE-Internal/rcabench/dto"
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

func GetCachedAlgorithmItemsFromRedis(ctx context.Context, traceID string) ([]dto.AlgorithmItem, error) {
	itemsJSON, err := client.GetRedisClient().HGet(ctx, InjectAlgorithmsKey, traceID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("no cached algorithms found for trace id %s", traceID)
		}
		return nil, fmt.Errorf("failed to get cached algorithms for trace id %s: %v", traceID, err)
	}

	var items []dto.AlgorithmItem
	if err := json.Unmarshal([]byte(itemsJSON), &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached algorithm items for trace id %s: %v", traceID, err)
	}

	return items, nil
}

func SetAlgorithmItemsToRedis(ctx context.Context, traceID string, items []dto.AlgorithmItem) error {
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("failed to marshal algorithm items to JSON: %v", err)
	}

	if _, err := client.GetRedisClient().Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, InjectAlgorithmsKey, traceID, itemsJSON)
		return nil
	}); err != nil {
		return fmt.Errorf("failed to cache algorithm items for trace id %s: %v", traceID, err)
	}

	return nil
}
