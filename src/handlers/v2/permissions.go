package v2

import (
	"net/http"
	"strconv"

	"aegis/dto"
	"aegis/service"

	"github.com/gin-gonic/gin"
)

// CreatePermission handles permission creation
//
//	@Summary Create a new permission
//	@Description Create a new permission with specified resource and action
//	@Tags Permissions
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.CreatePermissionRequest true "Permission creation request"
//	@Success 201 {object} dto.GenericResponse[dto.PermissionResponse] "Permission created successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request format or parameters"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 409 {object} dto.GenericResponse[any] "Permission already exists"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/permissions [post]
func CreatePermission(c *gin.Context) {
	var req dto.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := service.CreatePermission(&req)
	if handleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusCreated, "Permission created successfully", resp)
}

// DeletePermission handles permission deletion
//
//	@Summary Delete permission
//	@Description Delete a permission (soft delete by setting status to -1)
//	@Tags Permissions
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Permission ID"
//	@Success 204 {object} dto.GenericResponse[any] "Permission deleted successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid permission ID"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Permission not found"
//	@Failure 403 {object} dto.GenericResponse[any] "Cannot delete system permission"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/permissions/{id} [delete]
func DeletePermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	err = service.DeletePermission(id)
	if handleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Permission deleted successfully", nil)
}

// GetPermission handles getting a single permission by ID
//
//	@Summary Get permission by ID
//	@Description Get detailed information about a specific permission
//	@Tags Permissions
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Permission ID"
//	@Success 200 {object} dto.GenericResponse[dto.PermissionDetailResponse] "Permission retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid permission ID"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Permission not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/permissions/{id} [get]
func GetPermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	resp, err := service.GetPermissionDetail(id)
	if handleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListPermissions handles listing permissions with pagination and filtering
//
//	@Summary List permissions
//	@Description Get paginated list of permissions with optional filtering
//	@Tags Permissions
//	@Produce json
//	@Security BearerAuth
//	@Param page query int false "Page number" default(1)
//	@Param size query int false "Page size" default(20)
//	@Param action query string false "Filter by action"
//	@Param is_system query bool false "Filter by system permission"
//	@Param status query int false "Filter by status"
//	@Success 200 {object} dto.GenericResponse[dto.ListPermissionResponse] "Permissions retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request format or parameters"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/permissions [get]
func ListPermissions(c *gin.Context) {
	var req dto.ListPermissionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	response, err := service.ListPermissions(&req)
	if handleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, response)
}

// SearchPermissions handles complex permission search
//
//	@Summary Search permissions
//	@Description Search permissions with complex filtering and sorting
//	@Tags Permissions
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.SearchPermissionRequest true "Permission search request"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.PermissionResponse]] "Permissions retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/permissions/search [post]
func SearchPermissions(c *gin.Context) {
	var req dto.SearchPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Convert to SearchRequest
	searchReq := req.ConvertToSearchRequest()

	// Validate search request
	if err := searchReq.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid search parameters: "+err.Error())
		return
	}

	resp, err := service.SearchPermissions(searchReq)
	if handleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// UpdatePermission handles permission updates
//
//	@Summary Update permission
//	@Description Update permission information (partial update supported)
//	@Tags Permissions
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Permission ID"
//	@Param request body dto.UpdatePermissionRequest true "Permission update request"
//	@Success 202 {object} dto.GenericResponse[dto.PermissionResponse] "Permission updated successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid permission ID or request format/parameters"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Permission not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/permissions/{id} [put]
func UpdatePermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	var req dto.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := service.UpdatePermission(&req, id)
	if handleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Permission updated successfully", resp)
}
