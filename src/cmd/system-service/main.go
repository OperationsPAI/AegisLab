package main

import (
	"flag"

	systemapp "aegis/app/system"

	"go.uber.org/fx"
)

func main() {
	conf := flag.String("conf", "/etc/rcabench/config.prod.toml", "path to configuration file")
	flag.Parse()

	fx.New(systemapp.Options(*conf)).Run()
}
