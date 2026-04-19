package main

import (
	"flag"

	runtimeapp "aegis/app/runtime"

	"go.uber.org/fx"
)

func main() {
	conf := flag.String("conf", "/etc/rcabench/config.prod.toml", "path to configuration file")
	flag.Parse()

	fx.New(runtimeapp.Options(*conf)).Run()
}
