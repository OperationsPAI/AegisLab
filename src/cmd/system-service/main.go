package main

import (
	"flag"

	system "aegis/app/system"

	"go.uber.org/fx"
)

func main() {
	conf := flag.String("conf", "/etc/rcabench/config.prod.toml", "path to configuration file")
	flag.Parse()

	fx.New(system.Options(*conf)).Run()
}
