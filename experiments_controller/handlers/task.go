package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/CUHK-SE-Group/rcabench/repository"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func GetTaskDetail(c *gin.Context) {
	var req dto.TaskReq
	if err := c.BindUri(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid URI")
		return
	}

	logEntry := logrus.WithField("task_id", req.TaskID)

	taskItem, err := repository.FindTaskItemByID(req.TaskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			message := "task not found"
			logEntry.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusNotFound, message)
		} else {
			message := "failed to retrieve task of injection"
			logEntry.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
		}

		return
	}

	logKey := fmt.Sprintf("task:%s:logs", taskItem.ID)
	logs, err := client.GetRedisClient().LRange(c.Request.Context(), logKey, 0, -1).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			message := "Failed to retrieve logs"
			logrus.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		logs = []string{}
	}

	dto.SuccessResponse(c, dto.TaskDetailResp{Task: *taskItem, Logs: logs})
}

func GetQueuedTasks(c *gin.Context) {
	// Set default values and bind pagination parameters
	req := dto.TaskQueuePaginationRequest{
		Page:     1,
		PageSize: 10,
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid pagination parameters")
		return
	}

	ctx := c.Request.Context()
	redisCli := client.GetRedisClient()
	var tasks []executor.UnifiedTask

	// Get tasks from ready queue (immediate execution)
	readyTasks, err := redisCli.LRange(ctx, executor.ReadyQueueKey, 0, -1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		logrus.Errorf("Failed to get ready tasks: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get ready tasks")
		return
	}

	for _, taskData := range readyTasks {
		var task executor.UnifiedTask
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			logrus.Warnf("Invalid task data: %v", err)
			continue
		}

		tasks = append(tasks, task)
	}

	// Get tasks from delayed queue (scheduled execution)
	delayedTasksWithScore, err := redisCli.ZRangeByScoreWithScores(ctx, executor.DelayedQueueKey, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  100,
	}).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		logrus.Errorf("Failed to get delayed tasks: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get delayed tasks")
		return
	}

	for _, z := range delayedTasksWithScore {
		taskData, ok := z.Member.(string)
		if !ok {
			continue
		}

		var task executor.UnifiedTask
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			logrus.Warnf("Invalid delayed task data: %v", err)
			continue
		}

		tasks = append(tasks, task)
	}

	// Sort tasks by execution time
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ExecuteTime < tasks[j].ExecuteTime
	})

	// Apply pagination
	totalTasks := len(tasks)
	totalPages := (totalTasks + req.PageSize - 1) / req.PageSize

	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if start >= totalTasks {
		// Return empty array if page is out of range
		tasks = []executor.UnifiedTask{}
	} else if end > totalTasks {
		// Adjust end if it exceeds the total number of tasks
		tasks = tasks[start:]
	} else {
		tasks = tasks[start:end]
	}

	// Return paginated response with metadata
	dto.SuccessResponse(c, gin.H{
		"tasks": tasks,
		"pagination": gin.H{
			"page":       req.Page,
			"pageSize":   req.PageSize,
			"totalItems": totalTasks,
			"totalPages": totalPages,
		},
	})
}
