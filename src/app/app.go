package app

import (
	buildkit "aegis/infra/buildkit"
	config "aegis/infra/config"
	db "aegis/infra/db"
	etcd "aegis/infra/etcd"
	harbor "aegis/infra/harbor"
	helm "aegis/infra/helm"
	logger "aegis/infra/logger"
	loki "aegis/infra/loki"
	redis "aegis/infra/redis"
	tracing "aegis/infra/tracing"

	"go.uber.org/fx"
)

func BaseOptions(confPath string) fx.Option {
	return fx.Options(
		fx.Supply(config.Params{Path: confPath}),
		config.Module,
		logger.Module,
	)
}

func ObserveOptions() fx.Option {
	return fx.Options(
		loki.Module,
		tracing.Module,
	)
}

func DataOptions() fx.Option {
	return fx.Options(
		db.Module,
		redis.Module,
	)
}

func CoordinationOptions() fx.Option {
	return fx.Options(
		etcd.Module,
	)
}

func BuildInfraOptions() fx.Option {
	return fx.Options(
		harbor.Module,
		helm.Module,
		buildkit.Module,
	)
}

func CommonOptions(confPath string) fx.Option {
	return fx.Options(
		BaseOptions(confPath),
		ObserveOptions(),
		DataOptions(),
		CoordinationOptions(),
		BuildInfraOptions(),
	)
}
