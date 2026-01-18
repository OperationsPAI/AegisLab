package producer

import (
	"aegis/client"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ReadNotificationStreamMessages reads messages from the notification stream
func ReadNotificationStreamMessages(ctx context.Context, streamKey, lastID string, count int64, block time.Duration) ([]redis.XStream, error) {
	if lastID == "" {
		lastID = "0"
	}

	messages, err := client.RedisXRead(ctx, []string{streamKey, lastID}, count, block)
	if err != nil {
		return nil, fmt.Errorf("failed to read notification stream messages: %w", err)
	}
	return messages, nil
}
