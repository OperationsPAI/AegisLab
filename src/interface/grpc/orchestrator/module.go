package grpcorchestrator

import (
	project "aegis/module/project"

	"go.uber.org/fx"
)

var Module = fx.Module("grpc_orchestrator",
	fx.Provide(
		project.NewRepository,
		newProjectStatisticsReader,
		newTaskQueueController,
		newOrchestratorServer,
		newLifecycle,
	),
	fx.Invoke(registerLifecycle),
)
