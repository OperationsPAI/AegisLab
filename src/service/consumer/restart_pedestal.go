package consumer

import (
	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	helminfra "aegis/infra/helm"
	redisinfra "aegis/infra/redis"
	"aegis/service/common"
	"aegis/tracing"
	"aegis/utils"
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type restartPayload struct {
	pedestal      dto.ContainerVersionItem
	interval      int
	faultDuration int
	injectPayload map[string]any
}

// executeRestartPedestal handles the execution of a restart pedestal task
func executeRestartPedestal(ctx context.Context, task *dto.UnifiedTask, deps RuntimeDeps) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)
		span.AddEvent(fmt.Sprintf("Starting restarting pedestal attempt %d", task.ReStartNum+1))
		logEntry := logrus.WithFields(logrus.Fields{
			"task_id":  task.TaskID,
			"trace_id": task.TraceID,
		})
		helmGateway := deps.HelmGateway
		if helmGateway == nil {
			return handleExecutionError(span, logEntry, "helm gateway not initialized", fmt.Errorf("helm gateway not initialized"))
		}
		redisGateway := deps.RedisGateway
		if redisGateway == nil {
			return handleExecutionError(span, logEntry, "redis gateway not initialized", fmt.Errorf("redis gateway not initialized"))
		}

		rateLimiter := deps.RestartRateLimiter
		if rateLimiter == nil {
			return handleExecutionError(span, logEntry, "restart pedestal rate limiter not initialized", errors.New("restart pedestal rate limiter not initialized"))
		}
		acquired, err := rateLimiter.AcquireToken(childCtx, task.TaskID, task.TraceID)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to acquire rate limit token", err)
		}

		if !acquired {
			span.AddEvent("no token available, waiting")
			logEntry.Warn("No restart pedestal token available, waiting...")

			acquired, err = rateLimiter.WaitForToken(childCtx, task.TaskID, task.TraceID)
			if err != nil {
				return handleExecutionError(span, logEntry, "failed to wait for token", err)
			}

			if !acquired {
				if err := rescheduleRestartPedestalTask(childCtx, deps.DB, redisGateway, task, "rate limited, retrying later"); err != nil {
					return err
				}
				return nil
			}
		}

		payload, err := parseRestartPayload(task.Payload)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to parse restart payload", err)
		}

		system := chaos.SystemType(payload.pedestal.ContainerName)
		if !system.IsValid() {
			return handleExecutionError(span, logEntry, fmt.Sprintf("invalid pedestal system type: %s", payload.pedestal.Name), fmt.Errorf("invalid pedestal system type: %s", payload.pedestal.Name))
		}

		cfg, exists := config.GetChaosSystemConfigManager().Get(system)
		if !exists {
			return handleExecutionError(span, logEntry, fmt.Sprintf("no configuration found for system type: %s", system), fmt.Errorf("no configuration found for system type: %s", system))
		}

		monitor := deps.Monitor
		if monitor == nil {
			return handleExecutionError(span, logEntry, "monitor not initialized", errors.New("monitor not initialized"))
		}

		toReleased := false

		var namespace string
		defer func() {
			if acquired {
				if releaseErr := rateLimiter.ReleaseToken(childCtx, task.TaskID, task.TraceID); releaseErr != nil {
					logEntry.Errorf("failed to release restart pedestal token: %v", releaseErr)
				}
			}
			if toReleased && namespace != "" {
				if err := monitor.ReleaseLock(childCtx, namespace, task.TraceID); err != nil {
					if err := handleExecutionError(span, logEntry, fmt.Sprintf("failed to release lock for namespace %s", namespace), err); err != nil {
						logEntry.Error(err)
						return
					}
				}
			}
		}()

		t := time.Now()
		deltaTime := time.Duration(payload.interval) * consts.DefaultTimeUnit
		namespace = monitor.GetNamespaceToRestart(t.Add(deltaTime), cfg.NsPattern, task.TraceID)
		if namespace == "" {
			// Failed to acquire namespace lock, immediately release rate limit token
			if releaseErr := rateLimiter.ReleaseToken(childCtx, task.TaskID, task.TraceID); releaseErr != nil {
				logEntry.Errorf("failed to release restart pedestal token after namespace lock failure: %v", releaseErr)
			}

			acquired = false
			if err := rescheduleRestartPedestalTask(childCtx, deps.DB, redisGateway, task, "failed to acquire lock for namespace, retrying"); err != nil {
				return err
			}

			return nil
		}

		deltaTime = time.Duration(payload.interval-payload.faultDuration) * consts.DefaultTimeUnit
		injectTime := t.Add(deltaTime)

		index, err := cfg.ExtractNsNumber(namespace)
		if err != nil {
			toReleased = true
			return handleExecutionError(span, logEntry, "failed to read namespace index", err)
		}

		updateTaskState(childCtx,
			newTaskStateUpdate(
				task.TraceID,
				task.TaskID,
				consts.TaskTypeRestartPedestal,
				consts.TaskRunning,
				fmt.Sprintf("Restarting pedestal in namespace %s", namespace),
			).withSimpleEvent(consts.EventRestartPedestalStarted).withDB(deps.DB).withRedis(redisGateway),
		)

		if payload.pedestal.Extra == nil {
			toReleased = true
			publishEvent(redisGateway, childCtx, fmt.Sprintf(consts.StreamTraceLogKey, task.TraceID), dto.TraceStreamEvent{
				TaskID:    task.TaskID,
				TaskType:  consts.TaskTypeRestartPedestal,
				EventName: consts.EventRestartPedestalFailed,
				Payload:   "missing extra info in pedestal item",
			})

			return handleExecutionError(span, logEntry, "missing extra info in pedestal item", fmt.Errorf("missing extra info in pedestal item"))
		}

		if err := installPedestal(childCtx, helmGateway, namespace, index, payload.pedestal.Extra); err != nil {
			toReleased = true
			publishEvent(redisGateway, childCtx, fmt.Sprintf(consts.StreamTraceLogKey, task.TraceID), dto.TraceStreamEvent{
				TaskID:    task.TaskID,
				TaskType:  consts.TaskTypeRestartPedestal,
				EventName: consts.EventRestartPedestalFailed,
				Payload:   err.Error(),
			})

			return handleExecutionError(span, logEntry, fmt.Sprintf("failed to install pedestal of system %s", system), err)
		}

		message := fmt.Sprintf("Injection start at %s, duration %dm", injectTime.Local().String(), payload.faultDuration)
		updateTaskState(childCtx,
			newTaskStateUpdate(
				task.TraceID,
				task.TaskID,
				consts.TaskTypeRestartPedestal,
				consts.TaskCompleted,
				message,
			).withEvent(consts.EventRestartPedestalCompleted, message).withDB(deps.DB).withRedis(redisGateway),
		)

		tracing.SetSpanAttribute(childCtx, consts.TaskStateKey, consts.GetTaskStateName(consts.TaskCompleted))

		payload.injectPayload[consts.InjectNamespace] = namespace
		payload.injectPayload[consts.InjectPedestal] = system
		payload.injectPayload[consts.InjectPedestalID] = payload.pedestal.ID

		if err := common.ProduceFaultInjectionTasksWithDB(childCtx, deps.DB, deps.RedisGateway, task, injectTime, payload.injectPayload); err != nil {
			toReleased = true
			return handleExecutionError(span, logEntry, "failed to submit inject task", err)
		}

		return nil
	})
}

