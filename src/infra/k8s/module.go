package k8s

import (
	"k8s.io/client-go/rest"

	"go.uber.org/fx"
)

var Module = fx.Module("k8s",
	fx.Provide(ProvideController),
	fx.Provide(NewGateway),
	fx.Provide(ProvideRestConfig),
)

func ProvideController() *Controller {
	return getK8sController()
}

func ProvideRestConfig() *rest.Config {
	return getK8sRestConfig()
}
