package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"regexp"
	"strconv"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/LGU-SE-Internal/rcabench/client"
	"github.com/LGU-SE-Internal/rcabench/client/k8s"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/LGU-SE-Internal/rcabench/tracing"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type nsConfig struct {
	Port      string `json:"port"`
	ChartName string `json:"chart_name"`
	ImageName string `json:"image_name"`
	ImageTag  string `json:"image_tag"`
	RepoName  string `json:"repo_name"`
	RepoURL   string `json:"repo_url"`
}

type injectionPayload struct {
	benchmark   string
	faultType   int
	namespace   string
	preDuration int
	displayData string
	conf        *chaos.InjectionConf
	node        *chaos.Node
	labels      []dto.LabelItem
}

type restartPayload struct {
	interval      int
	faultDuration int
	injectPayload map[string]any
}

// 执行故障注入任务
func executeFaultInjection(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(ctx)

		payload, err := parseInjectionPayload(childCtx, task.Payload)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to parse injection payload")
			return err
		}

		monitor := k8s.GetMonitor()
		if err := monitor.CheckNamespaceToInject(payload.namespace, time.Now(), task.TraceID); err != nil {
			monitor.ReleaseLock(payload.namespace, task.TraceID)
			span.RecordError(fmt.Errorf("failed to get namespace to inject fault: %v", err))
			span.AddEvent("failed to get namespace to inject fault")
			return err
		}

		annotations, err := getAnnotations(childCtx, task)
		if err != nil {
			monitor.ReleaseLock(payload.namespace, task.TraceID)
			span.RecordError(err)
			span.AddEvent("failed to get annotations")
			return err
		}

		name, err := payload.conf.Create(
			childCtx,
			payload.namespace,
			annotations,
			map[string]string{
				consts.CRDTaskID:      task.TaskID,
				consts.CRDTraceID:     task.TraceID,
				consts.CRDGroupID:     task.GroupID,
				consts.CRDBenchmark:   payload.benchmark,
				consts.CRDPreDuration: strconv.Itoa(payload.preDuration),
			})
		if err != nil {
			monitor.ReleaseLock(payload.namespace, task.TraceID)
			span.RecordError(err)
			span.AddEvent("failed to inject fault")
			return fmt.Errorf("failed to inject fault: %v", err)
		}

		repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeFaultInjection,
			EventName: consts.EventFaultInjectionStarted,
		})

		engineData, err := json.Marshal(payload.node)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to marshal injection spec to engine config")
			return fmt.Errorf("failed to marshal injection spec to engine config: %v", err)
		}

		faultRecord := database.FaultInjectionSchedule{
			TaskID:        task.TaskID,
			FaultType:     payload.faultType,
			DisplayConfig: payload.displayData,
			EngineConfig:  string(engineData),
			PreDuration:   payload.preDuration,
			Status:        consts.DatasetInitial,
			Description:   fmt.Sprintf("Fault for task %s", task.TaskID),
			Benchmark:     payload.benchmark,
			InjectionName: name,
			Labels:        make(database.LabelsMap),
		}

		for _, label := range payload.labels {
			faultRecord.Labels[label.Key] = label.Value
		}

		if err = database.DB.Create(&faultRecord).Error; err != nil {
			span.RecordError(err)
			span.AddEvent("failed to write fault injection schedule to database")
			logrus.Errorf("failed to write fault injection schedule to database: %v", err)
			return fmt.Errorf("failed to write to database")
		}

		return nil
	})
}

func executeRestartService(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(ctx)
		span.AddEvent(fmt.Sprintf("Starting retry attempt %d", task.ReStartNum+1))

		payload, err := parseRestartPayload(childCtx, task.Payload)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to parse restart payload")
			return err
		}

		monitor := k8s.GetMonitor()

		t := time.Now()
		deltaTime := time.Duration(payload.interval) * consts.DefaultTimeUnit
		namespace := monitor.GetNamespaceToRestart(t.Add(deltaTime), task.TraceID)
		if namespace == "" {
			randomFactor := 0.3 + rand.Float64()*0.7 // Random factor between 0.3 and 1.0
			deltaTime = time.Duration(math.Min(math.Pow(2, float64(task.ReStartNum)), 5.0)*randomFactor) * consts.DefaultTimeUnit
			executeTime := time.Now().Add(deltaTime)

			tracing.SetSpanAttribute(ctx, consts.TaskStatusKey, string(consts.TaskStautsRescheduled))
			logrus.WithFields(logrus.Fields{
				"task_id":  task.TaskID,
				"trace_id": task.TraceID,
			}).Warnf("Failed to acquire lock for namespace, retrying at in %v", executeTime.String())
			span.AddEvent("failed to acquire lock for namespace, retrying")

			repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
				TaskID:    task.TaskID,
				TaskType:  consts.TaskTypeRestartService,
				EventName: consts.EventNoNamespaceAvailable,
				Payload:   executeTime.String(),
			})

			updateTaskStatus(
				ctx,
				task.TraceID,
				task.TaskID,
				"failed to acquire lock for namespace, retrying",
				consts.TaskStautsRescheduled,
				consts.TaskTypeRestartService,
			)

			if _, _, err := SubmitTask(childCtx, &dto.UnifiedTask{
				Type:         consts.TaskTypeRestartService,
				Immediate:    false,
				ExecuteTime:  executeTime.Unix(),
				ReStartNum:   task.ReStartNum + 1,
				Payload:      task.Payload,
				TraceID:      task.TraceID,
				GroupID:      task.GroupID,
				TraceCarrier: task.TraceCarrier,
			}); err != nil {
				span.RecordError(err)
				span.AddEvent("failed to submit restart task")
				return fmt.Errorf("failed to submit restart task: %v", err)
			}

			return nil
		}

		payload.injectPayload[consts.InjectNamespace] = namespace
		deltaTime = time.Duration(payload.interval-payload.faultDuration) * consts.DefaultTimeUnit
		injectTime := t.Add(deltaTime)

		nsPrefix, index, err := extractNamespace(namespace)
		if err != nil {
			monitor.ReleaseLock(namespace, task.TraceID)
			span.RecordError(err)
			span.AddEvent("failed to read namespace index")
			return fmt.Errorf("failed to read namespace index: %v", err)
		}

		repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeRestartService,
			EventName: consts.EventRestartServiceStarted,
		})

		if err := installTS(childCtx, namespace, nsPrefix, index); err != nil {
			monitor.ReleaseLock(namespace, task.TraceID)
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

		tracing.SetSpanAttribute(ctx, consts.TaskStatusKey, string(consts.TaskStatusScheduled))

		injectTask := &dto.UnifiedTask{
			Type:         consts.TaskTypeFaultInjection,
			Payload:      payload.injectPayload,
			Immediate:    false,
			ExecuteTime:  injectTime.Unix(),
			TraceID:      task.TraceID,
			GroupID:      task.GroupID,
			TraceCarrier: task.TraceCarrier,
		}
		if _, _, err := SubmitTask(childCtx, injectTask); err != nil {
			monitor.ReleaseLock(namespace, task.TraceID)
			span.RecordError(err)
			span.AddEvent("failed to submit inject task")
			return fmt.Errorf("failed to submit inject task: %v", err)
		}

		return nil
	})
}

