package httpapi

import (
	"aegis/middleware"
	"aegis/router"

	"go.uber.org/fx"
)

var Module = fx.Module("http",
	fx.Provide(
		middleware.NewService,
		router.New,
		NewServer,
	),
	fx.Invoke(registerServerLifecycle),
)
