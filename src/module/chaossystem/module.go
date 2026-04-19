package chaossystem

import "go.uber.org/fx"

var Module = fx.Module("chaos_system",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(AsHandlerService),
	fx.Provide(NewHandler),
)
