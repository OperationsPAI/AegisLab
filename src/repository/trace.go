package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"aegis/client"
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/utils"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type EventConf struct {
	CallerLevel int
}

type EventConfOption func(*EventConf)

type StreamProcessor struct {
	hasIssues      bool
	isCompleted    bool
	detectorTaskID string
	algorithms     map[string]struct{}
	finishedCount  int
}

func WithCallerLevel(level int) func(*EventConf) {
	return func(c *EventConf) {
		c.CallerLevel = level
	}
}

func PublishEvent(ctx context.Context, stream string, event dto.StreamEvent, opts ...EventConfOption) {
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

	_, err := client.GetRedisClient().XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: 10000,
		Approx: true,
		ID:     "*",
		Values: event.ToRedisStream(),
	}).Result()
	if err != nil {
		logrus.Errorf("Failed to publish event to Redis stream %s: %v", stream, err)
	}
}

func ReadStreamEvents(ctx context.Context, stream string, lastID string, count int64, block time.Duration) ([]redis.XStream, error) {
	if lastID == "" {
		lastID = "0"
	}

	return client.GetRedisClient().XRead(ctx, &redis.XReadArgs{
		Streams: []string{stream, lastID},
		Count:   count,
		Block:   block,
	}).Result()
}

func NewStreamProcessor(algorithmItems []dto.AlgorithmItem) *StreamProcessor {
	algorithms := make(map[string]struct{}, len(algorithmItems))
	for _, item := range algorithmItems {
		algorithms[item.Name] = struct{}{}
	}

	return &StreamProcessor{
		hasIssues:      false,
		isCompleted:    false,
		detectorTaskID: "",
		algorithms:     algorithms,
		finishedCount:  0,
	}
}

func (sp *StreamProcessor) IsCompleted() bool {
	return sp.isCompleted
}

func (sp *StreamProcessor) ProcessMessageForSSE(msg redis.XMessage) (string, *dto.StreamEvent, error) {
	streamEvent, err := parseStreamEvent(msg.Values)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse stream message value: %v", err)
	}

	switch streamEvent.EventName {
	case consts.EventImageBuildSucceed:
		sp.isCompleted = true

	case consts.EventDatapackNoAnomaly, consts.EventDatapackNoDetectorData:
		sp.detectorTaskID = streamEvent.TaskID
		sp.hasIssues = false

	case consts.EventDatapackResultCollection:
		sp.detectorTaskID = streamEvent.TaskID
		sp.hasIssues = true

	case consts.EventAlgoRunSucceed, consts.EventAlgoRunFailed:
		payload, ok := streamEvent.Payload.(*dto.ExecutionOptions)
		if !ok {
			return "", nil, fmt.Errorf("invalid payload type for task status update event: %T", streamEvent.Payload)
		}

		if payload.Algorithm.Name != config.GetString("algo.detector") {
			if len(sp.algorithms) != 0 {
				if _, exists := sp.algorithms[payload.Algorithm.Name]; exists {
					sp.finishedCount++
				}
			} else {
				sp.finishedCount++
			}
		}

	case consts.EventTaskStatusUpdate:
		payload, ok := streamEvent.Payload.(*dto.InfoPayloadTemplate)
		if !ok {
			return "", nil, fmt.Errorf("invalid payload type for task status update event: %T", streamEvent.Payload)
		}

		switch payload.Status {
		case consts.TaskStatusError:
			sp.isCompleted = true
		case consts.TaskStatusCompleted:
			if streamEvent.TaskType == consts.TaskTypeCollectResult {
				if sp.detectorTaskID != "" && streamEvent.TaskID == sp.detectorTaskID {
					sp.isCompleted = !sp.hasIssues || len(sp.algorithms) == 0
				} else {
					sp.isCompleted = len(sp.algorithms) == 0 || sp.finishedCount == len(sp.algorithms)
				}
			}
		}
	}

	return msg.ID, streamEvent, nil
}

func GetTraceEvents(ctx context.Context, traceID string, firstTaskType consts.TaskType, startTime, endTime time.Time) ([]*dto.StreamEvent, error) {
	historicalMessages, err := ReadStreamEvents(ctx, fmt.Sprintf(consts.StreamLogKey, traceID), "0", 200, -1)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if len(historicalMessages) != 1 {
		return nil, fmt.Errorf("expected exactly one stream for trace %s, got %d", traceID, len(historicalMessages))
	}

	events := make([]*dto.StreamEvent, 0)
	stream := historicalMessages[0]
	for idx, msg := range stream.Messages {
		streamEvent, err := parseStreamEvent(msg.Values)
		if err != nil {
			return nil, err
		}

		streamEvent.TimeStamp, err = strconv.Atoi(strings.Split(msg.ID, "-")[0])
		if err != nil {
			return nil, err
		}

		if idx == 0 {
			if firstTaskType != consts.TaskType("") && streamEvent.TaskType != firstTaskType {
				break
			}

			eventTime := time.UnixMilli(int64(streamEvent.TimeStamp))
			if !startTime.IsZero() && eventTime.Before(startTime) {
				break
			}
			if !endTime.IsZero() && eventTime.After(endTime) {
				break
			}
		}

		events = append(events, streamEvent)
	}

	return events, nil
}

func ListTraceIDs(opts *dto.TimeFilterOptions) ([]string, error) {
	startTime, endTime := opts.GetTimeRange()

	var tasks []database.Task
	if err := database.DB.Model(&database.Task{}).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to query tasks from database: %v", err)
	}

	traceIDSet := make(map[string]struct{})
	for _, task := range tasks {
		traceIDSet[task.TraceID] = struct{}{}
	}

	var traceIDs []string
	for traceID := range traceIDSet {
		traceIDs = append(traceIDs, traceID)
	}

	return traceIDs, nil
}

// parseStreamEvent parses StreamEvent from Redis Stream message values
func parseStreamEvent(values map[string]any) (*dto.StreamEvent, error) {
	message := "missing or invalid key %s in redis stream message values"

	taskID, ok := values[consts.RdbEventTaskID].(string)
	if !ok || taskID == "" {
		return nil, fmt.Errorf(message, consts.RdbEventTaskID)
	}
	event := &dto.StreamEvent{
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

	if _, exists := values[consts.RdbEventPayload]; exists {
		payloadStr, ok := values[consts.RdbEventPayload].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventPayload)
		}

		payload, _, err := parsePayloadByEventType(event.EventName, payloadStr)
		if err != nil {
			return nil, fmt.Errorf(message, consts.RdbEventPayload)
		}

		event.Payload = payload
	}

	return event, nil
}

// parsePayloadByEventType dynamically parses payload based on event type and
// returns the parsed payload as any, caller should do type assertion
func parsePayloadByEventType(eventType consts.EventType, payloadStr string) (any, bool, error) {
	payloadType, exists := dto.PayloadTypeRegistry[eventType]
	if !exists {
		return nil, false, nil
	}

	// Create new instance (pointer)
	valuePtr := reflect.New(payloadType)

	// Unmarshal JSON into the instance
	if err := json.Unmarshal([]byte(payloadStr), valuePtr.Interface()); err != nil {
		return nil, true, fmt.Errorf("failed to unmarshal payload for event %s: %w", eventType, err)
	}

	return valuePtr.Interface(), true, nil
}
