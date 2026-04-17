package helminfra

import "go.uber.org/fx"

var Module = fx.Module("helm",
	fx.Provide(NewGateway),
)
