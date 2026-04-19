package container

import "go.uber.org/fx"

var Module = fx.Module("container",
	fx.Provide(NewRepository),
	fx.Provide(NewBuildGateway),
	fx.Provide(NewHelmFileStore),
	fx.Provide(NewService),
	fx.Provide(AsHandlerService),
	fx.Provide(NewHandler),
)
