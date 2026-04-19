package app

import (
	"context"

	chaos "aegis/infra/chaos"
	etcd "aegis/infra/etcd"
	k8s "aegis/infra/k8s"
	redis "aegis/infra/redis"
	httpapi "aegis/interface/http"
	commonservice "aegis/service/common"
	"aegis/service/initialization"
	"aegis/utils"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

func ProducerOptions(confPath string, port string) fx.Option {
	return fx.Options(
		CommonOptions(confPath),
		chaos.Module,
		k8s.Module,
		ProducerHTTPOptions(port),
	)
}

func ProducerHTTPOptions(port string) fx.Option {
	return fx.Options(
		fx.Provide(newProducerInitializer),
		fx.Invoke(registerProducerInitialization),
		ProducerHTTPModules(),
		fx.Supply(httpapi.ServerConfig{Addr: normalizeAddr(port)}),
		httpapi.Module,
	)
}

type ProducerInitializer struct {
	etcd      *etcd.Gateway
	redis     *redis.Gateway
	db        *gorm.DB
	StartFunc func(context.Context) error
}

func newProducerInitializer(etcd *etcd.Gateway, redis *redis.Gateway, db *gorm.DB) *ProducerInitializer {
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
