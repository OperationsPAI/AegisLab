package router

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
)

type Handlers struct {
	Auth         *authmodule.Handler
	Project      *projectmodule.Handler
	Task         *taskmodule.Handler
	Injection    *injectionmodule.Handler
	Execution    *executionmodule.Handler
	Container    *containermodule.Handler
	Dataset      *datasetmodule.Handler
	Evaluation   *evaluationmodule.Handler
	Trace        *tracemodule.Handler
	Group        *groupmodule.Handler
	Metric       *metricmodule.Handler
	User         *usermodule.Handler
	RBAC         *rbacmodule.Handler
	SDK          *sdkmodule.Handler
	System       *systemmodule.Handler
	Notification *notificationmodule.Handler
	ChaosSystem  *chaossystemmodule.Handler
	Team         *teammodule.Handler
	Label        *labelmodule.Handler
	SystemMetric *systemmetricmodule.Handler
}

func NewHandlers(
	auth *authmodule.Handler,
	project *projectmodule.Handler,
	task *taskmodule.Handler,
	injection *injectionmodule.Handler,
	execution *executionmodule.Handler,
	container *containermodule.Handler,
	dataset *datasetmodule.Handler,
	evaluation *evaluationmodule.Handler,
	trace *tracemodule.Handler,
	group *groupmodule.Handler,
	metric *metricmodule.Handler,
	user *usermodule.Handler,
	rbac *rbacmodule.Handler,
	sdk *sdkmodule.Handler,
	system *systemmodule.Handler,
	notification *notificationmodule.Handler,
	chaosSystem *chaossystemmodule.Handler,
	team *teammodule.Handler,
	label *labelmodule.Handler,
	systemMetric *systemmetricmodule.Handler,
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
