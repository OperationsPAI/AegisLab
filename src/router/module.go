package router

import "go.uber.org/fx"

var Module = fx.Module("router",
	fx.Provide(NewHandlers),
)
