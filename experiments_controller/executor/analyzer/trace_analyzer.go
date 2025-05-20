package analyzer

import (
	"context"
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

type DurationStats struct {
	Total       int
	AvgDuration float64
	MinDuration float64
	MaxDuration float64
}

type Statistics struct {
	Total       int
	AvgDuration float64
	MinDuration float64
	MaxDuration float64

	AnomalyTraceMap    map[string]string
	NoAnomalyTraceMap  map[string]string
	EndNameMap         map[consts.EventType]int
	StatusMetaMap      map[consts.EventType]DurationStats
	TraceStatusTimeMap map[string]map[consts.EventType]float64
	TraceRunningList   []string
	TraceErrorMap      map[string]any
}

func AnalyzeTrace(ctx context.Context, opts dto.TraceAnalyzeFilterOptions) (*Statistics, error) {
	// 初始化统计结构
	stats := &Statistics{
		MinDuration:        math.MaxFloat64,
		EndNameMap:         make(map[consts.EventType]int),
		StatusMetaMap:      map[consts.EventType]DurationStats{},
		TraceStatusTimeMap: make(map[string]map[consts.EventType]float64),
		TraceErrorMap:      make(map[string]any),
		AnomalyTraceMap:    make(map[string]string),
		NoAnomalyTraceMap:  make(map[string]string),
	}

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

	type traceResult struct {
		TraceID       string
		Stat          *repository.TraceStatistic
		IsAnomaly     bool
		IsFinished    bool
		IsFailed      bool
		StatusTimeMap map[consts.EventType]float64
		EndEventName  consts.EventType
		TotalDuration float64
		Error         error
		Payload       any
	}

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

			events, err := repository.GetTraceEvents(ctx, traceID)
			if err != nil {
				logrus.WithField("trace_id", traceID).Errorf("failed to get trace events: %v", err)
				return
			}
			if len(events) == 0 {
				logrus.WithField("trace_id", traceID).Warn("no events found")
				return
			}

			var filterEvents []*dto.StreamEvent
			for _, event := range events {
				if event.EventName != consts.EventTaskStatusUpdate {
					filterEvents = append(filterEvents, event)
				}
			}

			if len(filterEvents) == 0 {
				return
			}

			firstEventTime := time.UnixMilli(int64(filterEvents[0].TimeStamp))
			if !startTime.IsZero() && firstEventTime.Before(startTime) {
				logrus.WithField("trace_id", traceID).Debug("no valid events found")
				return
			}
			if !endTime.IsZero() && firstEventTime.After(endTime) {
				logrus.WithField("trace_id", traceID).Debug("event time is out of range")
				return
			}

			if opts.EventName != consts.EventType("") && opts.EventName != filterEvents[len(filterEvents)-1].EventName {
				logrus.WithField("trace_id", traceID).Debug("event name does not match")
				return
			}

			stat, err := repository.GetTraceStatistic(ctx, filterEvents)
			if err != nil {
				logrus.WithField("trace_id", traceID).Errorf("failed to get trace statistic: %v", err)
				return
			}

			result := traceResult{
				TraceID:       traceID,
				Stat:          stat,
				IsAnomaly:     stat.DetectAnomaly && stat.Finished,
				IsFinished:    stat.Finished,
				IsFailed:      stat.IntermediateFailed,
				StatusTimeMap: stat.StatusTimeMap,
				TotalDuration: stat.TotalDuration,
				Payload:       stat.Payload,
			}

			if stat.EndEvent != nil {
				result.EndEventName = stat.EndEvent.EventName
			}

			resultChan <- result
		}(traceID)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	var anomalyTraces, noAnomalyTraces []string
	totalDuration := 0.0
	validTraces := 0

	for result := range resultChan {
		stats.Total++

		if result.IsFinished {
			if result.IsAnomaly {
				anomalyTraces = append(anomalyTraces, result.TraceID)
			} else {
				noAnomalyTraces = append(noAnomalyTraces, result.TraceID)
			}
		} else {
			if result.IsFailed {
				stats.TraceErrorMap[result.TraceID] = result.Payload
			} else {
				stats.TraceRunningList = append(stats.TraceRunningList, result.TraceID)
			}
		}

		if result.EndEventName != "" && result.EndEventName != consts.EventTaskStarted {
			stats.EndNameMap[result.EndEventName]++
		}

		stats.TraceStatusTimeMap[result.TraceID] = result.StatusTimeMap

		for eventName, duration := range result.StatusTimeMap {
			if _, ok := stats.StatusMetaMap[eventName]; !ok {
				stats.StatusMetaMap[eventName] = DurationStats{
					Total:       0,
					AvgDuration: 0,
					MinDuration: math.MaxFloat64,
					MaxDuration: 0,
				}
			}

			current := stats.StatusMetaMap[eventName]

			current.Total++
			current.AvgDuration = current.AvgDuration + (duration-current.AvgDuration)/float64(current.Total)
			current.MinDuration = math.Min(current.MinDuration, duration)
			current.MaxDuration = math.Max(current.MaxDuration, duration)

			stats.StatusMetaMap[eventName] = current
		}

		if result.TotalDuration > 0 {
			totalDuration += result.TotalDuration
			validTraces++

			stats.MinDuration = math.Min(stats.MinDuration, result.TotalDuration)
			stats.MaxDuration = math.Max(stats.MaxDuration, result.TotalDuration)
		}
	}

	// 批量获取display config以减少数据库操作
	if len(anomalyTraces) > 0 {
		anomalyTraceMap, err := repository.GetDisplayConfigByTraceIDs(anomalyTraces)
		if err != nil {
			return nil, fmt.Errorf("failed to get display config for anomaly traces: %v", err)
		}
		stats.AnomalyTraceMap = anomalyTraceMap
	}

	if len(noAnomalyTraces) > 0 {
		noAnomalyTraceMap, err := repository.GetDisplayConfigByTraceIDs(noAnomalyTraces)
		if err != nil {
			return nil, fmt.Errorf("failed to get display config for noAnomaly traces: %v", err)
		}
		stats.NoAnomalyTraceMap = noAnomalyTraceMap
	}

	// 计算平均值
	if validTraces > 0 {
		stats.AvgDuration = totalDuration / float64(validTraces)
	} else {
		stats.MinDuration = 0
	}

	return stats, nil
}
