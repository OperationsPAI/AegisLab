package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"aegis/config"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Singleton pattern etcd client
var (
	etcdClient *clientv3.Client
	etcdOnce   sync.Once
)

// GetEtcdClient returns the singleton etcd client instance
// It initializes the client on first call using configuration from config package
func GetEtcdClient() *clientv3.Client {
	etcdOnce.Do(func() {
		endpoints := config.GetStringSlice("etcd.endpoints")
		if len(endpoints) == 0 {
			endpoints = []string{"localhost:2379"}
			logrus.Warn("etcd.endpoints not configured, using default: localhost:2379")
		}

		logrus.Infof("Connecting to etcd endpoints: %v", endpoints)

		var err error
		etcdClient, err = clientv3.New(clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: 5 * time.Second,
			Username:    config.GetString("etcd.username"),
			Password:    config.GetString("etcd.password"),
		})
		if err != nil {
			logrus.Fatalf("Failed to connect to etcd: %v", err)
		}

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if _, err := etcdClient.Status(ctx, endpoints[0]); err != nil {
			logrus.Fatalf("Failed to verify etcd connection: %v", err)
		}

		logrus.Info("Successfully connected to etcd")
	})
	return etcdClient
}

// CloseEtcdClient closes the etcd client connection
// Should be called during application shutdown
func CloseEtcdClient() error {
	if etcdClient != nil {
		logrus.Info("Closing etcd client connection")
		return etcdClient.Close()
	}
	return nil
}

// EtcdPut writes a key-value pair to etcd with optional TTL
func EtcdPut(ctx context.Context, key, value string, ttl time.Duration) error {
	client := GetEtcdClient()

	if ttl > 0 {
		// Create lease for TTL
		lease, err := client.Grant(ctx, int64(ttl.Seconds()))
		if err != nil {
			return fmt.Errorf("failed to create lease: %w", err)
		}

		_, err = client.Put(ctx, key, value, clientv3.WithLease(lease.ID))
		if err != nil {
			return fmt.Errorf("failed to put key with lease: %w", err)
		}
	} else {
		_, err := client.Put(ctx, key, value)
		if err != nil {
			return fmt.Errorf("failed to put key: %w", err)
		}
	}

	return nil
}

// EtcdGet retrieves a value from etcd by key
func EtcdGet(ctx context.Context, key string) (string, error) {
	client := GetEtcdClient()

	resp, err := client.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get key: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("key not found: %s", key)
	}

	return string(resp.Kvs[0].Value), nil
}

// EtcdDelete deletes a key from etcd
func EtcdDelete(ctx context.Context, key string) error {
	client := GetEtcdClient()

	_, err := client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	return nil
}

// EtcdWatch watches for changes on a key or prefix
// Returns a channel that receives watch events
func EtcdWatch(ctx context.Context, key string, withPrefix bool) clientv3.WatchChan {
	client := GetEtcdClient()

	var opts []clientv3.OpOption
	if withPrefix {
		opts = append(opts, clientv3.WithPrefix())
	}

	return client.Watch(ctx, key, opts...)
}

// EtcdGetWithRevision retrieves a value and its revision
func EtcdGetWithRevision(ctx context.Context, key string) (string, int64, error) {
	client := GetEtcdClient()

	resp, err := client.Get(ctx, key)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get key: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return "", 0, fmt.Errorf("key not found: %s", key)
	}

	return string(resp.Kvs[0].Value), resp.Kvs[0].ModRevision, nil
}

// EtcdWatchFromRevision watches for changes starting from a specific revision
func EtcdWatchFromRevision(ctx context.Context, key string, revision int64, withPrefix bool) clientv3.WatchChan {
	client := GetEtcdClient()

	var opts []clientv3.OpOption
	if withPrefix {
		opts = append(opts, clientv3.WithPrefix())
	}
	opts = append(opts, clientv3.WithRev(revision))

	return client.Watch(ctx, key, opts...)
}
