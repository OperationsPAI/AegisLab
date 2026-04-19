package iam

import (
	"aegis/app"
	grpciam "aegis/interface/grpc/iam"
	"aegis/internalclient/resourceclient"
	"aegis/middleware"
	auth "aegis/module/auth"
	rbac "aegis/module/rbac"
	team "aegis/module/team"
	user "aegis/module/user"

	"go.uber.org/fx"
)

// Options builds the dedicated IAM service runtime.
func Options(confPath string) fx.Option {
	return fx.Options(
		app.BaseOptions(confPath),
		app.ObserveOptions(),
		app.DataOptions(),
		app.RequireConfiguredTargets(
			"iam-service",
			app.RequiredConfigTarget{Name: "resource-service", PrimaryKey: "clients.resource.target", LegacyKey: "resource.grpc.target"},
		),
		resourceclient.Module,
		team.RemoteProjectReaderOption(),
		auth.Module,
		rbac.Module,
		team.Module,
		user.Module,
		fx.Provide(middleware.NewService),
		grpciam.Module,
	)
}
