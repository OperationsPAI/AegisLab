package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	chaos "github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/sirupsen/logrus"
)

type injectionPayload struct {
	benchmark   string
	faultType   int
	preDuration int
	rawConf     map[string]any
	conf        *chaos.InjectionConf
}

// 执行故障注入任务
func executeFaultInjection(ctx context.Context, task *UnifiedTask) error {
	logrus.Info(task)

	payload, err := parseInjectionPayload(task.Payload)
	if err != nil {
		return err
	}

	conf, name, err := payload.conf.Create()
	if err != nil {
		return fmt.Errorf("failed to inject fault: %v", err)
	}

	updateTaskStatus(task.TaskID, task.TraceID,
		fmt.Sprintf("executing fault injection for task %s", task.TaskID),
		map[string]any{
			consts.RdbMsgStatus:   consts.TaskStatusRunning,
			consts.RdbMsgTaskType: consts.TaskTypeFaultInjection,
		})

	addDatasetIndex(task.TaskID, name)
	addTaskMeta(task.TaskID,
		consts.MetaBenchmark, payload.benchmark,
		consts.MetaPreDuration, payload.preDuration,
		consts.MetaTraceID, task.TraceID,
		consts.MetaGroupID, task.GroupID,
	)

	engineData, err := json.Marshal(payload.rawConf)
	if err != nil {
		return fmt.Errorf("failed to marshal injection spec to engine config: %v", err)
	}

	displayData, err := json.Marshal(conf)
	if err != nil {
		return fmt.Errorf("failed to marshal injection spec to display config: %v", err)
	}

	faultRecord := database.FaultInjectionSchedule{
		TaskID:        task.TaskID,
		FaultType:     payload.faultType,
		DisplayConfig: string(displayData),
		EngineConfig:  string(engineData),
		PreDuration:   payload.preDuration,
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

func parseInjectionPayload(payload map[string]any) (*injectionPayload, error) {
	message := "invalid or missing '%s' in task payload"

	benchmark, ok := payload[consts.InjectBenchmark].(string)
	if !ok {
		return nil, fmt.Errorf(message, consts.InjectBenchmark)
	}

	faultTypeFloat, ok := payload[consts.InjectFaultType].(float64)
	if !ok || faultTypeFloat <= 0 {
		return nil, fmt.Errorf(message, consts.InjectFaultType)
	}
	faultType := int(faultTypeFloat)

	preDurationFloat, ok := payload[consts.InjectPreDuration].(float64)
	if !ok || preDurationFloat <= 0 {
		return nil, fmt.Errorf(message, consts.InjectPreDuration)
	}
	preDuration := int(preDurationFloat)

	rawConf, ok := payload[consts.InjectRawConf].(map[string]any)
	if !ok {
		return nil, fmt.Errorf(message, consts.InjectRawConf)
	}

	m, ok := payload[consts.InjectConf].(map[string]any)
	if !ok {
		return nil, fmt.Errorf(message, consts.InjectConf)
	}

	jsonData, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", fmt.Sprintf(message, consts.InjectConf), err)
	}

	var conf chaos.InjectionConf
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return nil, fmt.Errorf("%s: %v", fmt.Sprintf(message, consts.InjectConf), err)
	}

	return &injectionPayload{
		benchmark:   benchmark,
		faultType:   faultType,
		preDuration: preDuration,
		rawConf:     rawConf,
		conf:        &conf,
	}, nil
}
