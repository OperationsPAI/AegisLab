package app

import "go.uber.org/fx"

func ProducerOptions(confPath string, port string) fx.Option {
	return fx.Options(
		CommonOptions(confPath),
		ProducerCompatibilityOptions(port),
	)
}
