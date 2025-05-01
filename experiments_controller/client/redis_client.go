package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// 单例模式的 Redis 客户端
var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

type StreamEvent struct {
	TaskID   string
	TaskType consts.TaskType
	Status   string
	FileName string
	Line     int
	Name     string
	Payload  map[string]any
}

func (s *StreamEvent) ToRedisStream() map[string]any {
	return map[string]any{
		consts.RdbEventTaskID:   s.TaskID,
		consts.RdbEventTaskType: s.TaskType,
		consts.RdbEventStatus:   s.Status,
		consts.RdbEventFileName: s.FileName,
		consts.RdbEventLine:     s.Line,
		consts.RdbEventName:     s.Name,
		consts.RdbEventPayload:  s.Payload,
	}
}

func (s *StreamEvent) ToSSE() (string, error) {
	message := map[string]any{
		consts.RdbEventTaskID:   s.TaskID,
		consts.RdbEventTaskType: s.TaskType,
		consts.RdbEventStatus:   s.Status,
		consts.RdbEventPayload:  s.Payload,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return "", nil
	}

	return string(jsonData), nil
}

// 获取 Redis 客户端
func GetRedisClient() *redis.Client {
	redisOnce.Do(func() {
		logrus.Infof("Connecting to Redis %s", config.GetString("redis.host"))
		redisClient = redis.NewClient(&redis.Options{
			Addr:     config.GetString("redis.host"),
			Password: "",
			DB:       0,
		})
		if err := errors.Join(redisotel.InstrumentTracing(redisClient), redisotel.InstrumentMetrics(redisClient)); err != nil {
			log.Fatal(err)
		}
		ctx := context.Background()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			logrus.Fatalf("Failed to connect to Redis: %v", err)
		}
	})
	return redisClient
}

func PublishEvent(ctx context.Context, stream string, event StreamEvent) (string, error) {
	return GetRedisClient().XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: 10000,
		Approx: true,
		ID:     "*",
		Values: event.ToRedisStream(),
	}).Result()
}

func ReadStreamEvents(ctx context.Context, stream string, lastID string, count int64, block time.Duration) ([]redis.XStream, error) {
	if lastID == "" {
		lastID = "0"
	}

	return GetRedisClient().XRead(ctx, &redis.XReadArgs{
		Streams: []string{stream, lastID},
		Count:   count,
		Block:   block,
	}).Result()
}

// CreateConsumerGroup 创建 Redis Stream 消费者组
func CreateConsumerGroup(ctx context.Context, stream, group, startID string) error {
	err := GetRedisClient().XGroupCreate(ctx, stream, group, startID).Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}
	return nil
}

// ConsumeStreamEvents 使用消费者组消费 Redis Stream 事件
func ConsumeStreamEvents(ctx context.Context, stream, group, consumer string, count int64, block time.Duration) ([]redis.XStream, error) {
	return GetRedisClient().XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    block,
	}).Result()
}

func AcknowledgeMessage(ctx context.Context, stream, group, id string) error {
	return GetRedisClient().XAck(ctx, stream, group, id).Err()
}

// ParseEventFromValues 从 Redis Stream 消息解析事件
func ParseEventFromValues(values map[string]any) (*StreamEvent, error) {
	message := "missing or invalid key %s in redis stream message values"

	taskID, ok := values[consts.RdbEventTaskID].(string)
	if !ok || taskID == "" {
		return nil, fmt.Errorf(message, consts.RdbEventTaskID)
	}

	taskType, ok := values[consts.RdbEventTaskType].(consts.TaskType)
	if !ok || taskID == "" {
		return nil, fmt.Errorf(message, consts.RdbEventTaskType)
	}

	status, ok := values[consts.RdbEventStatus].(string)
	if !ok || status == "" {
		return nil, fmt.Errorf(message, consts.RdbEventStatus)
	}

	event := &StreamEvent{
		TaskID:   taskID,
		TaskType: taskType,
		Status:   status,
	}

	_, exists := values[consts.RdbEventPayload]
	if exists {
		payload, ok := values[consts.RdbEventPayload].(map[string]any)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventPayload)
		}

		event.Payload = payload
	}

	_, exists = values[consts.RdbEventFileName]
	if exists {
		fileName, ok := values[consts.RdbEventFileName].(string)
		if !ok || fileName == "" {
			return nil, fmt.Errorf(message, consts.RdbEventTaskID)
		}

		event.FileName = fileName
	}

	_, exists = values[consts.RdbEventLine]
	if exists {
		lineInt64, ok := values[consts.RdbEventLine].(int64)
		if !ok || lineInt64 == 0 {
			return nil, fmt.Errorf(message, consts.RdbEventLine)
		}

		event.Line = int(lineInt64)
	}

	return event, nil
}
