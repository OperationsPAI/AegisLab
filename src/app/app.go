package app

import (
	buildkitinfra "aegis/infra/buildkit"
	configinfra "aegis/infra/config"
	dbinfra "aegis/infra/db"
	etcdinfra "aegis/infra/etcd"
	harborinfra "aegis/infra/harbor"
	helminfra "aegis/infra/helm"
	loggerinfra "aegis/infra/logger"
	lokiinfra "aegis/infra/loki"
	redisinfra "aegis/infra/redis"
	tracinginfra "aegis/infra/tracing"

	"go.uber.org/fx"
)

func BaseOptions(confPath string) fx.Option {
	return fx.Options(
		fx.Supply(configinfra.Params{Path: confPath}),
		configinfra.Module,
		loggerinfra.Module,
	)
}

func ObserveOptions() fx.Option {
	return fx.Options(
		lokiinfra.Module,
		tracinginfra.Module,
	)
}

func DataOptions() fx.Option {
	return fx.Options(
		dbinfra.Module,
		redisinfra.Module,
	)
}

func CoordinationOptions() fx.Option {
	return fx.Options(
		etcdinfra.Module,
	)
}

func BuildInfraOptions() fx.Option {
	return fx.Options(
		harborinfra.Module,
		helminfra.Module,
		buildkitinfra.Module,
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
