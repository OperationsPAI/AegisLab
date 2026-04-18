package resourceapp

import (
	"aegis/app"
	grpcresourceinterface "aegis/interface/grpcresource"
	"aegis/internalclient/orchestratorclient"
	chaossystemmodule "aegis/module/chaossystem"
	containermodule "aegis/module/container"
	datasetmodule "aegis/module/dataset"
	evaluationmodule "aegis/module/evaluation"
	labelmodule "aegis/module/label"
	projectmodule "aegis/module/project"

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
		evaluationmodule.RemoteQueryOption(),
		projectmodule.RemoteStatisticsOption(),
		chaossystemmodule.Module,
		containermodule.Module,
		datasetmodule.Module,
		evaluationmodule.Module,
		labelmodule.Module,
		projectmodule.Module,
		grpcresourceinterface.Module,
	)
}
