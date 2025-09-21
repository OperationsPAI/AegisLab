package v2

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"aegis/dto"
	"aegis/middleware"
	"aegis/repository"

	"github.com/gin-gonic/gin"
)

// GetHealth handles system health check
//
//	@Summary System health check
//	@Description Get system health status and service information
//	@Tags System
//	@Produce json
//	@Success 200 {object} dto.GenericResponse[dto.HealthCheckResponse] "Health check successful"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/health [get]
func GetHealth(c *gin.Context) {
	// Mock service health information
	services := map[string]dto.ServiceInfo{
		"database": {
			Status:       "healthy",
			LastChecked:  time.Now(),
			ResponseTime: "5ms",
		},
		"redis": {
			Status:       "healthy",
			LastChecked:  time.Now(),
			ResponseTime: "2ms",
		},
		"kubernetes": {
			Status:       "healthy",
			LastChecked:  time.Now(),
			ResponseTime: "10ms",
		},
	}

	response := dto.HealthCheckResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",                                              // Should be from build info
		Uptime:    time.Since(time.Now().Add(-72 * time.Hour)).String(), // Mock uptime
		Services:  services,
	}

	dto.SuccessResponse(c, response)
}

// GetStatistics handles system statistics
//
//	@Summary Get system statistics
//	@Description Get comprehensive system statistics and metrics
//	@Tags System
//	@Produce json
//	@Success 200 {object} dto.GenericResponse[dto.SystemStatisticsResponse] "Statistics retrieved successfully"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/statistics [get]
func GetStatistics(c *gin.Context) {
	var response dto.SystemStatisticsResponse

	// Get real system memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Get real user statistics
	userStats := dto.UserStatistics{}
	if allUsers, _, err := repository.ListUsers(1, 10000, nil); err == nil {
		userStats.Total = len(allUsers)
		for _, user := range allUsers {
			if user.IsActive {
				userStats.Active++
			} else {
				userStats.Inactive++
			}
			// Check if user was created today
			if user.CreatedAt.After(time.Now().Add(-24 * time.Hour)) {
				userStats.NewToday++
			}
			// Check if user was created this week
			if user.CreatedAt.After(time.Now().Add(-7 * 24 * time.Hour)) {
				userStats.NewThisWeek++
			}
		}
	}

	// Get real role statistics - using available repository functions
	roleStats := dto.ProjectStatistics{} // Reusing structure for roles
	if allRoles, _, err := repository.ListRoles(1, 10000, "", nil); err == nil {
		roleStats.Total = len(allRoles)
		for _, role := range allRoles {
			if role.Status == 1 {
				roleStats.Active++
			} else {
				roleStats.Inactive++
			}
			if role.CreatedAt.After(time.Now().Add(-24 * time.Hour)) {
				roleStats.NewToday++
			}
		}
	}

	// Get real permission statistics
	permStats := dto.TaskStatistics{} // Reusing structure for permissions
	if allPerms, _, err := repository.ListPermissions(1, 10000, "", nil, nil); err == nil {
		permStats.Total = len(allPerms)
		for _, perm := range allPerms {
			if perm.Status == 1 {
				permStats.Completed++ // Using as "active"
			} else {
				permStats.Failed++ // Using as "inactive"
			}
		}
	}

	// Get container statistics
	containerStats, err := repository.GetContainerStatistics()
	if err != nil {
		// Log error but continue with default values
		containerStats = map[string]int64{"total": 0, "active": 0, "deleted": 0}
	}

	// Get dataset statistics
	datasetStats, err := repository.GetDatasetStatistics()
	if err != nil {
		// Log error but continue with default values
		datasetStats = map[string]int64{"total": 0, "active": 0, "deleted": 0}
	}

	// Injection statistics are now handled in the response section using GetInjectionDetailedStats

	// Get execution statistics
	executionStats, err := repository.GetExecutionStatistics()
	if err != nil {
		executionStats = map[string]int64{"total": 0, "pending": 0, "running": 0, "completed": 0, "failed": 0}
	}

	// Get task statistics
	taskStats, err := repository.GetTaskStatistics()
	if err != nil {
		taskStats = map[string]int64{"total": 0}
	}

	// Get project statistics
	projectStats, err := repository.GetProjectStatistics()
	if err != nil {
		projectStats = map[string]int64{"total": 0, "active": 0, "inactive": 0, "new_today": 0}
	}

	response = dto.SystemStatisticsResponse{
		Users: userStats,
		Projects: dto.ProjectStatistics{
			Total:    int(projectStats["total"]),
			Active:   int(projectStats["active"]),
			Inactive: int(projectStats["inactive"]),
			NewToday: int(projectStats["new_today"]),
		},
		Tasks: dto.TaskStatistics{
			Total:     int(taskStats["total"]),
			Pending:   int(taskStats["pending"]),
			Running:   int(taskStats["running"]),
			Completed: int(taskStats["completed"]),
			Failed:    int(taskStats["failed"]),
		},
		Containers: dto.ContainerStatistics{
			Total:   int(containerStats["total"]),
			Active:  int(containerStats["active"]),
			Deleted: int(containerStats["deleted"]),
		},
		Datasets: func() dto.DatasetStatistics {
			totalSize, err := repository.GetDatasetTotalSize()
			if err != nil {
				totalSize = 0
			}
			return dto.DatasetStatistics{
				Total:     int(datasetStats["total"]),
				Public:    int(datasetStats["active"]),
				Private:   int(datasetStats["deleted"]),
				TotalSize: totalSize,
			}
		}(),
		Injections: func() dto.InjectionStatistics {
			detailedStats, err := repository.GetInjectionDetailedStats()
			if err != nil {
				detailedStats = map[string]int64{"total": 0, "scheduled": 0, "running": 0, "completed": 0, "failed": 0}
			}
			return dto.InjectionStatistics{
				Total:     int(detailedStats["total"]),
				Scheduled: int(detailedStats["scheduled"]),
				Running:   int(detailedStats["running"]),
				Completed: int(detailedStats["completed"]),
				Failed:    int(detailedStats["failed"]),
			}
		}(),
		Executions: dto.ExecutionStatistics{
			Total:      int(executionStats["total"]),
			Successful: int(executionStats["completed"]),
			Failed:     int(executionStats["failed"]),
		},
		GeneratedAt: time.Now(),
	}

	dto.SuccessResponse(c, response)
}

