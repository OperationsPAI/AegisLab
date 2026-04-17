package rbacmodule

import "go.uber.org/fx"

var Module = fx.Module("rbac",
	fx.Provide(NewRepository),
	fx.Provide(NewService),
	fx.Provide(NewHandler),
)
