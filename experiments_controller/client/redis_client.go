package client

import (
	"context"
	"sync"

	"github.com/CUHK-SE-Group/rcabench/config"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// 单例模式的 Redis 客户端
var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

// 初始化函数
func init() {
	ctx := context.Background()
	initConsumerGroup(ctx)
}

// 获取 Redis 客户端
func GetRedisClient() *redis.Client {
	redisOnce.Do(func() {
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
