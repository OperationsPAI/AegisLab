package authmodule

import (
	"aegis/httpx"
	"net/http"

	"aegis/consts"
	"aegis/dto"
	"aegis/middleware"
	"aegis/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Login handles user authentication
//
//	@Summary		User login
//	@Description	Authenticate user with username and password
//	@Tags			Authentication
//	@ID				login
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginReq							true	"Login credentials"
//	@Success		200		{object}	dto.GenericResponse[LoginResp]		"Login successful"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request format"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Invalid user name or password"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/auth/login [post]
//	@x-api-type		{"portal":"true","admin":"true"}
func (h *Handler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	resp, err := h.service.Login(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Login successful", resp)
}

// Register handles user registration
//
//	@Summary		User registration
//	@Description	Register a new user account
//	@Tags			Authentication
//	@ID				register_user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RegisterReq							true	"Registration details"
//	@Success		201		{object}	dto.GenericResponse[UserInfo]		"Registration successful"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request format/parameters"
//	@Failure		409		{object}	dto.GenericResponse[any]			"User already exists"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/auth/register [post]
//	@x-api-type		{"portal":"true","admin":"true"}
func (h *Handler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.Register(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusCreated, "Registration successful", resp)
}

// RefreshToken handles JWT token refresh
//
//	@Summary		Refresh JWT token
//	@Description	Refresh an existing JWT token
//	@Tags			Authentication
//	@ID				refresh_auth_token
//	@Accept			json
//	@Produce		json
//	@Param			request	body		TokenRefreshReq								true	"Token refresh request"
//	@Success		200		{object}	dto.GenericResponse[TokenRefreshResp]		"Token refreshed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]					"Invalid request format"
//	@Failure		401		{object}	dto.GenericResponse[any]					"Invalid token"
//	@Failure		500		{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/auth/refresh [post]
//	@x-api-type		{"portal":"true","admin":"true"}
func (h *Handler) RefreshToken(c *gin.Context) {
	var req TokenRefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	resp, err := h.service.RefreshToken(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
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
//	@x-api-type		{"portal":"true","admin":"true"}
func (h *Handler) Logout(c *gin.Context) {
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

	err = h.service.Logout(c.Request.Context(), claims)
	if httpx.HandleServiceError(c, err) {
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
//	@Param			request	body		ChangePasswordReq			true	"Password change request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Password changed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/auth/change-password [post]
//	@x-api-type		{"portal":"true","admin":"true"}
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	err := h.service.ChangePassword(c.Request.Context(), &req, userID)
	if httpx.HandleServiceError(c, err) {
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
//	@Success		200	{object}	dto.GenericResponse[UserProfileResp]	"Profile retrieved successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		500	{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/auth/profile [get]
//	@x-api-type		{"portal":"true","admin":"true"}
func (h *Handler) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	resp, err := h.service.GetProfile(c.Request.Context(), userID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Profile retrieved successfully", resp)
}

// CreateAccessKey handles access key creation for the current user.
//
//	@Summary		Create access key
//	@Description	Create an AK/SK credential for the current authenticated user. This Portal response is the only time the `secret_key` is returned in plaintext, so callers must save it immediately.
//	@Tags			Authentication
//	@ID				create_access_key
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateAccessKeyReq								true	"Access key create request"
//	@Success		201		{object}	dto.GenericResponse[AccessKeyWithSecretResp]	"Access key created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]						"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		500		{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/access-keys [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) CreateAccessKey(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CreateAccessKeyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.CreateAccessKey(c.Request.Context(), userID, &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusCreated, "Access key created successfully", resp)
}

// ListAccessKeys lists access keys for the current user.
//
//	@Summary		List access keys
//	@Description	List AK/SK credentials owned by the current authenticated user
//	@Tags			Authentication
//	@ID				list_access_keys
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int	false	"Page number"
//	@Param			size	query		int	false	"Page size"
//	@Success		200		{object}	dto.GenericResponse[ListAccessKeyResp]	"Access keys listed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]					"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		500		{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/access-keys [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ListAccessKeys(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req ListAccessKeyReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ListAccessKeys(c.Request.Context(), userID, &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// GetAccessKey gets a single access key for the current user.
//
//	@Summary		Get access key detail
//	@Description	Get metadata for an AK/SK credential owned by the current authenticated user
//	@Tags			Authentication
//	@ID				get_access_key
//	@Produce		json
//	@Security		BearerAuth
//	@Param			access_key_id	path		int	true	"Access key ID"
//	@Success		200				{object}	dto.GenericResponse[AccessKeyInfo]	"Access key detail retrieved successfully"
//	@Failure		401				{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		404				{object}	dto.GenericResponse[any]				"Access key not found"
//	@Failure		500				{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/access-keys/{access_key_id} [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) GetAccessKey(c *gin.Context) {
	userID, accessKeyID, ok := parseCurrentUserAndAccessKeyID(c)
	if !ok {
		return
	}

	resp, err := h.service.GetAccessKey(c.Request.Context(), userID, accessKeyID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// DeleteAccessKey deletes an access key for the current user.
//
//	@Summary		Delete access key
//	@Description	Delete an AK/SK credential owned by the current authenticated user
//	@Tags			Authentication
//	@ID				delete_access_key
//	@Produce		json
//	@Security		BearerAuth
//	@Param			access_key_id	path		int	true	"Access key ID"
//	@Success		204				{object}	dto.GenericResponse[any]	"Access key deleted successfully"
//	@Failure		401				{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		404				{object}	dto.GenericResponse[any]	"Access key not found"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/access-keys/{access_key_id} [delete]
//	@x-api-type		{"portal":"true"}
func (h *Handler) DeleteAccessKey(c *gin.Context) {
	userID, accessKeyID, ok := parseCurrentUserAndAccessKeyID(c)
	if !ok {
		return
	}

	if httpx.HandleServiceError(c, h.service.DeleteAccessKey(c.Request.Context(), userID, accessKeyID)) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Access key deleted successfully", nil)
}

// DisableAccessKey disables an access key for the current user.
//
//	@Summary		Disable access key
//	@Description	Disable an AK/SK credential owned by the current authenticated user
//	@Tags			Authentication
//	@ID				disable_access_key
//	@Produce		json
//	@Security		BearerAuth
//	@Param			access_key_id	path		int	true	"Access key ID"
//	@Success		200				{object}	dto.GenericResponse[any]	"Access key disabled successfully"
//	@Failure		401				{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		404				{object}	dto.GenericResponse[any]	"Access key not found"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/access-keys/{access_key_id}/disable [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) DisableAccessKey(c *gin.Context) {
	userID, accessKeyID, ok := parseCurrentUserAndAccessKeyID(c)
	if !ok {
		return
	}

	if httpx.HandleServiceError(c, h.service.DisableAccessKey(c.Request.Context(), userID, accessKeyID)) {
		return
	}

	dto.JSONResponse[any](c, http.StatusOK, "Access key disabled successfully", nil)
}

// EnableAccessKey enables an access key for the current user.
//
//	@Summary		Enable access key
//	@Description	Enable an AK/SK credential owned by the current authenticated user
//	@Tags			Authentication
//	@ID				enable_access_key
//	@Produce		json
//	@Security		BearerAuth
//	@Param			access_key_id	path		int	true	"Access key ID"
//	@Success		200				{object}	dto.GenericResponse[any]	"Access key enabled successfully"
//	@Failure		401				{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		404				{object}	dto.GenericResponse[any]	"Access key not found"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/access-keys/{access_key_id}/enable [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) EnableAccessKey(c *gin.Context) {
	userID, accessKeyID, ok := parseCurrentUserAndAccessKeyID(c)
	if !ok {
		return
	}

	if httpx.HandleServiceError(c, h.service.EnableAccessKey(c.Request.Context(), userID, accessKeyID)) {
		return
	}

	dto.JSONResponse[any](c, http.StatusOK, "Access key enabled successfully", nil)
}

// RotateAccessKey rotates the secret key for an existing access key.
//
//	@Summary		Rotate access key secret
//	@Description	Rotate the secret half of an AK/SK credential owned by the current authenticated user
//	@Tags			Authentication
//	@ID				rotate_access_key
//	@Produce		json
//	@Security		BearerAuth
//	@Param			access_key_id	path		int												true	"Access key ID"
//	@Success		200				{object}	dto.GenericResponse[AccessKeyWithSecretResp]	"Access key rotated successfully"
//	@Failure		401				{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		404				{object}	dto.GenericResponse[any]						"Access key not found"
//	@Failure		500				{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/access-keys/{access_key_id}/rotate [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) RotateAccessKey(c *gin.Context) {
	userID, accessKeyID, ok := parseCurrentUserAndAccessKeyID(c)
	if !ok {
		return
	}

	resp, err := h.service.RotateAccessKey(c.Request.Context(), userID, accessKeyID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Access key rotated successfully", resp)
}

// ExchangeAccessKeyToken exchanges AK/SK for a bearer token.
//
//	@Summary		Exchange access key for token
//	@Description	Exchange an AK/SK signed request for a short-lived bearer token. Access keys are created in Portal, while SDK and CLI callers use this endpoint with `X-Access-Key`, `X-Timestamp`, `X-Nonce`, and `X-Signature`.
//	@Tags			Authentication
//	@ID				exchange_access_key_token
//	@Produce		json
//	@Param			X-Access-Key	header		string										true	"Access key ID"
//	@Param			X-Timestamp		header		string										true	"Unix timestamp in seconds"
//	@Param			X-Nonce			header		string										true	"Unique request nonce"
//	@Param			X-Signature		header		string										true	"Hex encoded HMAC-SHA256 signature of METHOD\\nPATH\\nACCESS_KEY\\nTIMESTAMP\\nNONCE"
//	@Success		200		{object}	dto.GenericResponse[AccessKeyTokenResp]	"Access key token issued successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]					"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]					"Invalid signature or replayed request"
//	@Failure		500		{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/auth/access-key/token [post]
//	@x-api-type		{"sdk":"true"}
func (h *Handler) ExchangeAccessKeyToken(c *gin.Context) {
	var req AccessKeyTokenReq
	req.AccessKey = c.GetHeader("X-Access-Key")
	req.Timestamp = c.GetHeader("X-Timestamp")
	req.Nonce = c.GetHeader("X-Nonce")
	req.Signature = c.GetHeader("X-Signature")
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ExchangeAccessKeyToken(c.Request.Context(), &req, c.Request.Method, c.Request.URL.Path)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Access key token issued successfully", resp)
}

func parseCurrentUserAndAccessKeyID(c *gin.Context) (int, int, bool) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return 0, 0, false
	}

	accessKeyID, ok := httpx.ParsePositiveID(c, c.Param("access_key_id"), consts.URLPathID)
	if !ok {
		return 0, 0, false
	}

	return userID, accessKeyID, true
}
