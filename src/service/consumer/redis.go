package consumer

import (
	"context"
	"fmt"

	"aegis/consts"
	"aegis/dto"
	redis "aegis/infra/redis"
)

func consumerDetachedContext() context.Context {
	return context.TODO()
}

type redisStreamEvent interface {
	ToRedisStream() map[string]any
}

func publishRedisStreamEvent(gateway *redis.Gateway, ctx context.Context, stream string, event redisStreamEvent) error {
	if gateway == nil {
		return fmt.Errorf("redis gateway is nil")
	}
	if err := gateway.XAdd(ctx, stream, event.ToRedisStream()); err != nil {
		return fmt.Errorf("failed to publish redis stream event: %w", err)
	}
	return nil
}

func publishTraceStreamEvent(gateway *redis.Gateway, ctx context.Context, stream string, event *dto.TraceStreamEvent) error {
	if event == nil {
		return nil
	}
	return publishRedisStreamEvent(gateway, ctx, stream, event)
}

func loadCachedInjectionAlgorithms(gateway *redis.Gateway, ctx context.Context, groupID string) ([]dto.ContainerVersionItem, bool, error) {
	if gateway == nil {
		return nil, false, fmt.Errorf("redis gateway is nil")
	}
	if !gateway.CheckCachedField(ctx, consts.InjectionAlgorithmsKey, groupID) {
		return nil, false, nil
	}

	var algorithms []dto.ContainerVersionItem
	if err := gateway.GetHashField(ctx, consts.InjectionAlgorithmsKey, groupID, &algorithms); err != nil {
		return nil, false, err
	}
	return algorithms, true, nil
}
