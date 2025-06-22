package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/client"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const InjectAlgorithmsKey = "injection:algorithms"

func GetAlgorithmImageInfo(name string) (string, string, error) {
	var record database.Algorithm
	if err := database.DB.
		Where("name = ? AND status = ?", name, true).
		Order("created_at DESC").
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", fmt.Errorf("algorithm '%s' not found", name)
		}
		return "", "", fmt.Errorf("failed to query algorithm: %v", err)
	}

	return record.Image, record.Tag, nil
}

func GetAlgorithmByName(name string) (*database.Algorithm, error) {
	var algorithm database.Algorithm
	if err := database.DB.
		Where("name = ? AND status = ?", name, true).
		Order("created_at DESC").
		First(&algorithm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("algorithm '%s' not found", name)
		}

		return nil, fmt.Errorf("failed to query algorithm: %v", err)
	}

	return &algorithm, nil
}

func ListAlgorithms(status ...bool) ([]database.Algorithm, error) {
	query := database.DB.
		Order("created_at DESC")

	if len(status) == 0 {
	} else if len(status) == 1 {
		query = query.Where("status = ?", status[0])
	} else {
		return nil, fmt.Errorf("invalid status arguments")
	}

	var algorithms []database.Algorithm
	if err := query.
		Find(&algorithms).Error; err != nil {
		return nil, fmt.Errorf("failed to list all algorithms: %v", err)
	}

	return algorithms, nil
}

func ListAlgorithmByNames(names []string) ([]database.Algorithm, error) {
	var algorithms []database.Algorithm
	if err := database.DB.
		Where("name IN ? AND status = ?", names, true).
		Order("created_at DESC").
		Find(&algorithms).Error; err != nil {
		return nil, fmt.Errorf("failed to list algorithms by names: %v", err)
	}

	return algorithms, nil
}
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
