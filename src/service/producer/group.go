package producer

import (
	"aegis/client"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// GetGroupStats retrieves statistics for a group of traces
func GetGroupStats(req *dto.GetGroupStatsReq) (*dto.GroupStats, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Query all traces belonging to this group
	traces, err := repository.GetTracesByGroupID(database.DB, req.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to query traces for group %s: %w", req.GroupID, err)
	}

	if len(traces) == 0 {
		return dto.NewDefaultGroupStats(), nil
	}

	durations := make([]float64, 0, len(traces))
	totalDuration := 0.0
	for _, trace := range traces {
		if trace.EndTime != nil {
			duration := trace.EndTime.Sub(trace.StartTime).Seconds()
			durations = append(durations, duration)
			totalDuration += duration
		}
	}

	traceStateMap := make(map[string][]dto.TraceStatsItem, 4)
	for _, trace := range traces {
		stateName := consts.GetTraceStateName(trace.State)
		if _, exists := traceStateMap[stateName]; !exists {
			traceStateMap[stateName] = make([]dto.TraceStatsItem, 0)
		}

		traceStateMap[stateName] = append(traceStateMap[stateName], *dto.NewTraceStats(&trace))
	}

	return &dto.GroupStats{
		TotalTraces:   len(traces),
		AvgDuration:   totalDuration / float64(len(durations)),
		MinDuration:   slices.Min(durations),
		MaxDuration:   slices.Max(durations),
		TraceStateMap: traceStateMap,
	}, nil
}

// ===================== Group Stream Service =====================

// GroupStreamProcessor tracks group-level trace completion for SSE streaming.
// It counts how many traces have reached terminal states (Completed/Failed)
// and determines when the group stream should be considered complete.
type GroupStreamProcessor struct {
	totalTraces   int
	finishedCount int
}

// NewGroupStreamProcessor creates a processor that tracks progress for a group
func NewGroupStreamProcessor(groupID string) (*GroupStreamProcessor, error) {
	total, err := repository.CountTracesByGroupID(database.DB, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to count traces for group %s: %w", groupID, err)
	}

	if total == 0 {
		return nil, fmt.Errorf("the group %s does not exist", groupID)
	}

	return &GroupStreamProcessor{
		totalTraces:   int(total),
		finishedCount: 0,
	}, nil
}

// ProcessGroupMessage processes a single group stream Redis message and returns a GroupStreamEvent
func (p *GroupStreamProcessor) ProcessGroupMessage(msg redis.XMessage) (*dto.GroupStreamEvent, error) {
	traceID, ok := msg.Values[consts.RdbEventTraceID].(string)
	if !ok || traceID == "" {
		return nil, fmt.Errorf("missing or invalid %s in group stream message", consts.RdbEventTraceID)
	}

	stateStr, ok := msg.Values[consts.RdbEventTraceState].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid %s in group stream message", consts.RdbEventTraceState)
	}
	stateInt, err := strconv.Atoi(stateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid trace state value %s in group stream message: %w", stateStr, err)
	}
	state := consts.TraceState(stateInt)

	lastEventStr, ok := msg.Values[consts.RdbEventTraceLastEvent].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid %s in group stream message", consts.RdbEventTraceLastEvent)
	}
	lastEvent := consts.EventType(lastEventStr)

	p.finishedCount++

	return &dto.GroupStreamEvent{
		TraceID:   traceID,
		State:     state,
		LastEvent: lastEvent,
	}, nil
}

// IsCompleted returns true when all traces in the group have reached terminal states
func (p *GroupStreamProcessor) IsCompleted() bool {
	return p.totalTraces > 0 && p.finishedCount >= p.totalTraces
}

// ReadGroupStreamMessages reads messages from the group-level Redis stream
func ReadGroupStreamMessages(ctx context.Context, streamKey, lastID string, count int64, block time.Duration) ([]redis.XStream, error) {
	if lastID == "" {
		lastID = "0"
	}

	messages, err := client.RedisXRead(ctx, []string{streamKey, lastID}, count, block)
	if err != nil {
		return nil, fmt.Errorf("failed to read group stream messages: %w", err)
	}
	return messages, nil
}
