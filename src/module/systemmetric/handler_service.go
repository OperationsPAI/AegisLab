package systemmetricmodule

import "context"

// HandlerService captures the system metric operations consumed by the HTTP handler.
type HandlerService interface {
	GetSystemMetrics(context.Context) (*SystemMetricsResp, error)
	GetSystemMetricsHistory(context.Context) (*SystemMetricsHistoryResp, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
