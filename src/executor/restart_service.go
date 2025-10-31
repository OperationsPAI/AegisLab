package executor

import (
	"aegis/client"
	"aegis/client/k8s"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/tracing"
	"aegis/utils"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type installRelease = func(ctx context.Context, releaseName string, namespaceIdx int, containerVersion database.ContainerVersion, helmConfig database.HelmConfig) error

type restartPayload struct {
	containerVersion database.ContainerVersion
	helmConfig       database.HelmConfig
	interval         int
	faultDuration    int
	injectPayload    map[string]any
}

var installReleaseMap = map[string]installRelease{
	"ts": installTS,
}

func executeRestartService(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(ctx)
		span.AddEvent(fmt.Sprintf("Starting retry attempt %d", task.ReStartNum+1))

		rateLimiter := utils.NewRestartServiceRateLimiter()

		acquired, err := rateLimiter.AcquireToken(childCtx, task.TaskID, task.TraceID)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to acquire rate limit token")
			return fmt.Errorf("failed to acquire rate limit token: %v", err)
		}

		if !acquired {
			span.AddEvent("no token available, waiting")
			logrus.WithFields(logrus.Fields{
				"task_id":  task.TaskID,
				"trace_id": task.TraceID,
			}).Info("No restart service token available, waiting...")

			acquired, err = rateLimiter.WaitForToken(childCtx, task.TaskID, task.TraceID)
			if err != nil {
				span.RecordError(err)
				span.AddEvent("failed to wait for token")
				return fmt.Errorf("failed to wait for token: %v", err)
			}

			if !acquired {
				if err := rescheduleTask(childCtx, task, "rate limited, retrying later"); err != nil {
					return err
				}
				return nil
			}
		}

		var tokenAcquired = true
		defer func() {
			if tokenAcquired {
				if releaseErr := rateLimiter.ReleaseToken(ctx, task.TaskID, task.TraceID); releaseErr != nil {
					logrus.WithFields(logrus.Fields{
						"task_id":  task.TaskID,
						"trace_id": task.TraceID,
						"error":    releaseErr,
					}).Error("Failed to release restart service token")
				}
			}
		}()

		payload, err := parseRestartPayload(childCtx, task.Payload)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to parse restart payload")
			return err
		}

		monitor := k8s.GetMonitor()

		t := time.Now()
		deltaTime := time.Duration(payload.interval) * consts.DefaultTimeUnit
		nsPrefix := payload.helmConfig.NsPrefix
		namespace := monitor.GetNamespaceToRestart(t.Add(deltaTime), nsPrefix, task.TraceID)
		if namespace == "" {
			// Failed to acquire namespace lock, immediately release rate limit token
			if releaseErr := rateLimiter.ReleaseToken(ctx, task.TaskID, task.TraceID); releaseErr != nil {
				logrus.WithFields(logrus.Fields{
					"task_id":  task.TaskID,
					"trace_id": task.TraceID,
					"error":    releaseErr,
				}).Error("Failed to release restart service token after namespace lock failure")
			}
			tokenAcquired = false // Mark token as released to avoid duplicate release in defer

			if err := rescheduleTask(childCtx, task, "failed to acquire lock for namespace, retrying"); err != nil {
				return err
			}

			return nil
		}

		payload.injectPayload[consts.InjectNamespace] = namespace
		deltaTime = time.Duration(payload.interval-payload.faultDuration) * consts.DefaultTimeUnit
		injectTime := t.Add(deltaTime)

		_, index, err := extractNamespace(namespace)
		if err != nil {
			if err := monitor.ReleaseLock(namespace, task.TraceID); err != nil {
				logrus.Errorf("failed to release lock for namespace %s: %v", namespace, err)
			}
			span.RecordError(err)
			span.AddEvent("failed to read namespace index")
			return fmt.Errorf("failed to read namespace index: %v", err)
		}

		updateTaskStatus(
			ctx,
			task.TraceID,
			task.TaskID,
			fmt.Sprintf("Restarting service in namespace %s", namespace),
			consts.TaskStatusRunning,
			consts.TaskTypeRestartService,
		)

		repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeRestartService,
			EventName: consts.EventRestartServiceStarted,
		})

		installFunc, exists := installReleaseMap[nsPrefix]
		if !exists {
			if err := monitor.ReleaseLock(namespace, task.TraceID); err != nil {
				logrus.Errorf("failed to release lock for namespace %s: %v", namespace, err)
			}
			span.AddEvent("no install function for namespace prefix")
			return fmt.Errorf("no install function for namespace prefix: %s", nsPrefix)
		}

		if err := installFunc(childCtx, namespace, index, payload.containerVersion, payload.helmConfig); err != nil {
			if err := monitor.ReleaseLock(namespace, task.TraceID); err != nil {
				logrus.Errorf("failed to release lock for namespace %s: %v", namespace, err)
			}
			span.RecordError(err)
			span.AddEvent("failed to install Train Ticket")

			repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
				TaskID:    task.TaskID,
				TaskType:  consts.TaskTypeRestartService,
				EventName: consts.EventRestartServiceFailed,
				Payload:   err.Error(),
			})

			return err
		}

		message := fmt.Sprintf("Injection start at %s, duration %dm", injectTime.Local().String(), payload.faultDuration)
		repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeRestartService,
			EventName: consts.EventRestartServiceCompleted,
			Payload:   message,
		})

		updateTaskStatus(
			ctx,
			task.TraceID,
			task.TaskID,
			message,
			consts.TaskStatusCompleted,
			consts.TaskTypeRestartService,
		)

		tracing.SetSpanAttribute(ctx, consts.TaskStatusKey, string(consts.TaskStatusCompleted))

		injectTask := &dto.UnifiedTask{
			Type:         consts.TaskTypeFaultInjection,
			Payload:      payload.injectPayload,
			Immediate:    false,
			ExecuteTime:  injectTime.Unix(),
			TraceID:      task.TraceID,
			GroupID:      task.GroupID,
			ProjectID:    task.ProjectID,
			TraceCarrier: task.TraceCarrier,
		}
		if _, _, err := SubmitTask(childCtx, injectTask); err != nil {
			if err := monitor.ReleaseLock(namespace, task.TraceID); err != nil {
				logrus.Errorf("failed to release lock for namespace %s: %v", namespace, err)
			}
			span.RecordError(err)
			span.AddEvent("failed to submit inject task")
			return fmt.Errorf("failed to submit inject task: %v", err)
		}

		return nil
	})
}

