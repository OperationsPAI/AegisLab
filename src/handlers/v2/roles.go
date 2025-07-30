package v2

import (
	"net/http"
	"strconv"

	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
)

// CreateRole handles role creation
//	@Summary Create a new role
//	@Description Create a new role with specified permissions
//	@Tags Roles
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.CreateRoleRequest true "Role creation request"
//	@Success 201 {object} dto.GenericResponse[dto.RoleResponse] "Role created successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 409 {object} dto.GenericResponse[any] "Role already exists"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles [post]
func CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Check if role already exists
	if _, err := repository.GetRoleByName(req.Name); err == nil {
		dto.ErrorResponse(c, http.StatusConflict, "Role name already exists")
		return
	}

	role := &database.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Type:        req.Type,
		IsSystem:    false, // User-created roles are not system roles
		Status:      1,     // Active
	}

	if err := repository.CreateRole(role); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create role: "+err.Error())
		return
	}

	var response dto.RoleResponse
	response.ConvertFromRole(role)

	dto.JSONResponse(c, http.StatusCreated, "Role created successfully", response)
}

// GetRole handles getting a single role by ID
//	@Summary Get role by ID
//	@Description Get detailed information about a specific role
//	@Tags Roles
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Role ID"
//	@Success 200 {object} dto.GenericResponse[dto.RoleResponse] "Role retrieved successfully"
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

	role, err := repository.GetRoleByID(id)
	if err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Role not found")
		return
	}

	var response dto.RoleResponse
	response.ConvertFromRole(role)

	// Load role permissions
	if permissions, err := repository.GetRolePermissions(id); err == nil {
		response.Permissions = make([]dto.PermissionResponse, len(permissions))
		for i, permission := range permissions {
			response.Permissions[i].ConvertFromPermission(&permission)
		}
	}

	// Get user count for this role
	if users, err := repository.GetRoleUsers(id); err == nil {
		response.UserCount = len(users)
	}

	dto.SuccessResponse(c, response)
}

// ListRoles handles listing roles with pagination and filtering
//	@Summary List roles
//	@Description Get paginated list of roles with optional filtering
//	@Tags Roles
//	@Produce json
//	@Security BearerAuth
//	@Param page query int false "Page number" default(1)
//	@Param size query int false "Page size" default(20)
//	@Param type query string false "Filter by role type"
//	@Param status query int false "Filter by status"
//	@Param is_system query bool false "Filter by system role"
//	@Param name query string false "Filter by role name"
//	@Success 200 {object} dto.GenericResponse[dto.RoleListResponse] "Roles retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request parameters"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles [get]
func ListRoles(c *gin.Context) {
	var req dto.RoleListRequest
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

	roles, total, err := repository.ListRoles(req.Page, req.Size, req.Type, req.Status)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve roles: "+err.Error())
		return
	}

	var roleResponses []dto.RoleResponse
	for _, role := range roles {
		var response dto.RoleResponse
		response.ConvertFromRole(&role)

		// Get user count for each role
		if users, err := repository.GetRoleUsers(role.ID); err == nil {
			response.UserCount = len(users)
		}

		roleResponses = append(roleResponses, response)
	}

	pagination := dto.PaginationInfo{
		Page:       req.Page,
		Size:       req.Size,
		Total:      total,
		TotalPages: int((total + int64(req.Size) - 1) / int64(req.Size)),
	}

	response := dto.RoleListResponse{
		Items:      roleResponses,
		Pagination: pagination,
	}

	dto.SuccessResponse(c, response)
}

// UpdateRole handles role updates
//	@Summary Update role
//	@Description Update role information (partial update supported)
//	@Tags Roles
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Role ID"
//	@Param request body dto.UpdateRoleRequest true "Role update request"
//	@Success 200 {object} dto.GenericResponse[dto.RoleResponse] "Role updated successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 404 {object} dto.GenericResponse[any] "Role not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles/{id} [put]
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

	role, err := repository.GetRoleByID(id)
	if err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Role not found")
		return
	}

	// Prevent updating system roles
	if role.IsSystem {
		dto.ErrorResponse(c, http.StatusForbidden, "Cannot update system roles")
		return
	}

	// Update fields if provided
	if req.DisplayName != "" {
		role.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.Status != nil {
		role.Status = *req.Status
	}

	if err := repository.UpdateRole(role); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to update role: "+err.Error())
		return
	}

	var response dto.RoleResponse
	response.ConvertFromRole(role)

	dto.SuccessResponse(c, response)
}

