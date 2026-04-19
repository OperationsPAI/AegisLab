package system

import (
	"aegis/app"
	k8s "aegis/infra/k8s"
	grpcsystem "aegis/interface/grpc/system"
	"aegis/internalclient/runtimeclient"
	system "aegis/module/system"
	systemmetric "aegis/module/systemmetric"

	"go.uber.org/fx"
)

// Options builds the dedicated system service runtime.
func Options(confPath string) fx.Option {
	return fx.Options(
		app.BaseOptions(confPath),
		app.ObserveOptions(),
		app.DataOptions(),
		app.CoordinationOptions(),
		app.BuildInfraOptions(),
		app.RequireConfiguredTargets(
			"system-service",
			app.RequiredConfigTarget{Name: "runtime-worker-service", PrimaryKey: "clients.runtime.target", LegacyKey: "runtime_worker.grpc.target"},
		),
		system.RemoteRuntimeQueryOption(),
		k8s.Module,
		runtimeclient.Module,
		system.Module,
		systemmetric.Module,
		grpcsystem.Module,
	)
}
