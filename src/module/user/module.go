package usermodule

import "go.uber.org/fx"

var Module = fx.Module("user",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(NewHandler),
)
