package runtimeapp

import (
	"aegis/app"
	"aegis/service/consumer"

	"go.uber.org/fx"
)

// Options builds the dedicated runtime-worker-service runtime.
func Options(confPath string) fx.Option {
	return fx.Options(
		app.BaseOptions(confPath),
		app.ObserveOptions(),
		app.DataOptions(),
		app.CoordinationOptions(),
		app.BuildInfraOptions(),
		app.RuntimeWorkerStackOptions(),
		consumer.RemoteOwnerOptions(),
		app.RequireConfiguredTargets(
			"runtime-worker-service",
			app.RequiredConfigTarget{Name: "orchestrator-service", PrimaryKey: "clients.orchestrator.target", LegacyKey: "orchestrator.grpc.target"},
		),
	)
}
