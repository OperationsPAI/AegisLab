package gatewayapp

import (
	"aegis/app"
	"aegis/internalclient/iamclient"
	"aegis/internalclient/orchestratorclient"
	"aegis/internalclient/resourceclient"
	"aegis/internalclient/systemclient"
	"aegis/middleware"
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
	systemmodule "aegis/module/system"
	systemmetricmodule "aegis/module/systemmetric"
	taskmodule "aegis/module/task"
	teammodule "aegis/module/team"
	tracemodule "aegis/module/trace"
	usermodule "aegis/module/user"

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
		app.ProducerCompatibilityOptions(port),
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
		fx.Decorate(func(local authmodule.HandlerService, remote *iamclient.Client) authmodule.HandlerService {
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
		fx.Decorate(func(local usermodule.HandlerService, remote *iamclient.Client) usermodule.HandlerService {
			return remoteAwareUserService{
				HandlerService: local,
				iam:            remote,
			}
		}),
		fx.Decorate(func(local rbacmodule.HandlerService, remote *iamclient.Client) rbacmodule.HandlerService {
			return remoteAwareRBACService{
				HandlerService: local,
				iam:            remote,
			}
		}),
		fx.Decorate(func(local teammodule.HandlerService, remote *iamclient.Client) teammodule.HandlerService {
			return remoteAwareTeamService{
				HandlerService: local,
				iam:            remote,
			}
		}),
		fx.Decorate(func(local executionmodule.HandlerService, remote *orchestratorclient.Client) executionmodule.HandlerService {
			return remoteAwareExecutionService{
				HandlerService: local,
				orchestrator:   remote,
			}
		}),
		fx.Decorate(func(local injectionmodule.HandlerService, remote *orchestratorclient.Client) injectionmodule.HandlerService {
			return remoteAwareInjectionService{
				HandlerService: local,
				orchestrator:   remote,
			}
		}),
		fx.Decorate(func(local taskmodule.HandlerService, remote *orchestratorclient.Client) taskmodule.HandlerService {
			return remoteAwareTaskService{
				HandlerService: local,
				orchestrator:   remote,
			}
		}),
		fx.Decorate(func(local tracemodule.HandlerService, remote *orchestratorclient.Client) tracemodule.HandlerService {
			return remoteAwareTraceService{
				HandlerService: local,
				orchestrator:   remote,
			}
		}),
		fx.Decorate(func(local groupmodule.HandlerService, remote *orchestratorclient.Client) groupmodule.HandlerService {
			return remoteAwareGroupService{
				HandlerService: local,
				orchestrator:   remote,
			}
		}),
		fx.Decorate(func(local notificationmodule.HandlerService, remote *orchestratorclient.Client) notificationmodule.HandlerService {
			return remoteAwareNotificationService{
				HandlerService: local,
				orchestrator:   remote,
			}
		}),
		fx.Decorate(func(local projectmodule.HandlerService, remote *resourceclient.Client) projectmodule.HandlerService {
			return remoteAwareProjectService{
				HandlerService: local,
				resource:       remote,
			}
		}),
		fx.Decorate(func(local containermodule.HandlerService, remote *resourceclient.Client) containermodule.HandlerService {
			return remoteAwareContainerService{
				HandlerService: local,
				resource:       remote,
			}
		}),
		fx.Decorate(func(local datasetmodule.HandlerService, remote *resourceclient.Client) datasetmodule.HandlerService {
			return remoteAwareDatasetService{
				HandlerService: local,
				resource:       remote,
			}
		}),
		fx.Decorate(func(local evaluationmodule.HandlerService, remote *resourceclient.Client) evaluationmodule.HandlerService {
			return remoteAwareEvaluationService{
				HandlerService: local,
				resource:       remote,
			}
		}),
		fx.Decorate(func(local labelmodule.HandlerService, remote *resourceclient.Client) labelmodule.HandlerService {
			return remoteAwareLabelService{
				HandlerService: local,
				resource:       remote,
			}
		}),
		fx.Decorate(func(local chaossystemmodule.HandlerService, remote *resourceclient.Client) chaossystemmodule.HandlerService {
			return remoteAwareChaosSystemService{
				HandlerService: local,
				resource:       remote,
			}
		}),
		fx.Decorate(func(local metricmodule.HandlerService, orchestrator *orchestratorclient.Client, resource *resourceclient.Client) metricmodule.HandlerService {
			return remoteAwareMetricService{
				HandlerService: local,
				orchestrator:   orchestrator,
				resource:       resource,
			}
		}),
		fx.Decorate(func(local systemmodule.HandlerService, remote *systemclient.Client) systemmodule.HandlerService {
			return remoteAwareSystemService{
				HandlerService: local,
				system:         remote,
			}
		}),
		fx.Decorate(func(local systemmetricmodule.HandlerService, remote *systemclient.Client) systemmetricmodule.HandlerService {
			return remoteAwareSystemMetricService{
				HandlerService: local,
				system:         remote,
			}
		}),
	)
}
