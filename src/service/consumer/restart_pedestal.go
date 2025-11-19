package consumer

import (
	"aegis/client"
	"aegis/consts"
	"aegis/dto"
	"aegis/service/common"
	producer "aegis/service/prodcuer"
	"aegis/tracing"
	"aegis/utils"
	"context"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type restartPayload struct {
	pedestal      dto.ContainerVersionItem
	interval      int
	faultDuration int
	injectPayload map[string]any
}

type installRelease = func(ctx context.Context, releaseName string, namespaceIdx int, info *dto.PedestalInfo) error

var installReleaseMap = map[string]installRelease{
	"ts": installTS,
}

// executeRestartPedestal handles the execution of a restart pedestal task
func executeRestartPedestal(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)
		span.AddEvent(fmt.Sprintf("Starting restarting pedestal attempt %d", task.ReStartNum+1))
		logEntry := logrus.WithFields(logrus.Fields{
			"task_id":  task.TaskID,
			"trace_id": task.TraceID,
		})

		rateLimiter := GetRestartPedestalRateLimiter()
		acquired, err := rateLimiter.AcquireToken(childCtx, task.TaskID, task.TraceID)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to acquire rate limit token", err)
		}

		if !acquired {
			span.AddEvent("no token available, waiting")
			logEntry.Info("No restart pedestal token available, waiting...")

			acquired, err = rateLimiter.WaitForToken(childCtx, task.TaskID, task.TraceID)
			if err != nil {
				return handleExecutionError(span, logEntry, "failed to wait for token", err)
			}

			if !acquired {
				if err := rescheduleRestartPedestalTask(childCtx, task, "rate limited, retrying later"); err != nil {
					return err
				}
				return nil
			}
		}

		payload, err := parseRestartPayload(task.Payload)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to parse restart payload", err)
		}

		monitor := GetMonitor()
		toReleased := false

		var namespace string
		defer func() {
			if acquired {
				if releaseErr := rateLimiter.ReleaseToken(ctx, task.TaskID, task.TraceID); releaseErr != nil {
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
		nsPrefix := payload.pedestal.Extra.HelmConfig.NsPrefix
		namespace = monitor.GetNamespaceToRestart(t.Add(deltaTime), nsPrefix, task.TraceID)
		if namespace == "" {
			// Failed to acquire namespace lock, immediately release rate limit token
			if releaseErr := rateLimiter.ReleaseToken(ctx, task.TaskID, task.TraceID); releaseErr != nil {
				logEntry.Errorf("failed to release restart pedestal token after namespace lock failure: %v", releaseErr)
			}

			acquired = false
			if err := rescheduleRestartPedestalTask(childCtx, task, "failed to acquire lock for namespace, retrying"); err != nil {
				return err
			}

			return nil
		}

		deltaTime = time.Duration(payload.interval-payload.faultDuration) * consts.DefaultTimeUnit
		injectTime := t.Add(deltaTime)

		installFunc, exists := installReleaseMap[nsPrefix]
		if !exists {
			toReleased = true
			return handleExecutionError(span, logEntry, fmt.Sprintf("no install function for namespace prefix: %s", nsPrefix), fmt.Errorf("no install function for namespace prefix: %s", nsPrefix))
		}

		_, index, err := extractNamespace(namespace)
		if err != nil {
			toReleased = true
			return handleExecutionError(span, logEntry, "failed to read namespace index", err)
		}

		updateTaskState(
			ctx,
			task.TraceID,
			task.TaskID,
			fmt.Sprintf("Restarting pedestal in namespace %s", namespace),
			consts.TaskRunning,
			consts.TaskTypeRestartPedestal,
		)

		publishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeRestartPedestal,
			EventName: consts.EventRestartPedestalStarted,
		})

		if payload.pedestal.Extra == nil {
			toReleased = true
			publishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
				TaskID:    task.TaskID,
				TaskType:  consts.TaskTypeRestartPedestal,
				EventName: consts.EventRestartPedestalFailed,
				Payload:   "missing extra info in pedestal item",
			})

			return handleExecutionError(span, logEntry, "missing extra info in pedestal item", fmt.Errorf("missing extra info in pedestal item"))
		}

		if err := installFunc(childCtx, namespace, index, payload.pedestal.Extra); err != nil {
			toReleased = true
			publishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
				TaskID:    task.TaskID,
				TaskType:  consts.TaskTypeRestartPedestal,
				EventName: consts.EventRestartPedestalFailed,
				Payload:   err.Error(),
			})

			return handleExecutionError(span, logEntry, "failed to install Train Ticket", err)
		}

		message := fmt.Sprintf("Injection start at %s, duration %dm", injectTime.Local().String(), payload.faultDuration)
		publishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeRestartPedestal,
			EventName: consts.EventRestartPedestalCompleted,
			Payload:   message,
		})

		updateTaskState(
			ctx,
			task.TraceID,
			task.TaskID,
			message,
			consts.TaskCompleted,
			consts.TaskTypeRestartPedestal,
		)

		tracing.SetSpanAttribute(ctx, consts.TaskStateKey, consts.GetTaskStateName(consts.TaskCompleted))

		payload.injectPayload[consts.InjectNamespace] = namespace
		payload.injectPayload[consts.InjectPedestalID] = payload.pedestal.ID

		if err := producer.ProduceFaultInjectionTasks(childCtx, task, injectTime, payload.injectPayload); err != nil {
			toReleased = true
			return handleExecutionError(span, logEntry, "failed to submit inject task", err)
		}

		return nil
	})
}

