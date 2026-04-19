package buildkit

import "go.uber.org/fx"

var Module = fx.Module("buildkit",
	fx.Provide(NewGateway),
)
