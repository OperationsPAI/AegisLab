package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/utils"
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
	TaskID    string           `json:"task_id"`
	TaskType  consts.TaskType  `json:"task_type"`
	FileName  string           `json:"file_name"`
	FnName    string           `json:"function_name"`
	Line      int              `json:"line"`
	EventName consts.EventType `json:"event_name"`
	Payload   any              `json:"payload"`
}

type InfoPayloadTemplate struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

func (s *StreamEvent) ToRedisStream() map[string]any {
	payload, err := json.Marshal(s.Payload)
	if err != nil {
		logrus.Errorf("Failed to marshal payload: %v", err)
		return nil
	}

	return map[string]any{
		consts.RdbEventTaskID:   s.TaskID,
		consts.RdbEventTaskType: string(s.TaskType),
		consts.RdbEventFileName: s.FileName,
		consts.RdbEventFn:       s.FnName,
		consts.RdbEventLine:     s.Line,
		consts.RdbEventName:     string(s.EventName),
		consts.RdbEventPayload:  payload,
	}
}

func (s *StreamEvent) ToSSE() (string, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return "", err
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

type EventConf struct {
	CallerLevel int
}

type EventConfOption func(*EventConf)

func WithCallerLevel(level int) func(*EventConf) {
	return func(c *EventConf) {
		c.CallerLevel = level
	}
}

func PublishEvent(ctx context.Context, stream string, event StreamEvent, opts ...EventConfOption) {
	conf := &EventConf{
		CallerLevel: 2,
	}
	for _, opt := range opts {
		opt(conf)
	}

	file, line, fn := utils.GetCallerInfo(conf.CallerLevel)
	event.FileName = file
	event.Line = line
	event.FnName = fn

	res, err := GetRedisClient().XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: 10000,
		Approx: true,
		ID:     "*",
		Values: event.ToRedisStream(),
	}).Result()
	if err != nil {
		logrus.Errorf("Failed to publish event to Redis stream %s: %v", stream, err)
	}
	logrus.Infof("Published event to Redis stream %s: %s", stream, res)
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
	event := &StreamEvent{
		TaskID: taskID,
	}

	if _, exists := values[consts.RdbEventTaskType]; exists {
		taskType, ok := values[consts.RdbEventTaskType].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventTaskType)
		}
		event.TaskType = consts.TaskType(taskType)
	}

	if _, exists := values[consts.RdbEventFn]; exists {
		fnName, ok := values[consts.RdbEventFn].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventFn)
		}
		event.FnName = fnName
	}

	if _, exists := values[consts.RdbEventPayload]; exists {
		payload, ok := values[consts.RdbEventPayload]
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventPayload)
		}

		event.Payload = payload
	}

	if _, exists := values[consts.RdbEventFileName]; exists {
		fileName, ok := values[consts.RdbEventFileName].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventTaskID)
		}

		event.FileName = fileName
	}

	if _, exists := values[consts.RdbEventLine]; exists {
		lineInt64, ok := values[consts.RdbEventLine].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventLine)
		}

		line, err := strconv.Atoi(lineInt64)
		if err != nil {
			return nil, fmt.Errorf("invalid line number: %w", err)
		}
		event.Line = line
	}

	if _, exists := values[consts.RdbEventName]; exists {
		eventName, ok := values[consts.RdbEventName].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventName)
		}
		event.EventName = consts.EventType(eventName)
	}

	return event, nil
}
