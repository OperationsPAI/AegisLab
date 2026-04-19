package consumer

import (
	"context"
	"fmt"
	"strconv"

	"aegis/consts"
	redisinfra "aegis/infra/redis"
	goredis "github.com/redis/go-redis/v9"
)

type namespaceStatusStore struct {
	client *redisinfra.Gateway
}

func newNamespaceStatusStore(client *redisinfra.Gateway) namespaceStatusStore {
	return namespaceStatusStore{client: client}
}

func (s namespaceStatusStore) key(namespace string) string {
	return fmt.Sprintf(consts.NamespaceKeyPattern, namespace)
}

func (s namespaceStatusStore) get(ctx context.Context, namespace string) (consts.StatusType, error) {
	statusStr, err := s.client.HashGet(ctx, s.key(namespace), "status")
	if err == goredis.Nil {
		return consts.CommonEnabled, nil
	}
	if err != nil {
		return 0, err
	}

	status, err := strconv.Atoi(statusStr)
	if err != nil {
		return 0, fmt.Errorf("invalid status value: %w", err)
	}
	return consts.StatusType(status), nil
}

func (s namespaceStatusStore) set(ctx context.Context, namespace string, status consts.StatusType) error {
	return s.client.HashSet(ctx, s.key(namespace), map[string]any{"status": int(status)})
}
