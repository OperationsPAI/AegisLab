package ratelimiter

import "go.uber.org/fx"

var Module = fx.Module("ratelimiter",
	fx.Provide(NewService),
	fx.Provide(NewHandler),
)
