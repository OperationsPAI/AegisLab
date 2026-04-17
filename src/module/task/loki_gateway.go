package taskmodule

import (
	"context"
	"time"

	"aegis/dto"
	lokiinfra "aegis/infra/loki"
)

type LokiGateway struct {
	client *lokiinfra.Client
}

func NewLokiGateway(client *lokiinfra.Client) *LokiGateway {
	return &LokiGateway{client: client}
}

func (g *LokiGateway) QueryJobLogs(ctx context.Context, taskID string, start time.Time) ([]dto.LogEntry, error) {
	return g.client.QueryJobLogs(ctx, taskID, lokiinfra.QueryOpts{
		Start:     start,
		Direction: "forward",
	})
}
