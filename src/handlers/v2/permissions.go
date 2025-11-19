package v2

import (
	"aegis/consts"
	"net/http"
	"strconv"

	"aegis/dto"
	"aegis/handlers"
	producer "aegis/service/prodcuer"

	"github.com/gin-gonic/gin"
)

// CreatePermission handles permission creation
//
//	@Summary		Create a new permission
//	@Description	Create a new permission with specified resource and action
//	@Tags			Permissions
//	@ID				create_permission
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreatePermissionReq					true	"Permission creation request"
//	@Success		201		{object}	dto.GenericResponse[dto.PermissionResp]	"Permission created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		409		{object}	dto.GenericResponse[any]				"Permission already exists"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/permissions [post]
func CreatePermission(c *gin.Context) {
	var req dto.CreatePermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.CreatePermission(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusCreated, "Permission created successfully", resp)
}

// DeletePermission handles permission deletion
//
//	@Summary		Delete permission
//	@Description	Delete a permission (soft delete by setting status to -1)
//	@Tags			Permissions
//	@ID				delete_permission
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int							true	"Permission ID"
//	@Success		204	{object}	dto.GenericResponse[any]	"Permission deleted successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]	"Invalid permission ID"
//	@Failure		401	{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]	"Permission not found"
//	@Failure		403	{object}	dto.GenericResponse[any]	"Cannot delete system permission"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/permissions/{id} [delete]
func DeletePermission(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	err = producer.DeletePermission(id)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Permission deleted successfully", nil)
}

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
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	response, err := producer.ListPermissions(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, response)
}

// SearchPermissions handles complex permission search
//
//	@Summary		Search permissions
//	@Description	Search permissions with complex filtering and sorting
//	@Tags			Permissions
//	@ID				search_permissions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.SearchPermissionReq									true	"Permission search request"
//	@Success		200		{object}	dto.GenericResponse[dto.SearchResp[dto.PermissionResp]]	"Permissions retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]								"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/permissions/search [post]
func SearchPermissions(c *gin.Context) {
	var req dto.SearchPermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Convert to SearchRequest
	searchReq := req.ConvertToSearchRequest()

	// Validate search request
	if err := searchReq.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid search parameters: "+err.Error())
		return
	}

	resp, err := producer.SearchPermissions(searchReq)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// UpdatePermission handles permission updates
//
//	@Summary		Update permission
//	@Description	Update permission information (partial update supported)
//	@Tags			Permissions
//	@ID				update_permission
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int										true	"Permission ID"
//	@Param			request	body		dto.UpdatePermissionReq					true	"Permission update request"
//	@Success		202		{object}	dto.GenericResponse[dto.PermissionResp]	"Permission updated successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid permission ID or request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]				"Permission not found"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/permissions/{id} [put]
func UpdatePermission(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	var req dto.UpdatePermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.UpdatePermission(&req, id)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Permission updated successfully", resp)
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
