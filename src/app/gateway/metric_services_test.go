package gatewayapp

import (
	"context"
	"testing"
	"time"

	"aegis/consts"
	"aegis/dto"
	containermodule "aegis/module/container"
	metricmodule "aegis/module/metric"
)

type orchestratorMetricClientStub struct {
	injectionReqs []*metricmodule.GetMetricsReq
	executionReqs []*metricmodule.GetMetricsReq
	injection     *metricmodule.InjectionMetrics
	execution     map[int]metricmodule.ExecutionMetrics
	enabled       bool
}

func (s *orchestratorMetricClientStub) Enabled() bool {
	return s.enabled
}

func (s *orchestratorMetricClientStub) GetInjectionMetrics(_ context.Context, req *metricmodule.GetMetricsReq) (*metricmodule.InjectionMetrics, error) {
	s.injectionReqs = append(s.injectionReqs, req)
	return s.injection, nil
}

func (s *orchestratorMetricClientStub) GetExecutionMetrics(_ context.Context, req *metricmodule.GetMetricsReq) (*metricmodule.ExecutionMetrics, error) {
	s.executionReqs = append(s.executionReqs, req)
	if req != nil && req.AlgorithmID != nil {
		if metric, ok := s.execution[*req.AlgorithmID]; ok {
			result := metric
			return &result, nil
		}
	}
	return &metricmodule.ExecutionMetrics{}, nil
}

type resourceMetricClientStub struct {
	responses []*dto.ListResp[containermodule.ContainerResp]
	enabled   bool
	calls     int
}

func (s *resourceMetricClientStub) Enabled() bool {
	return s.enabled
}

func (s *resourceMetricClientStub) ListContainers(_ context.Context, _ *containermodule.ListContainerReq) (*dto.ListResp[containermodule.ContainerResp], error) {
	idx := s.calls
	s.calls++
	if idx >= len(s.responses) {
		return &dto.ListResp[containermodule.ContainerResp]{}, nil
	}
	return s.responses[idx], nil
}

func TestRemoteAwareMetricServiceGetInjectionMetricsRemoteOnly(t *testing.T) {
	service := remoteAwareMetricService{}
	_, err := service.GetInjectionMetrics(context.Background(), &metricmodule.GetMetricsReq{})
	if err == nil {
		t.Fatal("GetInjectionMetrics() error = nil, want missing dependency")
	}
}

func TestRemoteAwareMetricServiceGetAlgorithmMetricsBuildsFromRemoteSources(t *testing.T) {
	start := time.Now().Add(-time.Hour)
	end := time.Now()
	orchestrator := &orchestratorMetricClientStub{
		enabled: true,
		execution: map[int]metricmodule.ExecutionMetrics{
			1: {TotalCount: 3, SuccessCount: 2, FailedCount: 1, SuccessRate: 66.7, AvgDuration: 12.5},
			2: {TotalCount: 0},
			3: {TotalCount: 5, SuccessCount: 5, FailedCount: 0, SuccessRate: 100, AvgDuration: 8},
		},
	}
	resource := &resourceMetricClientStub{
		enabled: true,
		responses: []*dto.ListResp[containermodule.ContainerResp]{
			{
				Items: []containermodule.ContainerResp{
					{ID: 1, Name: "algo-a", Type: consts.GetContainerTypeName(consts.ContainerTypeAlgorithm)},
					{ID: 2, Name: "algo-b", Type: consts.GetContainerTypeName(consts.ContainerTypeAlgorithm)},
				},
				Pagination: &dto.PaginationInfo{Page: 1, Size: 100, Total: 3, TotalPages: 2},
			},
			{
				Items: []containermodule.ContainerResp{
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

	resp, err := service.GetAlgorithmMetrics(context.Background(), &metricmodule.GetMetricsReq{
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
