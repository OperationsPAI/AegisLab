package app

import (
	chaosinfra "aegis/infra/chaos"
	k8sinfra "aegis/infra/k8s"
	httpinterface "aegis/interface/http"

	"go.uber.org/fx"
)

func ProducerOptions(confPath string, port string) fx.Option {
	return fx.Options(
		CommonOptions(confPath),
		chaosinfra.Module,
		k8sinfra.Module,
		fx.Provide(newProducerInitializer),
		ProducerHTTPModules(),
		fx.Supply(httpinterface.ServerConfig{Addr: normalizeAddr(port)}),
		httpinterface.Module,
		fx.Invoke(registerProducerInitialization),
	)
}
