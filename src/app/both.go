package app

import "go.uber.org/fx"

func BothOptions(confPath string, port string) fx.Option {
	return fx.Options(
		CommonOptions(confPath),
		RuntimeWorkerStackOptions(),
		ProducerHTTPOptions(port),
	)
}