// ListAuditLogs handles audit log listing
//
//	@Summary List audit logs
//	@Description Get paginated list of audit logs with optional filtering
//	@Tags System
//	@Produce json
//	@Param page query int false "Page number" default(1)
//	@Param size query int false "Page size" default(20)
//	@Param user_id query int false "Filter by user ID"
//	@Param action query string false "Filter by action"
//	@Param resource query string false "Filter by resource"
//	@Param success query bool false "Filter by success status"
//	@Param start_date query string false "Filter from date (YYYY-MM-DD)"
//	@Param end_date query string false "Filter to date (YYYY-MM-DD)"
//	@Success 200 {object} dto.GenericResponse[dto.AuditLogListResponse] "Audit logs retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request parameters"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/audit [get]
func ListAuditLogs(c *gin.Context) {
	var req dto.AuditLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}

	// Parse optional filters
	var startTime, endTime *time.Time
	if !req.StartDate.IsZero() {
		startTime = &req.StartDate
	}
	if !req.EndDate.IsZero() {
		endTime = &req.EndDate
	}

	// Convert Success boolean to status string
	var status string
	if req.Success != nil {
		if *req.Success {
			status = "SUCCESS"
		} else {
			status = "FAILED"
		}
	}

	// Get audit logs from repository
	logs, total, err := repository.GetAuditLogs(
		req.Page, req.Size,
		req.UserID, req.Action, req.Resource, status,
		startTime, endTime,
	)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get audit logs: "+err.Error())
		return
	}

	// Convert to response DTOs
	var auditResponses []dto.AuditLogResponse
	for _, log := range logs {
		response := dto.AuditLogResponse{
			ID:         log.ID,
			UserID:     0, // Handle nullable user ID
			Username:   log.Username,
			Action:     log.Action,
			Resource:   log.Resource,
			ResourceID: fmt.Sprintf("%v", log.ResourceID),
			Details:    log.Details,
			IPAddress:  log.IPAddress,
			UserAgent:  log.UserAgent,
			Success:    log.Status == "SUCCESS",
			Error:      log.ErrorMsg,
			Timestamp:  log.CreatedAt,
		}

		if log.UserID != nil {
			response.UserID = *log.UserID
		}

		auditResponses = append(auditResponses, response)
	}

	pagination := dto.PaginationInfo{
		Page:       req.Page,
		Size:       req.Size,
		Total:      total,
		TotalPages: int((total + int64(req.Size) - 1) / int64(req.Size)),
	}

	response := dto.AuditLogListResponse{
		Items:      auditResponses,
		Pagination: pagination,
	}

	dto.SuccessResponse(c, response)
}

