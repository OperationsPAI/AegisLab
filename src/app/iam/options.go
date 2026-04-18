package iamapp

import (
	"aegis/app"
	grpciaminterface "aegis/interface/grpciam"
	"aegis/internalclient/resourceclient"
	"aegis/middleware"
	authmodule "aegis/module/auth"
	rbacmodule "aegis/module/rbac"
	teammodule "aegis/module/team"
	usermodule "aegis/module/user"

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
		teammodule.RemoteProjectReaderOption(),
		authmodule.Module,
		rbacmodule.Module,
		teammodule.Module,
		usermodule.Module,
		fx.Provide(middleware.NewService),
		grpciaminterface.Module,
	)
}
