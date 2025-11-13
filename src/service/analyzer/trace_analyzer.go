package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"aegis/client"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
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

type traceEventData struct {
	traceID string
	events  []*dto.StreamEvent
}

type traceStatistic struct {
	IntermediateFailed bool
	Finished           bool
	DetectAnomaly      bool

	CurrentTaskType consts.TaskType
	LastEndEvent    consts.EventType // Add last end event type

	TotalDuration float64
	StatusTimeMap map[consts.TaskType]float64

	RestartDuration  float64
	RestartWaitTimes int
	InjectDuration   float64

	ErrorMsgs []string
}

func (d *traceEventData) computeTraceStatistic() (*traceStatistic, error) {
	stat := &traceStatistic{
		IntermediateFailed: false,
		Finished:           false,
		DetectAnomaly:      false,
		StatusTimeMap:      make(map[consts.TaskType]float64),
		ErrorMsgs:          make([]string, 0),
	}

	startTime := time.UnixMilli(int64(d.events[0].TimeStamp))
	var endTime time.Time
	var taskStartTime time.Time
	var stageStartTime time.Time
	restartWaitTimes := 0

	for _, event := range d.events {
		eventTime := time.UnixMilli(int64(event.TimeStamp))

		switch event.EventName {
		case consts.EventTaskStarted:
			taskStartTime = eventTime
			stat.CurrentTaskType = event.TaskType

		// Restart pedestal related events
		case consts.EventNoNamespaceAvailable:
			restartWaitTimes++
		case consts.EventRestartPedestalStarted:
			stageStartTime = eventTime
		case consts.EventRestartPedestalCompleted:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(startTime).Seconds()
			stat.RestartDuration = eventTime.Sub(stageStartTime).Seconds()
			stat.RestartWaitTimes = restartWaitTimes
		case consts.EventRestartPedestalFailed:
			stat.IntermediateFailed = true
			endTime = eventTime

		// Fault injection related events
		case consts.EventFaultInjectionStarted:
			stageStartTime = eventTime
		case consts.EventFaultInjectionCompleted:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(taskStartTime).Seconds()
			stat.InjectDuration = eventTime.Sub(stageStartTime).Seconds()
			stat.Finished = true
			stat.LastEndEvent = consts.EventFaultInjectionCompleted
			endTime = eventTime

		case consts.EventFaultInjectionFailed:
			stat.IntermediateFailed = true
			stat.LastEndEvent = consts.EventFaultInjectionFailed
			endTime = eventTime

		// Dataset building related events
		case consts.EventDatapackBuildSucceed:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(taskStartTime).Seconds()

		// Algorithm execution related events
		case consts.EventAlgoRunSucceed:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(taskStartTime).Seconds()

		// Result collection related events
		case consts.EventDatapackNoAnomaly:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(taskStartTime).Seconds()
			stat.Finished = true
			stat.DetectAnomaly = false
			stat.LastEndEvent = consts.EventDatapackNoAnomaly
			endTime = eventTime
		case consts.EventDatapackResultCollection:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(taskStartTime).Seconds()
			stat.Finished = true
			stat.DetectAnomaly = true
			stat.LastEndEvent = consts.EventDatapackResultCollection
			endTime = eventTime
		case consts.EventDatapackNoDetectorData:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(taskStartTime).Seconds()
			stat.IntermediateFailed = true
			stat.LastEndEvent = consts.EventDatapackNoDetectorData
			endTime = eventTime
		case consts.EventTaskStateUpdate:
			if payload, ok := event.Payload.(string); ok {
				pl := dto.InfoPayloadTemplate{}
				err := json.Unmarshal([]byte(payload), &pl)
				if err != nil {
					logrus.Errorf("Failed to unmarshal payload: %v", err)
					continue
				}
				if pl.State == consts.GetTaskStateName(consts.TaskError) {
					stat.IntermediateFailed = true
					stat.ErrorMsgs = append(stat.ErrorMsgs, pl.Msg)
				}
			}
		}
	}

	if !endTime.IsZero() {
		stat.TotalDuration = endTime.Sub(startTime).Seconds()
	}

	return stat, nil
}

