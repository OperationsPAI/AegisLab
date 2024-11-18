package handlers

import (
	"dagger/rcabench/config"
	"dagger/rcabench/database"
	"dagger/rcabench/executor"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var validTaskTypes = map[string]bool{
	"FaultInjection":    true,
	"RunAlgorithm":      true,
	"EvaluateAlgorithm": true,
}

func SubmitTask(c *gin.Context) {
	taskType := c.Query("type")
	if taskType == "" {
		c.JSON(400, gin.H{"error": "Task type is required"})
		return
	}
	if !validTaskTypes[taskType] {
		c.JSON(400, gin.H{"error": "Invalid task type"})
		return
	}

	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON payload"})
		return
	}
	taskID := uuid.New().String()
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal payload"})
		return
	}

	ctx := c.Request.Context()

	_, err = executor.Rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: executor.StreamName,
		Values: map[string]interface{}{
			executor.RdbMsgTaskID:   taskID,
			executor.RdbMsgTaskType: taskType,
			executor.RdbMsgPayload:  jsonPayload,
		},
	}).Result()

	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to submit task, err: %s", err)})
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

	ctx := c.Request.Context()

	// 获取任务状态
	status, err := executor.Rdb.HGet(ctx, taskKey, "status").Result()
	if err == redis.Nil {
		c.JSON(404, gin.H{"error": "Task not found"})
		return
	} else if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve task status"})
		return
	}

	// 获取任务日志
	logKey := fmt.Sprintf("task:%s:logs", taskID) // 使用专用的日志键
	logs, err := executor.Rdb.LRange(ctx, logKey, 0, -1).Result()
	if err != nil && err != redis.Nil {
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

func GetTaskDetails(c *gin.Context) {
	taskID := c.Param("taskID")

	var task database.Task
	if err := database.DB.First(&task, "id = ?", taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": "Task not found"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to retrieve task"})
		}
		return
	}

	c.JSON(200, task)
}

func GetTaskLogs(c *gin.Context) {
	taskID := c.Param("taskID")
	logKey := fmt.Sprintf("task:%s:logs", taskID)

	ctx := c.Request.Context()

	logs, err := executor.Rdb.LRange(ctx, logKey, 0, -1).Result()
	if err == redis.Nil {
		logs = []string{}
	} else if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	c.JSON(200, gin.H{
		"taskID": taskID,
		"logs":   logs,
	})
}

func getSubFiles(dir string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

func GetAlgoBench(c *gin.Context) {
	pwd, err := os.Getwd()
	if err != nil {
		c.JSON(500, gin.H{"error": "Get work directory failed"})
		return
	}

	parentDir := filepath.Dir(pwd)
	benchPath := filepath.Join(parentDir, "benchmarks")
	algoPath := filepath.Join(parentDir, "algorithms")

	// 获取benchPath下的文件列表
	benchFiles, err := getSubFiles(benchPath)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to list files in %s: %v", benchPath, err)})
		return
	}

	// 获取algoPath下的文件列表
	algoFiles, err := getSubFiles(algoPath)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to list files in %s: %v", algoPath, err)})
		return
	}

	c.JSON(200, gin.H{
		"benchmarks": benchFiles,
		"algorithms": algoFiles,
	})
}

func GetInjectionPara(c *gin.Context) {
	choice := make(map[string][]handler.ActionSpace, 0)
	for tp, spec := range handler.SpecMap {
		actionSpace, err := handler.GenerateActionSpace(spec)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate action space"})
			return
		}
		name := handler.GetChaosTypeName(tp)
		choice[name] = actionSpace
	}
	c.JSON(200, gin.H{
		"specification": choice,
		"keymap":        handler.ChaosTypeMap,
	})
}

func GetDatasets(c *gin.Context) {
	var faultRecords []database.FaultInjectionSchedule

	// 获取当前时间
	currentTime := time.Now()

	// 查询所有非失败状态的记录，并满足 CreatedAt + Duration < 当前时间 的条件
	err := database.DB.
		Where("status != ?", database.DatasetFailed).
		Where("proposed_end_time < ?", currentTime).
		Find(&faultRecords).Error

	if err != nil {
		logrus.Errorf("Failed to query fault injection schedules: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve datasets"})
		return
	}

	// 用于存储最终成功的记录
	var successfulRecords []database.FaultInjectionSchedule

	for _, record := range faultRecords {
		datasetName := record.InjectionName
		var startTime, endTime time.Time

		// 如果状态为初始，查询 CRD 并更新记录
		if record.Status == database.DatasetInitial {
			startTime, endTime, err = client.QueryCRDByName("ts", datasetName)
			if err != nil {
				logrus.Errorf("Failed to QueryCRDByName for dataset %s: %v", datasetName, err)

				// 更新状态为失败
				if updateErr := database.DB.Model(&record).Where("injection_name = ?", datasetName).
					Update("status", database.DatasetFailed).Error; updateErr != nil {
					logrus.Errorf("Failed to update status to DatasetFailed for dataset %s: %v", datasetName, updateErr)
				}
				continue
			}

			// 更新数据库中的 start_time、end_time 和状态为成功
			if updateErr := database.DB.Model(&record).Where("injection_name = ?", datasetName).
				Updates(map[string]interface{}{
					"start_time": startTime,
					"end_time":   endTime,
					"status":     database.DatasetSuccess,
				}).Error; updateErr != nil {
				logrus.Errorf("Failed to update record for dataset %s: %v", datasetName, updateErr)
				continue
			}
			// 更新成功的记录状态到内存
			record.StartTime = startTime
			record.EndTime = endTime
			record.Status = database.DatasetSuccess
		}

		// 仅保留状态为成功的记录
		if record.Status == database.DatasetSuccess {
			successfulRecords = append(successfulRecords, record)
		}
	}

	datasetNames := []string{}
	for _, record := range successfulRecords {
		datasetNames = append(datasetNames, record.InjectionName)
	}

	// 返回成功的记录
	c.JSON(http.StatusOK, gin.H{
		"datasets": datasetNames,
	})
}

func GetNamespacePod(c *gin.Context) {
	res := make(map[string][]string)
	for _, ns := range config.GetStringSlice("injection.namespace") {
		labels, err := client.GetLabels(ns, config.GetString("injection.label"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get labels from namespace ts"})
		}
		res[ns] = labels
	}

	c.JSON(http.StatusOK, gin.H{
		"namespace_info": res,
	})
}
