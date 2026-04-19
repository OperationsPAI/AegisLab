package task

import (
	"context"
	"time"

	"aegis/dto"
	loki "aegis/infra/loki"
)

type LokiGateway struct {
	client *loki.Client
}

func NewLokiGateway(client *loki.Client) *LokiGateway {
	return &LokiGateway{client: client}
}

func (g *LokiGateway) QueryJobLogs(ctx context.Context, taskID string, start time.Time) ([]dto.LogEntry, error) {
	return g.client.QueryJobLogs(ctx, taskID, loki.QueryOpts{
		Start:     start,
		Direction: "forward",
	})
}
