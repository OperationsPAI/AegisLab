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
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type TraceStatistic struct {
	IntermediateFailed bool
	Finished           bool
	DetectAnomaly      bool

	CurrentTaskType consts.TaskType

	TotalDuration float64
	StatusTimeMap map[consts.TaskType]float64

	// Payload any 移除 payload，语义不明确，引起误导，且有重复赋值。
	RestartDuration  float64
	RestartWaitTimes int
	InjectDuration   float64

	ErrorMsgs []string
}

// TODO@wangrui: 增加判断机制 1. 在筛选 traceid 的时候根据第一条 event 的时间来筛选，不需要取所有的 traceid； 2. 需要规定 stream key 的类型，例如 uuid 格式的，是我们能取的，因为还有些 ts0 这样的 key 在存储 namespace 的信息
func GetTraceEvents(ctx context.Context, traceId string) ([]*dto.StreamEvent, error) {
	historicalMessages, err := ReadStreamEvents(ctx, fmt.Sprintf(consts.StreamLogKey, traceId), "0", 200, -1)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	events := make([]*dto.StreamEvent, 0)
	for _, stream := range historicalMessages {
		for _, msg := range stream.Messages {
			streamEvent, err := parseEventFromValues(msg.Values)
			if err != nil {
				return nil, err
			}

			streamEvent.TimeStamp, err = strconv.Atoi(strings.Split(msg.ID, "-")[0])
			if err != nil {
				return nil, err
			}

			events = append(events, streamEvent)
		}
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

		case consts.EventFaultInjectionFailed:
			stat.IntermediateFailed = true
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
			endTime = eventTime
		case consts.EventDatasetResultCollection:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(taskStartTime).Seconds()
			stat.Finished = true
			stat.DetectAnomaly = true
			endTime = eventTime
		case consts.EventDatasetNoConclusionFile:
			stat.StatusTimeMap[event.TaskType] = eventTime.Sub(taskStartTime).Seconds()
			stat.IntermediateFailed = true
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

func GetAllTraceIDsFromRedis(ctx context.Context) ([]string, error) {
	// 使用简单的SCAN命令遍历键
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
				traceIDs = append(traceIDs, parts[1])
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return traceIDs, nil
}
