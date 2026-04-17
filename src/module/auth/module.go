package authmodule

import (
	"go.uber.org/fx"
)

var Module = fx.Module("auth",
	fx.Provide(NewUserRepository),
	fx.Provide(NewRoleRepository),
	fx.Provide(NewAccessKeyRepository),
	fx.Provide(NewTokenStore),
	fx.Provide(NewService),
	fx.Provide(NewHandler),
)
