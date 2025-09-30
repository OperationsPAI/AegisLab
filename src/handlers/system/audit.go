package system

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"aegis/dto"
	"aegis/middleware"
	"aegis/repository"

	"github.com/gin-gonic/gin"
)

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
//	@Router /system/audit [post]
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
//	@Router /system/audit/{id} [get]
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
//	@Router /system/audit [get]
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
