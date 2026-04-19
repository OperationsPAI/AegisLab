package loki

import "go.uber.org/fx"

var Module = fx.Module("loki",
	fx.Provide(NewClient),
)