// rescheduleRestartPedestalTask reschedules a pedestal restart task with exponential backoff and jitter
func rescheduleRestartPedestalTask(ctx context.Context, db *gorm.DB, redisGateway *redisinfra.Gateway, task *dto.UnifiedTask, reason string) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(ctx)

		randomFactor := 0.3 + rand.Float64()*0.7 // Random factor between 0.3 and 1.0
		deltaTime := time.Duration(math.Min(math.Pow(2, float64(task.ReStartNum)), 5.0)*randomFactor) * consts.DefaultTimeUnit
		executeTime := time.Now().Add(deltaTime)

		span.AddEvent(fmt.Sprintf("rescheduling task: %s", reason))
		logrus.WithFields(logrus.Fields{
			"task_id":     task.TaskID,
			"trace_id":    task.TraceID,
			"delay_mins":  deltaTime.Minutes(),
			"retry_count": task.ReStartNum + 1,
		}).Warnf("%s: %s", reason, executeTime)

		tracing.SetSpanAttribute(ctx, consts.TaskStateKey, consts.GetTaskStateName(consts.TaskPending))

		updateTaskState(ctx,
			newTaskStateUpdate(
				task.TraceID,
				task.TaskID,
				consts.TaskTypeRestartPedestal,
				consts.TaskRescheduled,
				reason,
			).withEvent(consts.EventNoNamespaceAvailable, executeTime.String()).withDB(db).withRedis(redisGateway),
		)

		task.Reschedule(executeTime)
		if err := common.SubmitTaskWithDB(ctx, db, redisGateway, task); err != nil {
			span.RecordError(err)
			span.AddEvent("failed to submit rescheduled task")
			return fmt.Errorf("failed to submit rescheduled restart task: %w", err)
		}

		return nil
	})
}

