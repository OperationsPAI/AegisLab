package metric

import "context"

// HandlerService captures the metric operations consumed by the HTTP handler.
type HandlerService interface {
	GetInjectionMetrics(context.Context, *GetMetricsReq) (*InjectionMetrics, error)
	GetExecutionMetrics(context.Context, *GetMetricsReq) (*ExecutionMetrics, error)
	GetAlgorithmMetrics(context.Context, *GetMetricsReq) (*AlgorithmMetrics, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
