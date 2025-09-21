package v2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"aegis/client"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// GetTask handles getting a single task by ID
//
//	@Summary Get task by ID
//	@Description Get detailed information about a specific task including logs
//	@Tags Tasks
//	@Produce json
//	@Security BearerAuth
//	@Param id path string true "Task ID"
//	@Param include query []string false "Include additional data (logs)" collectionFormat(multi)
//	@Success 200 {object} dto.GenericResponse[dto.TaskDetailResponse] "Task retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid task ID"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Task not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/tasks/{id} [get]
func GetTask(c *gin.Context) {

	taskID := c.Param("id")
	includeList := c.QueryArray("include")

	logEntry := logrus.WithField("task_id", taskID)

	taskItem, err := repository.FindTaskItemByID(taskID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			message := "Task not found"
			logEntry.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusNotFound, message)
		} else {
			message := "Failed to get task"
			logEntry.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
		}
		return
	}

	// Convert to response DTO
	response := dto.TaskDetailResponse{
		TaskResponse: dto.TaskResponse{
			ID:        taskItem.ID,
			Type:      taskItem.Type,
			Status:    taskItem.Status,
			TraceID:   taskItem.TraceID,
			CreatedAt: taskItem.CreatedAt,
		},
	}

	// Include logs if requested
	if containsString(includeList, "logs") {
		logKey := fmt.Sprintf("task:%s:logs", taskItem.ID)
		logs, err := client.GetRedisClient().LRange(c.Request.Context(), logKey, 0, -1).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			logrus.Errorf("Failed to get task logs: %v", err)
		} else if err == nil {
			response.Logs = logs
		}
	}

	dto.SuccessResponse(c, response)
}

// ListTasks handles simple task listing
//
//	@Summary List tasks
//	@Description Get a simple list of tasks with basic filtering via query parameters
//	@Tags Tasks
//	@Produce json
//	@Security BearerAuth
//	@Param page query int false "Page number" default(1)
//	@Param size query int false "Page size" default(20)
//	@Param task_id query string false "Filter by task ID"
//	@Param trace_id query string false "Filter by trace ID"
//	@Param group_id query string false "Filter by group ID"
//	@Param task_type query string false "Filter by task type" Enums(RestartService,FaultInjection,BuildDataset,RunAlgorithm,CollectResult,BuildImage)
//	@Param status query string false "Filter by status" Enums(Pending,Running,Completed,Error,Cancelled,Scheduled,Rescheduled)
//	@Param immediate query bool false "Filter by immediate execution"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.TaskResponse]] "Tasks retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/tasks [get]
func ListTasks(c *gin.Context) {

	// Create a basic search request from query parameters
	req := dto.TaskSearchRequest{
		AdvancedSearchRequest: dto.AdvancedSearchRequest{
			SearchRequest: dto.SearchRequest{
				Page: 1,
				Size: 20,
			},
		},
	}

	// Parse pagination from query parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := parseIntParam(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}
	if sizeStr := c.Query("size"); sizeStr != "" {
		if size, err := parseIntParam(sizeStr); err == nil && size > 0 && size <= 1000 {
			req.Size = size
		}
	}

	// Parse filters from query parameters
	if taskID := c.Query("task_id"); taskID != "" {
		req.TaskID = &taskID
	}
	if traceID := c.Query("trace_id"); traceID != "" {
		req.TraceID = &traceID
	}
	if groupID := c.Query("group_id"); groupID != "" {
		req.GroupID = &groupID
	}
	if taskType := c.Query("task_type"); taskType != "" {
		req.TaskType = &taskType
	}
	if status := c.Query("status"); status != "" {
		req.Status = &status
	}
	if immediateStr := c.Query("immediate"); immediateStr != "" {
		if immediate, err := strconv.ParseBool(immediateStr); err == nil {
			req.Immediate = &immediate
		}
	}

	// Convert to SearchRequest
	searchReq := req.ConvertToSearchRequest()

	// Add default sorting by created_at desc
	searchReq.AddSort("created_at", dto.SortDESC)

	// Validate search request
	if err := searchReq.ValidateSearchRequest(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid search parameters: "+err.Error())
		return
	}

	// Execute search using query builder
	searchResult, err := repository.ExecuteSearch(database.DB, searchReq, database.Task{})
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get task list: "+err.Error())
		return
	}

	// Convert database tasks to response DTOs
	var taskResponses []dto.TaskResponse
	for _, task := range searchResult.Items {
		taskResponse := dto.TaskResponse{
			ID:        task.ID,
			Type:      task.Type,
			Status:    task.Status,
			TraceID:   task.TraceID,
			GroupID:   task.GroupID,
			Immediate: task.Immediate,
			CreatedAt: task.CreatedAt,
			UpdatedAt: task.UpdatedAt,
		}

		taskResponses = append(taskResponses, taskResponse)
	}

	// Build final response
	response := dto.SearchResponse[dto.TaskResponse]{
		Items:      taskResponses,
		Pagination: searchResult.Pagination,
		Filters:    searchResult.Filters,
		Sort:       searchResult.Sort,
	}

	dto.SuccessResponse(c, response)
}