func rescheduleTask(ctx context.Context, task *dto.UnifiedTask, reason string) error {
	span := trace.SpanFromContext(ctx)

	var executeTime time.Time

	randomFactor := 0.3 + rand.Float64()*0.7 // Random factor between 0.3 and 1.0
	deltaTime := time.Duration(math.Min(math.Pow(2, float64(task.ReStartNum)), 5.0)*randomFactor) * consts.DefaultTimeUnit
	executeTime = time.Now().Add(deltaTime)

	eventPayload := executeTime.String()

	span.AddEvent(fmt.Sprintf("rescheduling task: %s", reason))
	logrus.WithFields(logrus.Fields{
		"task_id":  task.TaskID,
		"trace_id": task.TraceID,
	}).Warnf("%s: %s", reason, executeTime)

	tracing.SetSpanAttribute(ctx, consts.TaskStatusKey, string(consts.TaskStatusPending))

	repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
		TaskID:    task.TaskID,
		TaskType:  consts.TaskTypeRestartService,
		EventName: consts.EventNoNamespaceAvailable,
		Payload:   eventPayload,
	})

	updateTaskStatus(
		ctx,
		task.TraceID,
		task.TaskID,
		reason,
		consts.TaskStautsRescheduled,
		consts.TaskTypeRestartService,
	)

	if _, _, err := SubmitTask(ctx, &dto.UnifiedTask{
		TaskID:       task.TaskID,
		Type:         consts.TaskTypeRestartService,
		Immediate:    false,
		ExecuteTime:  executeTime.Unix(),
		ReStartNum:   task.ReStartNum + 1,
		Payload:      task.Payload,
		Status:       consts.TaskStautsRescheduled,
		TraceID:      task.TraceID,
		GroupID:      task.GroupID,
		ProjectID:    task.ProjectID,
		UserID:       task.UserID,
		TraceCarrier: task.TraceCarrier,
	}); err != nil {
		span.RecordError(err)
		span.AddEvent("failed to submit rescheduled task")
		return fmt.Errorf("failed to submit rescheduled restart task: %v", err)
	}

	return nil
}

