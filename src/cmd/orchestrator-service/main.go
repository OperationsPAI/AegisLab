package main

import (
	"flag"

	orchestrator "aegis/app/orchestrator"

	"go.uber.org/fx"
)

func main() {
	conf := flag.String("conf", "/etc/rcabench/config.prod.toml", "path to configuration file")
	flag.Parse()

	fx.New(orchestrator.Options(*conf)).Run()
}
