package datasetmodule

import "go.uber.org/fx"

var Module = fx.Module("dataset",
	fx.Provide(NewRepository),
	fx.Provide(NewDatapackFileStore),
	fx.Provide(NewService),
	fx.Provide(AsHandlerService),
	fx.Provide(NewHandler),
)
