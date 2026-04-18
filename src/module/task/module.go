package taskmodule

import "go.uber.org/fx"

var Module = fx.Module("task",
	fx.Provide(NewRepository),
	fx.Provide(NewTaskQueueStore),
	fx.Provide(NewLokiGateway),
	fx.Provide(NewTaskLogService),
	fx.Provide(NewService),
	fx.Provide(AsHandlerService),
	fx.Provide(NewHandler),
)
