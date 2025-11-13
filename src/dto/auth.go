package dto

import (
	"fmt"
	"regexp"
	"time"

	"aegis/database"
)

const (
	usernamePattern = `^[a-zA-Z0-9_]{3,20}$`
)

// RegisterReq represents user registration request
type RegisterReq struct {
	Username string `json:"username" binding:"required" example:"newuser"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
}

// Validate validates the registration request
func (req *RegisterReq) Validate() error {
	// Username validation
	usernameRegex := regexp.MustCompile(usernamePattern)
	if !usernameRegex.MatchString(req.Username) {
		return fmt.Errorf("username must be 3-20 characters and contain only letters, numbers, and underscores")
	}

	// Password validation
	if len(req.Password) == 0 {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	return nil
}

// LoginReq represents user login request
type LoginReq struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"password123"`
}

func (req *LoginReq) Validate() error {
	usernameRegex := regexp.MustCompile(usernamePattern)
	if !usernameRegex.MatchString(req.Username) {
		return fmt.Errorf("Invalid username name or password")
	}
	if req.Password == "" {
		return fmt.Errorf("Invalid username name or password")
	}
	return nil
}

// TokenRefreshReq represents token refresh request
type TokenRefreshReq struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

func (req *TokenRefreshReq) Validate() error {
	if req.Token == "" {
		return fmt.Errorf("Invalid token")
	}
	return nil
}

// ChangePasswordReq represents password change request
type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpassword123"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"newpassword123"`
}

func (req *ChangePasswordReq) Validate() error {
	if req.OldPassword == "" {
		return fmt.Errorf("old_password is required")
	}
	if len(req.OldPassword) < 8 {
		return fmt.Errorf("old_password must be at least 8 characters long")
	}
	if req.NewPassword == "" {
		return fmt.Errorf("new_password is required")
	}
	if len(req.NewPassword) < 8 {
		return fmt.Errorf("new_password must be at least 8 characters long")
	}
	return nil
}

// LoginResp represents user login response
type LoginResp struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt time.Time `json:"expires_at" example:"2024-12-31T23:59:59Z"`
	User      UserInfo  `json:"user"`
}

// TokenRefreshResp represents token refresh response
type TokenRefreshResp struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt time.Time `json:"expires_at" example:"2024-12-31T23:59:59Z"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID       int    `json:"id" example:"1"`
	Username string `json:"username" example:"admin"`
	Avatar   string `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
}

func NewUserInfo(user *database.User) *UserInfo {
	return &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Avatar:   user.Avatar,
	}
}
