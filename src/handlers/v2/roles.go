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

// CreateRole handles role creation
//
//	@Summary		Create a new role
//	@Description	Create a new role with specified permissions
//	@Tags			Roles
//	@ID				create_role
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateRoleReq					true	"Role creation request"
//	@Success		201		{object}	dto.GenericResponse[dto.RoleResp]	"Role created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request format"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		409		{object}	dto.GenericResponse[any]			"Role already exists"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/roles [post]
func CreateRole(c *gin.Context) {
	var req dto.CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	resp, err := producer.CreateRole(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusCreated, "Role created successfully", resp)
}

// DeleteRole handles role deletion
//
//	@Summary		Delete role
//	@Description	Delete a role (soft delete by setting status to -1)
//	@Tags			Roles
//	@ID				delete_role
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int							true	"Role ID"
//	@Success		200	{object}	dto.GenericResponse[any]	"Role deleted successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]	"Invalid role ID"
//	@Failure		401	{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]	"Permission denied or cannot delete system role"
//	@Failure		404	{object}	dto.GenericResponse[any]	"Role not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/roles/{id} [delete]
func DeleteRole(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	err = producer.DeleteRole(id)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Role deleted successfully", nil)
}

// GetRole handles getting a single role by ID
//
//	@Summary		Get role by ID
//	@Description	Get detailed information about a specific role
//	@Tags			Roles
//	@ID				get_role_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int										true	"Role ID"
//	@Success		200	{object}	dto.GenericResponse[dto.RoleDetailResp]	"Role retrieved successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		400	{object}	dto.GenericResponse[any]				"Invalid role ID"
//	@Failure		404	{object}	dto.GenericResponse[any]				"Role not found"
//	@Failure		500	{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/roles/{id} [get]
func GetRole(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	resp, err := producer.GetRoleDetail(id)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListRoles handles listing roles with pagination and filtering
//
//	@Summary		List roles
//	@Description	Get paginated list of roles with optional filtering
//	@Tags			Roles
//	@ID				list_roles
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int										false	"Page number"	default(1)
//	@Param			size		query		int										false	"Page size"		default(20)
//	@Param			is_system	query		bool									false	"Filter by system role"
//	@Param			status		query		consts.StatusType						false	"Filter by status"
//	@Success		200			{object}	dto.GenericResponse[dto.ListRoleResp]	"Roles retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]				"Invalid request parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/roles [get]
func ListRoles(c *gin.Context) {
	var req dto.ListRoleReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListRoles(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// UpdateRole handles role updates
//
//	@Summary		Update role
//	@Description	Update role information (partial update supported)
//	@Tags			Roles
//	@ID				update_role
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int									true	"Role ID"
//	@Param			request	body		dto.UpdateRoleReq					true	"Role update request"
//	@Success		202		{object}	dto.GenericResponse[dto.RoleResp]	"Role updated successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]			"Role not found"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/roles/{id} [patch]
func UpdateRole(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	var req dto.UpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.UpdateRole(&req, id)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Role updated successfully", resp)
}

// ===================== Role-Permission API =====================

// AssignRolePermission handles role-permission assignment
//
//	@Summary		Assign permissions to role
//	@Description	Assign multiple permissions to a role
//	@Tags			Roles
//	@ID				grant_permissions_to_role
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			role_id	path		int							true	"Role ID"
//	@Param			request	body		dto.AssignRolePermissionReq	true	"Permission assignment request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Permissions assigned successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid role ID or request format"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Role not found"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/roles/{role_id}/permissions/assign [post]
func AssignRolePermission(c *gin.Context) {
	roleIdStr := c.Param(consts.URLPathRoleID)
	roleID, err := strconv.Atoi(roleIdStr)
	if err != nil || roleID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	var req dto.AssignRolePermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	err = producer.BatchAssignRolePermissions(req.PermissionIDs, roleID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusOK, "Permissions assigned successfully", nil)
}

// RemovePermissionsFromRole handles permission removal from role
//
//	@Summary		Remove permissions from role
//	@Description	Remove multiple permissions from a role
//	@Tags			Roles
//	@ID				revoke_permissions_from_role
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			role_id	path		int							true	"Role ID"
//	@Param			request	body		dto.RemoveRolePermissionReq	true	"Permission removal request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Permissions removed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid role ID or request format"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Role not found"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/roles/{role_id}/permissions/remove [post]
func RemovePermissionsFromRole(c *gin.Context) {
	roleIDStr := c.Param(consts.URLPathRoleID)
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil || roleID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	var req dto.RemoveRolePermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	err = producer.RemovePermissionsFromRole(req.PermissionIDs, roleID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusOK, "Permissions removed successfully", nil)
}
