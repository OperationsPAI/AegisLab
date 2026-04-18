package main

import (
	"flag"

	orchestratorapp "aegis/app/orchestrator"

	"go.uber.org/fx"
)

func main() {
	conf := flag.String("conf", "/etc/rcabench/config.prod.toml", "path to configuration file")
	flag.Parse()

	fx.New(orchestratorapp.Options(*conf)).Run()
}
