package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// 提交一个新任务到任务队列和数据库
func SubmitTask(ctx context.Context, taskType string, jsonPayload []byte) (string, bool) {
	taskID := uuid.New().String()

	// 提交任务到 Redis 任务队列
	_, err := client.GetRedisClient().XAdd(ctx, &redis.XAddArgs{
		Stream: StreamName,
		Values: map[string]interface{}{
			RdbMsgTaskID:   taskID,
			RdbMsgTaskType: taskType,
			RdbMsgPayload:  jsonPayload,
		},
	}).Result()
	if err != nil {
		logrus.Errorf("Failed to submit task, err: %s", err)
		return "Failed to submit task", false
	}

	// 保存任务到 SQLite 数据库
	task := database.Task{
		ID:      taskID,
		Type:    taskType,
		Payload: string(jsonPayload),
		Status:  "Pending",
	}
	if err := database.DB.Create(&task).Error; err != nil {
		logrus.Errorf("Failed to save task to database, err: %s", err)
		return "Failed to save task to database", false
	}

	content := map[string]interface{}{"task_id": taskID}

	var jsonContent []byte
	jsonContent, err = json.Marshal(content)
	if err != nil {
		return "Failed to marshal content", false
	}

	return string(jsonContent), true
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
