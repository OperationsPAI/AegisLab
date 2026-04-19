package trace

import (
	"context"
	"time"

	"aegis/dto"

	"github.com/redis/go-redis/v9"
)

// HandlerService captures trace operations consumed by HTTP handlers and gateway adapters.
type HandlerService interface {
	GetTrace(context.Context, string) (*TraceDetailResp, error)
	ListTraces(context.Context, *ListTraceReq) (*dto.ListResp[TraceResp], error)
	GetTraceStreamProcessor(context.Context, string) (*StreamProcessor, error)
	ReadTraceStreamMessages(context.Context, string, string, int64, time.Duration) ([]redis.XStream, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
