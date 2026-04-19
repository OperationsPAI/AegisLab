package app

import (
	auth "aegis/module/auth"
	chaossystem "aegis/module/chaossystem"
	container "aegis/module/container"
	dataset "aegis/module/dataset"
	evaluation "aegis/module/evaluation"
	execution "aegis/module/execution"
	group "aegis/module/group"
	injection "aegis/module/injection"
	label "aegis/module/label"
	metric "aegis/module/metric"
	notification "aegis/module/notification"
	pedestal "aegis/module/pedestal"
	project "aegis/module/project"
	ratelimiter "aegis/module/ratelimiter"
	rbac "aegis/module/rbac"
	sdk "aegis/module/sdk"
	system "aegis/module/system"
	systemmetric "aegis/module/systemmetric"
	task "aegis/module/task"
	team "aegis/module/team"
	trace "aegis/module/trace"
	user "aegis/module/user"
	"aegis/router"

	"go.uber.org/fx"
)

func ExecutionInjectionOwnerModules() fx.Option {
	return fx.Options(
		execution.Module,
		injection.Module,
	)
}

func ProducerHTTPModules() fx.Option {
	return fx.Options(
		auth.Module,
		chaossystem.Module,
		container.Module,
		dataset.Module,
		evaluation.Module,
		ExecutionInjectionOwnerModules(),
		group.Module,
		label.Module,
		metric.Module,
		notification.Module,
		pedestal.Module,
		project.Module,
		ratelimiter.Module,
		rbac.Module,
		sdk.Module,
		system.Module,
		systemmetric.Module,
		task.Module,
		team.Module,
		trace.Module,
		user.Module,
		router.Module,
	)
}
