package app

import (
	chaosinfra "aegis/infra/chaos"
	k8sinfra "aegis/infra/k8s"
	runtimeinfra "aegis/infra/runtime"
	controllerinterface "aegis/interface/controller"
	grpcruntimeinterface "aegis/interface/grpcruntime"
	receiverinterface "aegis/interface/receiver"
	workerinterface "aegis/interface/worker"
	"aegis/internalclient/orchestratorclient"
	"aegis/service/consumer"

	"go.uber.org/fx"
)

func RuntimeWorkerStackOptions() fx.Option {
	return fx.Options(
		runtimeinfra.Module,
		chaosinfra.Module,
		k8sinfra.Module,
		orchestratorclient.Module,
		RuntimeWorkerProviderOptions(),
		RuntimeWorkerInterfaceOptions(),
	)
}

func RuntimeWorkerProviderOptions() fx.Option {
	return fx.Provide(
		consumer.NewMonitor,
		fx.Annotate(consumer.NewRestartPedestalRateLimiter, fx.ResultTags(`name:"restart_limiter"`)),
		fx.Annotate(consumer.NewBuildContainerRateLimiter, fx.ResultTags(`name:"build_limiter"`)),
		fx.Annotate(consumer.NewAlgoExecutionRateLimiter, fx.ResultTags(`name:"algo_limiter"`)),
		consumer.NewFaultBatchManager,
		consumer.NewExecutionOwner,
		consumer.NewInjectionOwner,
	)
}

func RuntimeWorkerInterfaceOptions() fx.Option {
	return fx.Options(
		workerinterface.Module,
		controllerinterface.Module,
		grpcruntimeinterface.Module,
		receiverinterface.Module,
	)
}
