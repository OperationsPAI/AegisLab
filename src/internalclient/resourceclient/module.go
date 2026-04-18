package resourceclient

import "go.uber.org/fx"

var Module = fx.Module("resource_client",
	fx.Provide(NewClient),
)
