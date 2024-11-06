package executor

import (
	"context"
	"dagger/rcabench/database"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"dagger.io/dagger/dag"
	"github.com/go-redis/redis/v8"
	"github.com/k0kubun/pp"
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
			Streams: []string{StreamName, "$"},
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
			jsonPayload := entry.Values["payload"].(string)
			payload := map[string]interface{}{}
			err = json.Unmarshal([]byte(jsonPayload), &payload)
			if err != nil {
				logrus.Errorf("Unmarshaling %s failed", jsonPayload)
			}

			logrus.Infof("Executing %s", taskID)

			updateTaskStatus(taskID, "Running", fmt.Sprintf("Task %s started of type %s", taskID, taskType))

			switch taskType {
			case "FaultInjection":
				err = executeFaultInjection(taskID, payload)
			case "RunAlgorithm":
				err = executeAlgorithm(taskID, payload)
			case "EvaluateAlgorithm":
				err = evaluateAlgorithm(taskID, payload)
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
	updateTaskStatus(taskID, "Running", fmt.Sprintf("Executing fault injection for task %s", taskID))
	time.Sleep(2 * time.Second)
	return nil
}

func executeAlgorithm(taskID string, payload map[string]interface{}) error {
	pp.Print(payload)
	requiredKeys := []string{"benchmarks", "algorithms"}
	for _, key := range requiredKeys {
		if val, ok := payload[key].(string); !ok || val == "" {
			return fmt.Errorf("missing or invalid '%s' key in payload", key)
		}
	}
	bench := payload["benchmarks"].(string)
	algo := payload["algorithms"].(string)

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

func evaluateAlgorithm(taskID string, payload map[string]interface{}) error {
	pp.Print(payload)
	updateTaskStatus(taskID, "Running", fmt.Sprintf("Evaluating algorithm for task %s", taskID))
	time.Sleep(3 * time.Second)
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

	// 更新 SQLite 中的任务状态
	if err := database.DB.Model(&database.Task{}).Where("id = ?", taskID).Update("status", status).Error; err != nil {
		fmt.Printf("Failed to update task %s in SQLite: %v\n", taskID, err)
	}
}
