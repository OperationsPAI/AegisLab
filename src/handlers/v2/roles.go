package v2

import (
	"net/http"
	"strconv"

	"aegis/dto"
	"aegis/service"

	"github.com/gin-gonic/gin"
)

// CreateRole handles role creation
//
//	@Summary Create a new role
//	@Description Create a new role with specified permissions
//	@Tags Roles
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.CreateRoleRequest true "Role creation request"
//	@Success 201 {object} dto.GenericResponse[dto.RoleResponse] "Role created successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request format"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 409 {object} dto.GenericResponse[any] "Role already exists"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles [post]
func CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	resp, err := service.CreateRole(&req)
	if handleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusCreated, "Role created successfully", resp)
}

// DeleteRole handles role deletion
//
//	@Summary Delete role
//	@Description Delete a role (soft delete by setting status to -1)
//	@Tags Roles
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Role ID"
//	@Success 200 {object} dto.GenericResponse[any] "Role deleted successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid role ID"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied or cannot delete system role"
//	@Failure 404 {object} dto.GenericResponse[any] "Role not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles/{id} [delete]
func DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	err = service.DeleteRole(id)
	if handleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Role deleted successfully", nil)
}

// GetRole handles getting a single role by ID
//
//	@Summary Get role by ID
//	@Description Get detailed information about a specific role
//	@Tags Roles
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Role ID"
//	@Success 200 {object} dto.GenericResponse[dto.RoleDetailResponse] "Role retrieved successfully"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid role ID"
//	@Failure 404 {object} dto.GenericResponse[any] "Role not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles/{id} [get]
func GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	resp, err := service.GetRoleDetail(id)
	if handleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListRoles handles listing roles with pagination and filtering
//
//	@Summary List roles
//	@Description Get paginated list of roles with optional filtering
//	@Tags Roles
//	@Produce json
//	@Security BearerAuth
//	@Param page query int false "Page number" default(1)
//	@Param size query int false "Page size" default(20)
//	@Param is_system query bool false "Filter by system role"
//	@Param status query int false "Filter by status"
//	@Success 200 {object} dto.GenericResponse[dto.ListRoleResponse] "Roles retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request parameters"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles [get]
func ListRoles(c *gin.Context) {
	var req dto.ListRoleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := service.ListRoles(&req)
	if handleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// SearchRoles handles complex role search
//
//	@Summary Search roles
//	@Description Search roles with complex filtering and sorting
//	@Tags Roles
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.SearchRoleRequest true "Role search request"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.RoleResponse]] "Roles retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request format or search parameters"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles/search [post]
func SearchRoles(c *gin.Context) {
	var req dto.SearchRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	searchReq := req.ConvertToSearchRequest()
	if err := searchReq.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid search parameters: "+err.Error())
		return
	}

	resp, err := service.SearchRoles(searchReq)
	if handleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// UpdateRole handles role updates
//
//	@Summary Update role
//	@Description Update role information (partial update supported)
//	@Tags Roles
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Role ID"
//	@Param request body dto.UpdateRoleRequest true "Role update request"
//	@Success 202 {object} dto.GenericResponse[dto.RoleResponse] "Role updated successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Role not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles/{id} [patch]
func UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := service.UpdateRole(&req, id)
	if handleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Role updated successfully", resp)
}
