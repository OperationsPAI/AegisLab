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

func CheckCachedField(ctx context.Context, key, field string) bool {
	exists, err := client.GetRedisClient().HExists(ctx, key, field).Result()
	if err != nil {
		logrus.Errorf("failed to check if field %s exists in cache: %v", field, err)
		return false
	}

	return exists
}

func GetCachedAlgorithmItemsFromRedis(ctx context.Context, key, field string) ([]dto.AlgorithmItem, error) {
	itemsJSON, err := client.GetRedisClient().HGet(ctx, key, field).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("no cached algorithms found for field %s", field)
		}

		return nil, fmt.Errorf("failed to get cached algorithms for field %s: %v", field, err)
	}

	var items []dto.AlgorithmItem
	if err := json.Unmarshal([]byte(itemsJSON), &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached algorithm items for field %s: %v", field, err)
	}

	return items, nil
}

func SetAlgorithmItemsToRedis(ctx context.Context, key, field string, items []dto.AlgorithmItem) error {
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("failed to marshal algorithm items to JSON: %v", err)
	}

	if _, err := client.GetRedisClient().Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, key, field, itemsJSON)
		return nil
	}); err != nil {
		return fmt.Errorf("failed to cache algorithm items for field %s: %v", field, err)
	}

	return nil
}
