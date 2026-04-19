package systemmetric

import "go.uber.org/fx"

var Module = fx.Module("system_metric",
	fx.Provide(
		NewRepository,
		NewService,
		AsHandlerService,
		NewHandler,
	),
	fx.Invoke(RegisterMetricsCollector),
)
