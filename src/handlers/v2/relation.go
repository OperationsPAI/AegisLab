package v2

import (
	"net/http"
	"strconv"

	"aegis/dto"
	"aegis/repository"

	"github.com/gin-gonic/gin"
)

// AssignUserRole handles user-role assignment
//
//	@Summary Assign role to user
//	@Description Assign a role to a user (global role assignment)
//	@Tags Relations
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.AssignUserRoleRequest true "User role assignment request"
//	@Success 200 {object} dto.GenericResponse[any] "Role assigned successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 404 {object} dto.GenericResponse[any] "User or role not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/relations/user-roles [post]
func AssignUserRole(c *gin.Context) {
	var req dto.AssignUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Check if user exists
	if _, err := repository.GetUserByID(req.UserID); err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	// Check if role exists
	if _, err := repository.GetRoleByID(req.RoleID); err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Role not found")
		return
	}

	// Assign role to user
	if err := repository.AssignRoleToUser(req.UserID, req.RoleID); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to assign role: "+err.Error())
		return
	}

	dto.SuccessResponse(c, "Role assigned successfully")
}

// RemoveUserRole handles user-role removal
//
//	@Summary Remove role from user
//	@Description Remove a role from a user (global role removal)
//	@Tags Relations
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.RemoveUserRoleRequest true "User role removal request"
//	@Success 200 {object} dto.GenericResponse[any] "Role removed successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 404 {object} dto.GenericResponse[any] "User or role not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/relations/user-roles [delete]
func RemoveUserRole(c *gin.Context) {
	var req dto.RemoveUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Remove role from user
	if err := repository.RemoveRoleFromUser(req.UserID, req.RoleID); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove role: "+err.Error())
		return
	}

	dto.SuccessResponse(c, "Role removed successfully")
}

// AssignRolePermissions handles role-permission assignment
//
//	@Summary Assign permissions to role
//	@Description Assign multiple permissions to a role
//	@Tags Relations
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.AssignRolePermissionRequest true "Role permission assignment request"
//	@Success 200 {object} dto.GenericResponse[any] "Permissions assigned successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 404 {object} dto.GenericResponse[any] "Role or permission not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/relations/role-permissions [post]
func AssignRolePermissions(c *gin.Context) {
	var req dto.AssignRolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Check if role exists
	if _, err := repository.GetRoleByID(req.RoleID); err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Role not found")
		return
	}

	// Assign permissions to role
	for _, permissionID := range req.PermissionIDs {
		if err := repository.AssignPermissionToRole(req.RoleID, permissionID); err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to assign permission: "+err.Error())
			return
		}
	}

	dto.SuccessResponse(c, "Permissions assigned successfully")
}

// RemoveRolePermissions handles role-permission removal
//
//	@Summary Remove permissions from role
//	@Description Remove multiple permissions from a role
//	@Tags Relations
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.RemoveRolePermissionRequest true "Role permission removal request"
//	@Success 200 {object} dto.GenericResponse[any] "Permissions removed successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/relations/role-permissions [delete]
func RemoveRolePermissions(c *gin.Context) {
	var req dto.RemoveRolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Remove permissions from role
	for _, permissionID := range req.PermissionIDs {
		if err := repository.RemovePermissionFromRole(req.RoleID, permissionID); err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove permission: "+err.Error())
			return
		}
	}

	dto.SuccessResponse(c, "Permissions removed successfully")
}

// AssignUserPermission handles direct user-permission assignment
//
//	@Summary Assign permission to user
//	@Description Assign a permission directly to a user (with optional project scope)
//	@Tags Relations
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.AssignUserPermissionRequest true "User permission assignment request"
//	@Success 200 {object} dto.GenericResponse[any] "Permission assigned successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 404 {object} dto.GenericResponse[any] "User or permission not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/relations/user-permissions [post]
func AssignUserPermission(c *gin.Context) {
	var req dto.AssignUserPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Check if user exists
	if _, err := repository.GetUserByID(req.UserID); err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	// Assign permission to user
	if err := repository.GrantPermissionToUser(req.UserID, req.PermissionID, req.ProjectID); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to assign permission: "+err.Error())
		return
	}

	dto.SuccessResponse(c, "Permission assigned successfully")
}

// RemoveUserPermission handles direct user-permission removal
//
//	@Summary Remove permission from user
//	@Description Remove a permission directly from a user
//	@Tags Relations
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.RemoveUserPermissionRequest true "User permission removal request"
//	@Success 200 {object} dto.GenericResponse[any] "Permission removed successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/relations/user-permissions [delete]
func RemoveUserPermission(c *gin.Context) {
	var req dto.RemoveUserPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Remove permission from user
	if err := repository.RevokePermissionFromUser(req.UserID, req.PermissionID, req.ProjectID); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove permission: "+err.Error())
		return
	}

	dto.SuccessResponse(c, "Permission removed successfully")
}

