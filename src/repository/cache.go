package repository

import (
	"context"
	"fmt"

	"aegis/client"
)

const RedisDataInitKey = "app:system:data_init_done"
const InitCompleteValue = "true"

func IsInitialDataSeeded(ctx context.Context) bool {
	return client.GetRedisClient().Get(ctx, RedisDataInitKey).Val() == InitCompleteValue
}

func MarkDataSeedingComplete(ctx context.Context) error {
	if err := client.GetRedisClient().Set(ctx, RedisDataInitKey, InitCompleteValue, 0).Err(); err != nil {
		return fmt.Errorf("failed to set setup complete flag in Redis: %v", err)
	}
	return nil
}
