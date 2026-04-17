package evaluationmodule

import "go.uber.org/fx"

var Module = fx.Module("evaluation",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(NewHandler),
)
