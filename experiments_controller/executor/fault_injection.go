package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	chaosCli "github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/sirupsen/logrus"
)

// 故障注入任务的 Payload 结构
type FaultInjectionPayload struct {
	Duration   int            `json:"duration"`
	FaultType  int            `json:"faultType"`
	Namespace  string         `json:"injectNamespace"`
	Pod        string         `json:"injectPod"`
	InjectSpec map[string]int `json:"spec"`
	Benchmark  *string        `json:"benchmark"`
}

func ParseFaultInjectionPayload(payload map[string]interface{}) (*FaultInjectionPayload, error) {
	durationFloat, ok := payload[InjectDuration].(float64)
	if !ok || durationFloat <= 0 {
		return nil, fmt.Errorf("invalid or missing '%s' in payload", InjectDuration)
	}
	duration := int(durationFloat)

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

	var benchmark *string
	benchmarkStr, ok := payload[BuildBenchmark].(string)
	if ok && benchmarkStr != "" {
		benchmark = &benchmarkStr
	}

	return &FaultInjectionPayload{
		Namespace:  namespace,
		Pod:        pod,
		FaultType:  faultType,
		Duration:   duration,
		InjectSpec: injectSpec,
		Benchmark:  benchmark,
	}, nil
}

func checkExecutionTime(faultRecord database.FaultInjectionSchedule, namespace string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time

	if faultRecord.Status == DatasetSuccess {
		startTime = faultRecord.StartTime
		endTime = faultRecord.EndTime
	} else if faultRecord.Status == DatasetInitial {
		datasetName := faultRecord.InjectionName

		var err error
		startTime, endTime, err = chaosCli.QueryCRDByName(namespace, datasetName)
		if err != nil {
			return startTime, endTime, fmt.Errorf("failed to QueryCRDByName: %s, error: %v", datasetName, err)
		}

		if err := database.DB.Model(&faultRecord).Where("injection_name = ?", datasetName).
			Updates(map[string]interface{}{
				"start_time": startTime,
				"end_time":   endTime,
			}).Error; err != nil {
			return startTime, endTime, fmt.Errorf("failed to update start_time and end_time for dataset: %s, error: %v", datasetName, err)
		}
	}

	return startTime, endTime, nil
}

// 执行故障注入任务
func executeFaultInjection(ctx context.Context, task *UnifiedTask) error {
	logrus.Infof("Executing fault injection task %+v", task)

	fiPayload, err := ParseFaultInjectionPayload(task.Payload)
	logrus.Infof("Parsed fault injection payload: %+v", fiPayload)
	if err != nil {
		return err
	}

	// 更新任务状态
	updateTaskStatus(task.TaskID, TaskStatusRunning, fmt.Sprintf("Executing fault injection for task %s", task.TaskID))

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
	jsonData, err := json.Marshal(task.Payload)
	if err != nil {
		logrus.Errorf("marshal conf failed, conf: %+v, err: %s", conf, err)
		return err
	}

	faultRecord := database.FaultInjectionSchedule{
		TaskID:          task.TaskID,
		FaultType:       fiPayload.FaultType,
		Config:          string(jsonData),
		Duration:        fiPayload.Duration,
		Description:     fmt.Sprintf("Fault for task %s", task.TaskID),
		Status:          DatasetInitial,
		InjectionName:   name,
		ProposedEndTime: time.Now().Add(time.Duration(fiPayload.Duration+2) * time.Minute),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := database.DB.Create(&faultRecord).Error; err != nil {
		logrus.Errorf("Failed to write fault injection schedule to database: %v", err)
		return fmt.Errorf("failed to write to database: %v", err)
	}

	if fiPayload.Benchmark != nil {
		logrus.Info("Scheduling build dataset task")
		time.AfterFunc(time.Duration(fiPayload.Duration+2)*time.Minute, func() {
			startTime, endTime, err := checkExecutionTime(faultRecord, fiPayload.Namespace)
			logrus.Infof("checkExecutionTime for dataset %s, startTime: %v, endTime: %v", name, startTime, endTime)
			if err != nil {
				logrus.Errorf("Failed to checkExecutionTime for dataset %s: %v", name, err)
				return
			}

			updateTaskStatus(task.TaskID, TaskStatusCompleted, fmt.Sprintf("Task %s completed", task.TaskID))

			datasetPayload := map[string]interface{}{
				BuildBenchmark: *fiPayload.Benchmark,
				BuildDataset:   name,
				BuildNamespace: fiPayload.Namespace,
				BuildStartTime: &startTime,
				BuildEndTime:   &endTime,
			}

			if _, err := SubmitTask(context.Background(), &UnifiedTask{
				Type:      TaskTypeBuildDataset,
				Payload:   datasetPayload,
				Immediate: true,
				TraceID:   task.TraceID,
			}); err != nil {
				logrus.Error(err)
				return
			}
		})
	}

	return nil
}
