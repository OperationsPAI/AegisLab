package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

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
	benchmark   dto.ContainerVersionItem
	preDuration int
	node        *chaos.Node
	namespace   string
	pedestalID  int
	labels      []dto.LabelItem
}

// executeFaultInjection handles the injection of a fault task
func executeFaultInjection(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)
		logEntry := logrus.WithFields(logrus.Fields{
			"task_id":  task.TaskID,
			"trace_id": task.TraceID,
		})

		payload, err := parseInjectionPayload(task.Payload)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to parse injection payload", err)
		}

		injectionConf, err := chaos.NodeToStruct[chaos.InjectionConf](payload.node)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to convert node to injection conf", err)
		}

		displayMap, err := injectionConf.GetDisplayConfig()
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to get display config", err)
		}

		groundtruth, err := injectionConf.GetGroundtruth()
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to get groundtruth", err)
		}

		displayData, err := json.Marshal(displayMap)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to marshal injection spec to display config", err)
		}

		engineData, err := json.Marshal(payload.node)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to marshal injection spec to engine config", err)
		}

		injection := &database.FaultInjection{
			FaultType:     chaos.ChaosType(payload.node.Value),
			Description:   fmt.Sprintf("Fault for task %s", task.TaskID),
			DisplayConfig: utils.StringPtr(string(displayData)),
			EngineConfig:  string(engineData),
			Groundtruth:   database.NewDBGroundtruth(&groundtruth),
			PreDuration:   payload.preDuration,
			State:         consts.DatapackInitial,
			Status:        consts.CommonEnabled,
			TaskID:        &task.TaskID,
			BenchmarkID:   payload.benchmark.ID,
			PedestalID:    payload.pedestalID,
		}

		if err = repository.CreateInjection(database.DB, injection); err != nil {
			return handleExecutionError(span, logEntry, "failed to write fault injection schedule to database", err)
		}

		annotations, err := task.GetAnnotations(childCtx)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to get annotations", err)
		}

		itemJson, err := json.Marshal(payload.benchmark)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to marshal benchmark item", err)
		}
		annotations[consts.CRDAnnotationBenchmark] = string(itemJson)

		crdLabels := utils.MergeSimpleMaps(
			task.GetLabels(),
			map[string]string{
				consts.K8sLabelAppID:       consts.AppID,
				consts.CRDLabelInjectionID: strconv.Itoa(injection.ID),
			},
		)

		monitor := GetMonitor()
		toReleased := false
		if err := monitor.CheckNamespaceToInject(payload.namespace, time.Now(), task.TraceID); err != nil {
			toReleased = true
			return handleExecutionError(span, logEntry, "failed to get namespace to inject fault", err)
		}

		defer func() {
			if toReleased {
				if err := monitor.ReleaseLock(childCtx, payload.namespace, task.TraceID); err != nil {
					if err := handleExecutionError(span, logEntry, fmt.Sprintf("failed to release lock for namespace %s", payload.namespace), err); err != nil {
						logEntry.Error(err)
						return
					}
				}
			}
		}()

		name, err := injectionConf.Create(childCtx, payload.namespace, annotations, crdLabels)
		if err != nil {
			toReleased = true
			return handleExecutionError(span, logEntry, "failed to inject fault", err)
		}

		publishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeFaultInjection,
			EventName: consts.EventFaultInjectionStarted,
			Payload:   name,
		})

		return nil
	})
}

// parseInjectionPayload extracts and validates the injection payload from the task payload
func parseInjectionPayload(payload map[string]any) (*injectionPayload, error) {
	message := "invalid or missing '%s' in task payload"

	benchmark, err := utils.ConvertToType[dto.ContainerVersionItem](payload[consts.InjectBenchmark])
	if err != nil {
		return nil, fmt.Errorf("failed to convert benchmark: %w", err)
	}

	preDurationFloat, ok := payload[consts.InjectPreDuration].(float64)
	if !ok || preDurationFloat <= 0 {
		return nil, fmt.Errorf(message, consts.InjectPreDuration)
	}
	preDuration := int(preDurationFloat)

	node, err := utils.MapToStruct[chaos.Node](payload, consts.InjectNode, message)
	if err != nil {
		return nil, err
	}

	namespace, ok := payload[consts.InjectNamespace].(string)
	if !ok || namespace == "" {
		return nil, fmt.Errorf(message, consts.InjectNamespace)
	}

	pedestalIDFloat, ok := payload[consts.InjectPedestalID].(float64)
	if !ok || pedestalIDFloat <= 0 {
		return nil, fmt.Errorf(message, consts.InjectPedestalID)
	}
	pedestalID := int(pedestalIDFloat)

	labels, err := utils.ConvertToType[[]dto.LabelItem](payload[consts.InjectLabels])
	if err != nil {
		return nil, fmt.Errorf(message, consts.InjectLabels)
	}

	return &injectionPayload{
		benchmark:   benchmark,
		preDuration: preDuration,
		node:        node,
		namespace:   namespace,
		pedestalID:  pedestalID,
		labels:      labels,
	}, nil
}
