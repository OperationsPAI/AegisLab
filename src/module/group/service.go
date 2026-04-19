package group

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"aegis/consts"
	redisinfra "aegis/infra/redis"

	goredis "github.com/redis/go-redis/v9"
)

type Service struct {
	repo  *Repository
	redis *redisinfra.Gateway
}

func NewService(repo *Repository, redis *redisinfra.Gateway) *Service {
	return &Service{repo: repo, redis: redis}
}

func (s *Service) GetGroupStats(_ context.Context, req *GetGroupStatsReq) (*GroupStats, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	traces, err := s.repo.GetTracesByGroupID(req.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to query traces for group %s: %w", req.GroupID, err)
	}
	if len(traces) == 0 {
		return NewDefaultGroupStats(), nil
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

	traceStateMap := make(map[string][]TraceStatsItem, 4)
	for _, trace := range traces {
		stateName := consts.GetTraceStateName(trace.State)
		traceStateMap[stateName] = append(traceStateMap[stateName], *NewTraceStats(&trace))
	}

	return &GroupStats{
		TotalTraces:   len(traces),
		AvgDuration:   totalDuration / float64(len(durations)),
		MinDuration:   slices.Min(durations),
		MaxDuration:   slices.Max(durations),
		TraceStateMap: traceStateMap,
	}, nil
}

func (s *Service) NewGroupStreamProcessor(_ context.Context, groupID string) (*GroupStreamProcessor, error) {
	total, err := s.GetGroupTraceCount(groupID)
	if err != nil {
		return nil, err
	}
	return NewGroupStreamProcessor(int(total)), nil
}

func (s *Service) GetGroupTraceCount(groupID string) (int64, error) {
	total, err := s.repo.CountTracesByGroupID(groupID)
	if err != nil {
		return 0, fmt.Errorf("failed to count traces for group %s: %w", groupID, err)
	}
	if total == 0 {
		return 0, fmt.Errorf("the group %s does not exist", groupID)
	}
	return total, nil
}

func (s *Service) ReadGroupStreamMessages(ctx context.Context, streamKey, lastID string, count int64, block time.Duration) ([]goredis.XStream, error) {
	if lastID == "" {
		lastID = "0"
	}

	messages, err := s.redis.XRead(ctx, []string{streamKey, lastID}, count, block)
	if err != nil {
		return nil, fmt.Errorf("failed to read group stream messages: %w", err)
	}
	return messages, nil
}

type GroupStreamProcessor struct {
	totalTraces   int
	finishedCount int
}

func NewGroupStreamProcessor(totalTraces int) *GroupStreamProcessor {
	return &GroupStreamProcessor{
		totalTraces:   totalTraces,
		finishedCount: 0,
	}
}

func (p *GroupStreamProcessor) ProcessGroupMessage(msg goredis.XMessage) (*GroupStreamEvent, error) {
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

	lastEventStr, ok := msg.Values[consts.RdbEventTraceLastEvent].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid %s in group stream message", consts.RdbEventTraceLastEvent)
	}

	p.finishedCount++
	return &GroupStreamEvent{
		TraceID:   traceID,
		State:     consts.TraceState(stateInt),
		LastEvent: consts.EventType(lastEventStr),
	}, nil
}

func (p *GroupStreamProcessor) IsCompleted() bool {
	return p.totalTraces > 0 && p.finishedCount >= p.totalTraces
}
