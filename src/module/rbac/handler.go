package rbacmodule

import (
	"aegis/httpx"
	"net/http"
	"strconv"

	"aegis/consts"
	"aegis/dto"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateRole handles role creation
//
//	@Summary		Create a new role
//	@Description	Create a new role with specified permissions
//	@Tags			Roles
//	@ID				create_role
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateRoleReq						true	"Role creation request"
//	@Success		201		{object}	dto.GenericResponse[RoleResp]	"Role created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request format"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		409		{object}	dto.GenericResponse[any]			"Role already exists"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/roles [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) CreateRole(c *gin.Context) {
	var req CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	resp, err := h.service.CreateRole(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
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
//	@x-api-type		{"admin":"true"}
func (h *Handler) DeleteRole(c *gin.Context) {
	roleID, ok := parseID(c, consts.URLPathRoleID, "Invalid role ID")
	if !ok {
		return
	}
	if httpx.HandleServiceError(c, h.service.DeleteRole(c.Request.Context(), roleID)) {
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
//	@Success		200	{object}	dto.GenericResponse[RoleDetailResp]	"Role retrieved successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		400	{object}	dto.GenericResponse[any]				"Invalid role ID"
//	@Failure		404	{object}	dto.GenericResponse[any]				"Role not found"
//	@Failure		500	{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/roles/{id} [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) GetRole(c *gin.Context) {
	roleID, ok := parseID(c, consts.URLPathRoleID, "Invalid role ID")
	if !ok {
		return
	}
	resp, err := h.service.GetRole(c.Request.Context(), roleID)
	if httpx.HandleServiceError(c, err) {
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
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[RoleResp]]	"Roles retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]				"Invalid request parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/roles [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) ListRoles(c *gin.Context) {
	var req ListRoleReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.ListRoles(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
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
//	@Param			request	body		UpdateRoleReq						true	"Role update request"
//	@Success		202		{object}	dto.GenericResponse[RoleResp]	"Role updated successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]			"Role not found"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/roles/{id} [patch]
//	@x-api-type		{"admin":"true"}
func (h *Handler) UpdateRole(c *gin.Context) {
	roleID, ok := parseID(c, consts.URLPathRoleID, "Invalid role ID")
	if !ok {
		return
	}
	var req UpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.UpdateRole(c.Request.Context(), &req, roleID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse[any](c, http.StatusAccepted, "Role updated successfully", resp)
}

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
//	@Param			request	body		AssignRolePermissionReq		true	"Permission assignment request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Permissions assigned successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid role ID or request format"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Role not found"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/roles/{role_id}/permissions/assign [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) AssignRolePermissions(c *gin.Context) {
	roleID, ok := parseID(c, consts.URLPathRoleID, "Invalid role ID")
	if !ok {
		return
	}
	var req AssignRolePermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if httpx.HandleServiceError(c, h.service.AssignRolePermissions(c.Request.Context(), req.PermissionIDs, roleID)) {
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
//	@Param			request	body		RemoveRolePermissionReq		true	"Permission removal request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Permissions removed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid role ID or request format"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Role not found"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/roles/{role_id}/permissions/remove [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) RemoveRolePermissions(c *gin.Context) {
	roleID, ok := parseID(c, consts.URLPathRoleID, "Invalid role ID")
	if !ok {
		return
	}
	var req RemoveRolePermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if httpx.HandleServiceError(c, h.service.RemoveRolePermissions(c.Request.Context(), req.PermissionIDs, roleID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusOK, "Permissions removed successfully", nil)
}

// ListUsersFromRole handles listing users assigned to a role
//
//	@Summary		List users from role
//	@Description	Get list of users assigned to a specific role
//	@Tags			Roles
//	@ID				list_users_by_role
//	@Produce		json
//	@Security		BearerAuth
//	@Param			role_id	path		int									true	"Role ID"
//	@Success		200		{object}	dto.GenericResponse[[]UserListItem]	"Users retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid role ID"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]			"Role not found"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/roles/{role_id}/users [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) ListUsersFromRole(c *gin.Context) {
	roleID, ok := parseID(c, consts.URLPathRoleID, "Invalid role ID")
	if !ok {
		return
	}
	resp, err := h.service.ListUsersFromRole(c.Request.Context(), roleID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
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
//	@Success		200	{object}	dto.GenericResponse[PermissionDetailResp]	"Permission retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]						"Invalid permission ID"
//	@Failure		401	{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]						"Permission not found"
//	@Failure		500	{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/permissions/{id} [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) GetPermission(c *gin.Context) {
	permissionID, ok := parseID(c, consts.URLPathPermissionID, "Invalid permission ID")
	if !ok {
		return
	}
	resp, err := h.service.GetPermission(c.Request.Context(), permissionID)
	if httpx.HandleServiceError(c, err) {
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
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[PermissionResp]]	"Permissions retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]				"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/permissions [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) ListPermissions(c *gin.Context) {
	var req ListPermissionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.ListPermissions(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// ListRolesFromPermission handles listing roles assigned to a permission
//
//	@Summary		List roles from permission
//	@Description	Get list of roles assigned to a specific permission
//	@Tags			Permissions
//	@ID				list_roles_with_permission
//	@Produce		json
//	@Security		BearerAuth
//	@Param			permission_id	path		int									true	"Permission ID"
//	@Success		200				{object}	dto.GenericResponse[[]RoleResp]	"Roles retrieved successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]			"Invalid permission ID"
//	@Failure		401				{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]			"Permission not found"
//	@Failure		500				{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/permissions/{permission_id}/roles [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) ListRolesFromPermission(c *gin.Context) {
	permissionID, ok := parseID(c, consts.URLPathPermissionID, "Invalid permission ID")
	if !ok {
		return
	}
	resp, err := h.service.ListRolesFromPermission(c.Request.Context(), permissionID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// GetResourceDetail handles getting a single resource by ID
//
//	@Summary		Get resource by ID
//	@Description	Get detailed information about a specific resource
//	@Tags			Resources
//	@ID				get_resource_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int										true	"Resource ID"
//	@Success		200	{object}	dto.GenericResponse[ResourceResp]	"Resource retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]				"Invalid resource ID"
//	@Failure		401	{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]				"Resource not found"
//	@Failure		500	{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/resources/{id} [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) GetResource(c *gin.Context) {
	resourceID, ok := parseID(c, consts.URLPathResourceID, "Invalid resource ID")
	if !ok {
		return
	}
	resp, err := h.service.GetResource(c.Request.Context(), resourceID)
	if httpx.HandleServiceError(c, err) {
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
//	@Param			type		query		consts.ResourceType									false	"Filter by resource type"
//	@Param			category	query		consts.ResourceCategory								false	"Filter by resource category"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[ResourceResp]]	"Resources retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]							"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/resources [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) ListResources(c *gin.Context) {
	var req ListResourceReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.ListResources(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
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
//	@Success		200	{object}	dto.GenericResponse[[]PermissionResp]	"Permissions retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]					"Invalid resource ID or request form"
//	@Failure		401	{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]					"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]					"Resource not found"
//	@Failure		500	{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/resources/{id}/permissions [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) ListResourcePermissions(c *gin.Context) {
	resourceID, ok := parseID(c, consts.URLPathResourceID, "Invalid resource ID")
	if !ok {
		return
	}
	resp, err := h.service.ListResourcePermissions(c.Request.Context(), resourceID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

func parseID(c *gin.Context, param, message string) (int, bool) {
	value := c.Param(param)
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, message)
		return 0, false
	}
	return id, true
}
