package grpcsystem

import (
	"context"
	"errors"
	"testing"
	"time"

	"aegis/consts"
	"aegis/dto"
	system "aegis/module/system"
	systemmetric "aegis/module/systemmetric"
	systemv1 "aegis/proto/system/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type systemReaderStub struct {
	health  *system.HealthCheckResp
	metrics *system.MonitoringMetricsResp
	info    *system.SystemInfo
	locks   *system.ListNamespaceLockResp
	queued  *system.QueuedTasksResp
	audit   *system.AuditLogDetailResp
	audits  *dto.ListResp[system.AuditLogResp]
	config  *system.ConfigDetailResp
	configs *dto.ListResp[system.ConfigResp]
	err     error
}

func (s systemReaderStub) GetHealth(context.Context) (*system.HealthCheckResp, error) {
	return s.health, s.err
}
func (s systemReaderStub) GetMetrics(context.Context) (*system.MonitoringMetricsResp, error) {
	return s.metrics, s.err
}
func (s systemReaderStub) GetSystemInfo(context.Context) (*system.SystemInfo, error) {
	return s.info, s.err
}
func (s systemReaderStub) ListNamespaceLocks(context.Context) (*system.ListNamespaceLockResp, error) {
	return s.locks, s.err
}
func (s systemReaderStub) ListQueuedTasks(context.Context) (*system.QueuedTasksResp, error) {
	return s.queued, s.err
}
func (s systemReaderStub) GetAuditLog(_ context.Context, id int) (*system.AuditLogDetailResp, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.audit, s.err
}
func (s systemReaderStub) ListAuditLogs(context.Context, *system.ListAuditLogReq) (*dto.ListResp[system.AuditLogResp], error) {
	return s.audits, s.err
}
func (s systemReaderStub) GetConfig(_ context.Context, id int) (*system.ConfigDetailResp, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.config, s.err
}
func (s systemReaderStub) ListConfigs(context.Context, *system.ListConfigReq) (*dto.ListResp[system.ConfigResp], error) {
	return s.configs, s.err
}

type metricsReaderStub struct {
	current *systemmetric.SystemMetricsResp
	history *systemmetric.SystemMetricsHistoryResp
	err     error
}

func (s metricsReaderStub) GetSystemMetrics(context.Context) (*systemmetric.SystemMetricsResp, error) {
	return s.current, s.err
}
func (s metricsReaderStub) GetSystemMetricsHistory(context.Context) (*systemmetric.SystemMetricsHistoryResp, error) {
	return s.history, s.err
}

func TestSystemServerGetHealth(t *testing.T) {
	server := &systemServer{
		system: systemReaderStub{
			health: &system.HealthCheckResp{
				Status:    "healthy",
				Timestamp: time.Now(),
				Version:   "v1",
				Uptime:    "1m",
				Services: map[string]system.ServiceInfo{
					"redis": {Status: "healthy"},
				},
			},
			metrics: &system.MonitoringMetricsResp{},
			info:    &system.SystemInfo{},
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
			configs: &dto.ListResp[system.ConfigResp]{
				Items: []system.ConfigResp{{ID: 1, Key: "demo.key"}},
				Pagination: &dto.PaginationInfo{
					Page: 1, Size: 20, Total: 1, TotalPages: 1,
				},
			},
			metrics: &system.MonitoringMetricsResp{},
			info:    &system.SystemInfo{},
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
			metrics: &system.MonitoringMetricsResp{},
			info:    &system.SystemInfo{},
		},
		metrics: metricsReaderStub{
			history: &systemmetric.SystemMetricsHistoryResp{
				CPU: []systemmetric.MetricValue{{Value: 1}},
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
