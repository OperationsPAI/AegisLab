package handlers

import (
	"context"
	"dagger/rcabench/database"
	"dagger/rcabench/executor"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

func SubmitTask(c *gin.Context) {
	taskType := c.Query("type")
	if taskType != "FaultInjection" && taskType != "RunAlgorithm" && taskType != "EvaluateAlgorithm" {
		c.JSON(400, gin.H{"error": "Invalid task type"})
		return
	}

	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON payload"})
		return
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal payload"})
		return
	}
	taskID := uuid.New().String()

	// 将任务添加到 Redis Stream 队列
	_, err = executor.Rdb.XAdd(c, &redis.XAddArgs{
		Stream: executor.StreamName,
		Values: map[string]interface{}{
			"taskID":   taskID,
			"taskType": taskType,
			"payload":  jsonPayload,
		},
	}).Result()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to submit task"})
		return
	}

	// 保存任务到 SQLite 数据库
	task := database.Task{
		ID:      taskID,
		Type:    taskType,
		Payload: string(jsonPayload),
		Status:  "Pending",
	}
	if err := database.DB.Create(&task).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to save task to database"})
		return
	}

	c.JSON(202, gin.H{"taskID": taskID, "message": "Task submitted successfully"})
}

func GetTaskStatus(c *gin.Context) {
	taskID := c.Param("taskID")
	taskKey := fmt.Sprintf("task:%s:status", taskID) // 使用专用的状态键

	// 获取任务状态
	status, err := executor.Rdb.HGet(context.Background(), taskKey, "status").Result()
	if err != nil {
		c.JSON(404, gin.H{"error": "Task not found"})
		return
	}

	// 获取任务日志
	logKey := fmt.Sprintf("task:%s:logs", taskID) // 使用专用的日志键
	logs, err := executor.Rdb.LRange(context.Background(), logKey, 0, -1).Result()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	// 返回任务状态和日志
	c.JSON(200, gin.H{
		"taskID": taskID,
		"status": status,
		"logs":   logs,
	})
}

func ShowAllTasks(c *gin.Context) {
	var tasks []database.Task
	if err := database.DB.Find(&tasks).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve tasks"})
		return
	}

	c.HTML(http.StatusOK, "tasks.html", gin.H{
		"Tasks": tasks,
	})
}

func parseLogTimestamp(log string) (time.Time, error) {
	parts := strings.SplitN(log, " ", 2)
	return time.Parse(time.RFC3339, parts[0])
}

func GetTaskDetails(c *gin.Context) {
	taskID := c.Param("taskID")

	var task database.Task
	if err := database.DB.First(&task, "id = ?", taskID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(200, task)
}
func GetTaskLogs(c *gin.Context) {
	taskID := c.Param("taskID")
	logKey := fmt.Sprintf("task:%s:logs", taskID)

	logs, err := executor.Rdb.LRange(context.Background(), logKey, 0, -1).Result()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	c.JSON(200, gin.H{
		"taskID": taskID,
		"logs":   logs,
	})
}
