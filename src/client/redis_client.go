package client

import (
	"context"
	"sync"

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

		ctx := context.Background()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			logrus.Fatalf("Failed to connect to Redis: %v", err)
		}
	})
	return redisClient
}
