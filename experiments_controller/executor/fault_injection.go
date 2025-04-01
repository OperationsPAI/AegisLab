package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cli "github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/sirupsen/logrus"
)

// 故障注入任务的元数据
type InjectionMeta struct {
	Duration   int
	FaultType  int
	Namespace  string
	Pod        string
	InjectSpec map[string]int
}

// 下游数据采集任务配置
type downstreamConfig struct {
	Benchmark   string
	PreDuration int
}

func getInjectionMetaFromPayload(payload map[string]any) (*InjectionMeta, error) {
	message := "invalid or missing '%s' in payload"

	faultTypeFloat, ok := payload[consts.InjectFaultType].(float64)
	if !ok || faultTypeFloat <= 0 {
		return nil, fmt.Errorf(message, consts.InjectFaultType)
	}
	faultType := int(faultTypeFloat)

	namespace, ok := payload[consts.InjectNamespace].(string)
	if !ok || namespace == "" {
		return nil, fmt.Errorf(message, consts.InjectNamespace)
	}

	pod, ok := payload[consts.InjectPod].(string)
	if !ok || pod == "" {
		return nil, fmt.Errorf(message, consts.InjectPod)
	}

	injectSpecMap, ok := payload[consts.InjectSpec].(map[string]any)
	if !ok {
		return nil, fmt.Errorf(message, consts.InjectSpec)
	}
	injectSpec := make(map[string]int)
	for k, v := range injectSpecMap {
		floatVal, ok := v.(float64)
		if !ok {
			return nil, fmt.Errorf("invalid value for key '%s' in injectSpec", k)
		}
		injectSpec[k] = int(floatVal)
	}

	durationFloat, ok := payload[consts.InjectFaultDuration].(float64)
	if !ok || durationFloat <= 0 {
		return nil, fmt.Errorf(message, consts.InjectFaultDuration)
	}
	duration := int(durationFloat)

	return &InjectionMeta{
		Duration:   duration,
		Namespace:  namespace,
		Pod:        pod,
		FaultType:  faultType,
		InjectSpec: injectSpec,
	}, nil
}

func getDownstreamConfig(payload map[string]any) (*downstreamConfig, error) {
	message := "invalid or missing '%s' in payload"

	var benchmark string
	if _, exists := payload[consts.InjectBenchmark]; !exists {
		return nil, nil
	}
	benchmark, ok := payload[consts.InjectBenchmark].(string)
	if !ok {
		return nil, fmt.Errorf(message, consts.InjectBenchmark)
	}

	preDurationFloat, ok := payload[consts.InjectPreDuration].(float64)
	if !ok || preDurationFloat <= 0 {
		return nil, fmt.Errorf(message, consts.InjectFaultDuration)
	}
	preDuration := int(preDurationFloat)

	return &downstreamConfig{
		Benchmark:   benchmark,
		PreDuration: preDuration,
	}, nil
}

// 执行故障注入任务
func executeFaultInjection(ctx context.Context, task *UnifiedTask) error {
	logrus.Info(task)

	spec, ok := task.Payload["spec"].(map[string]any)
	if !ok {
		return fmt.Errorf("failed to read injection spec")
	}

	node, err := handler.MapToNode(spec)
	if err != nil {
		return err
	}

	var key int
	for key = range node.Children {
	}

	conf, err := handler.NodeToStruct[handler.InjectionConf](node)
	if err != nil {
		return err
	}

	name := conf.Create(cli.NewK8sClient())
	updateTaskStatus(task.TaskID, task.TraceID,
		fmt.Sprintf("executing fault injection for task %s", task.TaskID),
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusRunning,
			consts.RdbMsgTaskType: consts.TaskTypeFaultInjection,
		})

	addDatasetIndex(task.TaskID, name)

	config, err := getDownstreamConfig(task.Payload)
	if err != nil {
		return err
	}

	addTaskMeta(task.TaskID,
		consts.MetaBenchmark, config.Benchmark,
		consts.MetaPreDuration, config.PreDuration,
		consts.MetaTraceID, task.TraceID,
		consts.MetaGroupID, task.GroupID,
	)

	jsonData, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to marshal injection spec")
	}

	faultRecord := database.FaultInjectionSchedule{
		TaskID:        task.TaskID,
		FaultType:     key,
		Config:        string(jsonData),
		Duration:      0,
		PreDuration:   config.PreDuration,
		Description:   fmt.Sprintf("Fault for task %s", task.TaskID),
		Status:        consts.DatasetInitial,
		InjectionName: name,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err = database.DB.Create(&faultRecord).Error; err != nil {
		logrus.Errorf("failed to write fault injection schedule to database: %v", err)
		return fmt.Errorf("failed to write to database")
	}

	return err
}

func ParseInjectionMeta(config string) (*InjectionMeta, error) {
	var meta InjectionMeta
	if err := json.Unmarshal([]byte(config), &meta); err != nil {
		return nil, fmt.Errorf("config unmarshal error: %w", err)
	}
	return &meta, nil
}
