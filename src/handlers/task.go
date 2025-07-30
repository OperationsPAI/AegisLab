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
//	@Summary		获取任务详情
//	@Description	根据任务ID获取任务详细信息,包括任务基本信息和执行日志
//	@Tags			task
//	@Produce		json
//	@Param			task_id	path		string	true	"任务ID"
//	@Success		200		{object}	dto.GenericResponse[dto.TaskDetailResp]
//	@Failure		400		{object}	dto.GenericResponse[any]	"无效的任务ID"
//	@Failure		404		{object}	dto.GenericResponse[any]	"任务不存在"
//	@Failure		500		{object}	dto.GenericResponse[any]	"服务器内部错误"
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

// GetQueuedTasks
//
//	@Summary		获取队列中的任务
//	@Description	分页获取队列中等待执行的任务列表
//	@Tags			task
//	@Produce		json
//	@Param			page_num	query		int	false	"页码"		default(1)
//	@Param			page_size	query		int	false	"每页大小"	default(10)
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
//	@Summary		获取任务列表
//	@Description	根据多种条件分页获取任务列表。支持按任务ID、跟踪ID、组ID进行精确查询，或按任务类型、状态等进行过滤查询
//	@Tags			task
//	@Produce		json
//	@Param			task_id				query		string	false	"任务ID - 精确匹配特定任务 (与trace_id、group_id互斥)"
//	@Param			trace_id			query		string	false	"跟踪ID - 查找属于同一跟踪的所有任务 (与task_id、group_id互斥)"
//	@Param			group_id			query		string	false	"组ID - 查找属于同一组的所有任务 (与task_id、trace_id互斥)"
//	@Param			task_type			query		string	false	"任务类型过滤"	Enums(RestartService, FaultInjection, BuildDataset, RunAlgorithm, CollectResult, BuildImage)
//	@Param			status				query		string	false	"任务状态过滤"	Enums(Pending, Running, Completed, Error, Cancelled, Scheduled, Rescheduled)
//	@Param			immediate			query		bool	false	"是否立即执行 - true:立即执行任务, false:延时执行任务"
//	@Param			sort_field			query		string	false	"排序字段，默认created_at" default(created_at)
//	@Param			sort_order			query		string	false	"排序方式，默认desc"	Enums(asc, desc)	default(desc)
//	@Param			limit				query		int		false	"结果数量限制，用于控制返回记录数量"	minimum(1)
//	@Param			lookback			query		string	false	"时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置"
//	@Param			custom_start_time	query		string	false	"自定义开始时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Param			custom_end_time		query		string	false	"自定义结束时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[dto.ListTasksResp]	"成功返回故障注入记录列表"
//	@Failure		400					{object}	dto.GenericResponse[any]	"请求参数错误，如参数格式不正确、验证失败等"
//	@Failure		500					{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/tasks	[get]
func ListTasks(c *gin.Context) {
	var req dto.ListTasksReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("invalid query parameters: %v", err)
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
