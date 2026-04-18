package systemmodule

import (
	"context"

	"aegis/dto"
)

// HandlerService captures the system operations consumed by the HTTP handler.
type HandlerService interface {
	GetHealth(context.Context) (*HealthCheckResp, error)
	GetMetrics(context.Context) (*MonitoringMetricsResp, error)
	GetSystemInfo(context.Context) (*SystemInfo, error)
	ListNamespaceLocks(context.Context) (*ListNamespaceLockResp, error)
	ListQueuedTasks(context.Context) (*QueuedTasksResp, error)
	GetAuditLog(context.Context, int) (*AuditLogDetailResp, error)
	ListAuditLogs(context.Context, *ListAuditLogReq) (*dto.ListResp[AuditLogResp], error)
	GetConfig(context.Context, int) (*ConfigDetailResp, error)
	ListConfigs(context.Context, *ListConfigReq) (*dto.ListResp[ConfigResp], error)
	RollbackConfigValue(context.Context, *RollbackConfigReq, int, int, string, string) error
	RollbackConfigMetadata(context.Context, *RollbackConfigReq, int, int, string, string) (*ConfigResp, error)
	UpdateConfigValue(context.Context, *UpdateConfigValueReq, int, int, string, string) error
	UpdateConfigMetadata(context.Context, *UpdateConfigMetadataReq, int, int, string, string) (*ConfigResp, error)
	ListConfigHistories(context.Context, *ListConfigHistoryReq, int) (*dto.ListResp[ConfigHistoryResp], error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
