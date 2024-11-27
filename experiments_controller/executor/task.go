package executor

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"dagger/rcabench/config"
	"dagger/rcabench/database"

	"dagger.io/dagger/dag"

	"github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// 定义任务类型
type TaskType string

const (
	TaskTypeRunAlgorithm   TaskType = "RunAlgorithm"
	TaskTypeFaultInjection TaskType = "FaultInjection"
)

// Redis 消息字段
const (
	RdbMsgTaskID       = "taskID"
	RdbMsgTaskType     = "taskType"
	RdbMsgPayload      = "payload"
	RdbMsgParentTaskID = "parentTaskID"
)

// 不同任务类型的 Payload 键
const (
	EvalPayloadAlgo    = "algorithm"
	EvalPayloadDataset = "dataset"
	EvalPayloadBench   = "benchmark"

	InjectFaultType = "faultType"
	InjectNamespace = "injectNamespace"
	InjectPod       = "injectPod"
	InjectDuration  = "duration"
	InjectSpec      = "spec"
)

// Redis 流和消费者组配置
const (
	StreamName   = "task_stream"
	GroupName    = "task_consumer_group"
	ConsumerName = "task_consumer"
)

// 单例模式的 Redis 客户端
var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

// 初始化函数
func init() {
	ctx := context.Background()
	client := GetRedisClient()
	initConsumerGroup(ctx, client)
}

// 获取 Redis 客户端
func GetRedisClient() *redis.Client {
	redisOnce.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     config.GetString("redis.host"),
			Password: "",
			DB:       0,
		})

		ctx := context.Background()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			logrus.Fatalf("Failed to connect to Redis: %v", err)
		}
	})
	return redisClient
}

// 初始化消费者组
func initConsumerGroup(ctx context.Context, client *redis.Client) {
	err := client.XGroupCreateMkStream(ctx, StreamName, GroupName, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		logrus.Fatalf("Failed to create consumer group: %v", err)
	}
}

// 消费任务
func ConsumeTasks() {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("ConsumeTasks panicked: %v", r)
		}
	}()

	ctx := context.Background()
	client := GetRedisClient()

	for {
		messages, err := readMessages(ctx, client, []string{StreamName, "0"}, 1, 5*time.Second)
		if err != nil {
			logrus.Errorf("Error reading from stream: %v", err)
			time.Sleep(time.Second)
			continue
		}
		if len(messages) == 0 {
			messages, err = readMessages(ctx, client, []string{StreamName, ">"}, 1, 5*time.Second)
			if err != nil {
				logrus.Errorf("Error reading from stream: %v", err)
				time.Sleep(time.Second)
				continue
			}
		}

		if len(messages) == 0 {
			time.Sleep(time.Second)
			continue
		}

		for _, msg := range messages {
			processTask(msg)
		}
	}
}

// 读取消息
func readMessages(ctx context.Context, client *redis.Client, streams []string, count int, block time.Duration) ([]redis.XMessage, error) {
	entries, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    GroupName,
		Consumer: ConsumerName,
		Streams:  streams,
		Count:    int64(count),
		Block:    block,
	}).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if len(entries) > 0 && len(entries[0].Messages) > 0 {
		return entries[0].Messages, nil
	}
	return nil, nil
}

// 任务消息结构
type TaskMessage struct {
	TaskID       string
	TaskType     TaskType
	Payload      map[string]interface{}
	ParentTaskID string
}

