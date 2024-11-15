package executor

import (
	"context"
	"dagger/rcabench/database"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"dagger.io/dagger/dag"
	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/go-redis/redis/v8"
	"github.com/k0kubun/pp"
	"github.com/sirupsen/logrus"
)

const (
	RunAlgo        = "RunAlgorithm"
	FaultInjection = "FaultInjection"

	RdbMsgTaskID       = "taskID"
	RdbMsgTaskType     = "taskType"
	RdbMsgPayload      = "payload"
	RdbMsgParentTaskID = "parentTaskID"

	EvalPayloadAlgo    = "algorithm"
	EvalPayloadDataset = "dataset"
	EvalPayloadBench   = "benchmark"

	InjectFaultTpye = "faultType"
	InjectStartTime = "start_time"
	InjectEndTime   = "end_time"
	InjectSpec      = "spec"
)

func initRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

const (
	StreamName = "task_stream"
)

var Rdb = initRedisClient()

func ConsumeTasks() {
	for {
		entries, err := Rdb.XRead(context.Background(), &redis.XReadArgs{
			Streams: []string{StreamName, "$"},
			Count:   1,
			Block:   0,
		}).Result()
		if err != nil {
			fmt.Println("Error reading from stream:", err)
			continue
		}

		for _, entry := range entries[0].Messages {
			taskID := entry.Values[RdbMsgTaskID].(string)
			taskType := entry.Values[RdbMsgTaskType].(string)
			jsonPayload := entry.Values[RdbMsgPayload].(string)
			payload := map[string]interface{}{}
			err = json.Unmarshal([]byte(jsonPayload), &payload)
			if err != nil {
				logrus.Errorf("Unmarshaling %s failed", jsonPayload)
			}

			logrus.Infof("Executing %s", taskID)

			updateTaskStatus(taskID, "Running", fmt.Sprintf("Task %s started of type %s", taskID, taskType))

			switch taskType {
			case FaultInjection:
				err = executeFaultInjection(taskID, payload)
			case RunAlgo:
				err = executeAlgorithm(taskID, payload)
			}
			if err != nil {
				updateTaskStatus(taskID, "Error", fmt.Sprintf("Task %s error, message: %s", taskID, err))
				logrus.Error(err)
			} else {
				updateTaskStatus(taskID, "Completed", fmt.Sprintf("Task %s completed", taskID))
				logrus.Infof("Task %s completed", taskID)
			}
			Rdb.XDel(context.Background(), StreamName, entry.ID)
		}
	}
}

func executeFaultInjection(taskID string, payload map[string]interface{}) error {
	logrus.Infof("injecting, taskID: %s", taskID)
	pp.Print("payload:", payload)

	// 从 payload 中提取字段
	faultType, err := strconv.Atoi(payload[InjectFaultTpye].(string))
	if err != nil {
		logrus.Error(err)
		return err
	}
	injectSpec := make(map[string]int)
	for k, v := range payload[InjectSpec].(map[string]interface{}) {
		injectSpec[k] = int(v.(float64))
	}

	// 更新任务状态
	updateTaskStatus(taskID, "Running", fmt.Sprintf("Executing fault injection for task %s", taskID))

	// 故障注入逻辑
	_ = handler.ChaosConfig{}
	spec := handler.SpecMap[handler.ChaosType(faultType)]
	actionSpace, err := handler.GenerateActionSpace(spec)
	if err != nil {
		logrus.Error(err)
	}
	err = handler.ValidateAction(injectSpec, actionSpace)
	if err != nil {
		logrus.Error("ValidateAction", err)
		return err
	}
	chaosSpec, err := handler.ActionToStruct(handler.ChaosType(faultType), injectSpec)
	if err != nil {
		logrus.Errorf("ActionToStruct, err: %s", err)
		return err
	}

	config := handler.ChaosConfig{
		Type: handler.ChaosType(faultType),
		Spec: chaosSpec,
	}
	// handler.Create(config)
	pp.Print("config", config)

	//需要对账系统来 check 故障注入是否成功，如果不成功，则把数据库里的条目删除，否则会出现假注入。

	// 创建新的故障注入记录
	faultRecord := database.FaultInjectionSchedule{
		ID:          taskID,                                   // 使用任务 ID 作为记录的主键
		FaultType:   faultType,                                // 故障类型
		Config:      fmt.Sprintf("%v", payload),               // 故障配置 (JSON 格式化字符串)
		LastTime:    time.Now(),                               // 开始时间
		Description: fmt.Sprintf("Fault for task %s", taskID), // 可选描述
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 写入数据库
	if err := database.DB.Create(&faultRecord).Error; err != nil {
		logrus.Errorf("Failed to write fault injection schedule to database: %v", err)
		return fmt.Errorf("failed to write to database: %v", err)
	}

	return nil
}

func executeAlgorithm(taskID string, payload map[string]interface{}) error {
	requiredKeys := []string{EvalPayloadBench, EvalPayloadAlgo, EvalPayloadDataset}
	for _, key := range requiredKeys {
		if val, ok := payload[key].(string); !ok || val == "" {
			return fmt.Errorf("missing or invalid '%s' key in payload", key)
		}
	}
	bench := payload[EvalPayloadBench].(string)
	algo := payload[EvalPayloadAlgo].(string)
	dataset := payload[EvalPayloadDataset].(string)
	_ = dataset //TODO
	updateTaskStatus(taskID, "Running", fmt.Sprintf("Running algorithm for task %s", taskID))

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	parentDir := filepath.Dir(pwd)

	benchPath := filepath.Join(parentDir, "benchmarks", bench)
	algoPath := filepath.Join(parentDir, "algorithms", algo)
	startScriptPath := filepath.Join(parentDir, "experiments", "run_exp.py")

	if _, err := os.Stat(benchPath); os.IsNotExist(err) {
		return errors.New("benchmark directory does not exist")
	}
	if _, err := os.Stat(algoPath); os.IsNotExist(err) {
		return errors.New("algorithm directory does not exist")
	}
	if _, err := os.Stat(startScriptPath); os.IsNotExist(err) {
		return errors.New("start script does not exist")
	}
	rc := &Rcabench{}
	con := rc.Evaluate(context.Background(), dag.Host().Directory(benchPath), dag.Host().Directory(algoPath), dag.Host().File(startScriptPath))

	_, err = con.Export(context.Background(), "./output")
	if err != nil {
		return fmt.Errorf("failed to export result, details: %s", err.Error())
	}

	return nil
}

func updateTaskStatus(taskID, status, message string) {
	// 更新 Redis 中的任务状态
	taskKey := fmt.Sprintf("task:%s:status", taskID)
	Rdb.HSet(context.Background(), taskKey, "status", status)
	Rdb.HSet(context.Background(), taskKey, "updated_at", time.Now().Format(time.RFC3339))

	// 添加日志到 Redis
	logKey := fmt.Sprintf("task:%s:logs", taskID)
	Rdb.RPush(context.Background(), logKey, fmt.Sprintf("%s - %s", time.Now().Format(time.RFC3339), message))
	logrus.Info(fmt.Sprintf("%s - %s", time.Now().Format(time.RFC3339), message))
	// 更新 SQLite 中的任务状态
	if err := database.DB.Model(&database.Task{}).Where("id = ?", taskID).Update("status", status).Error; err != nil {
		logrus.Errorf("Failed to update task %s in SQLite: %v\n", taskID, err)
	}
}
