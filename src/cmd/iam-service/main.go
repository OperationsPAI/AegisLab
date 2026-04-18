package main

import (
	"flag"

	iamapp "aegis/app/iam"

	"go.uber.org/fx"
)

func main() {
	conf := flag.String("conf", "/etc/rcabench/config.prod.toml", "path to configuration file")
	flag.Parse()

	fx.New(iamapp.Options(*conf)).Run()
}
