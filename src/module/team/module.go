package teammodule

import "go.uber.org/fx"

var Module = fx.Module("team",
	fx.Provide(NewRepository),
	fx.Provide(newProjectReader),
	fx.Provide(NewService),
	fx.Provide(AsHandlerService),
	fx.Provide(NewHandler),
)
