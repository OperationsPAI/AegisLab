package resource

import (
	"aegis/app"
	grpcresource "aegis/interface/grpc/resource"
	"aegis/internalclient/orchestratorclient"
	chaossystem "aegis/module/chaossystem"
	container "aegis/module/container"
	dataset "aegis/module/dataset"
	evaluation "aegis/module/evaluation"
	label "aegis/module/label"
	project "aegis/module/project"

	"go.uber.org/fx"
)

// Options builds the dedicated resource service runtime.
func Options(confPath string) fx.Option {
	return fx.Options(
		app.BaseOptions(confPath),
		app.ObserveOptions(),
		app.DataOptions(),
		app.RequireConfiguredTargets(
			"resource-service",
			app.RequiredConfigTarget{Name: "orchestrator-service", PrimaryKey: "clients.orchestrator.target", LegacyKey: "orchestrator.grpc.target"},
		),
		orchestratorclient.Module,
		evaluation.RemoteQueryOption(),
		project.RemoteStatisticsOption(),
		chaossystem.Module,
		container.Module,
		dataset.Module,
		evaluation.Module,
		label.Module,
		project.Module,
		grpcresource.Module,
	)
}
