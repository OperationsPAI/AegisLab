package gateway

import (
	"context"

	"aegis/dto"
	"aegis/internalclient/systemclient"
	system "aegis/module/system"
	systemmetric "aegis/module/systemmetric"
)

type remoteAwareSystemService struct {
	system.HandlerService
	system *systemclient.Client
}

func (s remoteAwareSystemService) GetHealth(ctx context.Context) (*system.HealthCheckResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.GetHealth(ctx)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) GetMetrics(ctx context.Context) (*system.MonitoringMetricsResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.GetMetrics(ctx)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) GetSystemInfo(ctx context.Context) (*system.SystemInfo, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.GetSystemInfo(ctx)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) ListNamespaceLocks(ctx context.Context) (*system.ListNamespaceLockResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.ListNamespaceLocks(ctx)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) ListQueuedTasks(ctx context.Context) (*system.QueuedTasksResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.ListQueuedTasks(ctx)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) GetAuditLog(ctx context.Context, id int) (*system.AuditLogDetailResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.GetAuditLog(ctx, id)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) ListAuditLogs(ctx context.Context, req *system.ListAuditLogReq) (*dto.ListResp[system.AuditLogResp], error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.ListAuditLogs(ctx, req)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) GetConfig(ctx context.Context, configID int) (*system.ConfigDetailResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.GetConfig(ctx, configID)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) ListConfigs(ctx context.Context, req *system.ListConfigReq) (*dto.ListResp[system.ConfigResp], error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.ListConfigs(ctx, req)
	}
	return nil, missingRemoteDependency("system-service")
}

type remoteAwareSystemMetricService struct {
	systemmetric.HandlerService
	system *systemclient.Client
}

func (s remoteAwareSystemMetricService) GetSystemMetrics(ctx context.Context) (*systemmetric.SystemMetricsResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.GetSystemMetrics(ctx)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemMetricService) GetSystemMetricsHistory(ctx context.Context) (*systemmetric.SystemMetricsHistoryResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.GetSystemMetricsHistory(ctx)
	}
	return nil, missingRemoteDependency("system-service")
}
