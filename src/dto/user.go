package dto

import (
	"time"

	"github.com/LGU-SE-Internal/rcabench/database"
)

// CreateUserRequest represents user creation request
type CreateUserRequest struct {
	Username string `json:"username" binding:"required" example:"newuser"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
	FullName string `json:"full_name" binding:"required" example:"John Doe"`
	Phone    string `json:"phone,omitempty" example:"+1234567890"`
	Avatar   string `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	Email    string `json:"email,omitempty" binding:"omitempty,email" example:"newemail@example.com"`
	FullName string `json:"full_name,omitempty" example:"Jane Doe"`
	Phone    string `json:"phone,omitempty" example:"+0987654321"`
	Avatar   string `json:"avatar,omitempty" example:"https://example.com/new-avatar.jpg"`
	Status   *int   `json:"status,omitempty" example:"1"`
	IsActive *bool  `json:"is_active,omitempty" example:"true"`
}

// UserListRequest represents user list query parameters
type UserListRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1" example:"1"`
	Size     int    `form:"size,default=20" binding:"min=1,max=100" example:"20"`
	Status   *int   `form:"status" example:"1"`
	IsActive *bool  `form:"is_active" example:"true"`
	Username string `form:"username" example:"admin"`
	Email    string `form:"email" example:"admin@example.com"`
	FullName string `form:"full_name" example:"Administrator"`
}