// BatchRelationOperations handles batch relationship operations
//
//	@Summary Batch relationship operations
//	@Description Perform multiple relationship operations in a single request
//	@Tags Relations
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.BatchRelationRequest true "Batch relation operations request"
//	@Success 200 {object} dto.GenericResponse[any] "Batch operations completed successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/relations/batch [post]
func BatchRelationOperations(c *gin.Context) {
	var req dto.BatchRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Process each operation
	for i, op := range req.Operations {
		var err error

		switch op.Type {
		case "user_role":
			if op.Action == "assign" {
				err = repository.AssignRoleToUser(op.SourceID, op.TargetID)
			} else if op.Action == "remove" {
				err = repository.RemoveRoleFromUser(op.SourceID, op.TargetID)
			}
		case "role_permission":
			if op.Action == "assign" {
				err = repository.AssignPermissionToRole(op.SourceID, op.TargetID)
			} else if op.Action == "remove" {
				err = repository.RemovePermissionFromRole(op.SourceID, op.TargetID)
			}
		default:
			dto.ErrorResponse(c, http.StatusBadRequest, "Unsupported relation type: "+op.Type)
			return
		}

		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed operation "+strconv.Itoa(i+1)+": "+err.Error())
			return
		}
	}

	dto.SuccessResponse(c, "Batch operations completed successfully")
}

// ListRelations handles listing relationships
//
//	@Summary List relationships
//	@Description Get paginated list of relationships with optional filtering
//	@Tags Relations
//	@Produce json
//	@Security BearerAuth
//	@Param page query int false "Page number" default(1)
//	@Param size query int false "Page size" default(20)
//	@Param type query string false "Relationship type"
//	@Param source_type query string false "Source entity type"
//	@Param target_type query string false "Target entity type"
//	@Param source_id query int false "Source entity ID"
//	@Param target_id query int false "Target entity ID"
//	@Success 200 {object} dto.GenericResponse[dto.RelationListResponse] "Relations retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request parameters"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/relations [get]
func ListRelations(c *gin.Context) {
	var req dto.RelationListRequest
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

	// Get relationships using repository
	relations, total, err := repository.GetAllRelations(req.Page, req.Size, req.Type)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get relations: "+err.Error())
		return
	}

	// Convert to response DTOs
	var relationResponses []dto.RelationResponse
	for i, relation := range relations {
		relationResponse := dto.RelationResponse{
			ID:   i + 1, // Simple ID assignment
			Type: relation.Type,
			Source: dto.RelationEntity{
				Type: relation.SourceType,
				ID:   relation.SourceID,
				Name: relation.SourceName,
			},
			Target: dto.RelationEntity{
				Type: relation.TargetType,
				ID:   relation.TargetID,
				Name: relation.TargetName,
			},
			CreatedAt: relation.CreatedAt,
			UpdatedAt: relation.CreatedAt, // Using CreatedAt as UpdatedAt placeholder
		}
		relationResponses = append(relationResponses, relationResponse)
	}

	pagination := dto.PaginationInfo{
		Page:       req.Page,
		Size:       req.Size,
		Total:      total,
		TotalPages: int((total + int64(req.Size) - 1) / int64(req.Size)),
	}

	response := dto.RelationListResponse{
		Items:      relationResponses,
		Pagination: pagination,
	}

	dto.SuccessResponse(c, response)
}

// GetRelationStatistics handles relationship statistics
//
//	@Summary Get relationship statistics
//	@Description Get statistics about all relationship types in the system
//	@Tags Relations
//	@Produce json
//	@Security BearerAuth
//	@Success 200 {object} dto.GenericResponse[dto.RelationStatisticsResponse] "Statistics retrieved successfully"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/relations/statistics [get]
func GetRelationStatistics(c *gin.Context) {
	// Get relationship statistics from repository
	stats, err := repository.GetRelationStatistics()
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get relation statistics: "+err.Error())
		return
	}

	// Convert to response format
	response := dto.RelationStatisticsResponse{
		UserRoles:            int(stats["user_roles"].(int64)),
		RolePermissions:      int(stats["role_permissions"].(int64)),
		UserPermissions:      int(stats["user_permissions"].(int64)),
		UserProjects:         int(stats["user_projects"].(int64)),
		DatasetLabels:        int(stats["dataset_labels"].(int64)),
		ContainerLabels:      int(stats["container_labels"].(int64)),
		ProjectLabels:        int(stats["project_labels"].(int64)),
		FaultInjectionLabels: int(stats["fault_injection_labels"].(int64)),
	}

	dto.SuccessResponse(c, response)
}
