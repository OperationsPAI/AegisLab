package dto

import (
	"fmt"
	"regexp"
	"time"

	"aegis/database"
)

// LoginRequest represents user login request
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse represents user login response
type LoginResponse struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt time.Time `json:"expires_at" example:"2024-12-31T23:59:59Z"`
	User      UserInfo  `json:"user"`
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"newuser"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
	FullName string `json:"full_name" binding:"required" example:"John Doe"`
	Phone    string `json:"phone,omitempty" example:"+1234567890"`
}

// Validate validates the registration request
func (r *RegisterRequest) Validate() error {
	// Username validation
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
	if !usernameRegex.MatchString(r.Username) {
		return fmt.Errorf("username must be 3-20 characters and contain only letters, numbers, and underscores")
	}

	// Password validation
	if len(r.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Phone validation (if provided)
	if r.Phone != "" {
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
		if !phoneRegex.MatchString(r.Phone) {
			return fmt.Errorf("invalid phone number format")
		}
	}

	return nil
}

// TokenRefreshRequest represents token refresh request
type TokenRefreshRequest struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// TokenRefreshResponse represents token refresh response
type TokenRefreshResponse struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt time.Time `json:"expires_at" example:"2024-12-31T23:59:59Z"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpassword123"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"newpassword123"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID        int       `json:"id" example:"1"`
	Username  string    `json:"username" example:"admin"`
	Email     string    `json:"email" example:"admin@example.com"`
	FullName  string    `json:"full_name" example:"Administrator"`
	Avatar    string    `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	Phone     string    `json:"phone,omitempty" example:"+1234567890"`
	Status    int       `json:"status" example:"1"`
	IsActive  bool      `json:"is_active" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// ConvertFromUser converts database User to UserInfo DTO
func (u *UserInfo) ConvertFromUser(user *database.User) {
	u.ID = user.ID
	u.Username = user.Username
	u.Email = user.Email
	u.FullName = user.FullName
	u.Avatar = user.Avatar
	u.Phone = user.Phone
	u.Status = user.Status
	u.IsActive = user.IsActive
	u.CreatedAt = user.CreatedAt
	u.UpdatedAt = user.UpdatedAt
}

// LogoutRequest represents user logout request
type LogoutRequest struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
