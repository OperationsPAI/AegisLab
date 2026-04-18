package grpcorchestratorinterface

import (
	projectmodule "aegis/module/project"

	"go.uber.org/fx"
)

var Module = fx.Module("grpc_orchestrator",
	fx.Provide(
		projectmodule.NewRepository,
		newProjectStatisticsReader,
		newTaskQueueController,
		newOrchestratorServer,
		newLifecycle,
	),
	fx.Invoke(registerLifecycle),
)
