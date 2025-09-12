package v2

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/LGU-SE-Internal/rcabench/utils"
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

// CreateInjection
//
//	@Summary Create injections
//	@Description Create one or multiple injection records with automatic labeling based on task_id
//	@Tags Injections
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param injections body dto.InjectionV2CreateReq true "Injection creation request"
//	@Success 200 {object} dto.GenericResponse[dto.InjectionV2CreateResponse] "Injections created successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/injections [post]
func CreateInjection(c *gin.Context) {
	var req dto.InjectionV2CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	// Create injections
	createdInjections, failedItems, err := repository.CreateInjectionsV2(req.Injections)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create injections: "+err.Error())
		return
	}

	// Convert to response DTOs
	createdItems := make([]dto.InjectionV2Response, len(createdInjections))
	for i, injection := range createdInjections {
		createdItems[i] = *dto.ToInjectionV2Response(&injection, false)
	}

	// Build response message
	message := fmt.Sprintf("Successfully created %d injection(s)", len(createdInjections))
	if len(failedItems) > 0 {
		message += fmt.Sprintf(", %d failed", len(failedItems))
	}

	response := dto.InjectionV2CreateResponse{
		CreatedCount: len(createdInjections),
		CreatedItems: createdItems,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
		Message:      message,
	}

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
//	@Param tags query []string false "Filter by tags (array of tag values)"
//	@Param sort_by query string false "Sort field (id,task_id,fault_type,status,benchmark,injection_name,created_at,updated_at)"
//	@Param sort_order query string false "Sort order (asc,desc)"
//	@Param include query string false "Include related data (task)"
//	@Success 200 {object} dto.GenericResponse[dto.InjectionSearchResponse] "Injections retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request parameters"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/injections [get]
func ListInjections(c *gin.Context) {
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
	injections, total, err := repository.ListInjectionsV2(req.Page, req.Size, req.TaskID, req.FaultType, req.Status, req.Benchmark, req.Search, req.Tags)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to list injections: "+err.Error())
		return
	}

	includeTask := strings.Contains(req.Include, "task")
	items, err := toInjectionV2ResponsesWithLabels(injections, includeTask)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to convert injections: "+err.Error())
		return
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

// BatchDeleteInjections
//
//	@Summary Batch delete injections
//	@Description Batch delete injections by IDs or labels with cascading deletion of related records
//	@Tags Injections
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param batch_delete body dto.InjectionV2BatchDeleteReq true "Batch delete request"
//	@Success 200 {object} dto.GenericResponse[dto.InjectionV2BatchDeleteResponse] "Injections deleted successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/injections/batch-delete [post]
func BatchDeleteInjections(c *gin.Context) {
	var req dto.InjectionV2BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	var response dto.InjectionV2BatchDeleteResponse
	var err error

	// Delete by IDs or by labels
	if len(req.IDs) > 0 {
		response, err = repository.BatchDeleteInjectionsV2(req.IDs)
	} else {
		response, err = repository.BatchDeleteInjectionsByLabelsV2(req.Labels)
	}

	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete injections: "+err.Error())
		return
	}

	dto.SuccessResponse(c, response)
}

// SearchInjection
//
//	@Summary Search injections
//	@Description Advanced search for injections with complex filtering including custom labels
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
	var req dto.InjectionV2SearchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	withPagination := req.Page != nil && req.Size != nil
	if withPagination {
		if *req.Page == 0 {
			req.Page = utils.IntPtr(1)
		}
		if *req.Size == 0 {
			req.Size = utils.IntPtr(20)
		}
	}

	// Call repository
	injections, total, err := repository.SearchInjectionsV2(&req)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to search injections: "+err.Error())
		return
	}

	var items []dto.InjectionV2Response
	if req.IncludeLabels {
		items, err = toInjectionV2ResponsesWithLabels(injections, req.IncludeTask)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to convert injections: "+err.Error())
			return
		}
	} else {
		for _, injection := range injections {
			items = append(items, *dto.ToInjectionV2Response(&injection, req.IncludeTask))
		}
	}

	response := dto.SearchResponse[dto.InjectionV2Response]{
		Items: items,
	}
	if withPagination {
		response.Pagination = &dto.PaginationInfo{
			Page:       *req.Page,
			Size:       *req.Size,
			Total:      total,
			TotalPages: int((total + int64(*req.Size) - 1) / int64(*req.Size)),
		}
	}

	dto.SuccessResponse(c, response)
}

