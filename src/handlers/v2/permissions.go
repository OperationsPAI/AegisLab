package v2

import (
	"aegis/consts"
	"net/http"
	"strconv"

	"aegis/dto"
	"aegis/handlers"
	producer "aegis/service/producer"

	"github.com/gin-gonic/gin"
)

// GetPermission handles getting a single permission by ID
//
//	@Summary		Get permission by ID
//	@Description	Get detailed information about a specific permission
//	@Tags			Permissions
//	@ID				get_permission_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int												true	"Permission ID"
//	@Success		200	{object}	dto.GenericResponse[dto.PermissionDetailResp]	"Permission retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]						"Invalid permission ID"
//	@Failure		401	{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]						"Permission not found"
//	@Failure		500	{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/permissions/{id} [get]
func GetPermission(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	resp, err := producer.GetPermissionDetail(id)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListPermissions handles listing permissions with pagination and filtering
//
//	@Summary		List permissions
//	@Description	Get paginated list of permissions with optional filtering
//	@Tags			Permissions
//	@ID				list_permissions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int										false	"Page number"	default(1)
//	@Param			size		query		int										false	"Page size"		default(20)
//	@Param			action		query		string									false	"Filter by action"
//	@Param			is_system	query		bool									false	"Filter by system permission"
//	@Param			status		query		consts.StatusType						false	"Filter by status"
//	@Success		200			{object}	dto.GenericResponse[dto.PermissionResp]	"Permissions retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]				"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/permissions [get]
func ListPermissions(c *gin.Context) {
	var req dto.ListPermissionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	response, err := producer.ListPermissions(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, response)
}

// ===================== Role-Permission API =====================

// ListRolesFromPermission handles listing roles assigned to a permission
//
//	@Summary		List roles from permission
//	@Description	Get list of roles assigned to a specific permission
//	@Tags			Permissions
//	@ID				list_roles_with_permission
//	@Produce		json
//	@Security		BearerAuth
//	@Param			permission_id	path		int									true	"Permission ID"
//	@Success		200				{object}	dto.GenericResponse[[]dto.RoleResp]	"Roles retrieved successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]			"Invalid permission ID"
//	@Failure		401				{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]			"Permission not found"
//	@Failure		500				{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/permissions/{permission_id}/roles [get]
//	@x-api-type		{"sdk":"true"}
func ListRolesFromPermission(c *gin.Context) {
	permissionIDStr := c.Param(consts.URLPathPermissionID)
	permissionID, err := strconv.Atoi(permissionIDStr)
	if err != nil || permissionID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	resp, err := producer.ListRolesFromPermission(permissionID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}