// DeleteRole handles role deletion
//	@Summary Delete role
//	@Description Delete a role (soft delete by setting status to -1)
//	@Tags Roles
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Role ID"
//	@Success 200 {object} dto.GenericResponse[any] "Role deleted successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid role ID"
//	@Failure 404 {object} dto.GenericResponse[any] "Role not found"
//	@Failure 403 {object} dto.GenericResponse[any] "Cannot delete system role"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles/{id} [delete]
func DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	role, err := repository.GetRoleByID(id)
	if err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Role not found")
		return
	}

	// Prevent deleting system roles
	if role.IsSystem {
		dto.ErrorResponse(c, http.StatusForbidden, "Cannot delete system roles")
		return
	}

	if err := repository.DeleteRole(id); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete role: "+err.Error())
		return
	}

	dto.SuccessResponse(c, "Role deleted successfully")
}

// SearchRoles handles complex role search
//	@Summary Search roles
//	@Description Search roles with complex filtering and sorting
//	@Tags Roles
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.RoleSearchRequest true "Role search request"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.RoleResponse]] "Roles retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles/search [post]
func SearchRoles(c *gin.Context) {
	var req dto.RoleSearchRequest
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
	searchResult, err := repository.ExecuteSearch(database.DB, searchReq, database.Role{})
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to search roles: "+err.Error())
		return
	}

	// Convert database roles to response DTOs
	var roleResponses []dto.RoleResponse
	for _, role := range searchResult.Items {
		var response dto.RoleResponse
		response.ConvertFromRole(&role)

		// Load related data if requested
		if searchReq.HasFilter("include") {
			includes := searchReq.GetFilter("include")
			if includes != nil && includes.Value != nil {
				includeList, ok := includes.Value.([]string)
				if ok {
					for _, include := range includeList {
						switch include {
						case "permissions":
							if permissions, err := repository.GetRolePermissions(role.ID); err == nil {
								response.Permissions = make([]dto.PermissionResponse, len(permissions))
								for i, permission := range permissions {
									response.Permissions[i].ConvertFromPermission(&permission)
								}
							}
						case "users":
							if users, err := repository.GetRoleUsers(role.ID); err == nil {
								response.UserCount = len(users)
							}
						}
					}
				}
			}
		}

		roleResponses = append(roleResponses, response)
	}

	// Build final response
	response := dto.SearchResponse[dto.RoleResponse]{
		Items:      roleResponses,
		Pagination: searchResult.Pagination,
		Filters:    searchResult.Filters,
		Sort:       searchResult.Sort,
	}

	dto.SuccessResponse(c, response)
}

// AssignPermissionsToRole handles permission assignment to role
//	@Summary Assign permissions to role
//	@Description Assign multiple permissions to a role
//	@Tags Roles
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Role ID"
//	@Param request body dto.AssignPermissionToRoleRequest true "Permission assignment request"
//	@Success 200 {object} dto.GenericResponse[any] "Permissions assigned successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 404 {object} dto.GenericResponse[any] "Role not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles/{id}/permissions [post]
func AssignPermissionsToRole(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	var req dto.AssignPermissionToRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Check if role exists
	if _, err := repository.GetRoleByID(roleID); err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Role not found")
		return
	}

	// Assign permissions to role
	for _, permissionID := range req.PermissionIDs {
		if err := repository.AssignPermissionToRole(roleID, permissionID); err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to assign permission: "+err.Error())
			return
		}
	}

	dto.SuccessResponse(c, "Permissions assigned successfully")
}

// RemovePermissionsFromRole handles permission removal from role
//	@Summary Remove permissions from role
//	@Description Remove multiple permissions from a role
//	@Tags Roles
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Role ID"
//	@Param request body dto.RemovePermissionFromRoleRequest true "Permission removal request"
//	@Success 200 {object} dto.GenericResponse[any] "Permissions removed successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 404 {object} dto.GenericResponse[any] "Role not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles/{id}/permissions [delete]
func RemovePermissionsFromRole(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	var req dto.RemovePermissionFromRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Remove permissions from role
	for _, permissionID := range req.PermissionIDs {
		if err := repository.RemovePermissionFromRole(roleID, permissionID); err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove permission: "+err.Error())
			return
		}
	}

	dto.SuccessResponse(c, "Permissions removed successfully")
}

// GetRoleUsers handles getting users assigned to a role
//	@Summary Get role users
//	@Description Get list of users assigned to a specific role
//	@Tags Roles
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Role ID"
//	@Success 200 {object} dto.GenericResponse[[]dto.UserResponse] "Users retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid role ID"
//	@Failure 404 {object} dto.GenericResponse[any] "Role not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles/{id}/users [get]
func GetRoleUsers(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	// Check if role exists
	if _, err := repository.GetRoleByID(roleID); err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Role not found")
		return
	}

	users, err := repository.GetRoleUsers(roleID)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get role users: "+err.Error())
		return
	}

	var userResponses []dto.UserResponse
	for _, user := range users {
		var response dto.UserResponse
		response.ConvertFromUser(&user)
		userResponses = append(userResponses, response)
	}

	dto.SuccessResponse(c, userResponses)
}
