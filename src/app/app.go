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

func CommonOptions(confPath string) fx.Option {
	return fx.Options(
		fx.Supply(configinfra.Params{Path: confPath}),
		loggerinfra.Module,
		configinfra.Module,
		dbinfra.Module,
		redisinfra.Module,
		etcdinfra.Module,
		harborinfra.Module,
		helminfra.Module,
		buildkitinfra.Module,
		lokiinfra.Module,
		tracinginfra.Module,
	)
}