// 处理任务
func processTask(msg redis.XMessage) {
	logrus.Infof("Processing message ID: %s", msg.ID)
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("processTask panicked: %v\n%s", r, debug.Stack())
		}
		ctx := context.Background()
		client := GetRedisClient()
		client.XAck(ctx, StreamName, GroupName, msg.ID)
		client.XDel(ctx, StreamName, msg.ID)
	}()

	taskMsg, err := parseTaskMessage(msg)
	if err != nil {
		logrus.Errorf("Failed to parse task message: %v", err)
		return
	}

	logrus.Infof("Executing task ID: %s", taskMsg.TaskID)
	updateTaskStatus(taskMsg.TaskID, "Running", fmt.Sprintf("Task %s started of type %s", taskMsg.TaskID, taskMsg.TaskType))

	var execErr error
	switch taskMsg.TaskType {
	case TaskTypeFaultInjection:
		execErr = executeFaultInjection(taskMsg.TaskID, taskMsg.Payload)
	case TaskTypeRunAlgorithm:
		execErr = executeAlgorithm(taskMsg.TaskID, taskMsg.Payload)
	default:
		execErr = fmt.Errorf("unknown task type: %s", taskMsg.TaskType)
	}

	if execErr != nil {
		updateTaskStatus(taskMsg.TaskID, "Error", fmt.Sprintf("Task %s error, message: %s", taskMsg.TaskID, execErr))
		logrus.Error(execErr)
	} else {
		updateTaskStatus(taskMsg.TaskID, "Completed", fmt.Sprintf("Task %s completed", taskMsg.TaskID))
		logrus.Infof("Task %s completed", taskMsg.TaskID)
	}
}

// 解析任务消息
func parseTaskMessage(msg redis.XMessage) (*TaskMessage, error) {
	taskID, ok := msg.Values[RdbMsgTaskID].(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing taskID in message")
	}
	taskTypeStr, ok := msg.Values[RdbMsgTaskType].(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing taskType in message")
	}
	taskType := TaskType(taskTypeStr)
	jsonPayload, ok := msg.Values[RdbMsgPayload].(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing payload in message")
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(jsonPayload), &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %v", err)
	}
	parentTaskID, _ := msg.Values[RdbMsgParentTaskID].(string)
	return &TaskMessage{
		TaskID:       taskID,
		TaskType:     taskType,
		Payload:      payload,
		ParentTaskID: parentTaskID,
	}, nil
}

// 故障注入任务的 Payload 结构
type FaultInjectionPayload struct {
	Namespace  string
	Pod        string
	FaultType  int
	Duration   int
	InjectSpec map[string]int
}

// 执行故障注入任务
func executeFaultInjection(taskID string, payload map[string]interface{}) error {
	logrus.Infof("Injecting fault, taskID: %s", taskID)

	fiPayload, err := parseFaultInjectionPayload(payload)
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

	config := handler.ChaosConfig{
		Type:     handler.ChaosType(fiPayload.FaultType),
		Spec:     chaosSpec,
		Duration: fiPayload.Duration,
	}
	name := handler.Create(fiPayload.Namespace, fiPayload.Pod, config)
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

	// 写入数据库
	if err := database.DB.Create(&faultRecord).Error; err != nil {
		logrus.Errorf("Failed to write fault injection schedule to database: %v", err)
		return fmt.Errorf("failed to write to database: %v", err)
	}

	return nil
}

