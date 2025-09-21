package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"aegis/consts"
	"aegis/database"
	"aegis/dto"

	"github.com/redis/go-redis/v9"
)

func GetTraceEvents(ctx context.Context, traceID string, firstTaskType consts.TaskType, startTime, endTime time.Time) ([]*dto.StreamEvent, error) {
	historicalMessages, err := ReadStreamEvents(ctx, fmt.Sprintf(consts.StreamLogKey, traceID), "0", 200, -1)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if len(historicalMessages) != 1 {
		return nil, fmt.Errorf("expected exactly one stream for trace %s, got %d", traceID, len(historicalMessages))
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

func ListTraceIDs(opts *dto.TimeFilterOptions) ([]string, error) {
	startTime, endTime := opts.GetTimeRange()

	var tasks []database.Task
	if err := database.DB.Model(&database.Task{}).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to query tasks from database: %v", err)
	}

	traceIDSet := make(map[string]struct{})
	for _, task := range tasks {
		traceIDSet[task.TraceID] = struct{}{}
	}

	var traceIDs []string
	for traceID := range traceIDSet {
		traceIDs = append(traceIDs, traceID)
	}

	return traceIDs, nil
}
