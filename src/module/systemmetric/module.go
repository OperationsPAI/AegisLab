package systemmetricmodule

import "go.uber.org/fx"

var Module = fx.Module("system_metric",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(NewHandler),
	fx.Invoke(RegisterMetricsCollector),
)
