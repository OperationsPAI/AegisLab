package systemclient

import "go.uber.org/fx"

var Module = fx.Module("system_client",
	fx.Provide(NewClient),
)
