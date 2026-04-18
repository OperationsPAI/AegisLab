package grpcsysteminterface

import (
	"context"
	"errors"
	"testing"
	"time"

	"aegis/consts"
	"aegis/dto"
	systemmodule "aegis/module/system"
	systemmetricmodule "aegis/module/systemmetric"
	systemv1 "aegis/proto/system/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type systemReaderStub struct {
	health  *systemmodule.HealthCheckResp
	metrics *systemmodule.MonitoringMetricsResp
	info    *systemmodule.SystemInfo
	locks   *systemmodule.ListNamespaceLockResp
	queued  *systemmodule.QueuedTasksResp
	audit   *systemmodule.AuditLogDetailResp
	audits  *dto.ListResp[systemmodule.AuditLogResp]
	config  *systemmodule.ConfigDetailResp
	configs *dto.ListResp[systemmodule.ConfigResp]
	err     error
}

func (s systemReaderStub) GetHealth(context.Context) (*systemmodule.HealthCheckResp, error) {
	return s.health, s.err
}
func (s systemReaderStub) GetMetrics(context.Context) (*systemmodule.MonitoringMetricsResp, error) {
	return s.metrics, s.err
}
func (s systemReaderStub) GetSystemInfo(context.Context) (*systemmodule.SystemInfo, error) {
	return s.info, s.err
}
func (s systemReaderStub) ListNamespaceLocks(context.Context) (*systemmodule.ListNamespaceLockResp, error) {
	return s.locks, s.err
}
func (s systemReaderStub) ListQueuedTasks(context.Context) (*systemmodule.QueuedTasksResp, error) {
	return s.queued, s.err
}
func (s systemReaderStub) GetAuditLog(_ context.Context, id int) (*systemmodule.AuditLogDetailResp, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.audit, s.err
}
func (s systemReaderStub) ListAuditLogs(context.Context, *systemmodule.ListAuditLogReq) (*dto.ListResp[systemmodule.AuditLogResp], error) {
	return s.audits, s.err
}
func (s systemReaderStub) GetConfig(_ context.Context, id int) (*systemmodule.ConfigDetailResp, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.config, s.err
}
func (s systemReaderStub) ListConfigs(context.Context, *systemmodule.ListConfigReq) (*dto.ListResp[systemmodule.ConfigResp], error) {
	return s.configs, s.err
}

type metricsReaderStub struct {
	current *systemmetricmodule.SystemMetricsResp
	history *systemmetricmodule.SystemMetricsHistoryResp
	err     error
}

func (s metricsReaderStub) GetSystemMetrics(context.Context) (*systemmetricmodule.SystemMetricsResp, error) {
	return s.current, s.err
}
func (s metricsReaderStub) GetSystemMetricsHistory(context.Context) (*systemmetricmodule.SystemMetricsHistoryResp, error) {
	return s.history, s.err
}

func TestSystemServerGetHealth(t *testing.T) {
	server := &systemServer{
		system: systemReaderStub{
			health: &systemmodule.HealthCheckResp{
				Status:    "healthy",
				Timestamp: time.Now(),
				Version:   "v1",
				Uptime:    "1m",
				Services: map[string]systemmodule.ServiceInfo{
					"redis": {Status: "healthy"},
				},
			},
			metrics: &systemmodule.MonitoringMetricsResp{},
			info:    &systemmodule.SystemInfo{},
		},
		metrics: metricsReaderStub{},
	}

	resp, err := server.GetHealth(context.Background(), &systemv1.PingRequest{})
	if err != nil {
		t.Fatalf("GetHealth() error = %v", err)
	}
	if resp.GetData().AsMap()["status"] != "healthy" {
		t.Fatalf("GetHealth() unexpected response: %+v", resp.GetData().AsMap())
	}
}

func TestSystemServerListConfigs(t *testing.T) {
	server := &systemServer{
		system: systemReaderStub{
			configs: &dto.ListResp[systemmodule.ConfigResp]{
				Items: []systemmodule.ConfigResp{{ID: 1, Key: "demo.key"}},
				Pagination: &dto.PaginationInfo{
					Page: 1, Size: 20, Total: 1, TotalPages: 1,
				},
			},
			metrics: &systemmodule.MonitoringMetricsResp{},
			info:    &systemmodule.SystemInfo{},
		},
		metrics: metricsReaderStub{},
	}

	query, err := structpb.NewStruct(map[string]any{"page": 1, "size": 20})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}

	resp, err := server.ListConfigs(context.Background(), &systemv1.ListConfigsRequest{Query: query})
	if err != nil {
		t.Fatalf("ListConfigs() error = %v", err)
	}
	if resp.GetData().AsMap()["items"] == nil {
		t.Fatalf("ListConfigs() unexpected response: %+v", resp.GetData().AsMap())
	}
}

func TestSystemServerGetAuditLogNotFound(t *testing.T) {
	server := &systemServer{
		system:  systemReaderStub{err: consts.ErrNotFound},
		metrics: metricsReaderStub{},
	}

	_, err := server.GetAuditLog(context.Background(), &systemv1.GetResourceRequest{Id: 1})
	if err == nil {
		t.Fatal("GetAuditLog() error = nil, want error")
	}
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetAuditLog() code = %s, want %s", status.Code(err), codes.NotFound)
	}
}

func TestSystemServerGetSystemMetricsHistory(t *testing.T) {
	server := &systemServer{
		system: systemReaderStub{
			metrics: &systemmodule.MonitoringMetricsResp{},
			info:    &systemmodule.SystemInfo{},
		},
		metrics: metricsReaderStub{
			history: &systemmetricmodule.SystemMetricsHistoryResp{
				CPU: []systemmetricmodule.MetricValue{{Value: 1}},
			},
		},
	}

	resp, err := server.GetSystemMetricsHistory(context.Background(), &systemv1.PingRequest{})
	if err != nil {
		t.Fatalf("GetSystemMetricsHistory() error = %v", err)
	}
	if resp.GetData().AsMap()["cpu"] == nil {
		t.Fatalf("GetSystemMetricsHistory() unexpected response: %+v", resp.GetData().AsMap())
	}
}
