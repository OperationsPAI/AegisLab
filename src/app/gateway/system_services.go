package gatewayapp

import (
	"context"

	"aegis/dto"
	"aegis/internalclient/systemclient"
	systemmodule "aegis/module/system"
	systemmetricmodule "aegis/module/systemmetric"
)

type remoteAwareSystemService struct {
	systemmodule.HandlerService
	system *systemclient.Client
}

func (s remoteAwareSystemService) GetHealth(ctx context.Context) (*systemmodule.HealthCheckResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.GetHealth(ctx)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) GetMetrics(ctx context.Context) (*systemmodule.MonitoringMetricsResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.GetMetrics(ctx)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) GetSystemInfo(ctx context.Context) (*systemmodule.SystemInfo, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.GetSystemInfo(ctx)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) ListNamespaceLocks(ctx context.Context) (*systemmodule.ListNamespaceLockResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.ListNamespaceLocks(ctx)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) ListQueuedTasks(ctx context.Context) (*systemmodule.QueuedTasksResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.ListQueuedTasks(ctx)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) GetAuditLog(ctx context.Context, id int) (*systemmodule.AuditLogDetailResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.GetAuditLog(ctx, id)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) ListAuditLogs(ctx context.Context, req *systemmodule.ListAuditLogReq) (*dto.ListResp[systemmodule.AuditLogResp], error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.ListAuditLogs(ctx, req)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) GetConfig(ctx context.Context, configID int) (*systemmodule.ConfigDetailResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.GetConfig(ctx, configID)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemService) ListConfigs(ctx context.Context, req *systemmodule.ListConfigReq) (*dto.ListResp[systemmodule.ConfigResp], error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.ListConfigs(ctx, req)
	}
	return nil, missingRemoteDependency("system-service")
}

type remoteAwareSystemMetricService struct {
	systemmetricmodule.HandlerService
	system *systemclient.Client
}

func (s remoteAwareSystemMetricService) GetSystemMetrics(ctx context.Context) (*systemmetricmodule.SystemMetricsResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.GetSystemMetrics(ctx)
	}
	return nil, missingRemoteDependency("system-service")
}

func (s remoteAwareSystemMetricService) GetSystemMetricsHistory(ctx context.Context) (*systemmetricmodule.SystemMetricsHistoryResp, error) {
	if s.system != nil && s.system.Enabled() {
		return s.system.GetSystemMetricsHistory(ctx)
	}
	return nil, missingRemoteDependency("system-service")
}
