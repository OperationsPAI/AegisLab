package orchestratorclient

import "go.uber.org/fx"

var Module = fx.Module("orchestrator_client",
	fx.Provide(NewClient),
)
