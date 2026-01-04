package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
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

// injectionPayload contains all necessary data for executing a fault injection batch
type injectionPayload struct {
	benchmark   dto.ContainerVersionItem
	preDuration int
	nodes       []chaos.Node
	namespace   string
	pedestal    chaos.SystemType
	pedestalID  int
	labels      []dto.LabelItem
}

type batchManager struct {
	mu              sync.RWMutex
	batchCounts     map[string]int
	batchInjections map[string][]string
}

var (
	batchManagerInstance *batchManager
	batchManagerOnce     sync.Once
)

func getBatchManager() *batchManager {
	batchManagerOnce.Do(func() {
		batchManagerInstance = &batchManager{
			batchCounts:     make(map[string]int),
			batchInjections: make(map[string][]string),
		}
	})
	return batchManagerInstance
}

func (bm *batchManager) deleteBatch(batchID string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	delete(bm.batchCounts, batchID)
	delete(bm.batchInjections, batchID)
}

func (bm *batchManager) incrementBatchCount(batchID string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.batchCounts[batchID]++
}

func (bm *batchManager) isFinished(batchID string) bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	count, exists := bm.batchCounts[batchID]
	if !exists {
		return true
	}
	injectionNames, exists := bm.batchInjections[batchID]
	if !exists {
		return true
	}

	return count >= len(injectionNames)
}

func (bm *batchManager) setBatchInjections(batchID string, injectionNames []string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.batchCounts[batchID] = 0
	bm.batchInjections[batchID] = injectionNames
}

// executeFaultInjection handles the injection of a fault batch with support for parallel fault injection
//
// The function processes multiple fault nodes simultaneously:
//   - Parses all fault nodes in the batch
//   - Converts each node to InjectionConf
//   - Generates display configs and groundtruth for each fault
//   - Stores the entire batch as a single database record with array-based configs
//   - Uses Chaos Mesh BatchCreate to inject all faults in parallel
//
// Storage format:
//   - engine_config: JSON array of all chaos.Node objects
//   - display_config: JSON array of display maps for each fault
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

		// Process all fault nodes in the batch
		injectionConfs := make([]chaos.InjectionConf, 0, len(payload.nodes))
		displayMaps := make([]map[string]any, 0, len(payload.nodes))
		groundtruths := make([]database.Groundtruth, 0, len(payload.nodes))

		for i, node := range payload.nodes {
			injectionConf, err := chaos.NodeToStruct[chaos.InjectionConf](childCtx, &node)
			if err != nil {
				return handleExecutionError(span, logEntry, fmt.Sprintf("failed to convert node %d to injection conf", i), err)
			}

			displayMap, err := injectionConf.GetDisplayConfig(childCtx)
			if err != nil {
				return handleExecutionError(span, logEntry, fmt.Sprintf("failed to get display config for node %d", i), err)
			}

			chaosGroundtruth, err := injectionConf.GetGroundtruth(childCtx)
			if err != nil {
				return handleExecutionError(span, logEntry, fmt.Sprintf("failed to get groundtruth for node %d", i), err)
			}

			injectionConfs = append(injectionConfs, *injectionConf)
			displayMaps = append(displayMaps, displayMap)
			groundtruths = append(groundtruths, *database.NewDBGroundtruth(&chaosGroundtruth))
		}

		// Marshal display config as array
		displayData, err := json.Marshal(displayMaps)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to marshal injection specs to display config", err)
		}

		// Marshal engine config as array
		engineData, err := json.Marshal(payload.nodes)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to marshal injection specs to engine config", err)
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

		batchID := fmt.Sprintf("batch-%s", utils.GenerateULID(nil))
		crdLabels := utils.MergeSimpleMaps(
			task.GetLabels(),
			map[string]string{
				consts.K8sLabelAppID:    consts.AppID,
				consts.CRDLabelBatchID:  batchID,
				consts.CRDLabelIsHybrid: strconv.FormatBool(len(payload.nodes) > 1),
			},
		)

		// Batch create all fault injections in parallel
		names, err := chaos.BatchCreate(childCtx, injectionConfs, chaos.SystemTrainTicket, payload.namespace, annotations, crdLabels)
		if err != nil {
			toReleased = true
			return handleExecutionError(span, logEntry, "failed to inject faults", err)
		}

		var name string
		var faultType chaos.ChaosType
		if len(names) > 1 {
			name = batchID
			faultType = consts.Hybrid
			getBatchManager().setBatchInjections(batchID, names)
		} else {
			name = names[0]
			faultType = chaos.ChaosType(payload.nodes[0].Value)
		}

		injection := &database.FaultInjection{
			Name:          name,
			FaultType:     faultType,
			Category:      payload.pedestal,
			Description:   fmt.Sprintf("Fault batch for task %s (%d faults)", task.TaskID, len(payload.nodes)),
			DisplayConfig: utils.StringPtr(string(displayData)),
			EngineConfig:  string(engineData),
			Groundtruths:  groundtruths,
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

		return nil
	})
}

// parseInjectionPayload extracts and validates the injection payload from the task payload
//
// The payload now supports multiple fault nodes for parallel injection:
//   - Validates that at least one fault node is provided
//   - Parses the nodes array (not a single node)
//   - Ensures all required fields are present and valid
//
// Returns injectionPayload containing all parsed data for fault injection execution
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

	// Parse nodes array - now supports multiple fault nodes
	nodes, err := utils.ConvertToType[[]chaos.Node](payload[consts.InjectNodes])
	if err != nil {
		return nil, fmt.Errorf(message, consts.InjectNodes)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("at least one fault node is required in %s", consts.InjectNodes)
	}

	namespace, ok := payload[consts.InjectNamespace].(string)
	if !ok || namespace == "" {
		return nil, fmt.Errorf(message, consts.InjectNamespace)
	}

	pedestalStr, ok := payload[consts.InjectPedestal].(string)
	if !ok || pedestalStr == "" {
		return nil, fmt.Errorf(message, consts.InjectPedestal)
	}
	pedestal := chaos.SystemType(pedestalStr)
	if !pedestal.IsValid() {
		return nil, fmt.Errorf("invalid pedestal type: %s", pedestalStr)
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
		nodes:       nodes,
		namespace:   namespace,
		pedestal:    pedestal,
		pedestalID:  pedestalID,
		labels:      labels,
	}, nil
}
