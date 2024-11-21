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
	"strings"
	"time"

	"dagger.io/dagger/dag"
	"github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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

	InjectFaultType = "faultType"
	InjectNamespace = "injectNamespace"
	InjectPod       = "injectPod"
	InjectDuration  = "duration"
	InjectSpec      = "spec"
)
const (
	StreamName   = "task_stream"
	GroupName    = "task_consumer_group"
	ConsumerName = "task_consumer"
)

func initRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// 检查 Redis 连接是否成功
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		logrus.Fatalf("Failed to connect to Redis: %v", err)
	}

	return client
}

var Rdb = initRedisClient()

func init() {

	initConsumerGroup()
}

func initConsumerGroup() {
	ctx := context.Background()
	err := Rdb.XGroupCreateMkStream(ctx, StreamName, GroupName, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		logrus.Fatalf("Failed to create consumer group: %v", err)
	}
}

func ConsumeTasks() {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("ConsumeTasks panicked: %v", r)
		}
	}()

	ctx := context.Background()
	for {
		// 读取待处理的消息（Pending Entries）
		entries, err := Rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    GroupName,
			Consumer: ConsumerName,
			Streams:  []string{StreamName, "0"},
			Count:    1,
			Block:    time.Second * 5, // 阻塞等待最多5秒
		}).Result()
		if err != nil && err != redis.Nil {
			logrus.Errorf("Error reading from stream: %v", err)
			time.Sleep(time.Second)
			continue
		}

		if len(entries) > 0 && len(entries[0].Messages) > 0 {
			for _, entry := range entries[0].Messages {
				processTask(entry)
				err := Rdb.XAck(ctx, StreamName, GroupName, entry.ID).Err()
				if err != nil {
					logrus.Errorf("Failed to acknowledge message %v: %v", entry.ID, err)
				}
			}
			continue
		}

		// 如果没有待处理的消息，读取新的消息
		entries, err = Rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    GroupName,
			Consumer: ConsumerName,
			Streams:  []string{StreamName, ">"},
			Count:    1,
			Block:    time.Second * 5, // 阻塞等待最多5秒
		}).Result()
		if err != nil && err != redis.Nil {
			logrus.Errorf("Error reading from stream: %v", err)
			time.Sleep(time.Second)
			continue
		}

		if len(entries) > 0 && len(entries[0].Messages) > 0 {
			for _, entry := range entries[0].Messages {
				processTask(entry)
				// 确认消息已被处理
				err := Rdb.XAck(ctx, StreamName, GroupName, entry.ID).Err()
				if err != nil {
					logrus.Errorf("Failed to acknowledge message %v: %v", entry.ID, err)
				}
			}
		} else {
			// 如果没有消息，稍等一会儿
			time.Sleep(time.Second)
		}
	}
}

func processTask(entry redis.XMessage) {
	logrus.Infof("processing %s", entry.ID)
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("processTask panicked: %v", r)
		}
		// 确认并删除消息
		Rdb.XAck(context.Background(), StreamName, GroupName, entry.ID)
		Rdb.XDel(context.Background(), StreamName, entry.ID)
	}()

	taskID, ok := entry.Values[RdbMsgTaskID].(string)
	if !ok {
		logrus.Error("Invalid taskID in message")
		return
	}
	taskType, ok := entry.Values[RdbMsgTaskType].(string)
	if !ok {
		logrus.Error("Invalid taskType in message")
		return
	}
	jsonPayload, ok := entry.Values[RdbMsgPayload].(string)
	if !ok {
		logrus.Error("Invalid payload in message")
		return
	}
	payload := map[string]interface{}{}
	err := json.Unmarshal([]byte(jsonPayload), &payload)
	if err != nil {
		logrus.Errorf("Unmarshaling %s failed: %v", jsonPayload, err)
		return
	}

	logrus.Infof("Executing %s", taskID)
	updateTaskStatus(taskID, "Running", fmt.Sprintf("Task %s started of type %s", taskID, taskType))

	var execErr error
	switch taskType {
	case FaultInjection:
		execErr = executeFaultInjection(taskID, payload)
	case RunAlgo:
		execErr = executeAlgorithm(taskID, payload)
	default:
		execErr = fmt.Errorf("unknown task type: %s", taskType)
	}

	if execErr != nil {
		updateTaskStatus(taskID, "Error", fmt.Sprintf("Task %s error, message: %s", taskID, execErr))
		logrus.Error(execErr)
	} else {
		updateTaskStatus(taskID, "Completed", fmt.Sprintf("Task %s completed", taskID))
		logrus.Infof("Task %s completed", taskID)
	}
}

