package labelmodule

import "go.uber.org/fx"

var Module = fx.Module("label",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(NewHandler),
)
