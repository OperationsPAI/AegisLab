package injectionmodule

import "go.uber.org/fx"

var Module = fx.Module("injection",
	fx.Provide(NewRepository),
	fx.Provide(NewDatapackStore),
	fx.Provide(NewService),
	fx.Provide(NewHandler),
)
