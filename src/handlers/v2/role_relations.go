package v2

import (
	"aegis/dto"
	"aegis/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AssignPermissionsToRole handles permission assignment to role
//
//	@Summary Assign permissions to role
//	@Description Assign multiple permissions to a role
//	@Tags Roles
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Role ID"
//	@Param request body dto.AssignPermissionToRoleRequest true "Permission assignment request"
//	@Success 200 {object} dto.GenericResponse[any] "Permissions assigned successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid role ID or request format"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
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

	err = service.AssginPermissionsToRole(roleID, req.PermissionIDs)
	if handleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, "Permissions assigned successfully")
}

// RemovePermissionsFromRole handles permission removal from role
//
//	@Summary Remove permissions from role
//	@Description Remove multiple permissions from a role
//	@Tags Roles
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Role ID"
//	@Param request body dto.RemovePermissionFromRoleRequest true "Permission removal request"
//	@Success 200 {object} dto.GenericResponse[any] "Permissions removed successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid role ID or request format"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
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

	err = service.RemovePermissionsFromRole(roleID, req.PermissionIDs)
	if handleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, "Permissions removed successfully")
}

// ListUsersFromRole handles listing users assigned to a role
//
//	@Summary List users from role
//	@Description Get list of users assigned to a specific role
//	@Tags Roles
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Role ID"
//	@Success 200 {object} dto.GenericResponse[[]dto.UserResponse] "Users retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid role ID"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Role not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/roles/{id}/users [get]
func ListUsersFromRole(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	userResponses, err := service.ListUsersFromRole(roleID)
	if handleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, userResponses)
}
