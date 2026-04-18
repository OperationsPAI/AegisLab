package main

import (
	"flag"

	resourceapp "aegis/app/resource"

	"go.uber.org/fx"
)

func main() {
	conf := flag.String("conf", "/etc/rcabench/config.prod.toml", "path to configuration file")
	flag.Parse()

	fx.New(resourceapp.Options(*conf)).Run()
}
