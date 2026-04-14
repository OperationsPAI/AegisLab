package client

import (
	"fmt"
	"time"
)

// loginRequest matches dto.LoginReq.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// loginResponseData matches dto.LoginResp.
type loginResponseData struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar,omitempty"`
		Role     string `json:"role,omitempty"`
	} `json:"user"`
}

// tokenRefreshRequest matches dto.TokenRefreshReq.
type tokenRefreshRequest struct {
	Token string `json:"token"`
}

// tokenRefreshResponseData matches dto.TokenRefreshResp.
type tokenRefreshResponseData struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// LoginResult contains the result of a successful login.
type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	Username  string
}

// Login authenticates against the server and returns a token.
func Login(server, username, password string) (*LoginResult, error) {
	c := NewClient(server, "", 30*time.Second)

	var resp APIResponse[loginResponseData]
	err := c.Post("/api/v2/auth/login", loginRequest{
		Username: username,
		Password: password,
	}, &resp)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return &LoginResult{
		Token:     resp.Data.Token,
		ExpiresAt: resp.Data.ExpiresAt,
		Username:  resp.Data.User.Username,
	}, nil
}

// RefreshToken refreshes an existing JWT token.
func RefreshToken(server, currentToken string) (string, time.Time, error) {
	c := NewClient(server, currentToken, 30*time.Second)

	var resp APIResponse[tokenRefreshResponseData]
	err := c.Post("/api/v2/auth/refresh", tokenRefreshRequest{
		Token: currentToken,
	}, &resp)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token refresh failed: %w", err)
	}

	return resp.Data.Token, resp.Data.ExpiresAt, nil
}

// ProfileData represents the user profile returned by GET /api/v2/auth/profile.
type ProfileData struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar,omitempty"`
	Role     string `json:"role,omitempty"`
}

// GetProfile fetches the current user's profile.
func GetProfile(server, token string) (*ProfileData, error) {
	c := NewClient(server, token, 30*time.Second)

	var resp APIResponse[ProfileData]
	if err := c.Get("/api/v2/auth/profile", &resp); err != nil {
		return nil, fmt.Errorf("get profile failed: %w", err)
	}

	return &resp.Data, nil
}

// IsTokenExpired checks whether the stored token expiry has passed.
func IsTokenExpired(expiry time.Time) bool {
	if expiry.IsZero() {
		return false // unknown expiry — assume valid
	}
	return time.Now().After(expiry)
}