// UserResponse represents user response with role and project information
type UserResponse struct {
	ID           int                   `json:"id" example:"1"`
	Username     string                `json:"username" example:"admin"`
	Email        string                `json:"email" example:"admin@example.com"`
	FullName     string                `json:"full_name" example:"Administrator"`
	Avatar       string                `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	Phone        string                `json:"phone,omitempty" example:"+1234567890"`
	Status       int                   `json:"status" example:"1"`
	IsActive     bool                  `json:"is_active" example:"true"`
	LastLoginAt  *time.Time            `json:"last_login_at,omitempty" example:"2024-01-01T12:00:00Z"`
	CreatedAt    time.Time             `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt    time.Time             `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	GlobalRoles  []RoleResponse        `json:"global_roles,omitempty"`
	ProjectRoles []UserProjectResponse `json:"project_roles,omitempty"`
	Permissions  []PermissionResponse  `json:"permissions,omitempty"`
}

// UserListResponse represents paginated user list response
type UserListResponse struct {
	Items      []UserResponse `json:"items"`
	Pagination PaginationInfo `json:"pagination"`
}

// UserProjectResponse represents user-project relationship
type UserProjectResponse struct {
	ProjectID   int          `json:"project_id" example:"1"`
	ProjectName string       `json:"project_name" example:"Project Alpha"`
	Role        RoleResponse `json:"role"`
	JoinedAt    time.Time    `json:"joined_at" example:"2024-01-01T00:00:00Z"`
	Status      int          `json:"status" example:"1"`
}

// AssignUserToProjectRequest represents user-project assignment request
type AssignUserToProjectRequest struct {
	ProjectID int `json:"project_id" binding:"required" example:"1"`
	RoleID    int `json:"role_id" binding:"required" example:"2"`
}

// AssignRoleToUserRequest represents role assignment request
type AssignRoleToUserRequest struct {
	RoleID int `json:"role_id" binding:"required" example:"1"`
}

// UserSearchRequest represents advanced user search with complex filtering
type UserSearchRequest struct {
	AdvancedSearchRequest

	// User-specific filter shortcuts
	UsernamePattern string     `json:"username_pattern,omitempty"` // Username fuzzy match
	EmailPattern    string     `json:"email_pattern,omitempty"`    // Email fuzzy match
	FullNamePattern string     `json:"fullname_pattern,omitempty"` // Full name fuzzy match
	RoleIDs         []int      `json:"role_ids,omitempty"`         // Role ID filter
	ProjectIDs      []int      `json:"project_ids,omitempty"`      // Project ID filter
	Departments     []string   `json:"departments,omitempty"`      // Department filter
	LastLoginRange  *DateRange `json:"last_login_range,omitempty"` // Last login time range
}

// ConvertToSearchRequest converts UserSearchRequest to SearchRequest with user-specific filters
func (usr *UserSearchRequest) ConvertToSearchRequest() *SearchRequest {
	sr := usr.ConvertAdvancedToSearch()

	// Add user-specific filters
	if usr.UsernamePattern != "" {
		sr.AddFilter("username", OpLike, usr.UsernamePattern)
	}

	if usr.EmailPattern != "" {
		sr.AddFilter("email", OpLike, usr.EmailPattern)
	}

	if usr.FullNamePattern != "" {
		sr.AddFilter("full_name", OpLike, usr.FullNamePattern)
	}

	if len(usr.RoleIDs) > 0 {
		values := make([]interface{}, len(usr.RoleIDs))
		for i, v := range usr.RoleIDs {
			values[i] = v
		}
		sr.Filters = append(sr.Filters, SearchFilter{
			Field:    "role_id",
			Operator: OpIn,
			Values:   values,
		})
	}

	if len(usr.ProjectIDs) > 0 {
		values := make([]interface{}, len(usr.ProjectIDs))
		for i, v := range usr.ProjectIDs {
			values[i] = v
		}
		sr.Filters = append(sr.Filters, SearchFilter{
			Field:    "project_id",
			Operator: OpIn,
			Values:   values,
		})
	}

	if len(usr.Departments) > 0 {
		values := make([]interface{}, len(usr.Departments))
		for i, v := range usr.Departments {
			values[i] = v
		}
		sr.Filters = append(sr.Filters, SearchFilter{
			Field:    "department",
			Operator: OpIn,
			Values:   values,
		})
	}

	if usr.LastLoginRange != nil {
		if usr.LastLoginRange.From != nil && usr.LastLoginRange.To != nil {
			sr.AddFilter("last_login_at", OpDateBetween, []interface{}{usr.LastLoginRange.From, usr.LastLoginRange.To})
		} else if usr.LastLoginRange.From != nil {
			sr.AddFilter("last_login_at", OpDateAfter, usr.LastLoginRange.From)
		} else if usr.LastLoginRange.To != nil {
			sr.AddFilter("last_login_at", OpDateBefore, usr.LastLoginRange.To)
		}
	}

	return sr
}

// UserSearchFilters represents simple search filters for backward compatibility
type UserSearchFilters struct {
	Username    []string `json:"username,omitempty"`
	Email       []string `json:"email,omitempty"`
	FullName    []string `json:"full_name,omitempty"`
	Status      []int    `json:"status,omitempty"`
	IsActive    []bool   `json:"is_active,omitempty"`
	RoleIDs     []int    `json:"role_ids,omitempty"`
	ProjectIDs  []int    `json:"project_ids,omitempty"`
	Departments []string `json:"departments,omitempty"`
}

// ConvertFromUser converts database User to UserResponse DTO
func (u *UserResponse) ConvertFromUser(user *database.User) {
	u.ID = user.ID
	u.Username = user.Username
	u.Email = user.Email
	u.FullName = user.FullName
	u.Avatar = user.Avatar
	u.Phone = user.Phone
	u.Status = user.Status
	u.IsActive = user.IsActive
	u.LastLoginAt = user.LastLoginAt
	u.CreatedAt = user.CreatedAt
	u.UpdatedAt = user.UpdatedAt
}

// ConvertFromUserProject converts database UserProject to UserProjectResponse DTO
func (up *UserProjectResponse) ConvertFromUserProject(userProject *database.UserProject) {
	up.ProjectID = userProject.ProjectID
	up.JoinedAt = userProject.JoinedAt
	up.Status = userProject.Status

	if userProject.Project != nil {
		up.ProjectName = userProject.Project.Name
	}

	if userProject.Role != nil {
		up.Role.ConvertFromRole(userProject.Role)
	}
}
