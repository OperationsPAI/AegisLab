package v2

import (
	"context"
	"net/http"

	"aegis/dto"
	"aegis/handlers"
	"aegis/middleware"
	producer "aegis/service/prodcuer"
	"aegis/utils"

	"github.com/gin-gonic/gin"
)

// Register handles user registration
//
//	@Summary		User registration
//	@Description	Register a new user account
//	@Tags			Authentication
//	@ID				register_user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterReq						true	"Registration details"
//	@Success		201		{object}	dto.GenericResponse[dto.UserInfo]	"Registration successful"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request format/parameters"
//	@Failure		409		{object}	dto.GenericResponse[any]			"User already exists"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/auth/register [post]
func Register(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.Register(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusCreated, "Registration successful", resp)
}

// Login handles user authentication
//
//	@Summary		User login
//	@Description	Authenticate user with username and password
//	@Tags			Authentication
//	@ID				login
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginReq						true	"Login credentials"
//	@Success		200		{object}	dto.GenericResponse[dto.LoginResp]	"Login successful"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request format"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Invalid user name or password"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/auth/login [post]
func Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	resp, err := producer.Login(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Login successful", resp)
}

// RefreshToken handles JWT token refresh
//
//	@Summary		Refresh JWT token
//	@Description	Refresh an existing JWT token
//	@Tags			Authentication
//	@ID				refresh_auth_token
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.TokenRefreshReq							true	"Token refresh request"
//	@Success		200		{object}	dto.GenericResponse[dto.TokenRefreshResp]	"Token refreshed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]					"Invalid request format"
//	@Failure		401		{object}	dto.GenericResponse[any]					"Invalid token"
//	@Failure		500		{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/auth/refresh [post]
func RefreshToken(c *gin.Context) {
	var req dto.TokenRefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	resp, err := producer.RefreshToken(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Token refreshed successfully", resp)
}

// Logout handles user logout
//
//	@Summary		User logout
//	@Description	Logout user and invalidate token
//	@Tags			Authentication
//	@ID				logout
//	@Produce		json
//	@Success		200	{object}	dto.GenericResponse[any]	"Logout successful"
//	@Failure		400	{object}	dto.GenericResponse[any]	"Invalid authorization header"
//	@Failure		401	{object}	dto.GenericResponse[any]	"Invalid token"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/auth/logout [post]
func Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	token, err := utils.ExtractTokenFromHeader(authHeader)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid authorization header")
		return
	}

	claims, err := utils.ValidateToken(token)
	if err != nil {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Invalid token")
		return
	}

	err = producer.Logout(context.Background(), claims)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusOK, "Logged out successfully", nil)
}

// ChangePassword handles password change
//
//	@Summary		Change user password
//	@Description	Change password for authenticated user
//	@Tags			Authentication
//	@ID				change_password
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.ChangePasswordReq		true	"Password change request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Password changed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/auth/change-password [post]
func ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req dto.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	err := producer.ChangePassword(&req, userID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusOK, "Password changed successfully", nil)
}

// GetProfile handles getting current user profile
//
//	@Summary		Get current user profile
//	@Description	Get profile information for authenticated user
//	@Tags			Authentication
//	@ID				get_current_user_profile
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GenericResponse[dto.UserDetailResp]	"Profile retrieved successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		500	{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/auth/profile [get]
func GetProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	resp, err := producer.GetProfile(userID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Profile retrieved successfully", resp)
}
