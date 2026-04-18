package app

import "go.uber.org/fx"

func ConsumerOptions(confPath string) fx.Option {
	return fx.Options(
		CommonOptions(confPath),
		CompatibilityRuntimeOptions(),
	)
}
