package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/sirupsen/logrus"
)

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

		// Restart service related events
		case consts.EventNoNamespaceAvailable:
			restartWaitTimes++
		case consts.EventRestartServiceStarted:
			stageStartTime = eventTime
		case consts.EventRestartServiceCompleted:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(startTime).Seconds()
			stat.RestartDuration = eventTime.Sub(stageStartTime).Seconds()
			stat.RestartWaitTimes = restartWaitTimes
		case consts.EventRestartServiceFailed:
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
		case consts.EventTaskStatusUpdate:
			if payload, ok := event.Payload.(string); ok {
				pl := dto.InfoPayloadTemplate{}
				err := json.Unmarshal([]byte(payload), &pl)
				if err != nil {
					logrus.Errorf("Failed to unmarshal payload: %v", err)
					continue
				}
				if pl.Status == consts.TaskStatusError {
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

func AnalyzeTraces(ctx context.Context, req *dto.AnalyzeTracesReq) (*dto.TraceStats, error) {
	opts, err := req.TimeRangeQuery.Convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert time range query: %v", err)
	}

	datas, err := prepareTraceEventDatas(ctx, consts.TaskType(req.FirstTaskType), opts)
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
			return nil, fmt.Errorf("failed to get trace statistic with trace ID %s: %v", data.traceID, err)
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

func prepareTraceEventDatas(ctx context.Context, firstTaskType consts.TaskType, opts *dto.TimeFilterOptions) (chan traceEventData, error) {
	traceIDs, err := repository.ListTraceIDs(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get group to trace IDs map: %v", err)
	}

	startTime, endTime := opts.GetTimeRange()

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

			events, err := repository.GetTraceEvents(ctx, traceID, firstTaskType, startTime, endTime)
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