// AnalyzeTraces analyzes trace events and computes statistics
func AnalyzeTraces(ctx context.Context, req *dto.AnalyzeTracesReq) (*dto.TraceStats, error) {
	opts, err := req.TimeRangeQuery.Convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert time range query: %w", err)
	}

	datas, err := prepareTraceEventDatas(ctx, *req.FirstTaskType, opts)
	if err != nil {
		return nil, err
	}

	stats := &dto.TraceStats{
		MinDuration:          math.MaxFloat64,
		EndCountMap:          make(map[consts.TaskType]map[string]int),
		TraceStatusTimeMap:   make(map[string]map[consts.TaskType]float64),
		FaultInjectionTraces: make([]string, 0),
	}

	totalDuration := 0.0
	validTracesNum := 0

	traceErrorMap := make(map[consts.TaskType]map[string]any)

	for data := range datas {
		stats.Total++
		stat, err := data.computeTraceStatistic()
		if err != nil {
			return nil, fmt.Errorf("failed to get trace statistic with trace ID %s: %w", data.traceID, err)
		}

		if _, exists := stats.EndCountMap[stat.CurrentTaskType]; !exists {
			stats.EndCountMap[stat.CurrentTaskType] = make(map[string]int)
		}

		if stat.Finished {
			stats.EndCountMap[stat.CurrentTaskType]["completed"]++
			stats.TraceCompletedList = append(stats.TraceCompletedList, data.traceID)

			// Check if it ends with fault injection event
			if stat.LastEndEvent == consts.EventFaultInjectionCompleted {
				stats.FaultInjectionTraces = append(stats.FaultInjectionTraces, data.traceID)
			}
		} else {
			if stat.IntermediateFailed {
				stats.EndCountMap[stat.CurrentTaskType]["failed"]++
				if _, exists := traceErrorMap[stat.CurrentTaskType]; !exists {
					traceErrorMap[stat.CurrentTaskType] = make(map[string]any)
				}
				traceErrorMap[stat.CurrentTaskType][data.traceID] = stat.ErrorMsgs
			} else {
				stats.EndCountMap[stat.CurrentTaskType]["running"]++
			}
		}

		stats.TraceStatusTimeMap[data.traceID] = stat.StatusTimeMap

		if stat.TotalDuration > 0 {
			totalDuration += stat.TotalDuration
			validTracesNum++

			stats.MinDuration = math.Min(stats.MinDuration, stat.TotalDuration)
			stats.MaxDuration = math.Max(stats.MaxDuration, stat.TotalDuration)
		}
	}
	if validTracesNum > 0 {
		stats.AvgDuration = totalDuration / float64(validTracesNum)
	} else {
		stats.MinDuration = 0
	}

	// Assign error information to TraceErrors field
	stats.TraceErrors = traceErrorMap

	return stats, nil
}

// prepareTraceEventDatas prepares trace event data for analysis
func prepareTraceEventDatas(ctx context.Context, firstTaskType consts.TaskType, opts *dto.TimeFilterOptions) (chan traceEventData, error) {
	startTime, endTime := opts.GetTimeRange()

	traceIDs, err := repository.ListTraceIDs(database.DB, &startTime, &endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to list trace IDs: %w", err)
	}

	resultChans := make(chan traceEventData, len(traceIDs))
	var wg sync.WaitGroup

	maxWorkers := min(runtime.NumCPU()*2, 8)
	semaphore := make(chan struct{}, maxWorkers)
	for _, traceID := range traceIDs {
		wg.Add(1)
		go func(traceID string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			events, err := getTraceEvents(ctx, traceID, firstTaskType, &startTime, &endTime)
			if err != nil {
				logrus.WithField("trace_id", traceID).Errorf("%s, failed to get trace events: %v", traceID, err)
				return
			}

			if len(events) == 0 {
				return
			}

			resultChans <- traceEventData{
				traceID: traceID,
				events:  events,
			}
		}(traceID)
	}

	go func() {
		wg.Wait()
		close(resultChans)
	}()

	return resultChans, nil
}

// getTraceEvents retrieves trace events from Redis Stream for a given trace ID and time range
func getTraceEvents(ctx context.Context, traceID string, firstTaskType consts.TaskType, startTime, endTime *time.Time) ([]*dto.StreamEvent, error) {
	historicalMessages, err := client.RedisXRead(ctx, []string{fmt.Sprintf(consts.StreamLogKey, traceID), "0"}, 200, -1)
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
			return nil, fmt.Errorf("invalid message ID format: %w", err)
		}

		if idx == 0 {
			if streamEvent.TaskType != firstTaskType {
				break
			}
			eventTime := time.UnixMilli(int64(streamEvent.TimeStamp))
			if startTime != nil && eventTime.Before(*startTime) {
				break
			}
			if endTime != nil && eventTime.After(*endTime) {
				break
			}
		}

		events = append(events, streamEvent)
	}

	return events, nil
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
		taskTypeStr, ok := values[consts.RdbEventTaskType].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.RdbEventTaskType)
		}
		taskType, err := strconv.Atoi(taskTypeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid task type: %w", err)
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
	payloadType, exists := payloadTypeRegistry[eventType]
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
