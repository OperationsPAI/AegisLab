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
	"slices"
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
	consts.EventAlgoRunStarted: reflect.TypeFor[dto.ExecutionInfo](),
	consts.EventAlgoRunSucceed: reflect.TypeFor[dto.ExecutionResult](),
	consts.EventAlgoRunFailed:  reflect.TypeFor[dto.ExecutionResult](),

	// Dataset Build events
	consts.EventDatapackBuildStarted: reflect.TypeFor[dto.DatapackInfo](),
	consts.EventDatapackBuildSucceed: reflect.TypeFor[dto.DatapackResult](),
	consts.EventDatapackBuildFailed:  reflect.TypeFor[dto.DatapackResult](),

	// K8s Job events
	consts.EventJobSucceed: reflect.TypeFor[dto.JobMessage](),
	consts.EventJobFailed:  reflect.TypeFor[dto.JobMessage](),
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

// GetGroupStats retrieves statistics for a group of traces
func GetGroupStats(req *dto.GetGroupStatsReq) (*dto.GroupStats, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Query all traces belonging to this group
	traces, err := repository.GetTracesByGroupID(database.DB, req.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to query traces for group %s: %w", req.GroupID, err)
	}

	if len(traces) == 0 {
		return dto.NewDefaultGroupStats(), nil
	}

	durations := make([]float64, 0, len(traces))
	totalDuration := 0.0
	for _, trace := range traces {
		if trace.EndTime != nil {
			duration := trace.EndTime.Sub(trace.StartTime).Seconds()
			durations = append(durations, duration)
			totalDuration += duration
		}
	}

	traceStateMap := make(map[string][]dto.TraceStatsItem, 4)
	for _, trace := range traces {
		stateName := consts.GetTraceStateName(trace.State)
		if _, exists := traceStateMap[stateName]; !exists {
			traceStateMap[stateName] = make([]dto.TraceStatsItem, 0)
		}

		traceStateMap[stateName] = append(traceStateMap[stateName], *dto.NewTraceStats(&trace))
	}

	return &dto.GroupStats{
		TotalTraces:   len(traces),
		AvgDuration:   totalDuration / float64(len(durations)),
		MinDuration:   slices.Min(durations),
		MaxDuration:   slices.Max(durations),
		TraceStateMap: traceStateMap,
	}, nil
}

// ===================== Trace Stream Service =====================

type StreamProcessor struct {
	isCompleted   bool
	algorithmMap  map[string]struct{}
	finishedCount int
}

func NewStreamProcessor(algorithms []dto.ContainerVersionItem) *StreamProcessor {
	algorithmMap := make(map[string]struct{}, len(algorithms))
	for _, algorithm := range algorithms {
		algorithmMap[algorithm.ContainerName] = struct{}{}
	}

	return &StreamProcessor{
		isCompleted:   false,
		algorithmMap:  algorithmMap,
		finishedCount: 0,
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

	case consts.EventRestartPedestalFailed, consts.EventFaultInjectionFailed, consts.EventDatapackBuildFailed:
		sp.isCompleted = true

	case consts.EventDatapackNoAnomaly, consts.EventDatapackNoDetectorData:
		sp.isCompleted = true

	case consts.EventDatapackResultCollection:
		sp.isCompleted = len(sp.algorithmMap) == 0

	case consts.EventAlgoResultCollection, consts.EventAlgoRunFailed:
		payload, ok := streamEvent.Payload.(*dto.ExecutionResult)
		if !ok {
			return "", nil, fmt.Errorf("invalid payload type for task status update event: %T", streamEvent.Payload)
		}

		if payload.Algorithm != config.GetString(consts.DetectorKey) {
			if _, exists := sp.algorithmMap[payload.Algorithm]; exists {
				sp.finishedCount++
				if sp.finishedCount >= len(sp.algorithmMap) {
					sp.isCompleted = true
				}
			}
		}
	}

	return msg.ID, streamEvent, nil
}

// GetTraceStreamEvents retrieves trace events from Redis stream based on the provided TraceQuery
func GetTraceStreamEvents(ctx context.Context, query *dto.TraceQuery) ([]*dto.StreamEvent, error) {
	streamKey := fmt.Sprintf(consts.StreamLogKey, query.TraceID)

	historicalMessages, err := client.RedisXRead(ctx, []string{streamKey, "0"}, historicalCount, historicalBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to read trace stream messages: %w", err)
	}

	if len(historicalMessages) == 0 {
		logrus.Warnf("No messages found in Redis stream %s for trace ID %s", streamKey, query.TraceID)
		return nil, nil
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
	trace, err := repository.GetTraceByID(database.DB, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trace: %w", err)
	}

	var algorithms []dto.ContainerVersionItem
	if trace.Type == consts.TraceTypeFullPipeline {
		if client.CheckCachedField(ctx, consts.InjectionAlgorithmsKey, trace.GroupID) {
			err = client.GetHashField(ctx, consts.InjectionAlgorithmsKey, trace.GroupID, &algorithms)
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
	if err != nil {
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
		taskTypeStr, ok := values[consts.RdbEventTaskType].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventTaskType)
		}

		taskTypePtr := consts.GetTaskTypeByName(taskTypeStr)
		if taskTypePtr == nil {
			return nil, fmt.Errorf("unknown task type name: %s", taskTypeStr)
		}

		event.TaskType = *taskTypePtr
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
		if values[consts.RdbEventPayload] != nil {
			payloadStr, ok := values[consts.RdbEventPayload].(string)
			if !ok {
				return nil, fmt.Errorf(message, consts.RdbEventPayload)
			}

			payload, err := parsePayloadByEventType(event.EventName, payloadStr)
			if err != nil {
				return nil, fmt.Errorf(message, consts.RdbEventPayload)
			}
			event.Payload = payload
		}
	}

	return event, nil
}

// parsePayloadByEventType dynamically parses payload based on event type and
// returns the parsed payload as any, caller should do type assertion
func parsePayloadByEventType(eventType consts.EventType, payloadStr string) (any, error) {
	payloadType, exists := payloadTypeRegistry[eventType]
	if !exists {
		return nil, nil
	}

	valuePtr := reflect.New(payloadType)

	if err := json.Unmarshal([]byte(payloadStr), valuePtr.Interface()); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload for event %s: %w", eventType, err)
	}

	return valuePtr.Interface(), nil
}
