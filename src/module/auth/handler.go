package auth

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
	service HandlerService
}

func NewHandler(service HandlerService) *Handler {
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
//	@Param			request	body		LoginReq						true	"Login credentials"
//	@Success		200		{object}	dto.GenericResponse[LoginResp]	"Login successful"
//	@Failure		400		{object}	dto.GenericResponse[any]		"Invalid request format"
//	@Failure		401		{object}	dto.GenericResponse[any]		"Invalid user name or password"
//	@Failure		500		{object}	dto.GenericResponse[any]		"Internal server error"
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
//	@Param			request	body		RegisterReq						true	"Registration details"
//	@Success		201		{object}	dto.GenericResponse[UserInfo]	"Registration successful"
//	@Failure		400		{object}	dto.GenericResponse[any]		"Invalid request format/parameters"
//	@Failure		409		{object}	dto.GenericResponse[any]		"User already exists"
//	@Failure		500		{object}	dto.GenericResponse[any]		"Internal server error"
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
//	@Param			request	body		TokenRefreshReq							true	"Token refresh request"
//	@Success		200		{object}	dto.GenericResponse[TokenRefreshResp]	"Token refreshed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid request format"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Invalid token"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
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

// CreateAPIKey handles API key creation for the current user.
//
//	@Summary		Create API key
//	@Description	Create a Key ID / Key Secret credential for the current authenticated user. This Portal response is the only time the `key_secret` is returned in plaintext, so callers must save it immediately.
//	@Tags			Authentication
//	@ID				create_api_key
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateAPIKeyReq								true	"API key create request"
//	@Success		201		{object}	dto.GenericResponse[APIKeyWithSecretResp]	"API key created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]					"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		500		{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/api-keys [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) CreateAPIKey(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CreateAPIKeyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.CreateAPIKey(c.Request.Context(), userID, &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusCreated, "API key created successfully", resp)
}

// ListAPIKeys lists API keys for the current user.
//
//	@Summary		List API keys
//	@Description	List Key ID / Key Secret credentials owned by the current authenticated user
//	@Tags			Authentication
//	@ID				list_api_keys
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int									false	"Page number"
//	@Param			size	query		int									false	"Page size"
//	@Success		200		{object}	dto.GenericResponse[ListAPIKeyResp]	"API keys listed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/api-keys [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ListAPIKeys(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req ListAPIKeyReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ListAPIKeys(c.Request.Context(), userID, &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// GetAPIKey gets a single API key for the current user.
//
//	@Summary		Get API key detail
//	@Description	Get metadata for a Key ID / Key Secret credential owned by the current authenticated user
//	@Tags			Authentication
//	@ID				get_api_key
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int								true	"API key record ID"
//	@Success		200	{object}	dto.GenericResponse[APIKeyInfo]	"API key detail retrieved successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]		"Authentication required"
//	@Failure		404	{object}	dto.GenericResponse[any]		"API key not found"
//	@Failure		500	{object}	dto.GenericResponse[any]		"Internal server error"
//	@Router			/api/v2/api-keys/{id} [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) GetAPIKey(c *gin.Context) {
	userID, accessKeyID, ok := parseCurrentUserAndAPIKeyID(c)
	if !ok {
		return
	}

	resp, err := h.service.GetAPIKey(c.Request.Context(), userID, accessKeyID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// DeleteAPIKey deletes an API key for the current user.
//
//	@Summary		Delete API key
//	@Description	Delete a Key ID / Key Secret credential owned by the current authenticated user
//	@Tags			Authentication
//	@ID				delete_api_key
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int							true	"API key record ID"
//	@Success		204	{object}	dto.GenericResponse[any]	"API key deleted successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		404	{object}	dto.GenericResponse[any]	"API key not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/api-keys/{id} [delete]
//	@x-api-type		{"portal":"true"}
func (h *Handler) DeleteAPIKey(c *gin.Context) {
	userID, accessKeyID, ok := parseCurrentUserAndAPIKeyID(c)
	if !ok {
		return
	}

	if httpx.HandleServiceError(c, h.service.DeleteAPIKey(c.Request.Context(), userID, accessKeyID)) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "API key deleted successfully", nil)
}

// DisableAPIKey disables an API key for the current user.
//
//	@Summary		Disable API key
//	@Description	Disable a Key ID / Key Secret credential owned by the current authenticated user
//	@Tags			Authentication
//	@ID				disable_api_key
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int							true	"API key record ID"
//	@Success		200	{object}	dto.GenericResponse[any]	"API key disabled successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		404	{object}	dto.GenericResponse[any]	"API key not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/api-keys/{id}/disable [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) DisableAPIKey(c *gin.Context) {
	userID, accessKeyID, ok := parseCurrentUserAndAPIKeyID(c)
	if !ok {
		return
	}

	if httpx.HandleServiceError(c, h.service.DisableAPIKey(c.Request.Context(), userID, accessKeyID)) {
		return
	}

	dto.JSONResponse[any](c, http.StatusOK, "API key disabled successfully", nil)
}