// ManageInjectionTags manages injection tags
//
//	@Summary Manage injection tags
//	@Description Add or remove tags for an injection
//	@Tags Injections
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param name path string true "Injection Name"
//	@Param manage body dto.InjectionV2LabelManageReq true "Tag management request"
//	@Success 200 {object} dto.GenericResponse[dto.InjectionV2Response] "Tags managed successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Injection not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/injections/{name}/tags [patch]
func ManageInjectionTags(c *gin.Context) {
	injectionName := c.Param("name")
	if injectionName == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "Injection name is required")
		return
	}

	var req dto.InjectionV2LabelManageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Check if injection exists
	injection, err := repository.GetInjectionByNameV2(injectionName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			dto.ErrorResponse(c, http.StatusNotFound, "Injection not found")
		} else {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get injection: "+err.Error())
		}
		return
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Remove tags
	for _, tagValue := range req.RemoveTags {
		if err := repository.RemoveTagFromInjection(injection.ID, tagValue); err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove tag: "+err.Error())
			return
		}
	}

	// Add tags
	for _, tagValue := range req.AddTags {
		if err := repository.AddTagToInjection(injection.ID, tagValue); err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to add tag: "+err.Error())
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
		return
	}

	// Return updated injection with labels
	response := dto.ToInjectionV2Response(injection, false)

	// Load labels
	labels, err := repository.GetInjectionLabels(injection.ID)
	if err == nil {
		response.Labels = labels
	}

	dto.SuccessResponse(c, response)
}

// ManageInjectionCustomLabels manages injection custom labels (key-value pairs)
//
//	@Summary Manage injection custom labels
//	@Description Add or remove custom labels (key-value pairs) for an injection
//	@Tags Injections
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param name path string true "Injection Name"
//	@Param manage body dto.InjectionV2CustomLabelManageReq true "Custom label management request"
//	@Success 200 {object} dto.GenericResponse[dto.InjectionV2Response] "Custom labels managed successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Injection not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/injections/{name}/labels [patch]
func ManageInjectionCustomLabels(c *gin.Context) {
	injectionName := c.Param("name")
	if injectionName == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "Injection name is required")
		return
	}

	var req dto.InjectionV2CustomLabelManageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Validate request
	if len(req.AddLabels) == 0 && len(req.RemoveLabels) == 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "At least one operation (add or remove) must be specified")
		return
	}

	// Check if injection exists
	injection, err := repository.GetInjectionByNameV2(injectionName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			dto.ErrorResponse(c, http.StatusNotFound, "Injection not found")
		} else {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get injection: "+err.Error())
		}
		return
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Remove custom labels by key
	for _, key := range req.RemoveLabels {
		if err := repository.RemoveCustomLabelFromInjection(injection.ID, key); err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove custom label with key '"+key+"': "+err.Error())
			return
		}
	}

	// Add custom labels with override behavior
	for _, labelItem := range req.AddLabels {
		if err := repository.AddCustomLabelToInjectionWithOverride(injection.ID, labelItem.Key, labelItem.Value); err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to add custom label '"+labelItem.Key+"': "+err.Error())
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
		return
	}

	// Return updated injection with labels
	response := dto.ToInjectionV2Response(injection, false)

	// Load labels
	labels, err := repository.GetInjectionLabels(injection.ID)
	if err == nil {
		response.Labels = labels
	}

	dto.SuccessResponse(c, response)
}

func toInjectionV2ResponsesWithLabels(injections []database.FaultInjectionSchedule, includeTask bool) ([]dto.InjectionV2Response, error) {
	injectionIDs := make([]int, len(injections))
	for i, injection := range injections {
		injectionIDs[i] = injection.ID
	}

	// Batch load labels using optimized method (SEARCH)
	labelsMap, err := repository.GetInjectionLabelsMap(injectionIDs)
	if err != nil {
		return nil, fmt.Errorf("Failed to load injection labels: %v", err)
	}

	// Convert to response
	items := make([]dto.InjectionV2Response, len(injections))
	for i, injection := range injections {
		labels := labelsMap[injection.ID]
		items[i] = *dto.ToInjectionV2ResponseWithLabels(&injection, includeTask, labels)
	}

	return items, nil
}
