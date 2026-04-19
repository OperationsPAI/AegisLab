package notification

import (
	"context"
	"fmt"
	"time"

	redisinfra "aegis/infra/redis"

	goredis "github.com/redis/go-redis/v9"
)

type Service struct {
	repo  *Repository
	redis *redisinfra.Gateway
}

func NewService(repo *Repository, redis *redisinfra.Gateway) *Service {
	return &Service{repo: repo, redis: redis}
}

func (s *Service) ReadStreamMessages(ctx context.Context, streamKey, lastID string, count int64, block time.Duration) ([]goredis.XStream, error) {
	if lastID == "" {
		lastID = "0"
	}

	messages, err := s.redis.XRead(ctx, []string{streamKey, lastID}, count, block)
	if err != nil {
		return nil, fmt.Errorf("failed to read notification stream messages: %w", err)
	}
	return messages, nil
}
