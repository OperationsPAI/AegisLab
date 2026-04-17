package app

import (
	"context"

	etcdinfra "aegis/infra/etcd"
	redisinfra "aegis/infra/redis"
	commonservice "aegis/service/common"
	"aegis/service/initialization"
	"aegis/utils"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type ProducerInitializer struct {
	etcd      *etcdinfra.Gateway
	redis     *redisinfra.Gateway
	db        *gorm.DB
	StartFunc func(context.Context) error
}

func newProducerInitializer(etcd *etcdinfra.Gateway, redis *redisinfra.Gateway, db *gorm.DB) *ProducerInitializer {
	return &ProducerInitializer{etcd: etcd, redis: redis, db: db}
}

func (i *ProducerInitializer) start(ctx context.Context) error {
	if i.StartFunc != nil {
		return i.StartFunc(ctx)
	}
	if err := initialization.InitializeProducer(i.db, i.redis, commonservice.NewConfigUpdateListener(ctx, i.db, i.etcd)); err != nil {
		return err
	}
	utils.InitValidator()
	return nil
}

func registerProducerInitialization(lc fx.Lifecycle, initializer *ProducerInitializer) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return initializer.start(ctx)
		},
	})
}
