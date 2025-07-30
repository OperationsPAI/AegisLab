package v2

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
)

// GetInjection
//
//	@Summary Get injection by ID
//	@Description Get detailed information about a specific injection
//	@Tags Injections
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Injection ID"
//	@Param include query string false "Include related data (task)"
//	@Success 200 {object} dto.GenericResponse[dto.InjectionV2Response] "Injection retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid injection ID"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Injection not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/injections/{id} [get]
func GetInjection(c *gin.Context) {
	// Check permission
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canRead, err := checker.CanReadResource(consts.ResourceFaultInjection)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}
	if !canRead {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to read injections")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid injection ID")
		return
	}

	include := c.Query("include")
	includeTask := strings.Contains(include, "task")

	// Build query with preloads
	query := database.DB.Model(&database.FaultInjectionSchedule{})
	if includeTask {
		query = query.Preload("Task")
	}

	var injection database.FaultInjectionSchedule
	if err := query.First(&injection, id).Error; err != nil {
		if strings.Contains(err.Error(), "not found") {
			dto.ErrorResponse(c, http.StatusNotFound, "Injection not found")
		} else {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get injection: "+err.Error())
		}
		return
	}

	response := dto.ToInjectionV2Response(&injection, includeTask)
	dto.SuccessResponse(c, response)
}

// ListInjections
//
//	@Summary List injections
//	@Description Get a paginated list of injections with filtering and sorting
//	@Tags Injections
//	@Produce json
//	@Security BearerAuth
//	@Param page query int false "Page number (default 1)"
//	@Param size query int false "Page size (default 20, max 100)"
//	@Param task_id query string false "Filter by task ID"
//	@Param fault_type query int false "Filter by fault type"
//	@Param status query int false "Filter by status"
//	@Param benchmark query string false "Filter by benchmark"
//	@Param search query string false "Search in injection name and description"
//	@Param sort_by query string false "Sort field (id,task_id,fault_type,status,benchmark,injection_name,created_at,updated_at)"
//	@Param sort_order query string false "Sort order (asc,desc)"
//	@Param include query string false "Include related data (task)"
//	@Success 200 {object} dto.GenericResponse[dto.InjectionSearchResponse] "Injections retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request parameters"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/injections [get]
func ListInjections(c *gin.Context) {
	// Check permission
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canRead, err := checker.CanReadResource(consts.ResourceFaultInjection)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}
	if !canRead {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to read injections")
		return
	}

	var req dto.InjectionV2ListReq
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
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// Call repository
	injections, total, err := repository.ListInjectionsV2(req.Page, req.Size, req.TaskID, req.FaultType, req.Status, req.Benchmark, req.Search)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to list injections: "+err.Error())
		return
	}

	// Convert to response
	items := make([]dto.InjectionV2Response, len(injections))
	includeTask := strings.Contains(req.Include, "task")
	for i, injection := range injections {
		items[i] = *dto.ToInjectionV2Response(&injection, includeTask)
	}

	// Create pagination info
	pagination := dto.PaginationInfo{
		Page:       req.Page,
		Size:       req.Size,
		Total:      total,
		TotalPages: int((total + int64(req.Size) - 1) / int64(req.Size)),
	}

	response := dto.InjectionSearchResponse{
		Items:      items,
		Pagination: pagination,
	}

	dto.SuccessResponse(c, response)
}

