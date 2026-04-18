package authmodule

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	usermodule "aegis/module/user"
	"aegis/utils"
)

const usernamePattern = `^[a-zA-Z0-9_]{3,20}$`

type RegisterReq struct {
	Username string `json:"username" binding:"required" example:"newuser"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
}

func (req *RegisterReq) Validate() error {
	usernameRegex := regexp.MustCompile(usernamePattern)
	if !usernameRegex.MatchString(req.Username) {
		return fmt.Errorf("username must be 3-20 characters and contain only letters, numbers, and underscores")
	}
	if len(req.Password) == 0 {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	return nil
}

type LoginReq struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"password123"`
}

func (req *LoginReq) Validate() error {
	usernameRegex := regexp.MustCompile(usernamePattern)
	if !usernameRegex.MatchString(req.Username) {
		return fmt.Errorf("invalid username or password")
	}
	if req.Password == "" {
		return fmt.Errorf("invalid username or password")
	}
	return nil
}

type TokenRefreshReq struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

func (req *TokenRefreshReq) Validate() error {
	if req.Token == "" {
		return fmt.Errorf("invalid token")
	}
	return nil
}

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

type CreateAPIKeyReq struct {
	Name        string     `json:"name" binding:"required" example:"ci-bot"`
	Description string     `json:"description,omitempty" example:"SDK credential for CI pipeline"`
	Scopes      []string   `json:"scopes,omitempty" example:"[\"*\"]"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" example:"2026-12-31T23:59:59Z"`
}

func (req *CreateAPIKeyReq) Validate() error {
	if req == nil {
		return fmt.Errorf("request is required")
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(req.Name) > 128 {
		return fmt.Errorf("name must be no more than 128 characters long")
	}
	normalizedScopes, err := normalizeAPIKeyScopes(req.Scopes)
	if err != nil {
		return err
	}
	req.Scopes = normalizedScopes
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("expires_at must be in the future")
	}
	return nil
}

type ListAPIKeyReq struct {
	dto.PaginationReq
}

func (req *ListAPIKeyReq) Validate() error {
	if req == nil {
		return fmt.Errorf("request is required")
	}
	return req.PaginationReq.Validate()
}

type APIKeyTokenReq struct {
	KeyID     string `header:"X-Key-Id" example:"pk_1234567890abcdef"`
	Timestamp string `header:"X-Timestamp" example:"1713333333"`
	Nonce     string `header:"X-Nonce" example:"abc123"`
	Signature string `header:"X-Signature" example:"4cf2f2cbb93d..."`
}

func (req *APIKeyTokenReq) Validate() error {
	if req == nil {
		return fmt.Errorf("request is required")
	}
	req.KeyID = strings.TrimSpace(req.KeyID)
	req.Timestamp = strings.TrimSpace(req.Timestamp)
	req.Nonce = strings.TrimSpace(req.Nonce)
	req.Signature = strings.ToLower(strings.TrimSpace(req.Signature))
	if req.KeyID == "" || req.Timestamp == "" || req.Nonce == "" || req.Signature == "" {
		return fmt.Errorf("X-Key-Id, X-Timestamp, X-Nonce and X-Signature are required")
	}
	if _, err := strconv.ParseInt(req.Timestamp, 10, 64); err != nil {
		return fmt.Errorf("X-Timestamp must be a unix timestamp in seconds")
	}
	if len(req.Nonce) > 128 {
		return fmt.Errorf("X-Nonce must be no more than 128 characters long")
	}
	return nil
}

func (req *APIKeyTokenReq) TimestampUnix() (int64, error) {
	return strconv.ParseInt(req.Timestamp, 10, 64)
}

func (req *APIKeyTokenReq) CanonicalString(method, path string) string {
	return strings.Join([]string{
		strings.ToUpper(method),
		path,
		req.Timestamp,
		req.Nonce,
		utils.SHA256Hex(nil),
	}, "\n")
}

type LoginResp struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt time.Time `json:"expires_at" example:"2024-12-31T23:59:59Z"`
	User      UserInfo  `json:"user"`
}

type TokenRefreshResp struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt time.Time `json:"expires_at" example:"2024-12-31T23:59:59Z"`
}

