package iamclient

import "go.uber.org/fx"

var Module = fx.Module("iam_client",
	fx.Provide(NewClient),
)
