package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/LGU-SE-Internal/rcabench/client"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type TraceStatistic struct {
	IntermediateFailed bool
	Finished           bool
	DetectAnomaly      bool

	CurrentTaskType consts.TaskType
	LastEndEvent    consts.EventType // 添加最后一个结束事件类型

	TotalDuration float64
	StatusTimeMap map[consts.TaskType]float64

	RestartDuration  float64
	RestartWaitTimes int
	InjectDuration   float64

	ErrorMsgs []string
}

func GetTraceEvents(ctx context.Context, traceId string, firstTaskType consts.TaskType, startTime, endTime time.Time) ([]*dto.StreamEvent, error) {
	historicalMessages, err := ReadStreamEvents(ctx, fmt.Sprintf(consts.StreamLogKey, traceId), "0", 200, -1)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if len(historicalMessages) != 1 {
		return nil, fmt.Errorf("expected exactly one stream for trace %s, got %d", traceId, len(historicalMessages))
	}

	events := make([]*dto.StreamEvent, 0)
	stream := historicalMessages[0]
	for idx, msg := range stream.Messages {
		streamEvent, err := parseEventFromValues(msg.Values)
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

func GetTraceStatistic(ctx context.Context, events []*dto.StreamEvent) (*TraceStatistic, error) {
	stat := &TraceStatistic{
		IntermediateFailed: false,
		Finished:           false,
		DetectAnomaly:      false,
		StatusTimeMap:      make(map[consts.TaskType]float64),
		ErrorMsgs:          make([]string, 0),
	}

	startTime := time.UnixMilli(int64(events[0].TimeStamp))
	var endTime time.Time
	var taskStartTime time.Time
	var stageStartTime time.Time
	restartWaitTimes := 0

	for _, event := range events {
		eventTime := time.UnixMilli(int64(event.TimeStamp))

		switch event.EventName {
		case consts.EventTaskStarted:
			taskStartTime = eventTime
			stat.CurrentTaskType = event.TaskType

		// 重启服务相关事件
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

		// 故障注入相关事件
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

		// 数据集构建相关事件
		case consts.EventDatasetBuildSucceed:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(taskStartTime).Seconds()

		// 算法运行相关事件
		case consts.EventAlgoRunSucceed:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(taskStartTime).Seconds()

		// 结果收集相关事件
		case consts.EventDatasetNoAnomaly:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(taskStartTime).Seconds()
			stat.Finished = true
			stat.DetectAnomaly = false
			stat.LastEndEvent = consts.EventDatasetNoAnomaly
			endTime = eventTime
		case consts.EventDatasetResultCollection:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(taskStartTime).Seconds()
			stat.Finished = true
			stat.DetectAnomaly = true
			stat.LastEndEvent = consts.EventDatasetResultCollection
			endTime = eventTime
		case consts.EventDatasetNoConclusionFile:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(taskStartTime).Seconds()
			stat.IntermediateFailed = true
			stat.LastEndEvent = consts.EventDatasetNoConclusionFile
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

func GetAllTraceIDsFromRedis(ctx context.Context, opts dto.TraceAnalyzeFilterOptions) ([]string, error) {
	var cursor uint64
	var traceIDs []string

	for {
		keys, nextCursor, err := client.GetRedisClient().Scan(ctx, cursor, "trace:*:log", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan Redis keys: %v", err)
		}

		for _, key := range keys {
			parts := strings.Split(key, ":")
			if len(parts) == 3 && parts[0] == "trace" && parts[2] == "log" {
				if utils.IsValidUUID(parts[1]) {
					traceIDs = append(traceIDs, parts[1])
				}
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return traceIDs, nil
}
