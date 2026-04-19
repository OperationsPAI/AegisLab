package group

import "go.uber.org/fx"

var Module = fx.Module("group",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(AsHandlerService),
	fx.Provide(NewHandler),
)
