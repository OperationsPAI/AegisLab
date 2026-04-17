package sdkmodule

import "go.uber.org/fx"

var Module = fx.Module("sdk",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(NewHandler),
)
