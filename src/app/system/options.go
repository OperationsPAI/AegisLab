package systemapp

import (
	"aegis/app"
	k8sinfra "aegis/infra/k8s"
	grpcsysteminterface "aegis/interface/grpcsystem"
	"aegis/internalclient/runtimeclient"
	systemmodule "aegis/module/system"
	systemmetricmodule "aegis/module/systemmetric"

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
		systemmodule.RemoteRuntimeQueryOption(),
		k8sinfra.Module,
		runtimeclient.Module,
		systemmodule.Module,
		systemmetricmodule.Module,
		grpcsysteminterface.Module,
	)
}