// 解析故障注入任务的 Payload
func parseFaultInjectionPayload(payload map[string]interface{}) (*FaultInjectionPayload, error) {
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

// 算法执行任务的 Payload 结构
type AlgorithmExecutionPayload struct {
	Benchmark   string
	Algorithm   string
	DatasetName string
}

// 执行算法任务
func executeAlgorithm(taskID string, payload map[string]interface{}) error {
	algPayload, err := parseAlgorithmExecutionPayload(payload)
	if err != nil {
		return err
	}

	var faultRecord database.FaultInjectionSchedule
	err = database.DB.Where("injection_name = ?", algPayload.DatasetName).First(&faultRecord).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no matching fault injection record found for dataset: %s", algPayload.DatasetName)
		}
		return fmt.Errorf("failed to query database for dataset: %s, error: %v", algPayload.DatasetName, err)
	}

	var startTime, endTime time.Time
	if faultRecord.Status == database.DatasetSuccess {
		startTime = faultRecord.StartTime
		endTime = faultRecord.EndTime
	} else if faultRecord.Status == database.DatasetInitial {
		startTime, endTime, err = client.QueryCRDByName("ts", algPayload.DatasetName)
		if err != nil {
			return fmt.Errorf("failed to QueryCRDByName: %s, error: %v", algPayload.DatasetName, err)
		}
		if err := database.DB.Model(&faultRecord).Where("injection_name = ?", algPayload.DatasetName).
			Updates(map[string]interface{}{
				"start_time": startTime,
				"end_time":   endTime,
			}).Error; err != nil {
			return fmt.Errorf("failed to update start_time and end_time for dataset: %s, error: %v", algPayload.DatasetName, err)
		}
	}

	executionResult := database.ExecutionResult{
		Dataset: faultRecord.ID,
		TaskID:  taskID,
		Algo:    algPayload.Algorithm,
	}
	if err := database.DB.Create(&executionResult).Error; err != nil {
		return fmt.Errorf("failed to create execution result: %v", err)
	}

	updateTaskStatus(taskID, "Running", fmt.Sprintf("Running algorithm for task %s", taskID))

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}

	parentDir := filepath.Dir(pwd)

	benchPath := filepath.Join(parentDir, "benchmarks", algPayload.Benchmark)
	algoPath := filepath.Join(parentDir, "algorithms", algPayload.Algorithm)
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
		startTime, endTime, startTime.Add(-20*time.Minute), startTime)

	if config.GetBool("debug") {
		_, err = con.Directory("/app/output").Export(context.Background(), "./output")
		if err != nil {
			return fmt.Errorf("failed to export result, details: %s", err.Error())
		}
		_, err = con.Directory("/app/input").Export(context.Background(), "./input")
		if err != nil {
			return fmt.Errorf("failed to export result, details: %s", err.Error())
		}
	}

	content, err := con.File("/app/output/result.csv").Contents(context.Background())
	if err != nil {
		updateTaskStatus(taskID, "Running", "There is no result.csv file in /app/output, please check whether it is nomal")
	} else {
		results, err := readCSVContent2Result(content, executionResult.ID)
		if err != nil {
			return fmt.Errorf("convert result.csv to database struct failed: %v", err)
		}
		if err := database.DB.Create(&results).Error; err != nil {
			return fmt.Errorf("save result.csv to database failed: %v", err)
		}
	}

	conclusion, err := con.File("/app/output/conclusion.csv").Contents(context.Background())
	if err != nil {
		updateTaskStatus(taskID, "Running", "There is no conclusion.csv file in /app/output, please check whether it is nomal")

	} else {
		results, err := readDetectorCSV(conclusion, executionResult.ID)
		if err != nil {
			return fmt.Errorf("convert result.csv to database struct failed: %v", err)
		}
		fmt.Println(results)
		if err := database.DB.Create(&results).Error; err != nil {
			return fmt.Errorf("save conclusion.csv to database failed: %v", err)
		}
	}
	return nil
}

// 解析算法执行任务的 Payload
func parseAlgorithmExecutionPayload(payload map[string]interface{}) (*AlgorithmExecutionPayload, error) {
	benchmark, ok := payload[EvalPayloadBench].(string)
	if !ok || benchmark == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", EvalPayloadBench)
	}
	algorithm, ok := payload[EvalPayloadAlgo].(string)
	if !ok || algorithm == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", EvalPayloadAlgo)
	}
	datasetName, ok := payload[EvalPayloadDataset].(string)
	if !ok || datasetName == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", EvalPayloadDataset)
	}
	return &AlgorithmExecutionPayload{
		Benchmark:   benchmark,
		Algorithm:   algorithm,
		DatasetName: datasetName,
	}, nil
}

// 更新任务状态
func updateTaskStatus(taskID, status, message string) {
	ctx := context.Background()
	client := GetRedisClient()

	// 更新 Redis 中的任务状态
	taskKey := fmt.Sprintf("task:%s:status", taskID)
	if err := client.HSet(ctx, taskKey, "status", status).Err(); err != nil {
		logrus.Errorf("Failed to update task status in Redis for task %s: %v", taskID, err)
	}
	if err := client.HSet(ctx, taskKey, "updated_at", time.Now().Format(time.RFC3339)).Err(); err != nil {
		logrus.Errorf("Failed to update task updated_at in Redis for task %s: %v", taskID, err)
	}

	// 添加日志到 Redis
	logKey := fmt.Sprintf("task:%s:logs", taskID)
	if err := client.RPush(ctx, logKey, fmt.Sprintf("%s - %s", time.Now().Format(time.RFC3339), message)).Err(); err != nil {
		logrus.Errorf("Failed to push log to Redis for task %s: %v", taskID, err)
	}
	logrus.Info(fmt.Sprintf("%s - %s", time.Now().Format(time.RFC3339), message))
	// 更新 SQLite 中的任务状态
	if err := database.DB.Model(&database.Task{}).Where("id = ?", taskID).Update("status", status).Error; err != nil {
		logrus.Errorf("Failed to update task %s in SQLite: %v", taskID, err)
	}
}

