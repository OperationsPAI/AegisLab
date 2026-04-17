package chaosinfra

import (
	chaosCli "github.com/OperationsPAI/chaos-experiment/client"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
)

var Module = fx.Module("chaos",
	fx.Invoke(Initialize),
)

func Initialize(restConfig *rest.Config) {
	chaosCli.InitWithConfig(restConfig)
}