func parseRestartPayload(ctx context.Context, payload map[string]any) (*restartPayload, error) {
	return tracing.WithSpanReturnValue(ctx, func(childCtx context.Context) (*restartPayload, error) {
		message := "invalid or missing '%s' in task payload"

		containerVersion, err := utils.ConvertToType[database.ContainerVersion](payload[consts.RestartContainerVersion])
		if err != nil {
			return nil, fmt.Errorf(message, consts.RestartContainerVersion)
		}

		helmConfig, err := utils.ConvertToType[database.HelmConfig](payload[consts.RestartHelmConfig])
		if err != nil {
			return nil, fmt.Errorf(message, consts.RestartHelmConfig)
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
			containerVersion: containerVersion,
			helmConfig:       helmConfig,
			interval:         interval,
			faultDuration:    faultDuration,
			injectPayload:    injectPayload,
		}, nil
	})
}

func extractNamespace(namespace string) (string, int, error) {
	pattern := `^([a-zA-Z]+)(\d+)$`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(namespace)

	if len(match) < 3 {
		return "", 0, fmt.Errorf("failed to extract index from namespace %s", namespace)
	}

	num, err := strconv.Atoi(match[2])
	if err != nil {
		return "", 0, fmt.Errorf("failed to convert extracted index to integer: %v", err)
	}

	return match[1], num, nil
}

func installTS(ctx context.Context, releaseName string, namespaceIdx int, containerVersion database.ContainerVersion, helmConfig database.HelmConfig) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		helmClient, err := client.NewHelmClient(releaseName)
		if err != nil {
			return fmt.Errorf("failed to create Helm client: %v", err)
		}

		// Add Train Ticket repository
		if err := helmClient.AddRepo(helmConfig.RepoName, helmConfig.RepoURL); err != nil {
			return fmt.Errorf("failed to add repository: %v", err)
		}

		// Update repositories
		if err := helmClient.UpdateRepo(); err != nil {
			return fmt.Errorf("failed to update repositories: %v", err)
		}

		baseValues := map[string]any{
			"global": map[string]any{
				"image": map[string]any{
					"repository": fmt.Sprintf("%s/%s", containerVersion.Registry, containerVersion.Namespace),
					"tag":        containerVersion.Tag,
				},
			},
			"services": map[string]any{
				"tsUiDashboard": map[string]any{
					"nodePort": fmt.Sprintf(helmConfig.PortTemplate, namespaceIdx),
				},
			},
		}

		var dbValues map[string]any
		if helmConfig.Values != "" {
			if err := json.Unmarshal([]byte(helmConfig.Values), &dbValues); err != nil {
				return fmt.Errorf("failed to unmarshal database values: %v", err)
			}
		}

		values := baseValues
		if dbValues != nil {
			values = utils.DeepMergeClone(baseValues, dbValues)
		}

		if err := helmClient.Install(ctx,
			releaseName,
			helmConfig.FullChart,
			values,
			600*time.Second,
			360*time.Second,
		); err != nil {
			return fmt.Errorf("failed to install Train Ticket: %v", err)
		}

		logrus.Infof("Train Ticket installed successfully in namespace %s", releaseName)
		return nil
	})
}
