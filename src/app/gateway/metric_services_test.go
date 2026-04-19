package gateway

import (
	"context"
	"testing"
	"time"

	"aegis/consts"
	"aegis/dto"
	container "aegis/module/container"
	metric "aegis/module/metric"
)

type orchestratorMetricClientStub struct {
	injectionReqs []*metric.GetMetricsReq
	executionReqs []*metric.GetMetricsReq
	injection     *metric.InjectionMetrics
	execution     map[int]metric.ExecutionMetrics
	enabled       bool
}

func (s *orchestratorMetricClientStub) Enabled() bool {
	return s.enabled
}

func (s *orchestratorMetricClientStub) GetInjectionMetrics(_ context.Context, req *metric.GetMetricsReq) (*metric.InjectionMetrics, error) {
	s.injectionReqs = append(s.injectionReqs, req)
	return s.injection, nil
}

func (s *orchestratorMetricClientStub) GetExecutionMetrics(_ context.Context, req *metric.GetMetricsReq) (*metric.ExecutionMetrics, error) {
	s.executionReqs = append(s.executionReqs, req)
	if req != nil && req.AlgorithmID != nil {
		if metric, ok := s.execution[*req.AlgorithmID]; ok {
			result := metric
			return &result, nil
		}
	}
	return &metric.ExecutionMetrics{}, nil
}

type resourceMetricClientStub struct {
	responses []*dto.ListResp[container.ContainerResp]
	enabled   bool
	calls     int
}

func (s *resourceMetricClientStub) Enabled() bool {
	return s.enabled
}

func (s *resourceMetricClientStub) ListContainers(_ context.Context, _ *container.ListContainerReq) (*dto.ListResp[container.ContainerResp], error) {
	idx := s.calls
	s.calls++
	if idx >= len(s.responses) {
		return &dto.ListResp[container.ContainerResp]{}, nil
	}
	return s.responses[idx], nil
}

func TestRemoteAwareMetricServiceGetInjectionMetricsRemoteOnly(t *testing.T) {
	service := remoteAwareMetricService{}
	_, err := service.GetInjectionMetrics(context.Background(), &metric.GetMetricsReq{})
	if err == nil {
		t.Fatal("GetInjectionMetrics() error = nil, want missing dependency")
	}
}

func TestRemoteAwareMetricServiceGetAlgorithmMetricsBuildsFromRemoteSources(t *testing.T) {
	start := time.Now().Add(-time.Hour)
	end := time.Now()
	orchestrator := &orchestratorMetricClientStub{
		enabled: true,
		execution: map[int]metric.ExecutionMetrics{
			1: {TotalCount: 3, SuccessCount: 2, FailedCount: 1, SuccessRate: 66.7, AvgDuration: 12.5},
			2: {TotalCount: 0},
			3: {TotalCount: 5, SuccessCount: 5, FailedCount: 0, SuccessRate: 100, AvgDuration: 8},
		},
	}
	resource := &resourceMetricClientStub{
		enabled: true,
		responses: []*dto.ListResp[container.ContainerResp]{
			{
				Items: []container.ContainerResp{
					{ID: 1, Name: "algo-a", Type: consts.GetContainerTypeName(consts.ContainerTypeAlgorithm)},
					{ID: 2, Name: "algo-b", Type: consts.GetContainerTypeName(consts.ContainerTypeAlgorithm)},
				},
				Pagination: &dto.PaginationInfo{Page: 1, Size: 100, Total: 3, TotalPages: 2},
			},
			{
				Items: []container.ContainerResp{
					{ID: 3, Name: "algo-c", Type: consts.GetContainerTypeName(consts.ContainerTypeAlgorithm)},
				},
				Pagination: &dto.PaginationInfo{Page: 2, Size: 100, Total: 3, TotalPages: 2},
			},
		},
	}

	service := remoteAwareMetricService{
		orchestrator: orchestrator,
		resource:     resource,
	}

	resp, err := service.GetAlgorithmMetrics(context.Background(), &metric.GetMetricsReq{
		StartTime: &start,
		EndTime:   &end,
	})
	if err != nil {
		t.Fatalf("GetAlgorithmMetrics() error = %v", err)
	}
	if len(resp.Algorithms) != 2 {
		t.Fatalf("GetAlgorithmMetrics() algorithm count = %d, want 2", len(resp.Algorithms))
	}
	if resp.Algorithms[0].AlgorithmName != "algo-a" || resp.Algorithms[1].AlgorithmName != "algo-c" {
		t.Fatalf("GetAlgorithmMetrics() unexpected algorithms: %+v", resp.Algorithms)
	}
	if len(orchestrator.executionReqs) != 3 {
		t.Fatalf("GetAlgorithmMetrics() execution calls = %d, want 3", len(orchestrator.executionReqs))
	}
}
