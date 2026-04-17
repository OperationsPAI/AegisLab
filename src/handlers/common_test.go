package handlers

import (
	"aegis/consts"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type errorResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	return c, w
}

func parseErrorResp(t *testing.T, w *httptest.ResponseRecorder) errorResp {
	t.Helper()
	var resp errorResp
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "failed to parse response body")
	return resp
}

// ---------------------------------------------------------------------------
// HandleServiceError tests
// ---------------------------------------------------------------------------

func TestHandleServiceError_AuthenticationFailed(t *testing.T) {
	c, w := setupTestContext()
	wrapped := fmt.Errorf("login failed: %w", consts.ErrAuthenticationFailed)

	handled := HandleServiceError(c, wrapped)

	assert.True(t, handled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resp := parseErrorResp(t, w)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestHandleServiceError_BadRequest(t *testing.T) {
	c, w := setupTestContext()
	wrapped := fmt.Errorf("invalid input: %w", consts.ErrBadRequest)

	handled := HandleServiceError(c, wrapped)

	assert.True(t, handled)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseErrorResp(t, w)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestHandleServiceError_NotFound(t *testing.T) {
	c, w := setupTestContext()
	wrapped := fmt.Errorf("project not found: %w", consts.ErrNotFound)

	handled := HandleServiceError(c, wrapped)

	assert.True(t, handled)
	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseErrorResp(t, w)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestHandleServiceError_AlreadyExists(t *testing.T) {
	c, w := setupTestContext()
	wrapped := fmt.Errorf("duplicate entry: %w", consts.ErrAlreadyExists)

	handled := HandleServiceError(c, wrapped)

	assert.True(t, handled)
	assert.Equal(t, http.StatusConflict, w.Code)
	resp := parseErrorResp(t, w)
	assert.Equal(t, http.StatusConflict, resp.Code)
}

func TestHandleServiceError_PermissionDenied(t *testing.T) {
	c, w := setupTestContext()
	wrapped := fmt.Errorf("access denied: %w", consts.ErrPermissionDenied)

	handled := HandleServiceError(c, wrapped)

	assert.True(t, handled)
	assert.Equal(t, http.StatusForbidden, w.Code)
	resp := parseErrorResp(t, w)
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestHandleServiceError_Internal_SanitizesMessage(t *testing.T) {
	c, w := setupTestContext()
	wrapped := fmt.Errorf("db connection pool exhausted: %w", consts.ErrInternal)

	handled := HandleServiceError(c, wrapped)

	assert.True(t, handled)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := parseErrorResp(t, w)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, "Internal server error", resp.Message, "internal errors must be sanitized")
	assert.NotContains(t, resp.Message, "db connection pool", "internal details must not leak")
}

func TestHandleServiceError_WrappedInternal_SanitizesMessage(t *testing.T) {
	c, w := setupTestContext()
	innerErr := errors.New("redis connection refused")
	wrapped := fmt.Errorf("%w: %v", consts.ErrInternal, innerErr)

	handled := HandleServiceError(c, wrapped)

	assert.True(t, handled)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := parseErrorResp(t, w)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, "Internal server error", resp.Message, "wrapped internal errors must be sanitized")
	assert.NotContains(t, resp.Message, "redis", "internal details must not leak through wrapped errors")
}

func TestHandleServiceError_NilError_ReturnsFalse(t *testing.T) {
	c, w := setupTestContext()

	handled := HandleServiceError(c, nil)

	assert.False(t, handled)
	assert.Equal(t, http.StatusOK, w.Code, "response should not be written for nil error")
}

func TestHandleServiceError_UnknownError_Returns500(t *testing.T) {
	c, w := setupTestContext()
	unknown := errors.New("something completely unexpected")

	handled := HandleServiceError(c, unknown)

	assert.True(t, handled)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := parseErrorResp(t, w)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, "An unexpected error occurred", resp.Message,
		"unknown errors should return a generic message")
}

func TestHandleServiceError_WrappedMessage_UsesUserFriendly(t *testing.T) {
	c, w := setupTestContext()
	// Two-level wrap: outermost -> user-friendly message -> sentinel
	wrapped := fmt.Errorf("project 42 not found: %w", consts.ErrNotFound)

	handled := HandleServiceError(c, wrapped)

	assert.True(t, handled)
	resp := parseErrorResp(t, w)
	assert.Contains(t, resp.Message, "project 42 not found",
		"user-friendly wrapper message should be used")
}

// ---------------------------------------------------------------------------
// ParsePositiveID tests
// ---------------------------------------------------------------------------

func TestParsePositiveID_InvalidString(t *testing.T) {
	c, w := setupTestContext()

	id, ok := ParsePositiveID(c, "abc", "project_id")

	assert.False(t, ok)
	assert.Equal(t, 0, id)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseErrorResp(t, w)
	assert.Contains(t, resp.Message, "abc",
		"error message should include the rejected value")
}

func TestParsePositiveID_Zero(t *testing.T) {
	c, w := setupTestContext()

	id, ok := ParsePositiveID(c, "0", "project_id")

	assert.False(t, ok)
	assert.Equal(t, 0, id)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseErrorResp(t, w)
	assert.Contains(t, resp.Message, "0",
		"error message should include the rejected value")
}

func TestParsePositiveID_Negative(t *testing.T) {
	c, w := setupTestContext()

	id, ok := ParsePositiveID(c, "-1", "project_id")

	assert.False(t, ok)
	assert.Equal(t, 0, id)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseErrorResp(t, w)
	assert.Contains(t, resp.Message, "-1",
		"error message should include the rejected value")
}

func TestParsePositiveID_Valid(t *testing.T) {
	c, w := setupTestContext()

	id, ok := ParsePositiveID(c, "42", "project_id")

	assert.True(t, ok)
	assert.Equal(t, 42, id)
	assert.Equal(t, http.StatusOK, w.Code, "no error response should be written for valid ID")
}
