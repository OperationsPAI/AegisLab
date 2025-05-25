package analyzer

import (
	"context"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Statistics struct {
	Total       int     `json:"total"`
	AvgDuration float64 `json:"avg_duration"`
	MinDuration float64 `json:"min_duration"`
	MaxDuration float64 `json:"max_duration"`

	EndCountMap        map[consts.TaskType]map[string]int     `json:"end_count_map"`
	TraceStatusTimeMap map[string]map[consts.TaskType]float64 `json:"trace_status_time_map"`
	TraceCompletedList []string                               `json:"trace_completed_list"`
	TraceErrors        any                                    `json:"trace_errors"`
}

type traceResult struct {
	TraceID       string
	Stat          *repository.TraceStatistic
	IsAnomaly     bool
	IsCompleted   bool
	IsFailed      bool
	StatusTimeMap map[consts.TaskType]float64
	EndTaskType   consts.TaskType
	TotalDuration float64
	Error         error
	ErrorMsgs     []string
}

func AnalyzeTrace(ctx context.Context, opts dto.TraceAnalyzeFilterOptions) (*Statistics, error) {
	resultChan, err := getTraceResults(ctx, opts)
	if err != nil {
		return nil, err
	}

	stats := &Statistics{
		MinDuration:        math.MaxFloat64,
		EndCountMap:        make(map[consts.TaskType]map[string]int),
		TraceStatusTimeMap: make(map[string]map[consts.TaskType]float64),
	}

	// 收集结果
	totalDuration := 0.0
	validTracesNum := 0

	traceErrorMap := make(map[consts.TaskType]map[string]any)

	for result := range resultChan {
		stats.Total++

		if _, exists := stats.EndCountMap[result.EndTaskType]; !exists {
			stats.EndCountMap[result.EndTaskType] = make(map[string]int)
		}

		if result.IsCompleted {
			stats.EndCountMap[result.EndTaskType]["completed"]++
			stats.TraceCompletedList = append(stats.TraceCompletedList, result.TraceID)
		} else {
			if result.IsFailed {
				stats.EndCountMap[result.EndTaskType]["failed"]++
				if _, exists := traceErrorMap[result.EndTaskType]; !exists {
					traceErrorMap[result.EndTaskType] = make(map[string]any)
				}
				traceErrorMap[result.EndTaskType][result.TraceID] = result.ErrorMsgs
			} else {
				stats.EndCountMap[result.EndTaskType]["running"]++
			}
		}

		stats.TraceStatusTimeMap[result.TraceID] = result.StatusTimeMap

		if result.TotalDuration > 0 {
			totalDuration += result.TotalDuration
			validTracesNum++

			stats.MinDuration = math.Min(stats.MinDuration, result.TotalDuration)
			stats.MaxDuration = math.Max(stats.MaxDuration, result.TotalDuration)
		}
	}

	if opts.ErrorStruct == dto.ErrorStructMap {
		stats.TraceErrors = traceErrorMap
	} else {
		// TODO：不返回这个，没必要，参数也没有说明
		// traceErrorList := make([]string, 0, len(traceErrorMap))
		// for traceID := range traceErrorMap {
		// 	traceErrorList = append(traceErrorList, traceID)
		// }

		// stats.TraceErrors = traceErrorList
	}

	// 计算平均值
	if validTracesNum > 0 {
		stats.AvgDuration = totalDuration / float64(validTracesNum)
	} else {
		stats.MinDuration = 0
	}

	return stats, nil
}

func GetCompletedMap(ctx context.Context, opts dto.TraceAnalyzeFilterOptions) (map[consts.EventType]any, error) {
	resultChan, err := getTraceResults(ctx, opts)
	if err != nil {
		return nil, err
	}

	var anomalyTraces, noAnomalyTraces []string
	for result := range resultChan {
		if result.IsCompleted {
			if result.IsAnomaly {
				anomalyTraces = append(anomalyTraces, result.TraceID)
			} else {
				noAnomalyTraces = append(noAnomalyTraces, result.TraceID)
			}
		}
	}

	return map[consts.EventType]any{
		consts.EventDatasetResultCollection: anomalyTraces,
		consts.EventDatasetNoAnomaly:        noAnomalyTraces,
	}, nil
}

func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// TODO@wangrui:  此函数抽象不合理。 这里应该只做数据提取的操作。然后在外部 analyze trace 的时候统一统计信息。
func getTraceResults(ctx context.Context, opts dto.TraceAnalyzeFilterOptions) (chan traceResult, error) {
	traceIDs, err := repository.GetAllTraceIDsFromRedis(ctx)
	if err != nil {
		logrus.WithError(err).Error("failed to get group to trace IDs map")
		return nil, err
	}

	now := time.Now()
	var startTime, endTime time.Time
	if opts.UseCustomRange {
		startTime = opts.CustomStartTime
		endTime = opts.CustomEndTime
	} else if opts.Lookback != 0 {
		endTime = now
		startTime = now.Add(-opts.Lookback)
	} else {
		endTime = now
		startTime = time.Time{}
	}

	resultChan := make(chan traceResult, len(traceIDs))
	var wg sync.WaitGroup

	maxWorkers := min(runtime.NumCPU()*2, 8)
	semaphore := make(chan struct{}, maxWorkers)

	for _, traceID := range traceIDs {
		if !isValidUUID(traceID) {
			continue
		}
		wg.Add(1)
		go func(traceID string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			events, err := repository.GetTraceEvents(ctx, traceID)
			if err != nil {
				logrus.WithField("trace_id", traceID).Errorf("%s, failed to get trace events: %v", traceID, err)
				return
			}
			if len(events) == 0 {
				logrus.WithField("trace_id", traceID).Warn("no events found")
				return
			}

			if opts.FirstTaskType != consts.TaskType("") && events[0].TaskType != opts.FirstTaskType {
				return
			}

			firstEventTime := time.UnixMilli(int64(events[0].TimeStamp))
			if !startTime.IsZero() && firstEventTime.Before(startTime) {
				logrus.WithField("trace_id", traceID).Debug("no valid events found")
				return
			}
			if !endTime.IsZero() && firstEventTime.After(endTime) {
				logrus.WithField("trace_id", traceID).Debug("event time is out of range")
				return
			}
			// pp.Println(events)
			// TODO：即这里，应该提取到外部。收到一个 trace 信息的所有 event 之后一次性把他处理成一个统计结构。
			stat, err := repository.GetTraceStatistic(ctx, events)
			if err != nil {
				logrus.WithField("trace_id", traceID).Errorf("failed to get trace statistic: %v", err)
				return
			}

			result := traceResult{
				TraceID:       traceID,
				Stat:          stat,
				IsAnomaly:     stat.DetectAnomaly && stat.Finished,
				IsCompleted:   stat.Finished,
				IsFailed:      stat.IntermediateFailed,
				StatusTimeMap: stat.StatusTimeMap,
				TotalDuration: stat.TotalDuration,
				ErrorMsgs:     stat.ErrorMsgs,
				EndTaskType:   stat.CurrentTaskType,
			}

			resultChan <- result
		}(traceID)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan, nil
}
