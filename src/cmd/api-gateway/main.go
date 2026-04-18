package main

import (
	"flag"

	gatewayapp "aegis/app/gateway"

	"go.uber.org/fx"
)

func main() {
	conf := flag.String("conf", "/etc/rcabench/config.prod.toml", "path to configuration file")
	port := flag.String("port", "8080", "port to run the API gateway on")
	flag.Parse()

	fx.New(gatewayapp.Options(*conf, *port)).Run()
}
