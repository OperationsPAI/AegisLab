package tracemodule

import "go.uber.org/fx"

var Module = fx.Module("trace",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(NewHandler),
)
