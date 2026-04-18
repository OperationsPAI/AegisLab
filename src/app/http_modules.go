package app

import (
	authmodule "aegis/module/auth"
	chaossystemmodule "aegis/module/chaossystem"
	containermodule "aegis/module/container"
	datasetmodule "aegis/module/dataset"
	evaluationmodule "aegis/module/evaluation"
	executionmodule "aegis/module/execution"
	groupmodule "aegis/module/group"
	injectionmodule "aegis/module/injection"
	labelmodule "aegis/module/label"
	metricmodule "aegis/module/metric"
	notificationmodule "aegis/module/notification"
	projectmodule "aegis/module/project"
	rbacmodule "aegis/module/rbac"
	sdkmodule "aegis/module/sdk"
	systemmodule "aegis/module/system"
	systemmetricmodule "aegis/module/systemmetric"
	taskmodule "aegis/module/task"
	teammodule "aegis/module/team"
	tracemodule "aegis/module/trace"
	usermodule "aegis/module/user"
	"aegis/router"

	"go.uber.org/fx"
)

func ExecutionInjectionOwnerModules() fx.Option {
	return fx.Options(
		executionmodule.Module,
		injectionmodule.Module,
	)
}

func ProducerHTTPModules() fx.Option {
	return fx.Options(
		authmodule.Module,
		chaossystemmodule.Module,
		containermodule.Module,
		datasetmodule.Module,
		evaluationmodule.Module,
		ExecutionInjectionOwnerModules(),
		groupmodule.Module,
		labelmodule.Module,
		metricmodule.Module,
		notificationmodule.Module,
		projectmodule.Module,
		rbacmodule.Module,
		sdkmodule.Module,
		systemmodule.Module,
		systemmetricmodule.Module,
		taskmodule.Module,
		teammodule.Module,
		tracemodule.Module,
		usermodule.Module,
		router.Module,
	)
}
