package gatewayapp

import (
	"context"
	"slices"

	"aegis/consts"
	"aegis/dto"
	containermodule "aegis/module/container"
	metricmodule "aegis/module/metric"
)

type metricOrchestratorClient interface {
	Enabled() bool
	GetInjectionMetrics(context.Context, *metricmodule.GetMetricsReq) (*metricmodule.InjectionMetrics, error)
	GetExecutionMetrics(context.Context, *metricmodule.GetMetricsReq) (*metricmodule.ExecutionMetrics, error)
}

type metricResourceClient interface {
	Enabled() bool
	ListContainers(context.Context, *containermodule.ListContainerReq) (*dto.ListResp[containermodule.ContainerResp], error)
}

type remoteAwareMetricService struct {
	metricmodule.HandlerService
	orchestrator metricOrchestratorClient
	resource     metricResourceClient
}

func (s remoteAwareMetricService) GetInjectionMetrics(ctx context.Context, req *metricmodule.GetMetricsReq) (*metricmodule.InjectionMetrics, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.GetInjectionMetrics(ctx, req)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

func (s remoteAwareMetricService) GetExecutionMetrics(ctx context.Context, req *metricmodule.GetMetricsReq) (*metricmodule.ExecutionMetrics, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.GetExecutionMetrics(ctx, req)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

func (s remoteAwareMetricService) GetAlgorithmMetrics(ctx context.Context, req *metricmodule.GetMetricsReq) (*metricmodule.AlgorithmMetrics, error) {
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

	metrics := &metricmodule.AlgorithmMetrics{
		Algorithms: make([]metricmodule.AlgorithmMetricItem, 0, len(algorithms)),
	}
	for _, algorithm := range algorithms {
		algorithmID := algorithm.ID
		executionMetrics, err := s.orchestrator.GetExecutionMetrics(ctx, &metricmodule.GetMetricsReq{
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			AlgorithmID: &algorithmID,
		})
		if err != nil || executionMetrics == nil || executionMetrics.TotalCount == 0 {
			continue
		}
		metrics.Algorithms = append(metrics.Algorithms, metricmodule.AlgorithmMetricItem{
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

func (s remoteAwareMetricService) listAlgorithmContainers(ctx context.Context, req *metricmodule.GetMetricsReq) ([]containermodule.ContainerResp, error) {
	containerType := consts.ContainerTypeAlgorithm
	status := consts.CommonEnabled
	page := 1
	items := make([]containermodule.ContainerResp, 0)

	for {
		resp, err := s.resource.ListContainers(ctx, &containermodule.ListContainerReq{
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
	index := slices.IndexFunc(items, func(item containermodule.ContainerResp) bool {
		return item.ID == *req.AlgorithmID
	})
	if index < 0 {
		return []containermodule.ContainerResp{}, nil
	}
	return []containermodule.ContainerResp{items[index]}, nil
}