func executeFaultInjection(taskID string, payload map[string]interface{}) error {
	logrus.Infof("Injecting fault, taskID: %s", taskID)

	// 从 payload 中提取字段
	namespace, ok := payload[InjectNamespace].(string)
	if !ok {
		err := fmt.Errorf("invalid or missing '%s' in payload", InjectNamespace)
		logrus.Error(err)
		return err
	}
	targetPod, ok := payload[InjectPod].(string)
	if !ok {
		err := fmt.Errorf("invalid or missing '%s' in payload", InjectPod)
		logrus.Error(err)
		return err
	}

	faultTypeStr, ok := payload[InjectFaultType].(string)
	if !ok {
		err := fmt.Errorf("invalid or missing '%s' in payload", InjectFaultType)
		logrus.Error(err)
		return err
	}
	faultType, err := strconv.Atoi(faultTypeStr)
	if err != nil {
		logrus.Error(err)
		return err
	}
	duration, ok := payload[InjectDuration].(float64)
	if !ok {
		err := fmt.Errorf("invalid or missing '%s' in payload", InjectDuration)
		logrus.Error(err)
		return err
	}

	injectSpecMap, ok := payload[InjectSpec].(map[string]interface{})
	if !ok {
		err := fmt.Errorf("invalid or missing '%s' in payload", InjectSpec)
		logrus.Error(err)
		return err
	}
	injectSpec := make(map[string]int)
	for k, v := range injectSpecMap {
		floatVal, ok := v.(float64)
		if !ok {
			err := fmt.Errorf("invalid value for key '%s' in injectSpec", k)
			logrus.Error(err)
			return err
		}
		injectSpec[k] = int(floatVal)
	}

	// 更新任务状态
	updateTaskStatus(taskID, "Running", fmt.Sprintf("Executing fault injection for task %s", taskID))

	// 故障注入逻辑
	spec := handler.SpecMap[handler.ChaosType(faultType)]
	actionSpace, err := handler.GenerateActionSpace(spec)
	if err != nil {
		logrus.Error(err)
		return err
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
		Type:     handler.ChaosType(faultType),
		Spec:     chaosSpec,
		Duration: int(duration),
	}
	name := handler.Create(namespace, targetPod, config)
	if name == "" {
		return fmt.Errorf("create chaos failed, config: %+v", config)
	}
	jsonData, err := json.Marshal(config)
	if err != nil {
		logrus.Errorf("marshal config failed, config: %+v, err: %s", config, err)
		return err
	}

	// 创建新的故障注入记录
	faultRecord := database.FaultInjectionSchedule{
		ID:              taskID,           // 使用任务 ID 作为记录的主键
		FaultType:       faultType,        // 故障类型
		Config:          string(jsonData), // 故障配置 (JSON 格式化字符串)
		Duration:        int(duration),
		Description:     fmt.Sprintf("Fault for task %s", taskID),
		Status:          database.DatasetInitial,
		InjectionName:   name,
		ProposedEndTime: time.Now().Add(time.Duration(int(duration)+2) * time.Minute),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
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
		val, ok := payload[key].(string)
		if !ok || val == "" {
			return fmt.Errorf("missing or invalid '%s' key in payload", key)
		}
	}
	bench := payload[EvalPayloadBench].(string)
	algo := payload[EvalPayloadAlgo].(string)
	datasetName := payload[EvalPayloadDataset].(string)

	var faultRecord database.FaultInjectionSchedule
	err := database.DB.Where("injection_name = ?", datasetName).First(&faultRecord).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no matching fault injection record found for dataset: %s", datasetName)
		}
		return fmt.Errorf("failed to query database for dataset: %s, error: %v", datasetName, err)
	}

	var startTime time.Time
	var endTime time.Time
	if faultRecord.Status == database.DatasetSuccess {
		startTime = faultRecord.StartTime
		endTime = faultRecord.EndTime
	} else if faultRecord.Status == database.DatasetInitial {
		startTime, endTime, err = client.QueryCRDByName("ts", datasetName)
		if err != nil {
			return fmt.Errorf("failed to QueryCRDByName: %s, error: %v", datasetName, err)
		}
		if err := database.DB.Model(&faultRecord).Where("injection_name = ?", datasetName).
			Updates(map[string]interface{}{
				"start_time": startTime,
				"end_time":   endTime,
			}).Error; err != nil {
			return fmt.Errorf("failed to update start_time and end_time for dataset: %s, error: %v", datasetName, err)
		}
	}

	updateTaskStatus(taskID, "Running", fmt.Sprintf("Running algorithm for task %s", taskID))

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}

	parentDir := filepath.Dir(pwd)

	benchPath := filepath.Join(parentDir, "benchmarks", bench)
	algoPath := filepath.Join(parentDir, "algorithms", algo)
	startScriptPath := filepath.Join(parentDir, "experiments", "run_exp.py")

	if _, err := os.Stat(benchPath); os.IsNotExist(err) {
		return fmt.Errorf("benchmark directory does not exist: %s", benchPath)
	}
	if _, err := os.Stat(algoPath); os.IsNotExist(err) {
		return fmt.Errorf("algorithm directory does not exist: %s", algoPath)
	}
	if _, err := os.Stat(startScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("start script does not exist: %s", startScriptPath)
	}

	rc := &Rcabench{}
	con := rc.Evaluate(context.Background(), dag.Host().Directory(benchPath), dag.Host().Directory(algoPath), dag.Host().File(startScriptPath),
		startTime, endTime, startTime.Add(-1*time.Minute), startTime)

	_, err = con.Export(context.Background(), "./output")
	if err != nil {
		return fmt.Errorf("failed to export result, details: %s", err.Error())
	}

	return nil
}