// parseRestartPayload parses the payload for a restart pedestal task
func parseRestartPayload(payload map[string]any) (*restartPayload, error) {
	message := "invalid or missing '%s' in task payload"

	pedestal, err := utils.ConvertToType[dto.ContainerVersionItem](payload[consts.RestartPedestal])
	if err != nil {
		return nil, fmt.Errorf(message, consts.RestartPedestal)
	}

	intervalFloat, ok := payload[consts.RestartIntarval].(float64)
	if !ok || intervalFloat <= 0 {
		return nil, fmt.Errorf(message, consts.RestartIntarval)
	}
	interval := int(intervalFloat)

	faultDurationFloat, ok := payload[consts.RestartFaultDuration].(float64)
	if !ok || faultDurationFloat <= 0 {
		return nil, fmt.Errorf(message, consts.RestartFaultDuration)
	}
	faultDuration := int(faultDurationFloat)

	injectPayload, ok := payload[consts.RestartInjectPayload].(map[string]any)
	if !ok {
		return nil, fmt.Errorf(message, consts.RestartInjectPayload)
	}

	return &restartPayload{
		pedestal:      pedestal,
		interval:      interval,
		faultDuration: faultDuration,
		injectPayload: injectPayload,
	}, nil
}

// installPedestal installs or upgrades the pedestal using Helm
// Priority: Remote (if configured) -> Local fallback (if remote fails and LocalPath is set)
func installPedestal(ctx context.Context, gateway *helminfra.Gateway, releaseName string, namespaceIdx int, item *dto.HelmConfigItem) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)
		logEntry := logrus.WithFields(logrus.Fields{
			"release_name":  releaseName,
			"namespace_idx": namespaceIdx,
		})

		if item == nil {
			return handleExecutionError(span, logEntry, "missing helm config in container extra info", fmt.Errorf("missing helm config in container extra info"))
		}

		paramItems := item.DynamicValues
		for i := range paramItems {
			if paramItems[i].TemplateString != "" {
				paramItems[i].Value = fmt.Sprintf(paramItems[i].TemplateString, namespaceIdx)
			}
		}

		helmValues := item.GetValuesMap()

		// Determine chart source and installation strategy
		hasRemote := item.RepoURL != "" && item.RepoName != ""
		hasLocal := item.LocalPath != ""

		var installErr error

		if hasRemote {
			logEntry.Infof("Attempting to install chart from remote repository: %s/%s", item.RepoName, item.ChartName)

			if err := gateway.AddRepo(releaseName, item.RepoName, item.RepoURL); err != nil {
				logEntry.Warnf("Failed to add repository: %v", err)
				installErr = err
			} else if err := gateway.UpdateRepo(releaseName, item.RepoName); err != nil {
				logEntry.Warnf("Failed to update repository: %v", err)
				installErr = err
			} else {
				fullChart := fmt.Sprintf("%s/%s", item.RepoName, item.ChartName)

				logrus.WithFields(logrus.Fields{
					"release_name": releaseName,
					"chart":        fullChart,
					"version":      item.Version,
					"namespace":    releaseName,
				}).Infof("Installing Helm chart from remote with parameters: %+v", helmValues)

				if err := gateway.Install(ctx,
					releaseName,
					releaseName,
					fullChart,
					item.Version,
					helmValues,
					600*time.Second,
					300*time.Second,
				); err != nil {
					logEntry.Warnf("Failed to install chart from remote: %v", err)
					installErr = err
				} else {
					logEntry.Info("Helm chart installed successfully from remote repository")
					return nil
				}
			}
		}

		// Fallback to local chart if remote failed or not configured
		if hasLocal {
			if installErr != nil {
				logEntry.Infof("Remote installation failed, falling back to local chart: %s", item.LocalPath)
			} else {
				logEntry.Infof("Installing chart from local path: %s", item.LocalPath)
			}

			logrus.WithFields(logrus.Fields{
				"release_name": releaseName,
				"chart":        item.LocalPath,
				"namespace":    releaseName,
			}).Infof("Installing Helm chart from local path with parameters: %+v", helmValues)

			if err := gateway.Install(ctx,
				releaseName,
				releaseName,
				item.LocalPath,
				item.Version,
				helmValues,
				600*time.Second,
				360*time.Second,
			); err != nil {
				return fmt.Errorf("failed to install chart from local path %s: %w", item.LocalPath, err)
			}

			logEntry.Info("Helm chart installed successfully from local path")
			return nil
		}

		// No valid source available
		if installErr != nil {
			return fmt.Errorf("failed to install chart: remote installation failed and no local fallback available: %w", installErr)
		}

		return fmt.Errorf("no chart source configured (neither remote nor local)")
	})
}
