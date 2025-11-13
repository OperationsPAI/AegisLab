package v2

import (
	"aegis/consts"
	"net/http"
	"strconv"

	"aegis/database"
	"aegis/dto"
	"aegis/handlers"
	"aegis/repository"
	producer "aegis/service/prodcuer"

	"github.com/gin-gonic/gin"
)

// CreateUser handles user creation
//
//	@Summary		Create a new user
//	@Description	Create a new user account with specified details
//	@Tags			Users
//	@ID				create_user
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateUserReq					true	"User creation request"
//	@Success		201		{object}	dto.GenericResponse[dto.UserResp]	"User created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request format or parameters"
//	@Failure		409		{object}	dto.GenericResponse[any]			"User already exists"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/users [post]/
func CreateUser(c *gin.Context) {
	var req dto.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.CreateUser(&req)
	if handlers.HandleServiceError(c, err) {
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
//	@Router			/api/v2/users/{id} [delete]/
func DeleteUser(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	err = producer.DeleteUser(id)
	if handlers.HandleServiceError(c, err) {
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
//	@Success		200	{object}	dto.GenericResponse[dto.UserDetailResp]	"User retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]				"Invalid user ID"
//	@Failure		401	{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]				"User not found"
//	@Failure		500	{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/users/{id}/detail [get]
func GetUserDetailV2(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	resp, err := producer.GetUserDetail(id)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			status		query		int												false	"Filter by status"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[dto.UserResp]]	"Users retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]						"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/users [get]/
func ListUsersV2(c *gin.Context) {
	var req dto.ListUserReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListUsers(&req)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			request	body		dto.UpdateUserReq					true	"User update request"
//	@Success		202		{object}	dto.GenericResponse[dto.UserResp]	"User updated successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid user ID/request"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]			"User not found"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/users/{id} [patch]
func UpdateUser(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	resp, err := producer.UpdateUser(&req, id)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "User updated successfully", resp)
}

// SearchUsers handles complex user search
//
//	@Summary		Search users
//	@Description	Search users with complex filtering and sorting
//	@Tags			Users
//	@ID				search_users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.UserSearchReq									true	"User search request"
//	@Success		200		{object}	dto.GenericResponse[dto.SearchResp[dto.UserResp]]	"Users retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]							"Invalid request format or search parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/users/search [post]
func SearchUsers(c *gin.Context) {
	var req dto.UserSearchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	searchReq := req.ConvertToSearchReq()
	if err := searchReq.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid search parameters: "+err.Error())
		return
	}

	searchResult, err := repository.ExecuteSearch(database.DB, searchReq, database.User{})
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to search users: "+err.Error())
		return
	}

	var userResps []dto.UserResp
	for _, user := range searchResult.Items {
		userResps = append(userResps, *dto.NewUserResp(&user))
	}

	// Build final response
	response := dto.SearchResp[dto.UserResp]{
		Items:      userResps,
		Pagination: searchResult.Pagination,
		Filters:    searchResult.Filters,
		Sort:       searchResult.Sort,
	}

	dto.SuccessResponse(c, response)
}

// ===================== User-Role API =====================

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
func AssignUserRole(c *gin.Context) {
	userIDStr := c.Param(consts.URLPathUserID)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	roleIDStr := c.Param(consts.URLPathRoleID)
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil || roleID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	err = producer.AssignRoleToUser(userID, roleID)
	if handlers.HandleServiceError(c, err) {
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
func RemoveGlobalRole(c *gin.Context) {
	userIDStr := c.Param(consts.URLPathUserID)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	roleIDStr := c.Param(consts.URLPathRoleID)
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil || roleID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	err = producer.RemoveRoleFromUser(userID, roleID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Role removed successfully", nil)
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
//	@Success		200		{object}	dto.GenericResponse[[]dto.UserResp]	"Users retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid role ID"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]			"Role not found"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/roles/{role_id}/users [get]
func ListUsersFromRole(c *gin.Context) {
	roleIdStr := c.Param(consts.URLPathRoleID)
	roleID, err := strconv.Atoi(roleIdStr)
	if err != nil || roleID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	userResps, err := producer.ListUsersFromRole(roleID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, userResps)
}

// ===================== User-Permission API =====================

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
//	@Param			request	body		dto.AssignUserPermissionReq	true	"User permission assignment request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Permission assigned successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid user ID or invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failuer		404 {object} dto.GenericResponse[any] "Resource not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/users/{user_id}/permissions/assign [post]
func AssignUserPermission(c *gin.Context) {
	userIDStr := c.Param(consts.URLPathUserID)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.AssignUserPermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	err = producer.BatchAssignUserPermissions(&req, userID)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			request	body		dto.RemoveUserPermissionReq	true	"User permission removal request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Permission removed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid user or permission ID"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failuer		404 {object} dto.GenericResponse[any] "Resource not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/users/{user_id}/permissions/remove [post]
func RemoveUserPermission(c *gin.Context) {
	userIDStr := c.Param(consts.URLPathUserID)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.RemoveUserPermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	err = producer.BatchRemoveUserPermissions(&req, userID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusOK, "Permissions assigned successfully", nil)
}

// ===================== User-Container API =====================

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
func AssignUserContainer(c *gin.Context) {
	userIDStr := c.Param(consts.URLPathUserID)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	containerIDStr := c.Param(consts.URLPathContainerID)
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil || containerID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	roleIDStr := c.Param(consts.URLPathRoleID)
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil || roleID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	err = producer.AssignContainerToUser(userID, containerID, roleID)
	if handlers.HandleServiceError(c, err) {
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
func RemoveUserContainer(c *gin.Context) {
	userIDStr := c.Param(consts.URLPathUserID)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	containerIDStr := c.Param(consts.URLPathContainerID)
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	err = producer.RemoveContainerFromUser(userID, containerID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "User removed from container successfully", nil)
}

// ===================== User-Dataset API =====================

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
func AssignUserDataset(c *gin.Context) {
	userIDStr := c.Param(consts.URLPathUserID)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	datasetIDStr := c.Param(consts.URLPathDatasetID)
	datasetID, err := strconv.Atoi(datasetIDStr)
	if err != nil || datasetID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	roleIDStr := c.Param(consts.URLPathRoleID)
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil || roleID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	err = producer.AssignDatasetToUser(userID, datasetID, roleID)
	if handlers.HandleServiceError(c, err) {
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
func RemoveUserDataset(c *gin.Context) {
	userIDStr := c.Param(consts.URLPathUserID)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	datasetIDStr := c.Param(consts.URLPathDatasetID)
	datasetID, err := strconv.Atoi(datasetIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	err = producer.RemoveDatasetFromUser(userID, datasetID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "User removed from dataset successfully", nil)
}

// ===================== User-Project API =====================

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
func AssignUserProject(c *gin.Context) {
	userIDStr := c.Param(consts.URLPathUserID)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	projectIDStr := c.Param(consts.URLPathProjectID)
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil || projectID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	roleIDstr := c.Param(consts.URLPathRoleID)
	roleID, err := strconv.Atoi(roleIDstr)
	if err != nil || roleID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	err = producer.AssignProjectToUser(userID, projectID, roleID)
	if handlers.HandleServiceError(c, err) {
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
func RemoveUserProject(c *gin.Context) {
	userIDStr := c.Param(consts.URLPathUserID)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	projectIDStr := c.Param(consts.URLPathProjectID)
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	err = producer.RemoveProjectFromUser(userID, projectID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "User removed from project successfully", nil)
}