func updateTaskStatus(taskID, status, message string) {
	ctx := context.Background()
	// 更新 Redis 中的任务状态
	taskKey := fmt.Sprintf("task:%s:status", taskID)
	if err := Rdb.HSet(ctx, taskKey, "status", status).Err(); err != nil {
		logrus.Errorf("Failed to update task status in Redis for task %s: %v", taskID, err)
	}
	if err := Rdb.HSet(ctx, taskKey, "updated_at", time.Now().Format(time.RFC3339)).Err(); err != nil {
		logrus.Errorf("Failed to update task updated_at in Redis for task %s: %v", taskID, err)
	}

	// 添加日志到 Redis
	logKey := fmt.Sprintf("task:%s:logs", taskID)
	if err := Rdb.RPush(ctx, logKey, fmt.Sprintf("%s - %s", time.Now().Format(time.RFC3339), message)).Err(); err != nil {
		logrus.Errorf("Failed to push log to Redis for task %s: %v", taskID, err)
	}
	logrus.Info(fmt.Sprintf("%s - %s", time.Now().Format(time.RFC3339), message))
	// 更新 SQLite 中的任务状态
	if err := database.DB.Model(&database.Task{}).Where("id = ?", taskID).Update("status", status).Error; err != nil {
		logrus.Errorf("Failed to update task %s in SQLite: %v", taskID, err)
	}
}
