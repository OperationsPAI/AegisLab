package notification

import "go.uber.org/fx"

var Module = fx.Module("notification",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(AsHandlerService),
	fx.Provide(NewHandler),
)
