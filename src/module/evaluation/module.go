package evaluation

import "go.uber.org/fx"

var Module = fx.Module("evaluation",
	fx.Provide(NewRepository),
	fx.Provide(newExecutionQuerySource),
	fx.Provide(NewService),
	fx.Provide(AsHandlerService),
	fx.Provide(NewHandler),
)
