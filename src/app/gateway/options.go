package gateway

import (
	"aegis/app"
	chaos "aegis/infra/chaos"
	k8s "aegis/infra/k8s"
	"aegis/internalclient/iamclient"
	"aegis/internalclient/orchestratorclient"
	"aegis/internalclient/resourceclient"
	"aegis/internalclient/systemclient"
	"aegis/middleware"
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
	system "aegis/module/system"
	systemmetric "aegis/module/systemmetric"
	task "aegis/module/task"
	team "aegis/module/team"
	trace "aegis/module/trace"
	user "aegis/module/user"

	"go.uber.org/fx"
)

// Options builds the dedicated api-gateway runtime.
func Options(confPath, port string) fx.Option {
	return fx.Options(
		app.BaseOptions(confPath),
		app.ObserveOptions(),
		app.DataOptions(),
		app.CoordinationOptions(),
		app.BuildInfraOptions(),
		chaos.Module,
		k8s.Module,
		app.ProducerHTTPOptions(port),
		app.RequireConfiguredTargets(
			"api-gateway",
			app.RequiredConfigTarget{Name: "iam-service", PrimaryKey: "clients.iam.target", LegacyKey: "iam.grpc.target"},
			app.RequiredConfigTarget{Name: "orchestrator-service", PrimaryKey: "clients.orchestrator.target", LegacyKey: "orchestrator.grpc.target"},
			app.RequiredConfigTarget{Name: "resource-service", PrimaryKey: "clients.resource.target", LegacyKey: "resource.grpc.target"},
			app.RequiredConfigTarget{Name: "system-service", PrimaryKey: "clients.system.target", LegacyKey: "system.grpc.target"},
		),
		iamclient.Module,
		orchestratorclient.Module,
		resourceclient.Module,
		systemclient.Module,
		fx.Decorate(func(local auth.HandlerService, remote *iamclient.Client) auth.HandlerService {
			return remoteAwareAuthService{
				HandlerService: local,
				iam:            remote,
			}
		}),
		fx.Decorate(func(local middleware.Service, remote *iamclient.Client) middleware.Service {
			return remoteAwareMiddlewareService{
				base: local,
				iam:  remote,
			}
		}),
		fx.Decorate(func(local user.HandlerService, remote *iamclient.Client) user.HandlerService {
			return remoteAwareUserService{
				HandlerService: local,
				iam:            remote,
			}
		}),
		fx.Decorate(func(local rbac.HandlerService, remote *iamclient.Client) rbac.HandlerService {
			return remoteAwareRBACService{
				HandlerService: local,
				iam:            remote,
			}
		}),
		fx.Decorate(func(local team.HandlerService, remote *iamclient.Client) team.HandlerService {
			return remoteAwareTeamService{
				HandlerService: local,
				iam:            remote,
			}
		}),
		fx.Decorate(func(local execution.HandlerService, remote *orchestratorclient.Client) execution.HandlerService {
			return remoteAwareExecutionService{
				HandlerService: local,
				orchestrator:   remote,
			}
		}),
		fx.Decorate(func(local injection.HandlerService, remote *orchestratorclient.Client) injection.HandlerService {
			return remoteAwareInjectionService{
				HandlerService: local,
				orchestrator:   remote,
			}
		}),
		fx.Decorate(func(local task.HandlerService, remote *orchestratorclient.Client) task.HandlerService {
			return remoteAwareTaskService{
				HandlerService: local,
				orchestrator:   remote,
			}
		}),
		fx.Decorate(func(local trace.HandlerService, remote *orchestratorclient.Client) trace.HandlerService {
			return remoteAwareTraceService{
				HandlerService: local,
				orchestrator:   remote,
			}
		}),
		fx.Decorate(func(local group.HandlerService, remote *orchestratorclient.Client) group.HandlerService {
			return remoteAwareGroupService{
				HandlerService: local,
				orchestrator:   remote,
			}
		}),
		fx.Decorate(func(local notification.HandlerService, remote *orchestratorclient.Client) notification.HandlerService {
			return remoteAwareNotificationService{
				HandlerService: local,
				orchestrator:   remote,
			}
		}),
		fx.Decorate(func(local project.HandlerService, remote *resourceclient.Client) project.HandlerService {
			return remoteAwareProjectService{
				HandlerService: local,
				resource:       remote,
			}
		}),
		fx.Decorate(func(local container.HandlerService, remote *resourceclient.Client) container.HandlerService {
			return remoteAwareContainerService{
				HandlerService: local,
				resource:       remote,
			}
		}),
		fx.Decorate(func(local dataset.HandlerService, remote *resourceclient.Client) dataset.HandlerService {
			return remoteAwareDatasetService{
				HandlerService: local,
				resource:       remote,
			}
		}),
		fx.Decorate(func(local evaluation.HandlerService, remote *resourceclient.Client) evaluation.HandlerService {
			return remoteAwareEvaluationService{
				HandlerService: local,
				resource:       remote,
			}
		}),
		fx.Decorate(func(local label.HandlerService, remote *resourceclient.Client) label.HandlerService {
			return remoteAwareLabelService{
				HandlerService: local,
				resource:       remote,
			}
		}),
		fx.Decorate(func(local chaossystem.HandlerService, remote *resourceclient.Client) chaossystem.HandlerService {
			return remoteAwareChaosSystemService{
				HandlerService: local,
				resource:       remote,
			}
		}),
		fx.Decorate(func(local metric.HandlerService, orchestrator *orchestratorclient.Client, resource *resourceclient.Client) metric.HandlerService {
			return remoteAwareMetricService{
				HandlerService: local,
				orchestrator:   orchestrator,
				resource:       resource,
			}
		}),
		fx.Decorate(func(local system.HandlerService, remote *systemclient.Client) system.HandlerService {
			return remoteAwareSystemService{
				HandlerService: local,
				system:         remote,
			}
		}),
		fx.Decorate(func(local systemmetric.HandlerService, remote *systemclient.Client) systemmetric.HandlerService {
			return remoteAwareSystemMetricService{
				HandlerService: local,
				system:         remote,
			}
		}),
	)
}
