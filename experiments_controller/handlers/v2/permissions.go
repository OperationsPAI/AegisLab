package v2

import (
	"net/http"
	"strconv"

	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
)

// CreatePermission handles permission creation
// @Summary Create a new permission
// @Description Create a new permission with specified resource and action
// @Tags Permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreatePermissionRequest true "Permission creation request"
// @Success 201 {object} dto.GenericResponse[dto.PermissionResponse] "Permission created successfully"
// @Failure 400 {object} dto.GenericResponse[any] "Invalid request"
// @Failure 409 {object} dto.GenericResponse[any] "Permission already exists"
// @Failure 500 {object} dto.GenericResponse[any] "Internal server error"
// @Router /api/v2/permissions [post]
func CreatePermission(c *gin.Context) {
	var req dto.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Check if permission already exists
	if _, err := repository.GetPermissionByName(req.Name); err == nil {
		dto.ErrorResponse(c, http.StatusConflict, "Permission name already exists")
		return
	}

	permission := &database.Permission{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Action:      req.Action,
		ResourceID:  req.ResourceID,
		IsSystem:    false, // User-created permissions are not system permissions
		Status:      1,     // Active
	}

	if err := repository.CreatePermission(permission); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create permission: "+err.Error())
		return
	}

	var response dto.PermissionResponse
	response.ConvertFromPermission(permission)

	dto.JSONResponse(c, http.StatusCreated, "Permission created successfully", response)
}

// GetPermission handles getting a single permission by ID
// @Summary Get permission by ID
// @Description Get detailed information about a specific permission
// @Tags Permissions
// @Produce json
// @Security BearerAuth
// @Param id path int true "Permission ID"
// @Success 200 {object} dto.GenericResponse[dto.PermissionResponse] "Permission retrieved successfully"
// @Failure 400 {object} dto.GenericResponse[any] "Invalid permission ID"
// @Failure 404 {object} dto.GenericResponse[any] "Permission not found"
// @Failure 500 {object} dto.GenericResponse[any] "Internal server error"
// @Router /api/v2/permissions/{id} [get]
func GetPermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	permission, err := repository.GetPermissionByID(id)
	if err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Permission not found")
		return
	}

	var response dto.PermissionResponse
	response.ConvertFromPermission(permission)

	dto.SuccessResponse(c, response)
}

// ListPermissions handles listing permissions with pagination and filtering
// @Summary List permissions
// @Description Get paginated list of permissions with optional filtering
// @Tags Permissions
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(20)
// @Param action query string false "Filter by action"
// @Param resource_id query int false "Filter by resource ID"
// @Param status query int false "Filter by status"
// @Param is_system query bool false "Filter by system permission"
// @Param name query string false "Filter by permission name"
// @Success 200 {object} dto.GenericResponse[dto.PermissionListResponse] "Permissions retrieved successfully"
// @Failure 400 {object} dto.GenericResponse[any] "Invalid request parameters"
// @Failure 500 {object} dto.GenericResponse[any] "Internal server error"
// @Router /api/v2/permissions [get]
func ListPermissions(c *gin.Context) {
	var req dto.PermissionListRequest
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

	permissions, total, err := repository.ListPermissions(req.Page, req.Size, req.Action, req.ResourceID, req.Status)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve permissions: "+err.Error())
		return
	}

	var permissionResponses []dto.PermissionResponse
	for _, permission := range permissions {
		var response dto.PermissionResponse
		response.ConvertFromPermission(&permission)
		permissionResponses = append(permissionResponses, response)
	}

	pagination := dto.PaginationInfo{
		Page:       req.Page,
		Size:       req.Size,
		Total:      total,
		TotalPages: int((total + int64(req.Size) - 1) / int64(req.Size)),
	}

	response := dto.PermissionListResponse{
		Items:      permissionResponses,
		Pagination: pagination,
	}

	dto.SuccessResponse(c, response)
}

// UpdatePermission handles permission updates
// @Summary Update permission
// @Description Update permission information (partial update supported)
// @Tags Permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Permission ID"
// @Param request body dto.UpdatePermissionRequest true "Permission update request"
// @Success 200 {object} dto.GenericResponse[dto.PermissionResponse] "Permission updated successfully"
// @Failure 400 {object} dto.GenericResponse[any] "Invalid request"
// @Failure 404 {object} dto.GenericResponse[any] "Permission not found"
// @Failure 500 {object} dto.GenericResponse[any] "Internal server error"
// @Router /api/v2/permissions/{id} [put]
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

	permission, err := repository.GetPermissionByID(id)
	if err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Permission not found")
		return
	}

	// Prevent updating system permissions
	if permission.IsSystem {
		dto.ErrorResponse(c, http.StatusForbidden, "Cannot update system permissions")
		return
	}

	// Update fields if provided
	if req.DisplayName != "" {
		permission.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		permission.Description = req.Description
	}
	if req.Action != "" {
		permission.Action = req.Action
	}
	if req.ResourceID != nil {
		permission.ResourceID = *req.ResourceID
	}
	if req.Status != nil {
		permission.Status = *req.Status
	}

	if err := repository.UpdatePermission(permission); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to update permission: "+err.Error())
		return
	}

	var response dto.PermissionResponse
	response.ConvertFromPermission(permission)

	dto.SuccessResponse(c, response)
}

