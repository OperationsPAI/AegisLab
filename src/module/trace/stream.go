package tracemodule

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"aegis/consts"
	"aegis/dto"

	"github.com/redis/go-redis/v9"
)

var payloadTypeRegistry = map[consts.EventType]reflect.Type{
	consts.EventAlgoRunStarted:       reflect.TypeFor[dto.ExecutionInfo](),
	consts.EventAlgoRunSucceed:       reflect.TypeFor[dto.ExecutionResult](),
	consts.EventAlgoRunFailed:        reflect.TypeFor[dto.ExecutionResult](),
	consts.EventDatapackBuildStarted: reflect.TypeFor[dto.DatapackInfo](),
	consts.EventDatapackBuildSucceed: reflect.TypeFor[dto.DatapackResult](),
	consts.EventDatapackBuildFailed:  reflect.TypeFor[dto.DatapackResult](),
	consts.EventJobSucceed:           reflect.TypeFor[dto.JobMessage](),
	consts.EventJobFailed:            reflect.TypeFor[dto.JobMessage](),
}

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

func (sp *StreamProcessor) ProcessMessageForSSE(msg redis.XMessage) (string, *dto.TraceStreamEvent, error) {
	streamEvent, err := parseStreamEvent(msg.ID, msg.Values)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse stream message value: %v", err)
	}

	switch streamEvent.EventName {
	case consts.EventImageBuildSucceed, consts.EventRestartPedestalFailed, consts.EventFaultInjectionFailed, consts.EventDatapackBuildFailed, consts.EventDatapackNoAnomaly, consts.EventDatapackNoDetectorData:
		sp.isCompleted = true
	case consts.EventDatapackResultCollection:
		sp.isCompleted = len(sp.algorithmMap) == 0
	case consts.EventAlgoResultCollection, consts.EventAlgoRunFailed:
		payload, ok := streamEvent.Payload.(*dto.ExecutionResult)
		if !ok {
			return "", nil, fmt.Errorf("invalid payload type for task status update event: %T", streamEvent.Payload)
		}

		if len(sp.algorithmMap) == 0 {
			sp.isCompleted = true
			break
		}
		if _, exists := sp.algorithmMap[payload.Algorithm]; exists {
			sp.finishedCount++
			if sp.finishedCount >= len(sp.algorithmMap) {
				sp.isCompleted = true
			}
		}
	}

	return msg.ID, streamEvent, nil
}

func parseStreamEvent(id string, values map[string]any) (*dto.TraceStreamEvent, error) {
	message := "missing or invalid key %s in redis stream message values"

	taskID, ok := values[consts.RdbEventTaskID].(string)
	if !ok || taskID == "" {
		return nil, fmt.Errorf(message, consts.RdbEventTaskID)
	}

	timeStamp, err := strconv.Atoi(strings.Split(id, "-")[0])
	if err != nil {
		return nil, err
	}

	event := &dto.TraceStreamEvent{
		TimeStamp: timeStamp,
		TaskID:    taskID,
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

	if _, exists := values[consts.RdbEventPayload]; exists && values[consts.RdbEventPayload] != nil {
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

	return event, nil
}

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