// rescheduleRestartPedestalTask reschedules a pedestal restart task with exponential backoff and jitter
func rescheduleRestartPedestalTask(ctx context.Context, task *dto.UnifiedTask, reason string) error {
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

		publishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeRestartPedestal,
			EventName: consts.EventNoNamespaceAvailable,
			Payload:   executeTime.String(),
		})

		updateTaskState(
			ctx,
			task.TraceID,
			task.TaskID,
			reason,
			consts.TaskRescheduled,
			consts.TaskTypeRestartPedestal,
		)

		task.Reschedule(executeTime)
		if err := common.SubmitTask(ctx, task); err != nil {
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

// extractNamespace extracts the prefix and index from a namespace string
func extractNamespace(namespace string) (string, int, error) {
	pattern := `^([a-zA-Z]+)(\d+)$`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(namespace)

	if len(match) < 3 {
		return "", 0, fmt.Errorf("failed to extract index from namespace %s", namespace)
	}

	num, err := strconv.Atoi(match[2])
	if err != nil {
		return "", 0, fmt.Errorf("failed to convert extracted index to integer: %w", err)
	}

	return match[1], num, nil
}

// installTS installs the Train Ticket pedestal using Helm
func installTS(ctx context.Context, releaseName string, namespaceIdx int, info *dto.PedestalInfo) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)
		logEntry := logrus.WithFields(logrus.Fields{
			"release_name":  releaseName,
			"namespace_idx": namespaceIdx,
		})

		if info.HelmConfig == nil {
			return handleExecutionError(span, logEntry, "missing helm config in container extra info", fmt.Errorf("missing helm config in container extra info"))
		}

		helmClient, err := client.NewHelmClient(releaseName)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to create Helm client", err)
		}

		// Add Train Ticket repository
		if err := helmClient.AddRepo(info.HelmConfig.RepoName, info.HelmConfig.RepoURL); err != nil {
			return fmt.Errorf("failed to add repository: %w", err)
		}

		// Update repositories
		if err := helmClient.UpdateRepo(); err != nil {
			return fmt.Errorf("failed to update repositories: %w", err)
		}

		paramItems := info.HelmConfig.Values
		for i := range paramItems {
			if paramItems[i].TemplateString != "" {
				paramItems[i].Value = fmt.Sprintf(paramItems[i].TemplateString, namespaceIdx)
			}
		}

		if err := helmClient.Install(ctx,
			releaseName,
			info.HelmConfig.FullChart,
			buildNestedMap(paramItems),
			600*time.Second,
			360*time.Second,
		); err != nil {
			return fmt.Errorf("failed to install Train Ticket: %w", err)
		}

		logrus.Infof("Train Ticket installed successfully in namespace %s", releaseName)
		return nil
	})
}

// buildNestedMap constructs a nested map from a list of parameter items with dot-separated keys
func buildNestedMap(items []dto.ParameterItem) map[string]any {
	root := make(map[string]any)
	for _, item := range items {
		value := item.Value

		keys := strings.Split(item.Key, ".")
		cur := root

		for i, k := range keys {
			if i == len(keys)-1 {
				cur[k] = value
				break
			}
			if _, exists := cur[k]; !exists {
				cur[k] = make(map[string]any)
			}
			cur = cur[k].(map[string]any)
		}
	}

	return root
}
