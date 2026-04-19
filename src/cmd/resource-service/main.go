package main

import (
	"flag"

	resource "aegis/app/resource"

	"go.uber.org/fx"
)

func main() {
	conf := flag.String("conf", "/etc/rcabench/config.prod.toml", "path to configuration file")
	flag.Parse()

	fx.New(resource.Options(*conf)).Run()
}
