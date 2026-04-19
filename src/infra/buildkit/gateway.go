package buildkit

import (
	"context"
	"fmt"
	"net"
	"time"

	"aegis/config"

	buildkitclient "github.com/moby/buildkit/client"
)

type Gateway struct{}

func NewGateway() *Gateway {
	return &Gateway{}
}

func (g *Gateway) Address() string {
	return config.GetString("buildkit.address")
}

func (g *Gateway) Endpoint() string {
	address := g.Address()
	if address == "" {
		return ""
	}
	return fmt.Sprintf("tcp://%s", address)
}

func (g *Gateway) NewClient(ctx context.Context) (*buildkitclient.Client, error) {
	endpoint := g.Endpoint()
	if endpoint == "" {
		return nil, fmt.Errorf("buildkit address is not configured")
	}
	return buildkitclient.New(ctx, endpoint)
}

func (g *Gateway) CheckHealth(ctx context.Context, timeout time.Duration) error {
	address := g.Address()
	if address == "" {
		return fmt.Errorf("buildkit address is not configured")
	}

	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("cannot connect to BuildKit at %s: %w", address, err)
	}
	return conn.Close()
}