// DeletePermission handles permission deletion
// @Summary Delete permission
// @Description Delete a permission (soft delete by setting status to -1)
// @Tags Permissions
// @Produce json
// @Security BearerAuth
// @Param id path int true "Permission ID"
// @Success 200 {object} dto.GenericResponse[any] "Permission deleted successfully"
// @Failure 400 {object} dto.GenericResponse[any] "Invalid permission ID"
// @Failure 404 {object} dto.GenericResponse[any] "Permission not found"
// @Failure 403 {object} dto.GenericResponse[any] "Cannot delete system permission"
// @Failure 500 {object} dto.GenericResponse[any] "Internal server error"
// @Router /api/v2/permissions/{id} [delete]
func DeletePermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	permission, err := repository.GetPermissionByID(id)
	if err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Permission not found")
		return
	}

	// Prevent deleting system permissions
	if permission.IsSystem {
		dto.ErrorResponse(c, http.StatusForbidden, "Cannot delete system permissions")
		return
	}

	if err := repository.DeletePermission(id); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete permission: "+err.Error())
		return
	}

	dto.SuccessResponse(c, "Permission deleted successfully")
}

// SearchPermissions handles complex permission search
// @Summary Search permissions
// @Description Search permissions with complex filtering and sorting
// @Tags Permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.PermissionSearchRequest true "Permission search request"
// @Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.PermissionResponse]] "Permissions retrieved successfully"
// @Failure 400 {object} dto.GenericResponse[any] "Invalid request"
// @Failure 500 {object} dto.GenericResponse[any] "Internal server error"
// @Router /api/v2/permissions/search [post]
func SearchPermissions(c *gin.Context) {
	var req dto.PermissionSearchRequest
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
	searchResult, err := repository.ExecuteSearch(database.DB, searchReq, database.Permission{})
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to search permissions: "+err.Error())
		return
	}

	// Convert database permissions to response DTOs
	var permissionResponses []dto.PermissionResponse
	for _, permission := range searchResult.Items {
		var response dto.PermissionResponse
		response.ConvertFromPermission(&permission)

		// Load related data if requested
		if searchReq.HasFilter("include") {
			includes := searchReq.GetFilter("include")
			if includes != nil && includes.Value != nil {
				includeList, ok := includes.Value.([]string)
				if ok {
					for _, include := range includeList {
						switch include {
						case "roles":
							if roles, err := repository.GetPermissionRoles(permission.ID); err == nil {
								response.Roles = make([]dto.RoleResponse, len(roles))
								for i, role := range roles {
									response.Roles[i].ConvertFromRole(&role)
								}
							}
						case "resource":
							if resource, err := repository.GetResourceByID(permission.ResourceID); err == nil {
								response.Resource = &dto.ResourceResponse{
									ID:          resource.ID,
									Name:        resource.Name,
									DisplayName: resource.DisplayName,
								}
							}
						}
					}
				}
			}
		}

		permissionResponses = append(permissionResponses, response)
	}

	// Build final response
	response := dto.SearchResponse[dto.PermissionResponse]{
		Items:      permissionResponses,
		Pagination: searchResult.Pagination,
		Filters:    searchResult.Filters,
		Sort:       searchResult.Sort,
	}

	dto.SuccessResponse(c, response)
}

// GetPermissionRoles handles getting roles that have a specific permission
// @Summary Get permission roles
// @Description Get list of roles that have been assigned a specific permission
// @Tags Permissions
// @Produce json
// @Security BearerAuth
// @Param id path int true "Permission ID"
// @Success 200 {object} dto.GenericResponse[[]dto.RoleResponse] "Roles retrieved successfully"
// @Failure 400 {object} dto.GenericResponse[any] "Invalid permission ID"
// @Failure 404 {object} dto.GenericResponse[any] "Permission not found"
// @Failure 500 {object} dto.GenericResponse[any] "Internal server error"
// @Router /api/v2/permissions/{id}/roles [get]
func GetPermissionRoles(c *gin.Context) {
	idStr := c.Param("id")
	permissionID, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	// Check if permission exists
	if _, err := repository.GetPermissionByID(permissionID); err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Permission not found")
		return
	}

	// Get roles that have this permission
	roles, err := repository.GetPermissionRoles(permissionID)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get permission roles: "+err.Error())
		return
	}

	// Convert to response DTOs
	roleResponses := make([]dto.RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i].ConvertFromRole(&role)
	}

	dto.SuccessResponse(c, roleResponses)
}

// GetPermissionsByResource handles getting permissions for a specific resource
// @Summary Get permissions by resource
// @Description Get list of permissions associated with a specific resource
// @Tags Permissions
// @Produce json
// @Security BearerAuth
// @Param resource_id path int true "Resource ID"
// @Success 200 {object} dto.GenericResponse[[]dto.PermissionResponse] "Permissions retrieved successfully"
// @Failure 400 {object} dto.GenericResponse[any] "Invalid resource ID"
// @Failure 500 {object} dto.GenericResponse[any] "Internal server error"
// @Router /api/v2/permissions/resource/{resource_id} [get]
func GetPermissionsByResource(c *gin.Context) {
	resourceIDStr := c.Param("resource_id")
	resourceID, err := strconv.Atoi(resourceIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid resource ID")
		return
	}

	permissions, err := repository.GetPermissionsByResource(resourceID)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get permissions: "+err.Error())
		return
	}

	var permissionResponses []dto.PermissionResponse
	for _, permission := range permissions {
		var response dto.PermissionResponse
		response.ConvertFromPermission(&permission)
		permissionResponses = append(permissionResponses, response)
	}

	dto.SuccessResponse(c, permissionResponses)
}
