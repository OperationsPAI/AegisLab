package metric

import "go.uber.org/fx"

var Module = fx.Module("metric",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(AsHandlerService),
	fx.Provide(NewHandler),
)
