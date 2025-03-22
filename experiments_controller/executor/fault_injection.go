package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/sirupsen/logrus"
)

// 故障注入任务的 Payload 结构
type FaultInjectionPayload struct {
	FaultType  int            `json:"fault_type"`
	Namespace  string         `json:"inject_namespace"`
	Pod        string         `json:"inject_pod"`
	InjectSpec map[string]int `json:"spec"`

	PreDuration   int `json:"pre_duration"`
	FaultDuration int `json:"fault_duration"`

	Benchmark *string `json:"benchmark"`
}

func ParseFaultInjectionPayload(payload map[string]any) (*FaultInjectionPayload, error) {
	faultTypeFloat, ok := payload[InjectFaultType].(float64)
	if !ok || faultTypeFloat <= 0 {
		return nil, fmt.Errorf("invalid or missing '%s' in payload", InjectFaultType)
	}
	faultType := int(faultTypeFloat)

	namespace, ok := payload[InjectNamespace].(string)
	if !ok || namespace == "" {
		return nil, fmt.Errorf("invalid or missing '%s' in payload", InjectNamespace)
	}

	pod, ok := payload[InjectPod].(string)
	if !ok || pod == "" {
		return nil, fmt.Errorf("invalid or missing '%s' in payload", InjectPod)
	}

	injectSpecMap, ok := payload[InjectSpec].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid or missing '%s' in payload", InjectSpec)
	}
	injectSpec := make(map[string]int)
	for k, v := range injectSpecMap {
		floatVal, ok := v.(float64)
		if !ok {
			return nil, fmt.Errorf("invalid value for key '%s' in injectSpec", k)
		}
		injectSpec[k] = int(floatVal)
	}

	preDurationFloat, ok := payload[InjectPreDuration].(float64)
	if !ok || preDurationFloat <= 0 {
		return nil, fmt.Errorf("invalid or missing '%s' in payload", InjectPreDuration)
	}
	preDuration := int(preDurationFloat)

	faultDurationFloat, ok := payload[InjectFaultDuration].(float64)
	if !ok || faultDurationFloat <= 0 {
		return nil, fmt.Errorf("invalid or missing '%s' in payload", InjectFaultDuration)
	}
	faultDuration := int(faultDurationFloat)

	var benchmark *string
	benchmarkStr, ok := payload[BuildBenchmark].(string)
	if ok && benchmarkStr != "" {
		benchmark = &benchmarkStr
	}

	return &FaultInjectionPayload{
		Namespace:     namespace,
		Pod:           pod,
		FaultType:     faultType,
		InjectSpec:    injectSpec,
		PreDuration:   preDuration,
		FaultDuration: faultDuration,
		Benchmark:     benchmark,
	}, nil
}

// 执行故障注入任务
func executeFaultInjection(ctx context.Context, task *UnifiedTask) error {
	logrus.Info(task)

	fiPayload, err := ParseFaultInjectionPayload(task.Payload)
	logrus.Infof("Parsed fault injection payload: %+v", fiPayload)
	if err != nil {
		return err
	}

	// 故障注入逻辑
	var chaosSpec any
	spec := handler.SpecMap[handler.ChaosType(fiPayload.FaultType)]
	if spec != nil {
		actionSpace, err := handler.GenerateActionSpace(spec)
		if err != nil {
			logrus.Error("GenerateActionSpace: ", err)
			return err
		}
		err = handler.ValidateAction(fiPayload.InjectSpec, actionSpace)
		if err != nil {
			logrus.Error("ValidateAction: ", err)
			return err
		}
		chaosSpec, err = handler.ActionToStruct(handler.ChaosType(fiPayload.FaultType), fiPayload.InjectSpec)
		if err != nil {
			logrus.Errorf("ActionToStruct, err: %s", err)
			return err
		}
	}

	conf := handler.ChaosConfig{
		Type:     handler.ChaosType(fiPayload.FaultType),
		Spec:     chaosSpec,
		Duration: fiPayload.FaultDuration,
	}
	name := handler.Create(fiPayload.Namespace, fiPayload.Pod, conf)
	if name == "" {
		return fmt.Errorf("create chaos failed, conf: %+v", conf)
	}
	jsonData, err := json.Marshal(task.Payload)
	if err != nil {
		logrus.Errorf("Failed to marshal conf: %+v, err: %s", conf, err)
		return err
	}

	updateTaskStatus(task.TaskID, task.TraceID,
		fmt.Sprintf("Executing fault injection for task %s", task.TaskID),
		map[string]any{
			RdbMsgStatus:   TaskStatusRunning,
			RdbMsgTaskType: TaskTypeFaultInjection,
		})

	addDatasetIndex(task.TaskID, name)
	if fiPayload.Benchmark != nil {
		addTaskMeta(task.TaskID,
			MetaBenchmark, *fiPayload.Benchmark,
			MetaPreDuration, fiPayload.PreDuration,
			MetaTraceID, task.TraceID,
			MetaGroupID, task.GroupID,
		)
	}

	faultRecord := database.FaultInjectionSchedule{
		TaskID:          task.TaskID,
		FaultType:       fiPayload.FaultType,
		Config:          string(jsonData),
		Duration:        fiPayload.FaultDuration,
		Description:     fmt.Sprintf("Fault for task %s", task.TaskID),
		Status:          DatasetInitial,
		InjectionName:   name,
		ProposedEndTime: time.Now().Add(time.Duration(fiPayload.FaultDuration) * time.Minute),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := database.DB.Create(&faultRecord).Error; err != nil {
		logrus.Errorf("Failed to write fault injection schedule to database: %v", err)
		return fmt.Errorf("failed to write to database: %v", err)
	}

	return nil
}
