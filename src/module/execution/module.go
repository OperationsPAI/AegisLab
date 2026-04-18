package executionmodule

import "go.uber.org/fx"

var Module = fx.Module("execution",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(AsHandlerService),
	fx.Provide(NewHandler),
)
