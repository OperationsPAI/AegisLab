package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/LGU-SE-Internal/rcabench/client"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GetTaskDetail
//
//	@Summary		Get task detail
//	@Description	Get detailed information of a task by task ID, including basic info and execution logs
//	@Tags			task
//	@Produce		json
//	@Param			task_id	path		string	true	"Task ID"
//	@Success		200		{object}	dto.GenericResponse[dto.TaskDetailResp]
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid task ID"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Task not found"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v1/tasks/{task_id} [get]
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
			message := "Task not found"
			logEntry.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusNotFound, message)
		} else {
			message := "Failed to retrieve task of injection"
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

// GetQueuedTasks
//
//	@Summary		Get queued tasks
//	@Description	Paginate and get the list of tasks waiting in the queue
//	@Tags			task
//	@Produce		json
//	@Param			page_num	query		int	false	"Page number"		default(1)
//	@Param			page_size	query		int	false	"Page size"		default(10)
//	@Success		200			{object}	dto.GenericResponse[dto.PaginationResp[dto.UnifiedTask]]
//	@Failure		400			{object}	dto.GenericResponse[any]
//	@Failure		500			{object}	dto.GenericResponse[any]
//	@Router			/api/v1/tasks/queue [get]
func GetQueuedTasks(c *gin.Context) {
	req := dto.PaginationQuery{
		PageNum:  1,
		PageSize: 10,
	}
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid pagination parameters")
		return
	}

	ctx := c.Request.Context()
	redisCli := client.GetRedisClient()
	var tasks []dto.UnifiedTask

	// Get tasks from ready queue (immediate execution)
	readyTasks, err := redisCli.LRange(ctx, executor.ReadyQueueKey, 0, -1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		logrus.Errorf("Failed to get ready tasks: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get ready tasks")
		return
	}

	for _, taskData := range readyTasks {
		var task dto.UnifiedTask
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

		var task dto.UnifiedTask
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
	start := (req.PageNum - 1) * req.PageSize
	end := start + req.PageSize
	if start >= totalTasks {
		// Return empty array if page is out of range
		tasks = []dto.UnifiedTask{}
	} else if end > totalTasks {
		// Adjust end if it exceeds the total number of tasks
		tasks = tasks[start:]
	} else {
		tasks = tasks[start:end]
	}

	totalPages := (totalTasks + req.PageSize - 1) / req.PageSize
	dto.SuccessResponse(c, dto.PaginationResp[dto.UnifiedTask]{
		Total:      int64(totalTasks),
		TotalPages: int64(totalPages),
		Items:      tasks,
	})
}

// ListTasks
//
//	@Summary		Get task list
//	@Description	Paginate and get task list by multiple conditions. Supports exact query by task ID, trace ID, group ID, or filter by type, status, etc.
//	@Tags			task
//	@Produce		json
//	@Param			task_id				query		string	false	"Task ID - exact match (mutually exclusive with trace_id, group_id)"
//	@Param			trace_id			query		string	false	"Trace ID - find all tasks in the same trace (mutually exclusive with task_id, group_id)"
//	@Param			group_id			query		string	false	"Group ID - find all tasks in the same group (mutually exclusive with task_id, trace_id)"
//	@Param			task_type			query		string	false	"Task type filter"	Enums(RestartService, FaultInjection, BuildDataset, RunAlgorithm, CollectResult, BuildImage)
//	@Param			status				query		string	false	"Task status filter"	Enums(Pending, Running, Completed, Error, Cancelled, Scheduled, Rescheduled)
//	@Param			immediate			query		bool	false	"Immediate execution - true: immediate, false: delayed"
//	@Param			sort_field			query		string	false	"Sort field, default created_at" default(created_at)
//	@Param			sort_order			query		string	false	"Sort order, default desc"	Enums(asc, desc)	default(desc)
//	@Param			limit				query		int		false	"Result limit, controls number of records returned"	minimum(1)
//	@Param			lookback			query		string	false	"Time range query, supports relative time (1h/24h/7d) or custom, default unset"
//	@Param			custom_start_time	query		string	false	"Custom start time, RFC3339 format, required if lookback=custom"	Format(date-time)
//	@Param			custom_end_time		query		string	false	"Custom end time, RFC3339 format, required if lookback=custom"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[dto.ListTasksResp]	"Successfully returned fault injection record list"
//	@Failure		400					{object}	dto.GenericResponse[any]	"Request parameter error, e.g. invalid format or validation failed"
//	@Failure		500					{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v1/tasks	[get]
func ListTasks(c *gin.Context) {
	var req dto.ListTasksReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("Invalid query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	_, tasks, err := repository.ListTasks(&req)
	if err != nil {
		logrus.Errorf("failed to fetch tasks: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tasks")
		return
	}

	dto.SuccessResponse(c, dto.ListTasksResp(tasks))
}