// EnableAPIKey enables an API key for the current user.
//
//	@Summary		Enable API key
//	@Description	Enable a Key ID / Key Secret credential owned by the current authenticated user
//	@Tags			Authentication
//	@ID				enable_api_key
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int							true	"API key record ID"
//	@Success		200	{object}	dto.GenericResponse[any]	"API key enabled successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		404	{object}	dto.GenericResponse[any]	"API key not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/api-keys/{id}/enable [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) EnableAPIKey(c *gin.Context) {
	userID, accessKeyID, ok := parseCurrentUserAndAPIKeyID(c)
	if !ok {
		return
	}

	if httpx.HandleServiceError(c, h.service.EnableAPIKey(c.Request.Context(), userID, accessKeyID)) {
		return
	}

	dto.JSONResponse[any](c, http.StatusOK, "API key enabled successfully", nil)
}

// RevokeAPIKey permanently revokes an API key for the current user.
//
//	@Summary		Revoke API key
//	@Description	Permanently revoke a Key ID / Key Secret credential owned by the current authenticated user. Revoked API keys can no longer be re-enabled or used to exchange bearer tokens.
//	@Tags			Authentication
//	@ID				revoke_api_key
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int							true	"API key record ID"
//	@Success		200	{object}	dto.GenericResponse[any]	"API key revoked successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		404	{object}	dto.GenericResponse[any]	"API key not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/api-keys/{id}/revoke [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) RevokeAPIKey(c *gin.Context) {
	userID, accessKeyID, ok := parseCurrentUserAndAPIKeyID(c)
	if !ok {
		return
	}

	if httpx.HandleServiceError(c, h.service.RevokeAPIKey(c.Request.Context(), userID, accessKeyID)) {
		return
	}

	dto.JSONResponse[any](c, http.StatusOK, "API key revoked successfully", nil)
}

// RotateAPIKey rotates the key secret for an existing API key.
//
//	@Summary		Rotate API key secret
//	@Description	Rotate the key secret half of a Key ID / Key Secret credential owned by the current authenticated user
//	@Tags			Authentication
//	@ID				rotate_api_key
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int											true	"API key record ID"
//	@Success		200	{object}	dto.GenericResponse[APIKeyWithSecretResp]	"API key rotated successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		404	{object}	dto.GenericResponse[any]					"API key not found"
//	@Failure		500	{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/api-keys/{id}/rotate [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) RotateAPIKey(c *gin.Context) {
	userID, accessKeyID, ok := parseCurrentUserAndAPIKeyID(c)
	if !ok {
		return
	}

	resp, err := h.service.RotateAPIKey(c.Request.Context(), userID, accessKeyID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "API key rotated successfully", resp)
}

// ExchangeAPIKeyToken exchanges a signed API key request for a bearer token.
//
//	@Summary		Exchange API key for token
//	@Description	Exchange a signed Key ID / Key Secret request for a short-lived bearer token. SDK and CLI callers sign `METHOD\\nPATH\\nTIMESTAMP\\nNONCE\\nSHA256(BODY)` with the key secret and send the result via `X-Key-Id`, `X-Timestamp`, `X-Nonce`, and `X-Signature`.
//	@Tags			Authentication
//	@ID				exchange_api_key_token
//	@Produce		json
//	@Param			X-Key-Id	header		string									true	"Public key identifier"
//	@Param			X-Timestamp	header		string									true	"Unix timestamp in seconds"
//	@Param			X-Nonce		header		string									true	"Unique request nonce"
//	@Param			X-Signature	header		string									true	"Hex encoded HMAC-SHA256 signature of METHOD\\nPATH\\nTIMESTAMP\\nNONCE\\nSHA256(BODY)"
//	@Success		200			{object}	dto.GenericResponse[APIKeyTokenResp]	"API key token issued successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]				"Invalid request"
//	@Failure		401			{object}	dto.GenericResponse[any]				"Invalid signature or replayed request"
//	@Failure		500			{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/auth/api-key/token [post]
//	@x-api-type		{"sdk":"true"}
func (h *Handler) ExchangeAPIKeyToken(c *gin.Context) {
	var req APIKeyTokenReq
	req.KeyID = c.GetHeader("X-Key-Id")
	req.Timestamp = c.GetHeader("X-Timestamp")
	req.Nonce = c.GetHeader("X-Nonce")
	req.Signature = c.GetHeader("X-Signature")
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ExchangeAPIKeyToken(c.Request.Context(), &req, c.Request.Method, c.Request.URL.Path)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "API key token issued successfully", resp)
}

func parseCurrentUserAndAPIKeyID(c *gin.Context) (int, int, bool) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return 0, 0, false
	}

	accessKeyID, ok := httpx.ParsePositiveID(c, c.Param("id"), consts.URLPathID)
	if !ok {
		return 0, 0, false
	}

	return userID, accessKeyID, true
}
