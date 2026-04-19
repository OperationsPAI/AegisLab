package app

import (
	chaos "aegis/infra/chaos"
	k8s "aegis/infra/k8s"
	runtimeinfra "aegis/infra/runtime"
	controller "aegis/interface/controller"
	grpcruntime "aegis/interface/grpc/runtime"
	receiver "aegis/interface/receiver"
	worker "aegis/interface/worker"
	"aegis/internalclient/orchestratorclient"
	"aegis/service/consumer"

	"go.uber.org/fx"
)

func RuntimeWorkerStackOptions() fx.Option {
	return fx.Options(
		runtimeinfra.Module,
		chaos.Module,
		k8s.Module,
		orchestratorclient.Module,
		fx.Provide(
			consumer.NewMonitor,
			fx.Annotate(consumer.NewRestartPedestalRateLimiter, fx.ResultTags(`name:"restart_limiter"`)),
			fx.Annotate(consumer.NewBuildContainerRateLimiter, fx.ResultTags(`name:"build_limiter"`)),
			fx.Annotate(consumer.NewAlgoExecutionRateLimiter, fx.ResultTags(`name:"algo_limiter"`)),
			consumer.NewFaultBatchManager,
			consumer.NewExecutionOwner,
			consumer.NewInjectionOwner,
		),
		worker.Module,
		controller.Module,
		grpcruntime.Module,
		receiver.Module,
	)
}
