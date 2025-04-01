package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	cli "github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/sirupsen/logrus"
)

type downstreamConfig struct {
	Benchmark   string
	PreDuration int
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
		return nil, fmt.Errorf(message, consts.InjectPreDuration)
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

	spec, ok := task.Payload[consts.InjectSpec].(map[string]any)
	if !ok {
		return fmt.Errorf("failed to read injection spec")
	}

	node, err := handler.MapToNode(spec)
	if err != nil {
		return err
	}

	var key string
	for key = range node.Children {
	}

	intKey, err := strconv.Atoi(key)
	if err != nil {
		return err
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
		FaultType:     intKey,
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
