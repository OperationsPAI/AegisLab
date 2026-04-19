package harbor

import "go.uber.org/fx"

var Module = fx.Module("harbor",
	fx.Provide(NewGateway),
)
