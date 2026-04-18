package groupmodule

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// HandlerService captures group operations consumed by HTTP handlers and gateway adapters.
type HandlerService interface {
	GetGroupStats(context.Context, *GetGroupStatsReq) (*GroupStats, error)
	NewGroupStreamProcessor(context.Context, string) (*GroupStreamProcessor, error)
	ReadGroupStreamMessages(context.Context, string, string, int64, time.Duration) ([]redis.XStream, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
