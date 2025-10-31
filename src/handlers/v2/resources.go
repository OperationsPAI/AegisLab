package v2

import (
	"aegis/dto"
	"aegis/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListResourcePermissions handles listing permissions by resource
//
//	@Summary List permissions from resource
//	@Description Get list of permissions assigned to a specific resource
//	@Tags Resources
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Resource ID"
//	@Success 200 {object} dto.GenericResponse[[]dto.PermissionResponse] "Permissions retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid resource ID"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/resources/{id}/permissions [get]
func ListResourcePermissions(c *gin.Context) {
	resourceIDStr := c.Param("id")
	resourceID, err := strconv.Atoi(resourceIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid resource ID")
		return
	}

	resp, err := service.ListResourcePermissions(resourceID)
	if handleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}
