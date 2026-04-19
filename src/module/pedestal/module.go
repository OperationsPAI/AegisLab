package pedestal

import "go.uber.org/fx"

var Module = fx.Module("pedestal",
	fx.Provide(NewRepository),
	fx.Provide(NewHandler),
)