type APIKeyInfo struct {
	ID          int               `json:"id" example:"12"`
	Name        string            `json:"name" example:"ci-bot"`
	Description string            `json:"description,omitempty" example:"SDK credential for CI pipeline"`
	KeyID       string            `json:"key_id" example:"pk_1234567890abcdef"`
	Scopes      []string          `json:"scopes,omitempty" example:"[\"*\"]"`
	Status      consts.StatusType `json:"status" example:"1"`
	RevokedAt   *time.Time        `json:"revoked_at,omitempty" example:"2026-04-17T12:30:00Z"`
	LastUsedAt  *time.Time        `json:"last_used_at,omitempty" example:"2026-04-17T12:00:00Z"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty" example:"2026-12-31T23:59:59Z"`
	CreatedAt   time.Time         `json:"created_at" example:"2026-04-17T11:00:00Z"`
	UpdatedAt   time.Time         `json:"updated_at" example:"2026-04-17T11:00:00Z"`
}

func NewAPIKeyInfo(key *model.APIKey) *APIKeyInfo {
	if key == nil {
		return nil
	}
	return &APIKeyInfo{
		ID:          key.ID,
		Name:        key.Name,
		Description: key.Description,
		KeyID:       key.KeyID,
		Scopes:      append([]string(nil), key.Scopes...),
		Status:      key.Status,
		RevokedAt:   key.RevokedAt,
		LastUsedAt:  key.LastUsedAt,
		ExpiresAt:   key.ExpiresAt,
		CreatedAt:   key.CreatedAt,
		UpdatedAt:   key.UpdatedAt,
	}
}

type APIKeyWithSecretResp struct {
	APIKeyInfo
	KeySecret string `json:"key_secret" example:"ks_abcdefghijklmnopqrstuvwxyz123456"`
}

type ListAPIKeyResp struct {
	Items      []APIKeyInfo       `json:"items"`
	Pagination dto.PaginationInfo `json:"pagination"`
}

type APIKeyTokenResp struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.api_key.jwt"`
	TokenType string    `json:"token_type" example:"Bearer"`
	ExpiresAt time.Time `json:"expires_at" example:"2026-04-17T12:00:00Z"`
	AuthType  string    `json:"auth_type" example:"api_key"`
	KeyID     string    `json:"key_id" example:"pk_1234567890abcdef"`
}

const defaultAPIKeyScope = "*"

func normalizeAPIKeyScopes(scopes []string) ([]string, error) {
	if len(scopes) == 0 {
		return []string{defaultAPIKeyScope}, nil
	}

	normalized := make([]string, 0, len(scopes))
	seen := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			return nil, fmt.Errorf("scopes cannot contain empty items")
		}
		if len(scope) > 128 {
			return nil, fmt.Errorf("scope %q must be no more than 128 characters long", scope)
		}
		if _, exists := seen[scope]; exists {
			continue
		}
		seen[scope] = struct{}{}
		normalized = append(normalized, scope)
	}
	if len(normalized) == 0 {
		return []string{defaultAPIKeyScope}, nil
	}
	return normalized, nil
}

type UserProfileResp struct {
	ID          int        `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	FullName    string     `json:"full_name"`
	Avatar      string     `json:"avatar,omitempty"`
	Phone       string     `json:"phone,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`

	ContainerRoles []usermodule.UserContainerInfo `json:"container_roles,omitempty"`
	DatasetRoles   []usermodule.UserDatasetInfo   `json:"dataset_roles,omitempty"`
	ProjectRoles   []usermodule.UserProjectInfo   `json:"project_roles,omitempty"`
}

func NewUserProfileResp(user *model.User) *UserProfileResp {
	return &UserProfileResp{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FullName:    user.FullName,
		Avatar:      user.Avatar,
		Phone:       user.Phone,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
	}
}

type UserInfo struct {
	ID       int    `json:"id" example:"1"`
	Username string `json:"username" example:"admin"`
	Avatar   string `json:"avatar,omitempty"`
	Role     string `json:"role,omitempty"`
}

func NewUserInfo(user *model.User) *UserInfo {
	return &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Avatar:   user.Avatar,
	}
}
