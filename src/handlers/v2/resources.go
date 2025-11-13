package v2

import (
"aegis/consts"
	"aegis/dto"
	"aegis/handlers"
	producer "aegis/service/prodcuer"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetResourceDetail handles getting a single resource by ID
//
//	@Summary		Get resource by ID
//	@Description	Get detailed information about a specific resource
//	@Tags			Resources
//	@ID				get_resource_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int										true	"Resource ID"
//	@Success		200	{object}	dto.GenericResponse[dto.ResourceResp]	"Resource retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]				"Invalid resource ID"
//	@Failure		401	{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]				"Resource not found"
//	@Failure		500	{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/resources/{id} [get]
func GetResourceDetail(c *gin.Context) {
	resourceIDStr := c.Param(consts.URLPathID)
	resourceID, err := strconv.Atoi(resourceIDStr)
	if err != nil || resourceID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid resource ID")
		return
	}

	resp, err := producer.GetResourceDetail(resourceID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListResources handles listing resources with pagination and filtering
//
//	@Summary		List resources
//	@Description	Get paginated list of resources with filtering
//	@Tags			Resources
//	@ID				list_resources
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int													false	"Page number"	default(1)
//	@Param			size		query		int													false	"Page size"		default(20)
//	@Param			type		query		int													false	"Filter by resource type"
//	@Param			category	query		int													false	"Filter by resource category"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[dto.ResourceResp]]	"Resources retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]							"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/resources [get]
func ListResources(c *gin.Context) {
	var req dto.ListResourceReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListResources(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListResourcePermissions handles listing permissions by resource
//
//	@Summary		List permissions from resource
//	@Description	Get list of permissions assigned to a specific resource
//	@Tags			Resources
//	@ID				list_resource_permissions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int											true	"Resource ID"
//	@Success		200	{object}	dto.GenericResponse[[]dto.PermissionResp]	"Permissions retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]					"Invalid resource ID or request form"
//	@Failure		401	{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]					"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]					"Resource not found"
//	@Failure		500	{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/resources/{id}/permissions [get]
func ListResourcePermissions(c *gin.Context) {
	resourceIDStr := c.Param(consts.URLPathID)
	resourceID, err := strconv.Atoi(resourceIDStr)
	if err != nil || resourceID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid resource ID")
		return
	}

	resp, err := producer.ListResourcePermissions(resourceID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}
