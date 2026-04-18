package notificationmodule

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// HandlerService captures notification stream operations consumed by HTTP handlers and gateway adapters.
type HandlerService interface {
	ReadStreamMessages(context.Context, string, string, int64, time.Duration) ([]redis.XStream, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
