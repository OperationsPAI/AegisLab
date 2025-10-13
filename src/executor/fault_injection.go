package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"aegis/client/k8s"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/tracing"
	"aegis/utils"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
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
	labels      []dto.LabelItem
	userID      int
}

// Execute fault injection task
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
				consts.LabelTaskID:      task.TaskID,
				consts.LabelTraceID:     task.TraceID,
				consts.LabelGroupID:     task.GroupID,
				consts.LabelProjectID:   getProjectIDString(task.ProjectID),
				consts.LabelUserID:      strconv.Itoa(payload.userID),
				consts.LabelBenchmark:   payload.benchmark,
				consts.LabelPreDuration: strconv.Itoa(payload.preDuration),
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
			Status:        consts.DatapackInitial,
			Description:   fmt.Sprintf("Fault for task %s", task.TaskID),
			Benchmark:     payload.benchmark,
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

		userIDFloat, ok := payload[consts.InjectUserID].(float64)
		if !ok || userIDFloat <= 0 {
			return nil, fmt.Errorf(message, consts.InjectUserID)
		}
		userID := int(userIDFloat)

		return &injectionPayload{
			benchmark:   benchmark,
			faultType:   faultType,
			namespace:   namespace,
			preDuration: preDuration,
			displayData: displayData,
			conf:        conf,
			node:        node,
			labels:      labels,
			userID:      userID,
		}, nil
	})
}
