package systemmodule

import "go.uber.org/fx"

var Module = fx.Module("system",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(NewHandler),
)
