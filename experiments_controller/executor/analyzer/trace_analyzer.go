package analyzer

import (
	"context"
	"math"
	"runtime"
	"sync"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/sirupsen/logrus"
)

type Statistics struct {
	Total       int     `json:"total"`
	AvgDuration float64 `json:"avg_duration"`
	MinDuration float64 `json:"min_duration"`
	MaxDuration float64 `json:"max_duration"`

	EndCountMap          map[consts.TaskType]map[string]int     `json:"end_count_map"`
	TraceStatusTimeMap   map[string]map[consts.TaskType]float64 `json:"trace_status_time_map"`
	TraceCompletedList   []string                               `json:"trace_completed_list"`
	FaultInjectionTraces []string                               `json:"fault_injection_traces"`
	TraceErrors          any                                    `json:"trace_errors"`
}

type traceResult struct {
	traceID string
	events  []*dto.StreamEvent
}

func AnalyzeTrace(ctx context.Context, opts dto.TraceAnalyzeFilterOptions) (*Statistics, error) {
	resultChan, err := getTraceResults(ctx, opts)
	if err != nil {
		return nil, err
	}

	stats := &Statistics{
		MinDuration:          math.MaxFloat64,
		EndCountMap:          make(map[consts.TaskType]map[string]int),
		TraceStatusTimeMap:   make(map[string]map[consts.TaskType]float64),
		FaultInjectionTraces: make([]string, 0),
	}

	totalDuration := 0.0
	validTracesNum := 0

	traceErrorMap := make(map[consts.TaskType]map[string]any)

	for result := range resultChan {
		stats.Total++
		stat, err := repository.GetTraceStatistic(ctx, result.events)
		if err != nil {
			logrus.WithField("trace_id", result.traceID).Errorf("failed to get trace statistic: %v", err)
			return nil, nil
		}

		if _, exists := stats.EndCountMap[stat.CurrentTaskType]; !exists {
			stats.EndCountMap[stat.CurrentTaskType] = make(map[string]int)
		}

		if stat.Finished {
			stats.EndCountMap[stat.CurrentTaskType]["completed"]++
			stats.TraceCompletedList = append(stats.TraceCompletedList, result.traceID)

			// 检查是否以故障注入事件结束
			if stat.LastEndEvent == consts.EventFaultInjectionCompleted {
				stats.FaultInjectionTraces = append(stats.FaultInjectionTraces, result.traceID)
			}
		} else {
			if stat.IntermediateFailed {
				stats.EndCountMap[stat.CurrentTaskType]["failed"]++
				if _, exists := traceErrorMap[stat.CurrentTaskType]; !exists {
					traceErrorMap[stat.CurrentTaskType] = make(map[string]any)
				}
				traceErrorMap[stat.CurrentTaskType][result.traceID] = stat.ErrorMsgs
			} else {
				stats.EndCountMap[stat.CurrentTaskType]["running"]++
			}
		}

		stats.TraceStatusTimeMap[result.traceID] = stat.StatusTimeMap

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

	// 将错误信息赋值给 TraceErrors 字段
	stats.TraceErrors = traceErrorMap

	return stats, nil
}

func GetCompletedMap(ctx context.Context, opts dto.TraceAnalyzeFilterOptions) (map[consts.EventType]any, error) {
	results, err := getTraceResults(ctx, opts)
	if err != nil {
		return nil, err
	}

	var anomalyTraces, noAnomalyTraces []string
	for result := range results {
		stat, err := repository.GetTraceStatistic(ctx, result.events)
		if err != nil {
			logrus.WithField("trace_id", result.traceID).Errorf("failed to get trace statistic: %v", err)
			return nil, nil
		}

		if stat.Finished {
			if stat.DetectAnomaly && stat.Finished {
				anomalyTraces = append(anomalyTraces, result.traceID)
			} else {
				noAnomalyTraces = append(noAnomalyTraces, result.traceID)
			}
		}
	}

	return map[consts.EventType]any{
		consts.EventDatasetResultCollection: anomalyTraces,
		consts.EventDatasetNoAnomaly:        noAnomalyTraces,
	}, nil
}

func getTraceResults(ctx context.Context, opts dto.TraceAnalyzeFilterOptions) (chan traceResult, error) {
	traceIDs, err := repository.GetAllTraceIDsFromRedis(ctx, opts)
	if err != nil {
		logrus.WithError(err).Error("failed to get group to trace IDs map")
		return nil, err
	}

	startTime, endTime := opts.GetTimeRange()

	resultChan := make(chan traceResult, len(traceIDs))
	var wg sync.WaitGroup

	maxWorkers := min(runtime.NumCPU()*2, 8)
	semaphore := make(chan struct{}, maxWorkers)
	for _, traceID := range traceIDs {
		wg.Add(1)
		go func(traceID string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			events, err := repository.GetTraceEvents(ctx, traceID, opts.FirstTaskType, startTime, endTime)
			if err != nil {
				logrus.WithField("trace_id", traceID).Errorf("%s, failed to get trace events: %v", traceID, err)
				return
			}

			if len(events) == 0 {
				return
			}

			resultChan <- traceResult{
				traceID: traceID,
				events:  events,
			}
		}(traceID)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan, nil
}
