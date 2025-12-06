package producer

import (
	"aegis/client"
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/utils"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	historicalCount = 1000
	historicalBlock = 1 * time.Second
)

var payloadTypeRegistry = map[consts.EventType]reflect.Type{
	// Algorithm execution events
	consts.EventAlgoRunSucceed: reflect.TypeOf(dto.ExecutionResult{}),
	consts.EventAlgoRunFailed:  reflect.TypeOf(dto.ExecutionResult{}),

	// Dataset Build events
	consts.EventDatapackBuildSucceed: reflect.TypeOf(dto.DatapackResult{}),
	consts.EventDatapackBuildFailed:  reflect.TypeOf(dto.DatapackResult{}),

	// Task status events
	consts.EventTaskStateUpdate: reflect.TypeOf(dto.InfoPayloadTemplate{}),

	// K8s Job events
	consts.EventJobSucceed: reflect.TypeOf(dto.JobMessage{}),
	consts.EventJobFailed:  reflect.TypeOf(dto.JobMessage{}),
}

// ===================== Stream Processor =====================

type StreamProcessor struct {
	hasIssues      bool
	isCompleted    bool
	detectorTaskID string
	algorithmMap   map[string]struct{}
	finishedCount  int
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
		payload, ok := streamEvent.Payload.(*dto.ExecutionResult)
		if !ok {
			return "", nil, fmt.Errorf("invalid payload type for task status update event: %T", streamEvent.Payload)
		}

		if payload.Algorithm.Name != config.GetString(consts.DetectorKey) {
			if len(sp.algorithmMap) != 0 {
				if _, exists := sp.algorithmMap[payload.Algorithm.Name]; exists {
					sp.finishedCount++
				}
			} else {
				sp.finishedCount++
			}
		}

	case consts.EventTaskStateUpdate:
		payload, ok := streamEvent.Payload.(*dto.InfoPayloadTemplate)
		if !ok {
			return "", nil, fmt.Errorf("invalid payload type for task status update event: %T", streamEvent.Payload)
		}

		state := consts.GetTaskStateByName(payload.State)
		if state == nil {
			return "", nil, fmt.Errorf("invalid task state name in payload: %s", payload.State)
		}

		switch *state {
		case consts.TaskError:
			sp.isCompleted = true
		case consts.TaskCompleted:
			if streamEvent.TaskType == consts.TaskTypeCollectResult {
				if sp.detectorTaskID != "" && streamEvent.TaskID == sp.detectorTaskID {
					sp.isCompleted = !sp.hasIssues || len(sp.algorithmMap) == 0
				} else {
					sp.isCompleted = len(sp.algorithmMap) == 0 || sp.finishedCount == len(sp.algorithmMap)
				}
			}
		}
	}

	return msg.ID, streamEvent, nil
}

func NewStreamProcessor(algorithms []dto.ContainerVersionItem) *StreamProcessor {
	algorithmMap := make(map[string]struct{}, len(algorithms))
	for _, algorithm := range algorithms {
		algorithmMap[algorithm.Name] = struct{}{}
	}

	return &StreamProcessor{
		hasIssues:      false,
		isCompleted:    false,
		detectorTaskID: "",
		algorithmMap:   algorithmMap,
		finishedCount:  0,
	}
}

// ===================== Trace Service =====================

// ListTraceIDs lists unique trace IDs from tasks within the specified time range
func ListTraceIDs(opts *dto.TimeFilterOptions) ([]string, error) {
	startTime, endTime := opts.GetTimeRange()

	tasks, err := repository.ListTasksByTimeRange(database.DB, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks from database: %v", err)
	}

	var traceIDs []string
	for _, task := range tasks {
		traceIDs = append(traceIDs, task.TraceID)
	}

	traceIDs = utils.ToUniqueSlice(traceIDs)
	return traceIDs, nil
}

// ===================== Trace Stream Service =====================

// GetTraceStreamEvents retrieves trace events from Redis stream based on the provided TraceQuery
func GetTraceStreamEvents(ctx context.Context, query *dto.TraceQuery) ([]*dto.StreamEvent, error) {
	streamKey := fmt.Sprintf(consts.StreamLogKey, query.TraceID)

	historicalMessages, err := client.RedisXRead(ctx, []string{streamKey, "0"}, historicalCount, historicalBlock)
	if err != nil {
		if err != redis.Nil {
			return nil, fmt.Errorf("failed to read stream messages from Redis: %v", err)
		}
		logrus.Warnf("No messages found in Redis stream %s for trace ID %s", streamKey, query.TraceID)
	}

	if len(historicalMessages) != 1 {
		return nil, fmt.Errorf("expected exactly one stream for trace %s, got %d", query.TraceID, len(historicalMessages))
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
			if streamEvent.TaskType != query.FirstTaskType {
				break
			}

			eventTime := time.UnixMilli(int64(streamEvent.TimeStamp))
			if !query.StartTime.IsZero() && eventTime.Before(query.StartTime) {
				break
			}
			if !query.EndTime.IsZero() && eventTime.After(query.EndTime) {
				break
			}
		}

		events = append(events, streamEvent)
	}

	return events, nil
}

// GetTraceStreamProcessor creates and initializes a stream processor for the given trace
func GetTraceStreamProcessor(ctx context.Context, traceID string) (*StreamProcessor, error) {
	tasks, total, err := repository.ListTasks(database.DB, 1000, 0, &dto.ListTaskFilters{
		TraceID: traceID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks: %w", err)
	}

	if total == 0 {
		return nil, fmt.Errorf("no tasks found for trace ID: %s", traceID)
	}

	headTask := tasks[0]
	var algorithms []dto.ContainerVersionItem

	if consts.TaskType(headTask.Type) == consts.TaskTypeRestartPedestal {
		if client.CheckCachedField(ctx, consts.InjectionAlgorithmsKey, headTask.GroupID) {
			err = client.GetHashField(ctx, consts.InjectionAlgorithmsKey, headTask.GroupID, &algorithms)
			if err != nil {
				return nil, fmt.Errorf("failed to get algorithms from Redis: %w", err)
			}
		}
	}

	return NewStreamProcessor(algorithms), nil
}

// ReadTraceStreamMessages reads messages from the trace stream
func ReadTraceStreamMessages(ctx context.Context, streamKey, lastID string, count int64, block time.Duration) ([]redis.XStream, error) {
	if lastID == "" {
		lastID = "0"
	}

	messages, err := client.RedisXRead(ctx, []string{streamKey, lastID}, count, block)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to read stream messages: %w", err)
	}
	return messages, err
}

// parseStreamEvent parses a Redis stream message values into a StreamEvent
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
		taskTypeFloat, ok := values[consts.RdbEventTaskType].(float64)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventTaskType)
		}
		event.TaskType = consts.TaskType(int(taskTypeFloat))
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
	payloadType, exists := payloadTypeRegistry[eventType]
	if !exists {
		return nil, false, nil
	}

	valuePtr := reflect.New(payloadType)

	if err := json.Unmarshal([]byte(payloadStr), valuePtr.Interface()); err != nil {
		return nil, true, fmt.Errorf("failed to unmarshal payload for event %s: %w", eventType, err)
	}

	return valuePtr.Interface(), true, nil
}
