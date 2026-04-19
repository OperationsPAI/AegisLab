package label

import "go.uber.org/fx"

var Module = fx.Module("label",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(AsHandlerService),
	fx.Provide(NewHandler),
)