func installTS(ctx context.Context, namespace, nsPrefix string, namespaceIdx int) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		nsConfigMap, err := config.GetNsConfigMap()
		if err != nil {
			return fmt.Errorf("error getting namespace config map: %v", err)
		}

		payload, exists := nsConfigMap[nsPrefix]
		if !exists {
			return fmt.Errorf("namespace %s not found in config map", nsPrefix)
		}

		nsConfig, err := utils.ConvertToType[nsConfig](payload)
		if err != nil {
			return err
		}

		helmClient, err := client.NewHelmClient(namespace)
		if err != nil {
			return fmt.Errorf("error creating Helm client: %v", err)
		}

		// Add Train Ticket repository
		if err := helmClient.AddRepo(nsConfig.RepoName, nsConfig.RepoURL); err != nil {
			return fmt.Errorf("error adding repository: %v", err)
		}

		// Update repositories
		if err := helmClient.UpdateRepo(); err != nil {
			return fmt.Errorf("error updating repositories: %v", err)
		}

		port := fmt.Sprintf(nsConfig.Port, namespaceIdx)
		chartName := fmt.Sprintf("%s/%s", nsConfig.RepoName, nsConfig.ChartName)
		if err := helmClient.InstallTrainTicket(ctx,
			namespace,
			chartName,
			nsConfig.ImageName,
			nsConfig.ImageTag,
			port,
			500*time.Second,
			300*time.Second,
		); err != nil {
			return fmt.Errorf("error installing Train Ticket: %v", err)
		}

		logrus.Infof("Train Ticket installed successfully in namespace %s", namespace)
		return nil
	})
}

func parseInjectionPayload(ctx context.Context, payload map[string]any) (*injectionPayload, error) {
	return tracing.WithSpanReturnValue(ctx, func(childCtx context.Context) (*injectionPayload, error) {
		message := "invalid or missing '%s' in task payload"

		benchmark, ok := payload[consts.InjectBenchmark].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.InjectBenchmark)
		}

		faultTypeFloat, ok := payload[consts.InjectFaultType].(float64)
		if !ok || faultTypeFloat < 0 {
			return nil, fmt.Errorf(message, consts.InjectFaultType)
		}
		faultType := int(faultTypeFloat)

		namespace, ok := payload[consts.InjectNamespace].(string)
		if !ok || namespace == "" {
			return nil, fmt.Errorf(message, consts.InjectNamespace)
		}

		preDurationFloat, ok := payload[consts.InjectPreDuration].(float64)
		if !ok || preDurationFloat <= 0 {
			return nil, fmt.Errorf(message, consts.InjectPreDuration)
		}
		preDuration := int(preDurationFloat)

		displayData, ok := payload[consts.InjectDisplayData].(string)
		if !ok || displayData == "" {
			return nil, fmt.Errorf(message, consts.InjectDisplayData)
		}

		conf, err := utils.MapToStruct[chaos.InjectionConf](payload, consts.InjectConf, message)
		if err != nil {
			return nil, err
		}

		node, err := utils.MapToStruct[chaos.Node](payload, consts.InjectNode, message)
		if err != nil {
			return nil, err
		}

		labels, err := utils.ConvertToType[[]dto.LabelItem](payload[consts.InjectLabels])
		if err != nil {
			return nil, fmt.Errorf(message, consts.InjectLabels)
		}

		return &injectionPayload{
			benchmark:   benchmark,
			faultType:   faultType,
			namespace:   namespace,
			preDuration: preDuration,
			displayData: displayData,
			conf:        conf,
			node:        node,
			labels:      labels,
		}, nil
	})
}

func parseRestartPayload(ctx context.Context, payload map[string]any) (*restartPayload, error) {
	return tracing.WithSpanReturnValue(ctx, func(childCtx context.Context) (*restartPayload, error) {
		message := "invalid or missing '%s' in task payload"

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
			interval:      interval,
			faultDuration: faultDuration,
			injectPayload: injectPayload,
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
