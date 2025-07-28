package v2

import (
	"net/http"
	"strconv"

	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser handles user creation
//	@Summary Create a new user
//	@Description Create a new user account with specified details
//	@Tags Users
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.CreateUserRequest true "User creation request"
//	@Success 201 {object} dto.GenericResponse[dto.UserResponse] "User created successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 409 {object} dto.GenericResponse[any] "User already exists"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/users [post]
func CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Check if user already exists
	if _, err := repository.GetUserByUsername(req.Username); err == nil {
		dto.ErrorResponse(c, http.StatusConflict, "Username already exists")
		return
	}

	if _, err := repository.GetUserByEmail(req.Email); err == nil {
		dto.ErrorResponse(c, http.StatusConflict, "Email already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password: "+err.Error())
		return
	}

	user := &database.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		Status:   1,
		IsActive: true,
	}

	if err := repository.CreateUser(user); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user: "+err.Error())
		return
	}

	var response dto.UserResponse
	response.ConvertFromUser(user)

	dto.JSONResponse(c, http.StatusCreated, "User created successfully", response)
}

// GetUser handles getting a single user by ID
//	@Summary Get user by ID
//	@Description Get detailed information about a specific user
//	@Tags Users
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "User ID"
//	@Success 200 {object} dto.GenericResponse[dto.UserResponse] "User retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid user ID"
//	@Failure 404 {object} dto.GenericResponse[any] "User not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/users/{id} [get]
func GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := repository.GetUserByID(id)
	if err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	var response dto.UserResponse
	response.ConvertFromUser(user)

	// Load user roles
	if roles, err := repository.GetUserRoles(id); err == nil {
		response.GlobalRoles = make([]dto.RoleResponse, len(roles))
		for i, role := range roles {
			response.GlobalRoles[i].ConvertFromRole(&role)
		}
	}

	// Load user project roles
	if userProjects, err := repository.GetUserProjects(id); err == nil {
		response.ProjectRoles = make([]dto.UserProjectResponse, len(userProjects))
		for i, up := range userProjects {
			response.ProjectRoles[i].ConvertFromUserProject(&up)
		}
	}

	dto.SuccessResponse(c, response)
}

// ListUsers handles listing users with pagination and filtering
//	@Summary List users
//	@Description Get paginated list of users with optional filtering
//	@Tags Users
//	@Produce json
//	@Security BearerAuth
//	@Param page query int false "Page number" default(1)
//	@Param size query int false "Page size" default(20)
//	@Param status query int false "Filter by status"
//	@Param is_active query bool false "Filter by active status"
//	@Param username query string false "Filter by username"
//	@Param email query string false "Filter by email"
//	@Param full_name query string false "Filter by full name"
//	@Success 200 {object} dto.GenericResponse[dto.UserListResponse] "Users retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request parameters"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/users [get]
func ListUsers(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}

	users, total, err := repository.ListUsers(req.Page, req.Size, req.Status)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve users: "+err.Error())
		return
	}

	var userResponses []dto.UserResponse
	for _, user := range users {
		var response dto.UserResponse
		response.ConvertFromUser(&user)
		userResponses = append(userResponses, response)
	}

	pagination := dto.PaginationInfo{
		Page:       req.Page,
		Size:       req.Size,
		Total:      total,
		TotalPages: int((total + int64(req.Size) - 1) / int64(req.Size)),
	}

	response := dto.UserListResponse{
		Items:      userResponses,
		Pagination: pagination,
	}

	dto.SuccessResponse(c, response)
}

// UpdateUser handles user updates
//	@Summary Update user
//	@Description Update user information (partial update supported)
//	@Tags Users
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "User ID"
//	@Param request body dto.UpdateUserRequest true "User update request"
//	@Success 200 {object} dto.GenericResponse[dto.UserResponse] "User updated successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 404 {object} dto.GenericResponse[any] "User not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/users/{id} [put]
func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	user, err := repository.GetUserByID(id)
	if err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	// Update fields if provided
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := repository.UpdateUser(user); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to update user: "+err.Error())
		return
	}

	var response dto.UserResponse
	response.ConvertFromUser(user)

	dto.SuccessResponse(c, response)
}

