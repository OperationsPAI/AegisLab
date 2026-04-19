package config

import (
	"aegis/config"

	"go.uber.org/fx"
)

type Params struct {
	Path string
}

var Module = fx.Module("config",
	fx.Invoke(Init),
)

func Init(params Params) {
	config.Init(params.Path)
}
