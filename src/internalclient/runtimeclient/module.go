package runtimeclient

import "go.uber.org/fx"

var Module = fx.Module("runtime_client",
	fx.Provide(NewClient),
)
