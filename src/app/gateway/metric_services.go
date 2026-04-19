package gateway

import (
	"context"
	"slices"

	"aegis/consts"
	"aegis/dto"
	container "aegis/module/container"
	metric "aegis/module/metric"
)

type metricOrchestratorClient interface {
	Enabled() bool
	GetInjectionMetrics(context.Context, *metric.GetMetricsReq) (*metric.InjectionMetrics, error)
	GetExecutionMetrics(context.Context, *metric.GetMetricsReq) (*metric.ExecutionMetrics, error)
}

type metricResourceClient interface {
	Enabled() bool
	ListContainers(context.Context, *container.ListContainerReq) (*dto.ListResp[container.ContainerResp], error)
}

type remoteAwareMetricService struct {
	metric.HandlerService
	orchestrator metricOrchestratorClient
	resource     metricResourceClient
}

func (s remoteAwareMetricService) GetInjectionMetrics(ctx context.Context, req *metric.GetMetricsReq) (*metric.InjectionMetrics, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.GetInjectionMetrics(ctx, req)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

func (s remoteAwareMetricService) GetExecutionMetrics(ctx context.Context, req *metric.GetMetricsReq) (*metric.ExecutionMetrics, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.GetExecutionMetrics(ctx, req)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

func (s remoteAwareMetricService) GetAlgorithmMetrics(ctx context.Context, req *metric.GetMetricsReq) (*metric.AlgorithmMetrics, error) {
	if s.orchestrator == nil || !s.orchestrator.Enabled() {
		return nil, missingRemoteDependency("orchestrator-service")
	}
	if s.resource == nil || !s.resource.Enabled() {
		return nil, missingRemoteDependency("resource-service")
	}

	algorithms, err := s.listAlgorithmContainers(ctx, req)
	if err != nil {
		return nil, err
	}

	metrics := &metric.AlgorithmMetrics{
		Algorithms: make([]metric.AlgorithmMetricItem, 0, len(algorithms)),
	}
	for _, algorithm := range algorithms {
		algorithmID := algorithm.ID
		executionMetrics, err := s.orchestrator.GetExecutionMetrics(ctx, &metric.GetMetricsReq{
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			AlgorithmID: &algorithmID,
		})
		if err != nil || executionMetrics == nil || executionMetrics.TotalCount == 0 {
			continue
		}
		metrics.Algorithms = append(metrics.Algorithms, metric.AlgorithmMetricItem{
			AlgorithmID:    algorithm.ID,
			AlgorithmName:  algorithm.Name,
			ExecutionCount: executionMetrics.TotalCount,
			SuccessCount:   executionMetrics.SuccessCount,
			FailedCount:    executionMetrics.FailedCount,
			SuccessRate:    executionMetrics.SuccessRate,
			AvgDuration:    executionMetrics.AvgDuration,
		})
	}
	return metrics, nil
}

func (s remoteAwareMetricService) listAlgorithmContainers(ctx context.Context, req *metric.GetMetricsReq) ([]container.ContainerResp, error) {
	containerType := consts.ContainerTypeAlgorithm
	status := consts.CommonEnabled
	page := 1
	items := make([]container.ContainerResp, 0)

	for {
		resp, err := s.resource.ListContainers(ctx, &container.ListContainerReq{
			PaginationReq: dto.PaginationReq{
				Page: page,
				Size: consts.PageSizeXLarge,
			},
			Type:   &containerType,
			Status: &status,
		})
		if err != nil {
			return nil, err
		}
		items = append(items, resp.Items...)
		if resp.Pagination == nil || page >= resp.Pagination.TotalPages || len(resp.Items) == 0 {
			break
		}
		page++
	}

	if req.AlgorithmID == nil {
		return items, nil
	}
	index := slices.IndexFunc(items, func(item container.ContainerResp) bool {
		return item.ID == *req.AlgorithmID
	})
	if index < 0 {
		return []container.ContainerResp{}, nil
	}
	return []container.ContainerResp{items[index]}, nil
}
