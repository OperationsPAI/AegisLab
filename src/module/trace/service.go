package trace

import (
	"context"
	"fmt"
	"time"

	"aegis/config"
	"aegis/consts"
	"aegis/dto"
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

func (s *Service) GetTrace(_ context.Context, traceID string) (*TraceDetailResp, error) {
	trace, err := s.repo.GetTraceByID(traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace: %w", err)
	}
	return NewTraceDetailResp(trace), nil
}

func (s *Service) ListTraces(_ context.Context, req *ListTraceReq) (*dto.ListResp[TraceResp], error) {
	if req == nil {
		return nil, fmt.Errorf("list traces request is nil")
	}
	limit, offset := req.ToGormParams()
	filterOptions := req.ToFilterOptions()
	traces, total, err := s.repo.ListTraces(limit, offset, filterOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list traces: %w", err)
	}
	items := make([]TraceResp, 0, len(traces))
	for i := range traces {
		items = append(items, *NewTraceResp(&traces[i]))
	}
	return &dto.ListResp[TraceResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) GetTraceStreamProcessor(ctx context.Context, traceID string) (*StreamProcessor, error) {
	algorithms, err := s.GetTraceStreamAlgorithms(ctx, traceID)
	if err != nil {
		return nil, err
	}
	return NewStreamProcessor(algorithms), nil
}

func (s *Service) GetTraceStreamAlgorithms(ctx context.Context, traceID string) ([]dto.ContainerVersionItem, error) {
	trace, err := s.repo.GetTraceByID(traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trace: %w", err)
	}

	var algorithms []dto.ContainerVersionItem
	if trace.Type == consts.TraceTypeFullPipeline && s.redis.CheckCachedField(ctx, consts.InjectionAlgorithmsKey, trace.GroupID) {
		if err := s.redis.GetHashField(ctx, consts.InjectionAlgorithmsKey, trace.GroupID, &algorithms); err != nil {
			return nil, fmt.Errorf("failed to get algorithms from Redis: %w", err)
		}
	}

	if len(algorithms) == 0 {
		return nil, nil
	}

	filtered := algorithms[:0]
	for _, algorithm := range algorithms {
		if algorithm.ContainerName != config.GetDetectorName() {
			filtered = append(filtered, algorithm)
		}
	}
	return filtered, nil
}

func (s *Service) ReadTraceStreamMessages(ctx context.Context, streamKey, lastID string, count int64, block time.Duration) ([]goredis.XStream, error) {
	if lastID == "" {
		lastID = "0"
	}

	messages, err := s.redis.XRead(ctx, []string{streamKey, lastID}, count, block)
	if err != nil {
		return nil, fmt.Errorf("failed to read stream messages: %w", err)
	}
	return messages, nil
}
