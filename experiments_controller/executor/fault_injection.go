package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/CUHK-SE-Group/rcabench/database"

	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/sirupsen/logrus"
)

// 故障注入任务的 Payload 结构
type FaultInjectionPayload struct {
	Namespace  string
	Pod        string
	FaultType  int
	Duration   int
	InjectSpec map[string]int
}

// 执行故障注入任务
func executeFaultInjection(ctx context.Context, taskID string, payload map[string]interface{}) error {
	logrus.Infof("Injecting fault, taskID: %s", taskID)

	fiPayload, err := ParseFaultInjectionPayload(payload)
	if err != nil {
		logrus.Error(err)
		return err
	}

	// 更新任务状态
	updateTaskStatus(taskID, "Running", fmt.Sprintf("Executing fault injection for task %s", taskID))

	// 故障注入逻辑
	var chaosSpec interface{}
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
		Duration: fiPayload.Duration,
	}
	name := handler.Create(fiPayload.Namespace, fiPayload.Pod, conf)
	if name == "" {
		return fmt.Errorf("create chaos failed, conf: %+v", conf)
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		logrus.Errorf("marshal conf failed, conf: %+v, err: %s", conf, err)
		return err
	}

	faultRecord := database.FaultInjectionSchedule{
		TaskID:          taskID,
		FaultType:       fiPayload.FaultType,
		Config:          string(jsonData),
		Duration:        fiPayload.Duration,
		Description:     fmt.Sprintf("Fault for task %s", taskID),
		Status:          database.DatasetInitial,
		InjectionName:   name,
		ProposedEndTime: time.Now().Add(time.Duration(fiPayload.Duration+2) * time.Minute),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := database.DB.Create(&faultRecord).Error; err != nil {
		logrus.Errorf("Failed to write fault injection schedule to database: %v", err)
		return fmt.Errorf("failed to write to database: %v", err)
	}

	return nil
}

func ParseFaultInjectionPayload(payload map[string]interface{}) (*FaultInjectionPayload, error) {
	namespace, ok := payload[InjectNamespace].(string)
	if !ok || namespace == "" {
		return nil, fmt.Errorf("invalid or missing '%s' in payload", InjectNamespace)
	}
	pod, ok := payload[InjectPod].(string)
	if !ok || pod == "" {
		return nil, fmt.Errorf("invalid or missing '%s' in payload", InjectPod)
	}
	faultTypeStr, ok := payload[InjectFaultType].(string)
	if !ok || faultTypeStr == "" {
		return nil, fmt.Errorf("invalid or missing '%s' in payload", InjectFaultType)
	}
	faultType, err := strconv.Atoi(faultTypeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid faultType value: %v", err)
	}
	durationFloat, ok := payload[InjectDuration].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid or missing '%s' in payload", InjectDuration)
	}
	duration := int(durationFloat)
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
	return &FaultInjectionPayload{
		Namespace:  namespace,
		Pod:        pod,
		FaultType:  faultType,
		Duration:   duration,
		InjectSpec: injectSpec,
	}, nil
}
