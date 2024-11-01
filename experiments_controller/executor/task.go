package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
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
			Streams: []string{StreamName, "$"}, // 从最新位置开始读取
			Count:   1,
			Block:   0,
		}).Result()
		if err != nil {
			fmt.Println("Error reading from stream:", err)
			continue
		}

		for _, entry := range entries[0].Messages {
			taskID := entry.Values["taskID"].(string)
			taskType := entry.Values["taskType"].(string)

			logrus.Infof("Executing %s", taskID)

			updateTaskStatus(taskID, "Running", fmt.Sprintf("Task %s started of type %s", taskID, taskType))

			switch taskType {
			case "FaultInjection":
				executeFaultInjection(taskID)
			case "RunAlgorithm":
				executeAlgorithm(taskID)
			case "EvaluateAlgorithm":
				evaluateAlgorithm(taskID)
			}

			updateTaskStatus(taskID, "Completed", fmt.Sprintf("Task %s completed", taskID))
			Rdb.XDel(context.Background(), StreamName, entry.ID)
		}
	}
}

func executeFaultInjection(taskID string) {
	updateTaskStatus(taskID, "Running", fmt.Sprintf("Executing fault injection for task %s", taskID))
	time.Sleep(2 * time.Second)
}

func executeAlgorithm(taskID string) {
	updateTaskStatus(taskID, "Running", fmt.Sprintf("Running algorithm for task %s", taskID))
	time.Sleep(5 * time.Second)
}

func evaluateAlgorithm(taskID string) {
	updateTaskStatus(taskID, "Running", fmt.Sprintf("Evaluating algorithm for task %s", taskID))
	time.Sleep(3 * time.Second)
}

func updateTaskStatus(taskID, status, message string) {
	taskKey := fmt.Sprintf("task:%s:status", taskID)
	Rdb.HSet(context.Background(), taskKey, "status", status)
	Rdb.HSet(context.Background(), taskKey, "updated_at", time.Now().Format(time.RFC3339))

	logKey := fmt.Sprintf("task:%s:logs", taskID)
	Rdb.RPush(context.Background(), logKey, fmt.Sprintf("%s - %s", time.Now().Format(time.RFC3339), message))
}
