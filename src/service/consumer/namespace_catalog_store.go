package consumer

import (
	"context"
	"fmt"
	"time"

	"aegis/consts"
	redisinfra "aegis/infra/redis"
)

type namespaceCatalogStore struct {
	client *redisinfra.Gateway
}

func newNamespaceCatalogStore(client *redisinfra.Gateway) namespaceCatalogStore {
	return namespaceCatalogStore{client: client}
}

func (s namespaceCatalogStore) key(namespace string) string {
	return fmt.Sprintf(consts.NamespaceKeyPattern, namespace)
}

func (s namespaceCatalogStore) list(ctx context.Context) ([]string, error) {
	return s.client.SetMembers(ctx, consts.NamespacesKey)
}

func (s namespaceCatalogStore) exists(ctx context.Context, namespace string) (bool, error) {
	return s.client.Exists(ctx, s.key(namespace))
}

func (s namespaceCatalogStore) seed(ctx context.Context, namespace string, endTime time.Time) error {
	return s.client.SeedNamespaceState(ctx, s.key(namespace), namespace, endTime.Unix(), int(consts.CommonEnabled))
}