// UpdateInjection
//
//	@Summary Update injection
//	@Description Update injection information
//	@Tags Injections
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Injection ID"
//	@Param injection body dto.InjectionV2UpdateReq true "Injection update request"
//	@Success 200 {object} dto.GenericResponse[dto.InjectionV2Response] "Injection updated successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Injection not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/injections/{id} [put]
func UpdateInjection(c *gin.Context) {
	// Check permission
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canUpdate, err := checker.CanWriteResource(consts.ResourceFaultInjection)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}
	if !canUpdate {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to update injections")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid injection ID")
		return
	}

	var req dto.InjectionV2UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Check if injection exists
	_, err = repository.GetInjectionByIDV2(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			dto.ErrorResponse(c, http.StatusNotFound, "Injection not found")
		} else {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get injection: "+err.Error())
		}
		return
	}

	// Build update map
	updates := make(map[string]interface{})
	if req.TaskID != nil {
		updates["task_id"] = *req.TaskID
	}
	if req.FaultType != nil {
		updates["fault_type"] = *req.FaultType
	}
	if req.DisplayConfig != nil {
		updates["display_config"] = *req.DisplayConfig
	}
	if req.EngineConfig != nil {
		updates["engine_config"] = *req.EngineConfig
	}
	if req.PreDuration != nil {
		updates["pre_duration"] = *req.PreDuration
	}
	if req.StartTime != nil {
		updates["start_time"] = *req.StartTime
	}
	if req.EndTime != nil {
		updates["end_time"] = *req.EndTime
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Benchmark != nil {
		updates["benchmark"] = *req.Benchmark
	}
	if req.InjectionName != nil {
		updates["injection_name"] = *req.InjectionName
	}

	if len(updates) == 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "No fields to update")
		return
	}

	// Update injection
	if err := repository.UpdateInjectionV2(id, updates); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to update injection: "+err.Error())
		return
	}

	// Get updated injection
	injection, err := repository.GetInjectionByIDV2(id)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get updated injection: "+err.Error())
		return
	}

	response := dto.ToInjectionV2Response(injection, false)
	dto.SuccessResponse(c, response)
}

// DeleteInjection
//
//	@Summary Delete injection
//	@Description Soft delete an injection (sets status to -1)
//	@Tags Injections
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Injection ID"
//	@Success 200 {object} dto.GenericResponse[any] "Injection deleted successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid injection ID"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Injection not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/injections/{id} [delete]
func DeleteInjection(c *gin.Context) {
	// Check permission
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canDelete, err := checker.CanDeleteResource(consts.ResourceFaultInjection)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}
	if !canDelete {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to delete injections")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid injection ID")
		return
	}

	// Check if injection exists
	if _, err := repository.GetInjectionByIDV2(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			dto.ErrorResponse(c, http.StatusNotFound, "Injection not found")
		} else {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get injection: "+err.Error())
		}
		return
	}

	// Soft delete
	if err := repository.DeleteInjectionV2(id); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete injection: "+err.Error())
		return
	}

	dto.SuccessResponse(c, gin.H{"message": "Injection deleted successfully"})
}

// SearchInjection
//
//	@Summary Search injections
//	@Description Advanced search for injections with complex filtering
//	@Tags Injections
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param search body dto.InjectionV2SearchReq true "Search criteria"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.InjectionV2Response]] "Search results"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/injections/search [post]
func SearchInjections(c *gin.Context) {
	// Check permission
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canRead, err := checker.CanReadResource(consts.ResourceFaultInjection)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}
	if !canRead {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to read injections")
		return
	}

	var req dto.InjectionV2SearchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}

	// Call repository
	injections, total, err := repository.SearchInjectionsV2(&req)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to search injections: "+err.Error())
		return
	}

	// Convert to response
	items := make([]dto.InjectionV2Response, len(injections))
	includeTask := strings.Contains(req.Include, "task")
	for i, injection := range injections {
		items[i] = *dto.ToInjectionV2Response(&injection, includeTask)
	}

	// Create response using existing SearchResponse
	pagination := dto.PaginationInfo{
		Page:       req.Page,
		Size:       req.Size,
		Total:      total,
		TotalPages: int((total + int64(req.Size) - 1) / int64(req.Size)),
	}

	response := dto.SearchResponse[dto.InjectionV2Response]{
		Items:      items,
		Pagination: pagination,
	}

	dto.SuccessResponse(c, response)
}
