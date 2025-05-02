package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"reflect"
	"regexp"
	"strconv"
	"time"

	chaos "github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/client/k8s"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/tracing"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type injectionPayload struct {
	benchmark   string
	faultType   int
	namespace   string
	preDuration int
	displayData string
	conf        *chaos.InjectionConf
	node        *chaos.Node
}

type restartPayload struct {
	interval      int
	faultDuration int
	injectPayload map[string]any
}

// 执行故障注入任务
func executeFaultInjection(ctx context.Context, task *UnifiedTask) error {
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
			monitor.ReleaseLock(payload.namespace)
			span.RecordError(fmt.Errorf("failed to get namespace to inject fault: %v", err))
			span.AddEvent("failed to get namespace to inject fault")
			return err
		}

		annotations, err := getAnnotations(childCtx, task)
		if err != nil {
			monitor.ReleaseLock(payload.namespace)
			span.RecordError(err)
			span.AddEvent("failed to get annotations")
			return err
		}

		prefix, index, err := extractNamespace(payload.namespace)
		if err != nil {
			monitor.ReleaseLock(payload.namespace)
			span.RecordError(err)
			span.AddEvent("failed to read namespace index")
			return fmt.Errorf("failed to read namespace index: %v", err)
		}

		name, err := payload.conf.Create(
			childCtx,
			index,
			annotations,
			map[string]string{
				consts.CRDTaskID:      task.TaskID,
				consts.CRDTraceID:     task.TraceID,
				consts.CRDGroupID:     task.GroupID,
				consts.CRDBenchmark:   payload.benchmark,
				consts.CRDPreDuration: strconv.Itoa(payload.preDuration),
			})
		if err != nil {
			monitor.ReleaseLock(payload.namespace)
			span.RecordError(err)
			span.AddEvent("failed to inject fault")
			return fmt.Errorf("failed to inject fault: %v", err)
		}

		m, err := config.GetNsTargetMap()
		if err != nil {
			monitor.ReleaseLock(payload.namespace)
			span.RecordError(err)
			span.AddEvent("failed to get namespace target map in configuration")
			return fmt.Errorf("failed to get namespace target map in configuration: %v", err)
		}

		childNode := payload.node.Children[strconv.Itoa(payload.node.Value)]
		childNode.Children[strconv.Itoa(len(childNode.Children))] = &chaos.Node{
			Value: index % m[prefix],
		}

		engineConfig := chaos.NodeToMap(payload.node, true)
		engineData, err := json.Marshal(engineConfig)
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
			Description:   fmt.Sprintf("Fault for task %s", task.TaskID),
			Status:        consts.DatasetInitial,
			InjectionName: name,
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

