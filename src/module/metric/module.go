package metricmodule

import "go.uber.org/fx"

var Module = fx.Module("metric",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(NewHandler),
)
