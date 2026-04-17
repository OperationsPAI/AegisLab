package httpinterface

import (
	"aegis/middleware"

	"go.uber.org/fx"
)

var Module = fx.Module("http",
	fx.Provide(
		middleware.NewService,
		NewGinEngine,
		NewServer,
	),
	fx.Invoke(registerServerLifecycle),
)
