package orchestratorapp

import (
	"aegis/app"
	grpcorchestratorinterface "aegis/interface/grpcorchestrator"
	groupmodule "aegis/module/group"
	metricmodule "aegis/module/metric"
	notificationmodule "aegis/module/notification"
	taskmodule "aegis/module/task"
	tracemodule "aegis/module/trace"

	"go.uber.org/fx"
)

// Options builds the dedicated orchestrator service runtime.
func Options(confPath string) fx.Option {
	return fx.Options(
		app.BaseOptions(confPath),
		app.ObserveOptions(),
		app.DataOptions(),
		app.ExecutionInjectionOwnerModules(),
		groupmodule.Module,
		metricmodule.Module,
		notificationmodule.Module,
		taskmodule.Module,
		tracemodule.Module,
		grpcorchestratorinterface.Module,
	)
}
