package redisinfra

import "go.uber.org/fx"

var Module = fx.Module("redis",
	fx.Provide(NewGatewayWithLifecycle),
)
