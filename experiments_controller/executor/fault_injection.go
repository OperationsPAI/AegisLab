package executor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	chaos "github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/sirupsen/logrus"
)

type injectionPayload struct {
	benchmark   string
	faultType   int
	preDuration int
	rawConf     string
	conf        *chaos.InjectionConf
}

type restartPayload struct {
	namespace     string
	injectionTime time.Time
	injectionPayload
}

// 执行故障注入任务
// TODO 回退
func executeFaultInjection(ctx context.Context, task *UnifiedTask) error {
	payload, err := parseInjectionPayload(task.Payload)
	if err != nil {
		return err
	}

	annotations, err := getAnnotations(ctx, task)
	if err != nil {
		return err
	}

	config, name, err := payload.conf.Create(
		annotations,
		map[string]string{
			consts.CRDTaskID:      task.TaskID,
			consts.CRDTraceID:     task.TraceID,
			consts.CRDGroupID:     task.GroupID,
			consts.CRDBenchmark:   payload.benchmark,
			consts.CRDPreDuration: strconv.Itoa(payload.preDuration),
		})
	if err != nil {
		return fmt.Errorf("failed to inject fault: %v", err)
	}

	displayData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal injection spec to display config: %v", err)
	}

	faultRecord := database.FaultInjectionSchedule{
		TaskID:        task.TaskID,
		FaultType:     payload.faultType,
		DisplayConfig: string(displayData),
		EngineConfig:  payload.rawConf,
		PreDuration:   payload.preDuration,
		Description:   fmt.Sprintf("Fault for task %s", task.TaskID),
		Status:        consts.DatasetInitial,
		InjectionName: name,
	}
	if err = database.DB.Create(&faultRecord).Error; err != nil {
		logrus.Errorf("failed to write fault injection schedule to database: %v", err)
		return fmt.Errorf("failed to write to database")
	}

	return nil
}

func executeRestartService(ctx context.Context, task *UnifiedTask) error {
	payload, err := parseRestartPayload(task.Payload)
	if err != nil {
		return err
	}

	if err := executeCommand(fmt.Sprintf(config.GetString("injection.command"), config.GetString("workspace"), payload.namespace)); err != nil {
		return err
	}

	taskPayload := map[string]any{
		consts.InjectBenchmark:   payload.benchmark,
		consts.InjectFaultType:   payload.faultType,
		consts.InjectPreDuration: payload.preDuration,
		consts.InjectRawConf:     payload.rawConf,
		consts.InjectConf:        payload.conf,
	}

	injectionTask := &UnifiedTask{
		Type:         consts.TaskTypeFaultInjection,
		Payload:      taskPayload,
		Immediate:    false,
		ExecuteTime:  payload.injectionTime.Unix(),
		TraceID:      task.TraceID,
		GroupID:      task.GroupID,
		TraceCarrier: task.TraceCarrier,
	}
	if _, _, err := SubmitTask(ctx, injectionTask); err != nil {
		return fmt.Errorf("failed to submit injection task: %v", err)
	}

	return nil
}

func parseInjectionPayload(payload map[string]any) (*injectionPayload, error) {
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

	preDurationFloat, ok := payload[consts.InjectPreDuration].(float64)
	if !ok || preDurationFloat <= 0 {
		return nil, fmt.Errorf(message, consts.InjectPreDuration)
	}
	preDuration := int(preDurationFloat)

	rawConf, ok := payload[consts.InjectRawConf].(string)
	if !ok || rawConf == "" {
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

func parseRestartPayload(payload map[string]any) (*restartPayload, error) {
	message := "invalid or missing '%s' in task payload"

	intervalFloat, ok := payload[consts.RestartInterval].(float64)
	if !ok || intervalFloat <= 0 {
		return nil, fmt.Errorf(message, consts.RestartInterval)
	}
	interval := int(intervalFloat)

	_, executionTimeExists := payload[consts.RestartExecutionTime]

	var executionTime time.Time
	if executionTimeExists {
		executionTimePtr, err := parseTimePtrFromPayload(payload, consts.RestartExecutionTime)
		if err != nil {
			return nil, fmt.Errorf(message, consts.RestartExecutionTime)
		}

		executionTime = *executionTimePtr
	}

	injectionPayload, err := parseInjectionPayload(payload)
	if err != nil {
		return nil, err
	}

	message = "invalid or missing '%s' in injection config"
	_, config, err := injectionPayload.conf.GetActiveInjection()
	if err != nil {
		return nil, fmt.Errorf("failed to read config in injection conf: %v", err)
	}

	durationInt64, ok := config[consts.RestartDuration].(int64)
	if !ok || durationInt64 <= 0 {
		return nil, fmt.Errorf(message, consts.RestartDuration)
	}

	duration := int(durationInt64)

	namespace, ok := config[consts.RestartNamespace].(string)
	if !ok || namespace == "" {
		return nil, fmt.Errorf(message, consts.RestartNamespace)
	}

	deltaTime := time.Duration(interval-injectionPayload.preDuration-duration) * consts.DefaultTimeUnit
	injectionTime := executionTime.Add(deltaTime)

	return &restartPayload{
		namespace:        namespace,
		injectionTime:    injectionTime,
		injectionPayload: *injectionPayload,
	}, nil
}

func executeCommand(command string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("faied to get the command output pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get the command error pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	go func() {
		for stdoutScanner.Scan() {
			logrus.Info("STDOUT: ", stdoutScanner.Text())
		}
	}()

	go func() {
		for stderrScanner.Scan() {
			if strings.Contains(stderrScanner.Text(), "Warning") {
				logrus.Warn("STDERR: ", stderrScanner.Text())
			} else {
				logrus.Error("STDERR: ", stderrScanner.Text())
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	}

	return nil
}
