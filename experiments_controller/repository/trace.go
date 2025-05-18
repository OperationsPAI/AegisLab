package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/redis/go-redis/v9"
)

type TraceStatistic struct {
	DetectAnomaly    bool
	RestartWaitTimes int

	IntermediateFailed bool

	TotalDuration   float64
	RestartDuration float64
}

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

func GetTraceStatistic(ctx context.Context, traceId string) (*TraceStatistic, error) {
	events, err := GetTraceEvents(ctx, traceId)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("no events found for trace ID: %s", traceId)
	}

	totalStartTime := time.UnixMilli(int64(events[0].TimeStamp))
	totalEndTime := time.Time{}
	restartFormalBegin := time.Time{}
	restartFormalEnd := time.Time{}

	stat := &TraceStatistic{}
	for _, event := range events {
		switch event.EventName {
		case consts.EventDatasetNoAnomaly:
			stat.DetectAnomaly = false
			totalEndTime = time.UnixMilli(int64(event.TimeStamp))
		case consts.EventDatasetResultCollection:
			stat.DetectAnomaly = true
			totalEndTime = time.UnixMilli(int64(event.TimeStamp))
		case consts.EventDatasetNoConclusionFile:
			stat.IntermediateFailed = true
			totalEndTime = time.UnixMilli(int64(event.TimeStamp))
		case consts.EventNoNamespaceAvailable:
			stat.RestartWaitTimes++

		case consts.EventRestartServiceStarted:
			if restartFormalBegin.IsZero() {
				restartFormalBegin = time.UnixMilli(int64(event.TimeStamp))
			}
		case consts.EventRestartServiceCompleted:
			if restartFormalEnd.IsZero() {
				restartFormalEnd = time.UnixMilli(int64(event.TimeStamp))
			}
		case consts.EventRestartServiceFailed:
			stat.IntermediateFailed = true
		default:
			// logrus.WithField("event_name", event.EventName).Warn("Unknown event name")
		}
	}

	if !totalEndTime.IsZero() {
		stat.TotalDuration = totalEndTime.Sub(totalStartTime).Minutes()
	}
	if !restartFormalBegin.IsZero() && !restartFormalEnd.IsZero() {
		stat.RestartDuration = restartFormalEnd.Sub(restartFormalBegin).Minutes()
	}
	return stat, nil
}

func GetGroupToTraceIDsMap() (map[string][]string, error) {
	groupToTraceIDs := make(map[string][]string)

	type Result struct {
		GroupID string
		TraceID string
	}

	var results []Result

	err := database.DB.Model(&database.Task{}).
		Select("DISTINCT group_id, trace_id").
		Where("group_id <> ''").
		Where("trace_id <> ''").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get group to trace IDs map: %w", err)
	}

	seenTraceIDs := make(map[string]map[string]bool)

	for _, result := range results {
		if _, exists := seenTraceIDs[result.GroupID]; !exists {
			seenTraceIDs[result.GroupID] = make(map[string]bool)
			groupToTraceIDs[result.GroupID] = make([]string, 0)
		}

		if !seenTraceIDs[result.GroupID][result.TraceID] {
			groupToTraceIDs[result.GroupID] = append(groupToTraceIDs[result.GroupID], result.TraceID)
			seenTraceIDs[result.GroupID][result.TraceID] = true
		}
	}

	return groupToTraceIDs, nil
}