func executeRestartService(ctx context.Context, task *UnifiedTask) error {
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
		namespace := monitor.AcquireLock(t.Add(deltaTime), task.TraceID)
		if namespace == "" {
			randomFactor := 0.7 + rand.Float64()*0.6 // Random factor between 0.7 and 1.3
			deltaTime = time.Duration(math.Min(math.Pow(2, float64(task.ReStartNum)), 10.0)*randomFactor) * consts.DefaultTimeUnit
			executeTime := time.Now().Add(deltaTime)

			tracing.SetSpanAttribute(ctx, consts.TaskStatusKey, string(consts.TaskStautsRescheduled))
			logrus.WithFields(logrus.Fields{
				"task_id":  task.TaskID,
				"trace_id": task.TraceID,
			}).Warnf("Failed to acquire lock for namespace, retrying at in %v", executeTime.String())
			span.AddEvent("failed to acquire lock for namespace, retrying")

			client.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), client.StreamEvent{
				TaskID:    task.TaskID,
				TaskType:  consts.TaskTypeRestartService,
				EventName: consts.EventNoNamespaceAvailable,
				Payload:   executeTime.String(),
			})

			if _, _, err := SubmitTask(ctx, &UnifiedTask{
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

		_, index, err := extractNamespace(namespace)
		if err != nil {
			monitor.ReleaseLock(namespace)
			span.RecordError(err)
			span.AddEvent("failed to read namespace index")
			return fmt.Errorf("failed to read namespace index: %v", err)
		}

		client.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), client.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeRestartService,
			EventName: consts.EventRestartServiceStarted,
		})

		if err := installTS(
			childCtx,
			namespace,
			// TODO not hard code
			fmt.Sprintf("3009%d", index),
			config.GetString("injection.ts_image_tag"),
		); err != nil {
			monitor.ReleaseLock(namespace)
			span.RecordError(err)
			span.AddEvent("failed to install Train Ticket")
			client.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), client.StreamEvent{
				TaskID:    task.TaskID,
				TaskType:  consts.TaskTypeRestartService,
				EventName: consts.EventRestartServiceFailed,
				Payload:   err.Error(),
			})
			return err
		}

		client.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), client.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeRestartService,
			EventName: consts.EventRestartServiceCompleted,
			Payload:   fmt.Sprintf("Injection started at %s, fault duration %d seconds", injectTime.String(), payload.faultDuration),
		})

		tracing.SetSpanAttribute(ctx, consts.TaskStatusKey, string(consts.TaskStatusScheduled))

		injectTask := &UnifiedTask{
			Type:         consts.TaskTypeFaultInjection,
			Payload:      payload.injectPayload,
			Immediate:    false,
			ExecuteTime:  injectTime.Unix(),
			TraceID:      task.TraceID,
			GroupID:      task.GroupID,
			TraceCarrier: task.TraceCarrier,
		}
		if _, _, err := SubmitTask(childCtx, injectTask); err != nil {
			monitor.ReleaseLock(namespace)
			span.RecordError(err)
			span.AddEvent("failed to submit inject task")
			return fmt.Errorf("failed to submit inject task: %v", err)
		}

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

		conf, err := parseStructFromMap[chaos.InjectionConf](ctx, payload, consts.InjectConf, message)
		if err != nil {
			return nil, err
		}

		node, err := parseStructFromMap[chaos.Node](ctx, payload, consts.InjectNode, message)
		if err != nil {
			return nil, err
		}

		return &injectionPayload{
			benchmark:   benchmark,
			faultType:   faultType,
			namespace:   namespace,
			preDuration: preDuration,
			displayData: displayData,
			conf:        conf,
			node:        node,
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

func parseStructFromMap[T any](ctx context.Context, payload map[string]any, key string, errorMsgTemplate string) (*T, error) {
	return tracing.WithSpanReturnValue(ctx, func(ctx context.Context) (*T, error) {
		rawValue, ok := payload[key]
		if !ok {
			return nil, fmt.Errorf(errorMsgTemplate, key)
		}

		innerMap, ok := rawValue.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: expected map[string]any, got %T", fmt.Sprintf(errorMsgTemplate, key), rawValue)
		}

		jsonData, err := json.Marshal(innerMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal intermediate map for key '%s': %w", key, err)
		}

		var result T
		if err := json.Unmarshal(jsonData, &result); err != nil {
			typeName := reflect.TypeOf(result).Name()
			if typeName == "" {
				typeName = reflect.TypeOf(result).String()
			}
			return nil, fmt.Errorf("failed to unmarshal JSON for key '%s' into type %s: %w", key, typeName, err)
		}

		return &result, nil
	})
}

func installTS(ctx context.Context, namespace, port, imageTag string) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		client, err := client.NewHelmClient(namespace)
		if err != nil {
			return fmt.Errorf("error creating Helm client: %v", err)
		}

		// Add Train Ticket repository
		if err := client.AddRepo("train-ticket", "https://cuhk-se-group.github.io/train-ticket"); err != nil {
			return fmt.Errorf("error adding repository: %v", err)
		}

		// Update repositories
		if err := client.UpdateRepo(); err != nil {
			return fmt.Errorf("error updating repositories: %v", err)
		}

		if err := client.InstallTrainTicket(namespace, imageTag, port); err != nil {
			return fmt.Errorf("error installing Train Ticket: %v", err)
		}

		logrus.Infof("Train Ticket installed successfully in namespace %s", namespace)
		return nil
	})
}

func extractNamespace(namespace string) (string, int, error) {
	pattern := `^([a-zA-Z]+)(\d+)$`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(namespace)

	if len(match) < 3 {
		return "", 0, fmt.Errorf("failed to extract index from namespace %s", namespace)
	}

	if _, ok := config.GetMap("injection.namespace_target_map")[match[1]]; !ok {
		return "", 0, fmt.Errorf("namespace %s is not defined in configuration 'injection.namespace_target_map'", match[1])
	}

	num, err := strconv.Atoi(match[2])
	if err != nil {
		return "", 0, fmt.Errorf("failed to convert extracted index to integer: %v", err)
	}

	return match[1], num, nil
}
