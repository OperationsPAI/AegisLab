package handlers

import (
	"context"
	"dagger/rcabench/executor"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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
	taskData := map[string]interface{}{
		"taskID":   taskID,
		"taskType": taskType,
		"payload":  jsonPayload,
	}

	// 将任务添加到 Redis Stream 队列
	_, err = executor.Rdb.XAdd(c, &redis.XAddArgs{
		Stream: executor.StreamName, // 用单独的 Stream 键
		Values: taskData,
	}).Result()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to submit task"})
		return
	}

	// 设置任务状态初始值
	taskKey := fmt.Sprintf("task:%s:status", taskID)
	executor.Rdb.HSet(context.Background(), taskKey, "status", "Pending")

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
	var tasks []map[string]interface{}
	iter := executor.Rdb.Scan(context.Background(), 0, "task:*:status", 0).Iterator()

	for iter.Next(context.Background()) {
		taskKey := iter.Val()
		taskID := taskKey[len("task:") : len(taskKey)-len(":status")]

		status, err := executor.Rdb.HGet(context.Background(), taskKey, "status").Result()
		if err != nil && err != redis.Nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve task status"})
			return
		}

		logKey := "task:" + taskID + ":logs"
		logs, err := executor.Rdb.LRange(context.Background(), logKey, 0, -1).Result()
		if err != nil && err != redis.Nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve task logs"})
			return
		}

		var lastLogTime time.Time
		if len(logs) > 0 {
			lastLogTime, err = parseLogTimestamp(logs[len(logs)-1])
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to parse log timestamp"})
				return
			}
		}

		taskInfo := map[string]interface{}{
			"taskID":      taskID,
			"status":      status,
			"logs":        logs,
			"lastLogTime": lastLogTime,
		}
		tasks = append(tasks, taskInfo)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i]["lastLogTime"].(time.Time).After(tasks[j]["lastLogTime"].(time.Time))
	})

	c.HTML(http.StatusOK, "tasks.html", gin.H{
		"Tasks": tasks,
	})
}

func parseLogTimestamp(log string) (time.Time, error) {
	parts := strings.SplitN(log, " ", 2)
	return time.Parse(time.RFC3339, parts[0])
}