// GetMetrics handles monitoring metrics query
//
//	@Summary Get monitoring metrics
//	@Description Query monitoring metrics for system performance
//	@Tags System
//	@Accept json
//	@Produce json
//	@Param request body dto.MonitoringQueryRequest true "Metrics query request"
//	@Success 200 {object} dto.GenericResponse[dto.MonitoringMetricsResponse] "Metrics retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/monitor/metrics [post]
func GetMetrics(c *gin.Context) {
	var req dto.MonitoringQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// TODO: Implement real metrics querying from monitoring system
	// For now, return mock data
	metrics := map[string]dto.MetricValue{
		"cpu_usage": {
			Value:     25.5,
			Timestamp: time.Now(),
			Unit:      "percent",
		},
		"memory_usage": {
			Value:     60.2,
			Timestamp: time.Now(),
			Unit:      "percent",
		},
		"disk_usage": {
			Value:     45.8,
			Timestamp: time.Now(),
			Unit:      "percent",
		},
		"active_connections": {
			Value:     142,
			Timestamp: time.Now(),
			Unit:      "count",
		},
	}

	labels := map[string]string{
		"instance": "rcabench-01",
		"version":  "1.0.0",
	}

	response := dto.MonitoringMetricsResponse{
		Timestamp: time.Now(),
		Metrics:   metrics,
		Labels:    labels,
	}

	dto.SuccessResponse(c, response)
}

// GetSystemInfo handles basic system information
//
//	@Summary Get system information
//	@Description Get basic system information and status
//	@Tags System
//	@Produce json
//	@Success 200 {object} dto.GenericResponse[dto.SystemInfo] "System info retrieved successfully"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/monitor/info [get]
func GetSystemInfo(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	info := dto.SystemInfo{
		CPUUsage:    25.5, // Mock value - should get from system
		MemoryUsage: float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		DiskUsage:   45.8,            // Mock value - should get from system
		LoadAverage: "1.2, 1.5, 1.8", // Mock value - should get from system
	}

	dto.SuccessResponse(c, info)
}

// CreateAuditLog handles creating new audit log entries
//
//	@Summary Create audit log
//	@Description Create a new audit log entry
//	@Tags System
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param audit_log body dto.AuditLogRequest true "Audit log data"
//	@Success 201 {object} dto.GenericResponse[dto.AuditLogResponse] "Audit log created successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/audit [post]
func CreateAuditLog(c *gin.Context) {
	var req dto.AuditLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Get user info from context
	userID, _ := middleware.GetCurrentUserID(c)
	username, _ := middleware.GetCurrentUsername(c)

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Create audit log entry
	auditLog := &repository.AuditLog{
		UserID:     &userID,
		Username:   username,
		Action:     req.Action,
		Resource:   req.Resource,
		ResourceID: req.ResourceID,
		Details:    req.Details,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:     "SUCCESS",
	}

	err := repository.CreateAuditLog(auditLog)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create audit log: "+err.Error())
		return
	}

	// Convert to response DTO
	response := dto.AuditLogResponse{
		ID:         auditLog.ID,
		UserID:     userID,
		Username:   username,
		Action:     req.Action,
		Resource:   req.Resource,
		ResourceID: fmt.Sprintf("%v", req.ResourceID),
		Details:    req.Details,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Success:    true,
		Timestamp:  auditLog.CreatedAt,
	}

	dto.SuccessResponse(c, response)
}

// GetAuditLog handles single audit log retrieval
//
//	@Summary Get audit log by ID
//	@Description Get a specific audit log entry by ID
//	@Tags System
//	@Produce json
//	@Param id path int true "Audit log ID"
//	@Success 200 {object} dto.GenericResponse[dto.AuditLogResponse] "Audit log retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid ID"
//	@Failure 404 {object} dto.GenericResponse[any] "Audit log not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/audit/{id} [get]
func GetAuditLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid audit log ID")
		return
	}

	// TODO: Implement real audit log retrieval from repository
	// For now, return mock data
	if id == 1 {
		log := dto.AuditLogResponse{
			ID:         1,
			UserID:     1,
			Username:   "admin",
			Action:     "CREATE_USER",
			Resource:   "users",
			ResourceID: "123",
			Details:    `{"username":"newuser"}`,
			IPAddress:  "192.168.1.100",
			UserAgent:  "Mozilla/5.0...",
			Success:    true,
			Timestamp:  time.Now().Add(-1 * time.Hour),
		}
		dto.SuccessResponse(c, log)
		return
	}

	dto.ErrorResponse(c, http.StatusNotFound, "Audit log not found")
}