// 读取 CSV 内容并转换为结果
func readCSVContent2Result(csvContent string, executionID int) ([]database.GranularityResult, error) {
	reader := csv.NewReader(strings.NewReader(csvContent))

	// 读取表头
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	expectedHeader := []string{"level", "result", "rank", "confidence"}
	if len(header) != len(expectedHeader) {
		return nil, fmt.Errorf("unexpected header length: got %d, expected %d", len(header), len(expectedHeader))
	}
	for i, field := range header {
		if field != expectedHeader[i] {
			return nil, fmt.Errorf("unexpected header field at column %d: got '%s', expected '%s'", i+1, field, expectedHeader[i])
		}
	}

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV rows: %v", err)
	}

	var results []database.GranularityResult
	for i, row := range rows {
		if len(row) != len(expectedHeader) {
			return nil, fmt.Errorf("row %d has incorrect number of columns: got %d, expected %d", i+1, len(row), len(expectedHeader))
		}

		level := row[0]
		result := row[1]
		rank, err := strconv.Atoi(row[2])
		if err != nil {
			return nil, fmt.Errorf("invalid rank value in row %d: %v", i+1, err)
		}
		confidence, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid confidence value in row %d: %v", i+1, err)
		}

		results = append(results, database.GranularityResult{
			ExecutionID: executionID,
			Level:       level,
			Result:      result,
			Rank:        rank,
			Confidence:  confidence,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	return results, nil
}

func readDetectorCSV(csvContent string, executionID int) ([]database.Detector, error) {
	reader := csv.NewReader(strings.NewReader(csvContent))

	// 读取表头
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	expectedHeader := []string{"SpanName", "Issues", "AvgDuration", "SuccRate", "P90", "P95", "P99"}
	if len(header) != len(expectedHeader) {
		return nil, fmt.Errorf("unexpected header length: got %d, expected %d", len(header), len(expectedHeader))
	}
	for i, field := range header {
		if field != expectedHeader[i] {
			return nil, fmt.Errorf("unexpected header field at column %d: got '%s', expected '%s'", i+1, field, expectedHeader[i])
		}
	}

	// 读取所有行
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV rows: %v", err)
	}

	var results []database.Detector
	for i, row := range rows {
		if len(row) != len(expectedHeader) {
			return nil, fmt.Errorf("row %d has incorrect number of columns: got %d, expected %d", i+1, len(row), len(expectedHeader))
		}

		spanName := row[0]
		issues := row[1]

		// 处理空值
		var avgDuration, succRate, p90, p95, p99 *float64

		// 如果字段非空，转换为 float64，否则设置为 nil
		if row[2] != "" {
			val, err := strconv.ParseFloat(row[2], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid AvgDuration value in row %d: %v", i+1, err)
			}
			avgDuration = &val
		}
		if row[3] != "" {
			val, err := strconv.ParseFloat(row[3], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid SuccRate value in row %d: %v", i+1, err)
			}
			succRate = &val
		}
		if row[4] != "" {
			val, err := strconv.ParseFloat(row[4], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid P90 value in row %d: %v", i+1, err)
			}
			p90 = &val
		}
		if row[5] != "" {
			val, err := strconv.ParseFloat(row[5], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid P95 value in row %d: %v", i+1, err)
			}
			p95 = &val
		}
		if row[6] != "" {
			val, err := strconv.ParseFloat(row[6], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid P99 value in row %d: %v", i+1, err)
			}
			p99 = &val
		}

		// 将数据添加到结果
		results = append(results, database.Detector{
			ExecutionID: executionID,
			SpanName:    spanName,
			Issues:      issues,
			AvgDuration: avgDuration,
			SuccRate:    succRate,
			P90:         p90,
			P95:         p95,
			P99:         p99,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	return results, nil
}