// DeleteUser handles user deletion
//	@Summary Delete user
//	@Description Delete a user (soft delete by setting status to -1)
//	@Tags Users
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "User ID"
//	@Success 200 {object} dto.GenericResponse[any] "User deleted successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid user ID"
//	@Failure 404 {object} dto.GenericResponse[any] "User not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/users/{id} [delete]
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := repository.DeleteUser(id); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete user: "+err.Error())
		return
	}

	dto.SuccessResponse(c, "User deleted successfully")
}

// SearchUsers handles complex user search
//	@Summary Search users
//	@Description Search users with complex filtering and sorting
//	@Tags Users
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.UserSearchRequest true "User search request"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.UserResponse]] "Users retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/users/search [post]
func SearchUsers(c *gin.Context) {
	var req dto.UserSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Convert to SearchRequest
	searchReq := req.ConvertToSearchRequest()

	// Validate search request
	if err := searchReq.ValidateSearchRequest(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid search parameters: "+err.Error())
		return
	}

	// Execute search using query builder
	searchResult, err := repository.ExecuteSearch(database.DB, searchReq, database.User{})
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to search users: "+err.Error())
		return
	}

	// Convert database users to response DTOs
	var userResponses []dto.UserResponse
	for _, user := range searchResult.Items {
		var response dto.UserResponse
		response.ConvertFromUser(&user)

		// Load related data if requested
		if searchReq.HasFilter("include") {
			includes := searchReq.GetFilter("include")
			if includes != nil && includes.Value != nil {
				includeList, ok := includes.Value.([]string)
				if ok {
					for _, include := range includeList {
						switch include {
						case "roles":
							if roles, err := repository.GetUserRoles(user.ID); err == nil {
								response.GlobalRoles = make([]dto.RoleResponse, len(roles))
								for i, role := range roles {
									response.GlobalRoles[i].ConvertFromRole(&role)
								}
							}
						case "projects":
							if userProjects, err := repository.GetUserProjects(user.ID); err == nil {
								response.ProjectRoles = make([]dto.UserProjectResponse, len(userProjects))
								for i, up := range userProjects {
									response.ProjectRoles[i].ConvertFromUserProject(&up)
								}
							}
						case "permissions":
							if permissions, err := repository.GetUserPermissions(user.ID, nil); err == nil {
								response.Permissions = make([]dto.PermissionResponse, len(permissions))
								for i, permission := range permissions {
									response.Permissions[i].ConvertFromPermission(&permission)
								}
							}
						}
					}
				}
			}
		}

		userResponses = append(userResponses, response)
	}

	// Build final response
	response := dto.SearchResponse[dto.UserResponse]{
		Items:      userResponses,
		Pagination: searchResult.Pagination,
		Filters:    searchResult.Filters,
		Sort:       searchResult.Sort,
	}

	dto.SuccessResponse(c, response)
}

// AssignUserToProject handles user-project assignment
//	@Summary Assign user to project
//	@Description Assign a user to a project with a specific role
//	@Tags Users
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "User ID"
//	@Param request body dto.AssignUserToProjectRequest true "Project assignment request"
//	@Success 200 {object} dto.GenericResponse[any] "User assigned to project successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 404 {object} dto.GenericResponse[any] "User not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/users/{id}/projects [post]
func AssignUserToProject(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.AssignUserToProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Check if user exists
	if _, err := repository.GetUserByID(userID); err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	if err := repository.AddUserToProject(userID, req.ProjectID, req.RoleID); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to assign user to project: "+err.Error())
		return
	}

	dto.SuccessResponse(c, "User assigned to project successfully")
}

// RemoveUserFromProject handles user-project removal
//	@Summary Remove user from project
//	@Description Remove a user from a project
//	@Tags Users
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "User ID"
//	@Param project_id path int true "Project ID"
//	@Success 200 {object} dto.GenericResponse[any] "User removed from project successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid ID"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/users/{id}/projects/{project_id} [delete]
func RemoveUserFromProject(c *gin.Context) {
	userIDStr := c.Param("id")
	projectIDStr := c.Param("project_id")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	if err := repository.RemoveUserFromProject(userID, projectID); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove user from project: "+err.Error())
		return
	}

	dto.SuccessResponse(c, "User removed from project successfully")
}
