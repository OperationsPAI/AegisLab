package orchestrator

import (
	"aegis/app"
	grpcorchestrator "aegis/interface/grpc/orchestrator"
	group "aegis/module/group"
	metric "aegis/module/metric"
	notification "aegis/module/notification"
	task "aegis/module/task"
	trace "aegis/module/trace"

	"go.uber.org/fx"
)

// Options builds the dedicated orchestrator service runtime.
func Options(confPath string) fx.Option {
	return fx.Options(
		app.BaseOptions(confPath),
		app.ObserveOptions(),
		app.DataOptions(),
		app.ExecutionInjectionOwnerModules(),
		group.Module,
		metric.Module,
		notification.Module,
		task.Module,
		trace.Module,
		grpcorchestrator.Module,
	)
}