// SearchTasks handles complex task search with advanced filtering
//
//	@Summary Search tasks
//	@Description Search tasks with complex filtering, sorting and pagination
//	@Tags Tasks
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.TaskSearchRequest true "Task search request"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.TaskResponse]] "Tasks retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/tasks/search [post]
func SearchTasks(c *gin.Context) {

	var req dto.TaskSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Convert to SearchRequest
	searchReq := req.ConvertToSearchRequest()

	// Validate search request
	if err := searchReq.ValidateSearchRequest(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid search parameters: "+err.Error())
		return
	}

	// Execute search using query builder
	searchResult, err := repository.ExecuteSearch(database.DB, searchReq, database.Task{})
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to search tasks: "+err.Error())
		return
	}

	// Convert database tasks to response DTOs
	var taskResponses []dto.TaskResponse
	for _, task := range searchResult.Items {
		taskResponse := dto.TaskResponse{
			ID:        task.ID,
			Type:      task.Type,
			Status:    task.Status,
			TraceID:   task.TraceID,
			GroupID:   task.GroupID,
			Immediate: task.Immediate,
			CreatedAt: task.CreatedAt,
			UpdatedAt: task.UpdatedAt,
		}

		// Load logs if requested
		if searchReq.HasFilter("include") {
			includes := searchReq.GetFilter("include")
			if includes != nil && includes.Value != "" {
				if containsString(includes.Values, "logs") {
					logKey := fmt.Sprintf("task:%s:logs", task.ID)
					if logs, err := client.GetRedisClient().LRange(context.Background(), logKey, 0, -1).Result(); err == nil {
						taskResponse.Logs = logs
					}
				}
			}
		}

		taskResponses = append(taskResponses, taskResponse)
	}

	// Build final response
	response := dto.SearchResponse[dto.TaskResponse]{
		Items:      taskResponses,
		Pagination: searchResult.Pagination,
		Filters:    searchResult.Filters,
		Sort:       searchResult.Sort,
	}

	dto.SuccessResponse(c, response)
}

// GetQueuedTasks handles getting tasks in queue with pagination
//
//	@Summary Get queued tasks
//	@Description Get tasks in queue (ready and delayed) with pagination and filtering
//	@Tags Tasks
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.AdvancedSearchRequest true "Search request with pagination"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.TaskResponse]] "Queued tasks retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/tasks/queue [post]
func GetQueuedTasks(c *gin.Context) {
	var req dto.AdvancedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Convert and validate search request
	searchReq := req.ConvertAdvancedToSearch()
	if err := searchReq.ValidateSearchRequest(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid search parameters: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	redisCli := client.GetRedisClient()
	var tasks []dto.TaskResponse

	// Get tasks from ready queue (immediate execution)
	readyTasks, err := redisCli.LRange(ctx, "ready_queue", 0, -1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		logrus.Errorf("Failed to get ready queue tasks: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get ready queue tasks")
		return
	}

	for _, taskData := range readyTasks {
		var task database.Task
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			logrus.Warnf("Invalid task data: %v", err)
			continue
		}

		tasks = append(tasks, dto.TaskResponse{
			ID:        task.ID,
			Type:      task.Type,
			Status:    "Queued",
			TraceID:   task.TraceID,
			GroupID:   task.GroupID,
			Immediate: true,
			CreatedAt: task.CreatedAt,
			UpdatedAt: task.UpdatedAt,
		})
	}

	// Get tasks from delayed queue (scheduled execution)
	delayedTasksWithScore, err := redisCli.ZRangeByScoreWithScores(ctx, "delayed_queue", &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  1000, // Limit to avoid memory issues
	}).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		logrus.Errorf("Failed to get delayed queue tasks: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get delayed queue tasks")
		return
	}

	for _, z := range delayedTasksWithScore {
		taskData, ok := z.Member.(string)
		if !ok {
			continue
		}

		var task database.Task
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			logrus.Warnf("Invalid delayed task data: %v", err)
			continue
		}

		tasks = append(tasks, dto.TaskResponse{
			ID:        task.ID,
			Type:      task.Type,
			Status:    "Scheduled",
			TraceID:   task.TraceID,
			GroupID:   task.GroupID,
			Immediate: false,
			CreatedAt: task.CreatedAt,
			UpdatedAt: task.UpdatedAt,
		})
	}

	// Apply pagination
	totalTasks := int64(len(tasks))
	start := searchReq.GetOffset()
	end := start + searchReq.Size
	if start >= len(tasks) {
		tasks = []dto.TaskResponse{}
	} else if end > len(tasks) {
		tasks = tasks[start:]
	} else {
		tasks = tasks[start:end]
	}

	// Calculate pagination info
	totalPages := int((totalTasks + int64(searchReq.Size) - 1) / int64(searchReq.Size))
	pagination := &dto.PaginationInfo{
		Page:       searchReq.Page,
		Size:       searchReq.Size,
		Total:      totalTasks,
		TotalPages: totalPages,
	}

	response := dto.SearchResponse[dto.TaskResponse]{
		Items:      tasks,
		Pagination: pagination,
		Filters:    searchReq.Filters,
		Sort:       searchReq.Sort,
	}
	dto.SuccessResponse(c, response)
}

// Helper function to check if slice contains string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
