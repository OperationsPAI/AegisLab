package app

import (
	chaosinfra "aegis/infra/chaos"
	k8sinfra "aegis/infra/k8s"
	runtimeinfra "aegis/infra/runtime"
	controllerinterface "aegis/interface/controller"
	httpinterface "aegis/interface/http"
	receiverinterface "aegis/interface/receiver"
	workerinterface "aegis/interface/worker"
	"aegis/service/consumer"

	"go.uber.org/fx"
)

func BothOptions(confPath string, port string) fx.Option {
	return fx.Options(
		CommonOptions(confPath),
		runtimeinfra.Module,
		chaosinfra.Module,
		k8sinfra.Module,
		fx.Provide(
			consumer.NewMonitor,
			fx.Annotate(consumer.NewRestartPedestalRateLimiter, fx.ResultTags(`name:"restart_limiter"`)),
			fx.Annotate(consumer.NewBuildContainerRateLimiter, fx.ResultTags(`name:"build_limiter"`)),
			fx.Annotate(consumer.NewAlgoExecutionRateLimiter, fx.ResultTags(`name:"algo_limiter"`)),
			consumer.NewFaultBatchManager,
			newProducerInitializer,
		),
		ProducerHTTPModules(),
		fx.Supply(httpinterface.ServerConfig{Addr: normalizeAddr(port)}),
		httpinterface.Module,
		workerinterface.Module,
		controllerinterface.Module,
		receiverinterface.Module,
		fx.Invoke(registerProducerInitialization),
	)
}
