package router

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
	project "aegis/module/project"
	rbac "aegis/module/rbac"
	sdk "aegis/module/sdk"
	system "aegis/module/system"
	systemmetric "aegis/module/systemmetric"
	task "aegis/module/task"
	team "aegis/module/team"
	trace "aegis/module/trace"
	user "aegis/module/user"
)

type Handlers struct {
	Auth         *auth.Handler
	Project      *project.Handler
	Task         *task.Handler
	Injection    *injection.Handler
	Execution    *execution.Handler
	Container    *container.Handler
	Dataset      *dataset.Handler
	Evaluation   *evaluation.Handler
	Trace        *trace.Handler
	Group        *group.Handler
	Metric       *metric.Handler
	User         *user.Handler
	RBAC         *rbac.Handler
	SDK          *sdk.Handler
	System       *system.Handler
	Notification *notification.Handler
	ChaosSystem  *chaossystem.Handler
	Team         *team.Handler
	Label        *label.Handler
	SystemMetric *systemmetric.Handler
}

func NewHandlers(
	auth *auth.Handler,
	project *project.Handler,
	task *task.Handler,
	injection *injection.Handler,
	execution *execution.Handler,
	container *container.Handler,
	dataset *dataset.Handler,
	evaluation *evaluation.Handler,
	trace *trace.Handler,
	group *group.Handler,
	metric *metric.Handler,
	user *user.Handler,
	rbac *rbac.Handler,
	sdk *sdk.Handler,
	system *system.Handler,
	notification *notification.Handler,
	chaosSystem *chaossystem.Handler,
	team *team.Handler,
	label *label.Handler,
	systemMetric *systemmetric.Handler,
) *Handlers {
	return &Handlers{
		Auth:         auth,
		Project:      project,
		Task:         task,
		Injection:    injection,
		Execution:    execution,
		Container:    container,
		Dataset:      dataset,
		Evaluation:   evaluation,
		Trace:        trace,
		Group:        group,
		Metric:       metric,
		User:         user,
		RBAC:         rbac,
		SDK:          sdk,
		System:       system,
		Notification: notification,
		ChaosSystem:  chaosSystem,
		Team:         team,
		Label:        label,
		SystemMetric: systemMetric,
	}
}
