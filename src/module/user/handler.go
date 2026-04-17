package usermodule

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

// CreateUser handles user creation
//
//	@Summary		Create a new user
//	@Description	Create a new user account with specified details
//	@Tags			Users
//	@ID				create_user
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateUserReq						true	"User creation request"
//	@Success		201		{object}	dto.GenericResponse[UserResp]		"User created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request format or parameters"
//	@Failure		409		{object}	dto.GenericResponse[any]			"User already exists"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/users [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.CreateUser(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusCreated, "User created successfully", resp)
}

// DeleteUser handles user deletion
//
//	@Summary		Delete user
//	@Description	Delete a user
//	@Tags			Users
//	@ID				delete_user
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int							true	"User ID"
//	@Success		204	{object}	dto.GenericResponse[any]	"User deleted successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]	"Invalid user ID"
//	@Failure		401	{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]	"User not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/users/{id} [delete]
//	@x-api-type		{"admin":"true"}
func (h *Handler) DeleteUser(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}
	if httpx.HandleServiceError(c, h.service.DeleteUser(c.Request.Context(), userID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusNoContent, "User deleted successfully", nil)
}

// GetUserDetail handles getting a single user by ID (new CRUD version)
//
//	@Summary		Get user by ID
//	@Description	Get detailed information about a specific user
//	@Tags			Users
//	@ID				get_user_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int										true	"User ID"
//	@Success		200	{object}	dto.GenericResponse[UserDetailResp]	"User retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]				"Invalid user ID"
//	@Failure		401	{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]				"User not found"
//	@Failure		500	{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/users/{id}/detail [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) GetUserDetail(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.GetUserDetail(c.Request.Context(), userID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// ListUsers handles listing users with pagination and filtering
//
//	@Summary		List users
//	@Description	Get paginated list of users with filtering
//	@Tags			Users
//	@ID				list_users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int												false	"Page number"	default(1)
//	@Param			size		query		int												false	"Page size"		default(20)
//	@Param			username	query		string											false	"Filter by username"
//	@Param			email		query		string											false	"Filter by email"
//	@Param			is_active	query		bool											false	"Filter by active status"
//	@Param			status		query		consts.StatusType								false	"Filter by status"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[UserResp]]	"Users retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]						"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/users [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) ListUsers(c *gin.Context) {
	var req ListUserReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.ListUsers(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// UpdateUser handles user updates
//
//	@Summary		Update user
//	@Description	Update an existing user's information
//	@Tags			Users
//	@ID				update_user
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int									true	"User ID"
//	@Param			request	body		UpdateUserReq						true	"User update request"
//	@Success		202		{object}	dto.GenericResponse[UserResp]		"User updated successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid user ID/request"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]			"User not found"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/users/{id} [patch]
//	@x-api-type		{"admin":"true"}
func (h *Handler) UpdateUser(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}
	var req UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	resp, err := h.service.UpdateUser(c.Request.Context(), &req, userID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse[any](c, http.StatusAccepted, "User updated successfully", resp)
}

// AssignUserRole handles user-role assignment
//
//	@Summary		Assign global role to user
//	@Description	Assign a role to a user (global role assignment)
//	@Tags			Relations
//	@ID				assign_role_to_user
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path		int							true	"User ID"
//	@Param			role_id	path		int							true	"Role ID"
//	@Success		200		{object}	dto.GenericResponse[any]	"Role assigned successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid user ID or role ID"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Resource not found"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/users/{user_id}/role/{role_id} [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) AssignRole(c *gin.Context) {
	userID, roleID, ok := parseUserAndRoleIDs(c)
	if !ok {
		return
	}
	if httpx.HandleServiceError(c, h.service.AssignRole(c.Request.Context(), userID, roleID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusOK, "Role assigned successfully", nil)
}

// RemoveGlobalRole handles user-role removal
//
//	@Summary		Remove role from user
//	@Description	Remove a role from a user (global role removal)
//	@Tags			Relations
//	@ID				remove_role_from_user
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path		int							true	"User ID"
//	@Param			role_id	path		int							true	"Role ID"
//	@Success		204		{object}	dto.GenericResponse[any]	"Role removed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid user or role ID"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]	"User or role not found"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/users/{user_id}/roles/{role_id} [delete]
//	@x-api-type		{"admin":"true"}
func (h *Handler) RemoveRole(c *gin.Context) {
	userID, roleID, ok := parseUserAndRoleIDs(c)
	if !ok {
		return
	}
	if httpx.HandleServiceError(c, h.service.RemoveRole(c.Request.Context(), userID, roleID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusNoContent, "Role removed successfully", nil)
}

// AssignUserPermission handles direct user-permission assignment
//
//	@Summary		Assign permission to user
//	@Description	Assign permissions directly to a user (with optional container/dataset/project scope)
//	@Tags			Users
//	@ID				grant_user_permissions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path		int							true	"User ID"
//	@Param			request	body		AssignUserPermissionReq	true	"User permission assignment request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Permission assigned successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid user ID or invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failuer		404 {object} dto.GenericResponse[any] "Resource not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/users/{user_id}/permissions/assign [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) AssignPermissions(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}
	var req AssignUserPermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	if httpx.HandleServiceError(c, h.service.AssignPermissions(c.Request.Context(), &req, userID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusOK, "Permissions assigned successfully", nil)
}

// RemoveUserPermission handles direct user-permission removal
//
//	@Summary		Remove permission from user
//	@Description	Remove permissions directly from a user
//	@Tags			Users
//	@ID				revoke_user_permissions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path		int							true	"User ID"
//	@Param			request	body		RemoveUserPermissionReq	true	"User permission removal request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Permission removed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid user or permission ID"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failuer		404 {object} dto.GenericResponse[any] "Resource not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/users/{user_id}/permissions/remove [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) RemovePermissions(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}
	var req RemoveUserPermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	if httpx.HandleServiceError(c, h.service.RemovePermissions(c.Request.Context(), &req, userID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusOK, "Permissions assigned successfully", nil)
}

// AssignUserContainer handles user-container assignment
//
//	@Summary		Assign user to container
//	@Description	Assign a user to a container with a specific role
//	@Tags			Users
//	@ID				assign_user_to_container
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id			path		int							true	"User ID"
//	@Param			container_id	path		int							true	"Container ID"
//	@Param			role_id			path		int							true	"Role ID"
//	@Success		200				{object}	dto.GenericResponse[any]	"User assigned to container successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]	"Invalid user ID or container ID or role ID"
//	@Failure		401				{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]	"User or container or role not found"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/users/{user_id}/containers/{container_id}/roles/{role_id} [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) AssignContainer(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}
	containerID, ok := parsePathID(c, consts.URLPathContainerID, "Invalid container ID")
	if !ok {
		return
	}
	roleID, ok := parsePathID(c, consts.URLPathRoleID, "Invalid role ID")
	if !ok {
		return
	}
	if httpx.HandleServiceError(c, h.service.AssignContainer(c.Request.Context(), userID, containerID, roleID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusOK, "User assigned to container successfully", nil)
}

// RemoveUserContainer handles user-container removal
//
//	@Summary		Remove user from container
//	@Description	Remove a user from a container
//	@Tags			Users
//	@ID				remove_user_from_container
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id			path		int							true	"User ID"
//	@Param			container_id	path		int							true	"Container ID"
//	@Success		204				{object}	dto.GenericResponse[any]	"User removed from container successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]	"Invalid user or container ID"
//	@Failure		401				{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]	"User or container not found"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/users/{user_id}/containers/{container_id} [delete]
//	@x-api-type		{"admin":"true"}
func (h *Handler) RemoveContainer(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}
	containerID, ok := parsePathID(c, consts.URLPathContainerID, "Invalid container ID")
	if !ok {
		return
	}
	if httpx.HandleServiceError(c, h.service.RemoveContainer(c.Request.Context(), userID, containerID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusNoContent, "User removed from container successfully", nil)
}

// AssignUserDataset handles user-dataset assignment
//
//	@Summary		Assign user to dataset
//	@Description	Assign a user to a dataset with a specific role
//	@Tags			Users
//	@ID				assign_user_to_dataset
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id		path		int							true	"User ID"
//	@Param			dataset_id	path		int							true	"Dataset ID"
//	@Param			role_id		path		int							true	"Role ID"
//	@Success		200			{object}	dto.GenericResponse[any]	"User assigned to dataset successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid user ID or dataset ID or role ID"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]	"User or dataset or role not found"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/users/{user_id}/datasets/{dataset_id}/roles/{role_id} [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) AssignDataset(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}
	datasetID, ok := parsePathID(c, consts.URLPathDatasetID, "Invalid dataset ID")
	if !ok {
		return
	}
	roleID, ok := parsePathID(c, consts.URLPathRoleID, "Invalid role ID")
	if !ok {
		return
	}
	if httpx.HandleServiceError(c, h.service.AssignDataset(c.Request.Context(), userID, datasetID, roleID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusOK, "User assigned to dataset successfully", nil)
}

// RemoveUserDataset handles user-dataset removal
//
//	@Summary		Remove user from dataset
//	@Description	Remove a user from a dataset
//	@Tags			Users
//	@ID				remove_user_from_dataset
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id		path		int							true	"User ID"
//	@Param			dataset_id	path		int							true	"Dataset ID"
//	@Success		204			{object}	dto.GenericResponse[any]	"User removed from dataset successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid user or dataset ID"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]	"User or dataset not found"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/users/{user_id}/datasets/{dataset_id} [delete]
//	@x-api-type		{"admin":"true"}
func (h *Handler) RemoveDataset(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}
	datasetID, ok := parsePathID(c, consts.URLPathDatasetID, "Invalid dataset ID")
	if !ok {
		return
	}
	if httpx.HandleServiceError(c, h.service.RemoveDataset(c.Request.Context(), userID, datasetID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusNoContent, "User removed from dataset successfully", nil)
}

// AssignUserToProject handles user-project assignment
//
//	@Summary		Assign user to project
//	@Description	Assign a user to a project with a specific role
//	@Tags			Users
//	@ID				assign_user_to_project
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id		path		int							true	"User ID"
//	@Param			project_id	path		int							true	"Project ID"
//	@Param			role_id		path		int							true	"Role ID"
//	@Success		200			{object}	dto.GenericResponse[any]	"User assigned to project successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid user ID or project ID or role ID"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]	"User or project or role not found"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/users/{user_id}/projects/{project_id}/roles/{role_id} [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) AssignProject(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}
	projectID, ok := parsePathID(c, consts.URLPathProjectID, "Invalid project ID")
	if !ok {
		return
	}
	roleID, ok := parsePathID(c, consts.URLPathRoleID, "Invalid role ID")
	if !ok {
		return
	}
	if httpx.HandleServiceError(c, h.service.AssignProject(c.Request.Context(), userID, projectID, roleID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusOK, "User assigned to project successfully", nil)
}

// RemoveUserFromProject handles user-project removal
//
//	@Summary		Remove user from project
//	@Description	Remove a user from a project
//	@Tags			Users
//	@ID				remove_user_from_project
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id		path		int							true	"User ID"
//	@Param			project_id	path		int							true	"Project ID"
//	@Success		204			{object}	dto.GenericResponse[any]	"User removed from project successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid user or project ID"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]	"User or project not found"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/users/{user_id}/projects/{project_id} [delete]
//	@x-api-type		{"admin":"true"}
func (h *Handler) RemoveProject(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}
	projectID, ok := parsePathID(c, consts.URLPathProjectID, "Invalid project ID")
	if !ok {
		return
	}
	if httpx.HandleServiceError(c, h.service.RemoveProject(c.Request.Context(), userID, projectID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusNoContent, "User removed from project successfully", nil)
}

func parseUserID(c *gin.Context) (int, bool) {
	return parsePathID(c, consts.URLPathUserID, "Invalid user ID")
}

func parseUserAndRoleIDs(c *gin.Context) (int, int, bool) {
	userID, ok := parseUserID(c)
	if !ok {
		return 0, 0, false
	}
	roleID, ok := parsePathID(c, consts.URLPathRoleID, "Invalid role ID")
	if !ok {
		return 0, 0, false
	}
	return userID, roleID, true
}

func parsePathID(c *gin.Context, name, message string) (int, bool) {
	value := c.Param(name)
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, message)
		return 0, false
	}
	return id, true
}
