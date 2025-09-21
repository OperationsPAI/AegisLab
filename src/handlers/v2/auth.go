package v2

import (
	"net/http"

	"aegis/database"
	"aegis/dto"
	"aegis/middleware"
	"aegis/repository"
	"aegis/utils"

	"github.com/gin-gonic/gin"
)

// Login handles user authentication
//
//	@Summary User login
//	@Description Authenticate user with username and password
//	@Tags Authentication
//	@Accept json
//	@Produce json
//	@Param request body dto.LoginRequest true "Login credentials"
//	@Success 200 {object} dto.GenericResponse[dto.LoginResponse] "Login successful"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 401 {object} dto.GenericResponse[any] "Invalid credentials"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/auth/login [post]
func Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Get user by username
	user, err := repository.GetUserByUsername(req.Username)
	if err != nil {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Verify password
	if !utils.VerifyPassword(req.Password, user.Password) {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token
	token, expiresAt, err := utils.GenerateToken(user.ID, user.Username, user.Email, user.IsActive)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token: "+err.Error())
		return
	}

	// Update last login time
	if err := repository.UpdateUserLoginTime(user.ID); err != nil {
		// Log error but don't fail the login
	}

	var userInfo dto.UserInfo
	userInfo.ConvertFromUser(user)

	response := dto.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      userInfo,
	}

	dto.SuccessResponse(c, response)
}

// Register handles user registration
//
//	@Summary User registration
//	@Description Register a new user account
//	@Tags Authentication
//	@Accept json
//	@Produce json
//	@Param request body dto.RegisterRequest true "Registration details"
//	@Success 201 {object} dto.GenericResponse[dto.UserInfo] "Registration successful"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request or validation error"
//	@Failure 409 {object} dto.GenericResponse[any] "User already exists"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/auth/register [post]
func Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	// Check if user already exists
	if _, err := repository.GetUserByUsername(req.Username); err == nil {
		dto.ErrorResponse(c, http.StatusConflict, "Username is already taken")
		return
	}

	if _, err := repository.GetUserByEmail(req.Email); err == nil {
		dto.ErrorResponse(c, http.StatusConflict, "Email is already registered")
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Password hashing failed: "+err.Error())
		return
	}

	user := &database.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		FullName: req.FullName,
		Phone:    req.Phone,
		Status:   1, // Active
		IsActive: true,
	}

	if err := repository.CreateUser(user); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	var userInfo dto.UserInfo
	userInfo.ConvertFromUser(user)

	dto.JSONResponse(c, http.StatusCreated, "Registration successful", userInfo)
}

// RefreshToken handles JWT token refresh
//
//	@Summary Refresh JWT token
//	@Description Refresh an existing JWT token
//	@Tags Authentication
//	@Accept json
//	@Produce json
//	@Param request body dto.TokenRefreshRequest true "Token refresh request"
//	@Success 200 {object} dto.GenericResponse[dto.TokenRefreshResponse] "Token refreshed successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 401 {object} dto.GenericResponse[any] "Invalid token"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/auth/refresh [post]
func RefreshToken(c *gin.Context) {
	var req dto.TokenRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Refresh the token
	newToken, expiresAt, err := utils.RefreshToken(req.Token)
	if err != nil {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Token refresh failed: "+err.Error())
		return
	}

	response := dto.TokenRefreshResponse{
		Token:     newToken,
		ExpiresAt: expiresAt,
	}

	dto.SuccessResponse(c, response)
}

// Logout handles user logout
//
//	@Summary User logout
//	@Description Logout user and invalidate token
//	@Tags Authentication
//	@Accept json
//	@Produce json
//	@Param request body dto.LogoutRequest true "Logout request"
//	@Success 200 {object} dto.GenericResponse[any] "Logout successful"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/auth/logout [post]
func Logout(c *gin.Context) {
	// Extract token from header
	authHeader := c.GetHeader("Authorization")
	token, err := utils.ExtractTokenFromHeader(authHeader)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid authorization header")
		return
	}

	// Validate and extract claims from token
	claims, err := utils.ValidateToken(token)
	if err != nil {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Add token to blacklist
	err = repository.BlacklistToken(
		claims.ID, // JWT ID (jti claim)
		claims.UserID,
		claims.ExpiresAt.Time,
		"User logout",
	)
	if err != nil {
		// Log error but don't fail logout
		// In production, you might want to log this properly
	}

	dto.SuccessResponse(c, gin.H{"message": "Logged out successfully"})
}

// ChangePassword handles password change
//
//	@Summary Change user password
//	@Description Change password for authenticated user
//	@Tags Authentication
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.ChangePasswordRequest true "Password change request"
//	@Success 200 {object} dto.GenericResponse[any] "Password changed successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 401 {object} dto.GenericResponse[any] "Unauthorized or invalid old password"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/auth/change-password [post]
func ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Get user ID from JWT token context (using middleware)
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	user, err := repository.GetUserByID(userID)
	if err != nil {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not found")
		return
	}

	// Verify old password
	if !utils.VerifyPassword(req.OldPassword, user.Password) {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Invalid old password")
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Password hashing failed: "+err.Error())
		return
	}
	user.Password = hashedPassword

	if err := repository.UpdateUser(user); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to update password")
		return
	}

	dto.SuccessResponse(c, "Password changed successfully")
}

// GetProfile handles getting current user profile
//
//	@Summary Get current user profile
//	@Description Get profile information for authenticated user
//	@Tags Authentication
//	@Produce json
//	@Security BearerAuth
//	@Success 200 {object} dto.GenericResponse[dto.UserResponse] "Profile retrieved successfully"
//	@Failure 401 {object} dto.GenericResponse[any] "Unauthorized"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/auth/profile [get]
func GetProfile(c *gin.Context) {
	// Get user ID from JWT token context (using middleware)
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	user, err := repository.GetUserByID(userID)
	if err != nil {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not found")
		return
	}

	var userResponse dto.UserResponse
	userResponse.ConvertFromUser(user)

	// Load user roles and permissions
	if roles, err := repository.GetUserRoles(userID); err == nil {
		userResponse.GlobalRoles = make([]dto.RoleResponse, len(roles))
		for i, role := range roles {
			userResponse.GlobalRoles[i].ConvertFromRole(&role)
		}
	}

	if permissions, err := repository.GetUserPermissions(userID, nil); err == nil {
		userResponse.Permissions = make([]dto.PermissionResponse, len(permissions))
		for i, permission := range permissions {
			userResponse.Permissions[i].ConvertFromPermission(&permission)
		}
	}

	dto.SuccessResponse(c, userResponse)
}
