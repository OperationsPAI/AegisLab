package v2

import (
	"aegis/dto"
	"aegis/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AssignUserPermission handles direct user-permission assignment
//
//	@Summary Assign permission to user
//	@Description Assign a permission directly to a user (with optional project scope)
//	@Tags Relations
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "User ID"
//	@Param request body dto.AssignUserPermissionRequest true "User permission assignment request"
//	@Success 200 {object} dto.GenericResponse[any] "Permission assigned successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failuer 404 {object} dto.GenericResponse[any] "User or permission or project or container not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/users/{id}/permissions [post]
func AssignUserPermission(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.AssignUserPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	err = service.AssignPermissionToUser(&req, userID)
	if handleServiceError(c, err) {
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
//	@Param id path int true "User ID"
//	@Param permission_id path int true "Permission ID"
//	@Success 204 {object} dto.GenericResponse[any] "Permission removed successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid user or permission ID"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "User or permission not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/users/{id}/permissions/{permission_id} [delete]
func RemoveUserPermission(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	permissionIDStr := c.Param("permission_id")
	permissionID, err := strconv.Atoi(permissionIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid permissionID ID")
		return
	}

	err = service.RemovePermissionFromUser(userID, permissionID)
	if handleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Permission removed successfully", nil)
}

// AssignUserRole handles user-role assignment
//
//	@Summary Assign role to user
//	@Description Assign a role to a user (global role assignment)
//	@Tags Relations
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "User ID"
//	@Param request body dto.AssignRoleToUserRequest true "User role assignment request"
//	@Success 200 {object} dto.GenericResponse[any] "Role assigned successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid user ID or request format"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "User or role not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/users/{id}/roles [post]
func AssignGlobalRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.AssignRoleToUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	err = service.AssignRoleToUser(userID, req.RoleID)
	if handleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, "Role assigned successfully")
}

// RemoveGlobalRole handles user-role removal
//
//	@Summary Remove role from user
//	@Description Remove a role from a user (global role removal)
//	@Tags Relations
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "User ID"
//	@Param role_id path int true "Role ID"
//	@Success 204 {object} dto.GenericResponse[any] "Role removed successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid user or role ID"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "User or role not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/users/{id}/roles/{role_id} [delete]
func RemoveGlobalRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	roleIDStr := c.Param("role_id")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	err = service.RemoveRoleFromUser(userID, roleID)
	if handleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Role removed successfully", nil)
}

// AssignUserToProject handles user-project assignment
//
//	@Summary Assign user to project
//	@Description Assign a user to a project with a specific role
//	@Tags Users
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "User ID"
//	@Param request body dto.AssignUserToProjectRequest true "Project assignment request"
//	@Success 200 {object} dto.GenericResponse[any] "User assigned to project successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid user ID or request format"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "User or project not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/users/{id}/projects [post]
func AssignUserToProject(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.AssignUserToProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	err = service.AssignUserToProject(userID, req.ProjectID, req.RoleID)
	if handleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, "User assigned to project successfully")
}

// RemoveUserFromProject handles user-project removal
//
//	@Summary Remove user from project
//	@Description Remove a user from a project
//	@Tags Users
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "User ID"
//	@Param project_id path int true "Project ID"
//	@Success 204 {object} dto.GenericResponse[any] "User removed from project successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid user or project ID"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "User or project not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/users/{id}/projects/{project_id} [delete]
func RemoveUserFromProject(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	projectIDStr := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	err = service.RemoveUserFromProject(userID, projectID)
	if handleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "User removed from project successfully", nil)
}
