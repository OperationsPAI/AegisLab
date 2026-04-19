package etcd

import (
	"context"
	"fmt"
	"time"

	"aegis/config"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/fx"
)

type Gateway struct {
	client *clientv3.Client
}

func NewGateway(client *clientv3.Client) *Gateway {
	if client == nil {
		client = newClient()
	}
	return &Gateway{client: client}
}

func NewGatewayWithLifecycle(lc fx.Lifecycle) *Gateway {
	gateway := NewGateway(nil)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logrus.Info("Closing etcd client")
			return gateway.close()
		},
	})

	return gateway
}

func (g *Gateway) Put(ctx context.Context, key, value string, ttl time.Duration) error {
	client := g.clientOrInit()
	if ttl > 0 {
		lease, err := client.Grant(ctx, int64(ttl.Seconds()))
		if err != nil {
			return fmt.Errorf("failed to create lease: %w", err)
		}

		if _, err = client.Put(ctx, key, value, clientv3.WithLease(lease.ID)); err != nil {
			return fmt.Errorf("failed to put key with lease: %w", err)
		}
		return nil
	}

	if _, err := client.Put(ctx, key, value); err != nil {
		return fmt.Errorf("failed to put key: %w", err)
	}
	return nil
}

func (g *Gateway) Get(ctx context.Context, key string) (string, error) {
	resp, err := g.clientOrInit().Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get key: %w", err)
	}
	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return string(resp.Kvs[0].Value), nil
}

func (g *Gateway) Delete(ctx context.Context, key string) error {
	if _, err := g.clientOrInit().Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}
	return nil
}

func (g *Gateway) Watch(ctx context.Context, key string, withPrefix bool) clientv3.WatchChan {
	var opts []clientv3.OpOption
	if withPrefix {
		opts = append(opts, clientv3.WithPrefix())
	}
	return g.clientOrInit().Watch(ctx, key, opts...)
}

func (g *Gateway) clientOrInit() *clientv3.Client {
	if g.client == nil {
		g.client = newClient()
	}
	return g.client
}

func (g *Gateway) close() error {
	if g.client == nil {
		return nil
	}
	return g.client.Close()
}

func newClient() *clientv3.Client {
	endpoints := config.GetStringSlice("etcd.endpoints")
	if len(endpoints) == 0 {
		endpoints = []string{"localhost:2379"}
		logrus.Warn("etcd.endpoints not configured, using default: localhost:2379")
	}

	logrus.Infof("Connecting to etcd endpoints: %v", endpoints)

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
		Username:    config.GetString("etcd.username"),
		Password:    config.GetString("etcd.password"),
	})
	if err != nil {
		logrus.Fatalf("Failed to connect to etcd: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := client.Status(ctx, endpoints[0]); err != nil {
		logrus.Fatalf("Failed to verify etcd connection: %v", err)
	}

	logrus.Info("Successfully connected to etcd")
	return client
}
