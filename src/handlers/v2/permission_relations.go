package v2

import (
	"aegis/dto"
	"aegis/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListPermissionRoles handles listing roles assigned to a permission
//
//	@Summary List roles from permission
//	@Description Get list of roles assigned to a specific permission
//	@Tags Permissions
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Permission ID"
//	@Success 200 {object} dto.GenericResponse[[]dto.RoleResponse] "Roles retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid permission ID"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Permission not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/permissions/{id}/roles [get]
func ListPermissionRoles(c *gin.Context) {
	idStr := c.Param("id")
	permissionID, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	resp, err := service.ListPermissionRoles(permissionID)
	if handleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}
