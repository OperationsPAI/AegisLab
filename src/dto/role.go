package dto

import (
	"fmt"
	"time"

	"aegis/database"
)

// CreateRoleRequest represents role creation request
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required" example:"data_analyst"`
	DisplayName string `json:"display_name" binding:"required" example:"Data Analyst"`
	Description string `json:"description,omitempty" example:"Role for data analysis operations"`
	Type        string `json:"type" binding:"required,oneof=system custom" example:"custom"`
}

// UpdateRoleRequest represents role update request
type UpdateRoleRequest struct {
	DisplayName string `json:"display_name,omitempty" example:"Senior Data Analyst"`
	Description string `json:"description,omitempty" example:"Updated role description"`
	Status      *int   `json:"status,omitempty" example:"1"`
}

// RoleResponse represents role response
type RoleResponse struct {
	ID          int                  `json:"id" example:"1"`
	Name        string               `json:"name" example:"data_analyst"`
	DisplayName string               `json:"display_name" example:"Data Analyst"`
	Description string               `json:"description" example:"Role for data analysis operations"`
	Type        string               `json:"type" example:"custom"`
	IsSystem    bool                 `json:"is_system" example:"false"`
	Status      int                  `json:"status" example:"1"`
	CreatedAt   time.Time            `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time            `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
	UserCount   int                  `json:"user_count,omitempty" example:"5"`
}

// RoleListRequest represents role list query parameters
type RoleListRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1" example:"1"`
	Size     int    `form:"size,default=20" binding:"min=1,max=100" example:"20"`
	Type     string `form:"type" binding:"omitempty,oneof=system custom" example:"custom"`
	Status   *int   `form:"status" example:"1"`
	IsSystem *bool  `form:"is_system" example:"false"`
	Name     string `form:"name" example:"admin"`
}

// RoleListResponse represents paginated role list response
type RoleListResponse struct {
	Items      []RoleResponse `json:"items"`
	Pagination PaginationInfo `json:"pagination"`
}

// AssignPermissionToRoleRequest represents permission assignment to role request
type AssignPermissionToRoleRequest struct {
	PermissionIDs []int `json:"permission_ids" binding:"required,min=1" example:"[1,2,3]"`
}

// RemovePermissionFromRoleRequest represents permission removal from role request
type RemovePermissionFromRoleRequest struct {
	PermissionIDs []int `json:"permission_ids" binding:"required,min=1" example:"[1,2]"`
}

// RoleSearchRequest represents advanced role search with complex filtering
type RoleSearchRequest struct {
	AdvancedSearchRequest

	// Role-specific filter shortcuts
	NamePattern        string       `json:"name_pattern,omitempty"`         // Role name fuzzy match
	DisplayNamePattern string       `json:"display_name_pattern,omitempty"` // Display name fuzzy match
	DescriptionPattern string       `json:"description_pattern,omitempty"`  // Description fuzzy match
	Types              []string     `json:"types,omitempty"`                // Role type filter
	IsSystem           *bool        `json:"is_system,omitempty"`            // Whether system role
	PermissionIDs      []int        `json:"permission_ids,omitempty"`       // Permission ID filter
	UserCount          *NumberRange `json:"user_count,omitempty"`           // User count range
}

// ConvertToSearchRequest converts RoleSearchRequest to SearchRequest with role-specific filters
func (rsr *RoleSearchRequest) ConvertToSearchRequest() *SearchRequest {
	sr := rsr.ConvertAdvancedToSearch()

	// Add role-specific filters
	if rsr.NamePattern != "" {
		sr.AddFilter("name", OpLike, rsr.NamePattern)
	}

	if rsr.DisplayNamePattern != "" {
		sr.AddFilter("display_name", OpLike, rsr.DisplayNamePattern)
	}

	if rsr.DescriptionPattern != "" {
		sr.AddFilter("description", OpLike, rsr.DescriptionPattern)
	}

	if len(rsr.Types) > 0 {
		values := make([]string, len(rsr.Types))
		for i, v := range rsr.Types {
			values[i] = fmt.Sprintf("%v", v)
		}
		sr.Filters = append(sr.Filters, SearchFilter{
			Field:    "type",
			Operator: OpIn,
			Values:   values,
		})
	}

	if rsr.IsSystem != nil {
		sr.AddFilter("is_system", OpEqual, *rsr.IsSystem)
	}

	if len(rsr.PermissionIDs) > 0 {
		values := make([]string, len(rsr.PermissionIDs))
		for i, v := range rsr.PermissionIDs {
			values[i] = fmt.Sprintf("%v", v)
		}
		sr.Filters = append(sr.Filters, SearchFilter{
			Field:    "permission_id",
			Operator: OpIn,
			Values:   values,
		})
	}

	// Note: UserCount filtering would require a subquery or join
	// This could be implemented in the repository layer

	return sr
}

// RoleSearchFilters represents simple search filters for backward compatibility
type RoleSearchFilters struct {
	Name        []string `json:"name,omitempty"`
	DisplayName []string `json:"display_name,omitempty"`
	Type        []string `json:"type,omitempty"`
	Status      []int    `json:"status,omitempty"`
	IsSystem    []bool   `json:"is_system,omitempty"`
}

// ConvertFromRole converts database Role to RoleResponse DTO
func (r *RoleResponse) ConvertFromRole(role *database.Role) {
	r.ID = role.ID
	r.Name = role.Name
	r.DisplayName = role.DisplayName
	r.Description = role.Description
	r.Type = role.Type
	r.IsSystem = role.IsSystem
	r.Status = role.Status
	r.CreatedAt = role.CreatedAt
	r.UpdatedAt = role.UpdatedAt
}
