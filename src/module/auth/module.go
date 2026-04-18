package authmodule

import (
	"go.uber.org/fx"
)

var Module = fx.Module("auth",
	fx.Provide(NewUserRepository),
	fx.Provide(NewRoleRepository),
	fx.Provide(NewAPIKeyRepository),
	fx.Provide(NewTokenStore),
	fx.Provide(NewService),
	fx.Provide(AsHandlerService),
	fx.Provide(NewTokenVerifier),
	fx.Provide(NewHandler),
)
