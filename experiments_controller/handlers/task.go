package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/CUHK-SE-Group/rcabench/executor"

	"github.com/CUHK-SE-Group/rcabench/database"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/config"

	chaosCli "github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var validTaskTypes = map[string]bool{
	string(executor.TaskTypeFaultInjection): true,
	string(executor.TaskTypeBuildImages):    true,
	string(executor.TaskTypeRunAlgorithm):   true,
}

// GetTaskStatus 查询任务状态
//
//	@Summary		查询任务状态
//	@Description	通过任务 ID 查询任务的执行状态和日志
//	@Tags			tasks
//	@Produce		json
//	@Param			taskID	path		string					true	"任务 ID"
//	@Success		200		{object}	map[string]interface{}	"返回任务状态和日志"
//	@Failure		404		{object}	map[string]string		"任务未找到"
//	@Failure		500		{object}	map[string]string		"服务器内部错误"
//	@Router			/tasks/{taskID}/status [get]
func GetTaskStatus(c *gin.Context) {
	taskID := c.Param("taskID")
	taskKey := fmt.Sprintf("task:%s:status", taskID) // 使用专用的状态键

	ctx := c.Request.Context()

	// 获取任务状态
	status, err := client.GetRedisClient().HGet(ctx, taskKey, "status").Result()
	if err == redis.Nil {
		c.JSON(404, gin.H{"error": "Task not found"})
		return
	} else if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve task status"})
		return
	}

	// 获取任务日志
	logKey := fmt.Sprintf("task:%s:logs", taskID) // 使用专用的日志键
	logs, err := client.GetRedisClient().LRange(ctx, logKey, 0, -1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
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

// ShowAllTasks 显示所有任务
//
//	@Summary		显示所有任务
//	@Description	显示数据库中所有任务的记录
//	@Tags			tasks
//	@Produce		html
//	@Success		200	"返回 HTML 页面显示任务"
//	@Failure		500	{object}	map[string]string	"服务器内部错误"
//	@Router			/tasks [get]
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

// SubmitTask 提交一个新任务到任务队列和数据库
//
//	@Summary		提交任务
//	@Description	提交一个指定类型的任务到系统
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			type	query		string				true	"任务类型 (FaultInjection, RunAlgorithm, EvaluateAlgorithm)"
//	@Param			payload	body		object				true	"任务参数"
//	@Success		202		{object}	map[string]string	"返回任务 ID"
//	@Failure		400		{object}	map[string]string	"请求错误或无效参数"
//	@Failure		500		{object}	map[string]string	"服务器内部错误"
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

	_, err = client.GetRedisClient().XAdd(ctx, &redis.XAddArgs{
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

// GetTaskDetails 获取任务详情
//
//	@Summary		获取任务详情
//	@Description	根据任务 ID 查询任务的详细信息
//	@Tags			tasks
//	@Produce		json
//	@Param			taskID	path		string				true	"任务 ID"
//	@Success		200		{object}	database.Task		"返回任务详情"
//	@Failure		404		{object}	map[string]string	"任务未找到"
//	@Failure		500		{object}	map[string]string	"服务器内部错误"
//	@Router			/tasks/{taskID} [get]
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

// GetTaskLogs 获取任务的日志
//
//	@Summary		获取任务日志
//	@Description	根据任务 ID 查询任务执行过程中记录的日志
//	@Tags			tasks
//	@Produce		json
//	@Param			taskID	path		string					true	"任务 ID"
//	@Success		200		{object}	map[string]interface{}	"返回任务的日志列表"
//	@Failure		500		{object}	map[string]string		"服务器内部错误"
//	@Router			/tasks/{taskID}/logs [get]
func GetTaskLogs(c *gin.Context) {
	taskID := c.Param("taskID")
	logKey := fmt.Sprintf("task:%s:logs", taskID)

	ctx := c.Request.Context()

	logs, err := client.GetRedisClient().LRange(ctx, logKey, 0, -1).Result()
	if errors.Is(err, redis.Nil) {
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

// GetAlgoBench 获取算法和基准列表
//
//	@Summary		获取算法和基准列表
//	@Description	返回算法和基准测试的文件列表
//	@Tags			algorithms
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"返回算法和基准的文件列表"
//	@Failure		500	{object}	map[string]string		"服务器内部错误"
//	@Router			/algorithms/benchmarks [get]
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

// GetDatasets 获取数据集列表
//
//	@Summary		获取数据集列表
//	@Description	返回已成功的故障注入数据集列表
//	@Tags			datasets
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"返回数据集名称列表"
//	@Failure		500	{object}	map[string]string		"服务器内部错误"
//	@Router			/datasets [get]
func GetDatasets(c *gin.Context) {
	var faultRecords []database.FaultInjectionSchedule

	currentTime := time.Now()

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
			startTime, endTime, err = chaosCli.QueryCRDByName("ts", datasetName)
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

// GetNamespacePod 获取命名空间中的 Pod 标签
//
//	@Summary		获取命名空间中的 Pod 标签
//	@Description	返回指定命名空间中符合条件的 Pod 标签列表
//	@Tags			pods
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"返回命名空间和对应的 Pod 标签信息"
//	@Failure		500	{object}	map[string]string		"服务器内部错误，无法获取 Pod 标签"
//	@Router			/pods/namespaces [get]
func GetNamespacePod(c *gin.Context) {
	res := make(map[string][]string)
	for _, ns := range config.GetStringSlice("injection.namespace") {
		labels, err := chaosCli.GetLabels(ns, config.GetString("injection.label"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get labels from namespace ts"})
		}
		res[ns] = labels
	}

	c.JSON(http.StatusOK, gin.H{
		"namespace_info": res,
	})
}

// WithdrawTask 撤回任务
//
//	@Summary		撤回任务
//	@Description	通过任务 ID 取消正在执行的任务
//	@Tags			tasks
//	@Produce		json
//	@Param			taskID	path		string				true	"任务 ID"
//	@Success		200		{object}	map[string]string	"任务撤回成功"
//	@Failure		400		{object}	map[string]string	"任务 ID 无效或撤回失败"
//	@Router			/tasks/{taskID}/withdraw [delete]
func WithdrawTask(c *gin.Context) {
	id := c.Param("taskID")
	if id == "" {
		c.JSON(400, gin.H{"error": "Task id is required"})
		return
	}
	err := executor.CancelTask(id)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Cancel Task failed: %v", err)})
		return
	}
}
