package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/CUHK-SE-Group/rcabench/database"

	"github.com/CUHK-SE-Group/rcabench/client"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var (
	taskCancelFuncs      = make(map[string]context.CancelFunc)
	taskCancelFuncsMutex sync.Mutex
	taskSemaphore        = make(chan struct{}, 2) // 限制并发任务数为2
)

// 初始化函数
func init() {
	ctx := context.Background()
	initConsumerGroup(ctx)
}

// 初始化消费者组
func initConsumerGroup(ctx context.Context) {
	err := client.GetRedisClient().XGroupCreateMkStream(ctx, StreamName, GroupName, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		logrus.Fatalf("Failed to create consumer group: %v", err)
	}
}
func CancelTask(taskID string) error {
	taskCancelFuncsMutex.Lock()
	cancelFunc, exists := taskCancelFuncs[taskID]
	taskCancelFuncsMutex.Unlock()
	if !exists {
		return fmt.Errorf("no running task with ID %s", taskID)
	}
	cancelFunc()
	return nil
}

func ConsumeTasks() {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("ConsumeTasks panicked: %v", r)
		}
	}()

	ctx := context.Background()
	redisCli := client.GetRedisClient()

	consumerName := generateUniqueConsumerName()

	err := redisCli.XGroupCreateMkStream(ctx, StreamName, GroupName, "$").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		logrus.Errorf("Failed to create consumer group: %v", err)
		return
	}

	messages, err := readGroupMessages(ctx, redisCli, consumerName, []string{StreamName, "0"}, 10, 0)
	if err != nil {
		logrus.Panicf("Error reading from group messages: %v", err)
	}
	for _, msg := range messages {
		processMessage(ctx, redisCli, consumerName, msg)
	}

	for {
		messages, err = readGroupMessages(ctx, redisCli, consumerName, []string{StreamName, ">"}, 1, 5*time.Second)
		if err != nil {
			logrus.Errorf("Error reading from stream: %v", err)
			time.Sleep(time.Second)
			continue
		}

		if len(messages) == 0 {
			time.Sleep(time.Second)
			continue
		}

		for _, msg := range messages {
			processMessage(ctx, redisCli, consumerName, msg)
		}
	}
}

func processMessage(ctx context.Context, redisCli *redis.Client, consumerName string, msg redis.XMessage) {
	taskMsg, err := parseTaskMessage(msg)
	if err != nil {
		logrus.Errorf("Failed to parse task message: %v", err)
		redisCli.XAck(ctx, StreamName, GroupName, msg.ID)
		redisCli.XDel(ctx, StreamName, msg.ID)
		return
	}
	taskCtx, cancel := context.WithCancel(ctx)
	taskCancelFuncsMutex.Lock()
	taskCancelFuncs[taskMsg.TaskID] = cancel
	taskCancelFuncsMutex.Unlock()

	taskSemaphore <- struct{}{}
	go func(msg redis.XMessage, taskMsg *TaskMessage, ctx context.Context) {
		defer func() {
			taskCancelFuncsMutex.Lock()
			delete(taskCancelFuncs, taskMsg.TaskID)
			taskCancelFuncsMutex.Unlock()
			<-taskSemaphore
		}()
		processTaskWithContext(ctx, msg)
		redisCli.XAck(ctx, StreamName, GroupName, msg.ID)
	}(msg, taskMsg, taskCtx)
}

// 读取消费者组的消息
func readGroupMessages(ctx context.Context, redisCli *redis.Client, consumerName string, streams []string, count int64, block time.Duration) ([]redis.XMessage, error) {
	args := &redis.XReadGroupArgs{
		Group:    GroupName,
		Consumer: consumerName,
		Streams:  streams,
		Count:    count,
		Block:    block,
		NoAck:    false,
	}
	streamsMessages, err := redisCli.XReadGroup(ctx, args).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return []redis.XMessage{}, nil
		}
		return nil, err
	}
	var messages []redis.XMessage
	for _, streamMsg := range streamsMessages {
		messages = append(messages, streamMsg.Messages...)
	}
	return messages, nil
}

// 生成唯一的消费者名称
func generateUniqueConsumerName() string {
	return fmt.Sprintf("consumer-%s", uuid.New().String())
}

// 任务消息结构
type TaskMessage struct {
	TaskID       string
	TaskType     TaskType
	Payload      map[string]interface{}
	ParentTaskID string
}

func processTaskWithContext(ctx context.Context, msg redis.XMessage) {
	logrus.Infof("Processing message ID: %s", msg.ID)
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("processTask panicked: %v\n%s", r, debug.Stack())
		}
		ackCtx := context.Background()
		client := client.GetRedisClient()
		client.XAck(ackCtx, StreamName, GroupName, msg.ID)
		client.XDel(ackCtx, StreamName, msg.ID)
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
		execErr = executeFaultInjection(ctx, taskMsg.TaskID, taskMsg.Payload)
	case TaskTypeRunAlgorithm:
		execErr = executeAlgorithm(ctx, taskMsg.TaskID, taskMsg.Payload)
	case TaskTypeBuildImages:
		execErr = executeBuildImages(ctx, taskMsg.TaskID, taskMsg.Payload)
	case TaskTypeBuildDataset:
		execErr = executeBuildDataset(ctx, taskMsg.TaskID, taskMsg.Payload)
	case TaskTypeCollectResult:
		execErr = executeBuildDataset(ctx, taskMsg.TaskID, taskMsg.Payload)
	default:
		execErr = fmt.Errorf("unknown task type: %s", taskMsg.TaskType)
	}

	if execErr != nil {
		if errors.Is(execErr, context.Canceled) {
			updateTaskStatus(taskMsg.TaskID, "Canceled", fmt.Sprintf("Task %s was canceled", taskMsg.TaskID))
			logrus.Infof("Task %s was canceled", taskMsg.TaskID)
		} else {
			updateTaskStatus(taskMsg.TaskID, "Error", fmt.Sprintf("Task %s error, message: %s", taskMsg.TaskID, execErr))
			logrus.Error(execErr)
		}
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

// 更新任务状态
func updateTaskStatus(taskID, status, message string) {
	ctx := context.Background()
	client := client.GetRedisClient()

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
