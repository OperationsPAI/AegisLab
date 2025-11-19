package dto

import (
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/database"
)

// CreateRoleReq represents role creation request
type CreateRoleReq struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description,omitempty" binding:"omitempty"`
}

// ConvertToRole converts CreateRoleReq to database Role model
func (req *CreateRoleReq) ConvertToRole() *database.Role {
	return &database.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsSystem:    false,
		Status:      consts.CommonEnabled,
	}
}

// ListRoleReq represents role list query parameters
type ListRoleReq struct {
	PaginationReq
	IsSystem *bool              `form:"is_system" binding:"omitempty"`
	Status   *consts.StatusType `form:"status" binding:"omitempty"`
}

func (req *ListRoleReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	return validateStatusField(req.Status, false)
}

// SearchRoleReq represents advanced role search with complex filtering
type SearchRoleReq struct {
	AdvancedSearchReq

	// Role-specific filter shortcuts
	NamePattern        string       `json:"name_pattern" binding:"omitempty"`         // Role name fuzzy match
	DisplayNamePattern string       `json:"display_name_pattern" binding:"omitempty"` // Display name fuzzy match
	DescriptionPattern string       `json:"description_pattern" binding:"omitempty"`  // Description fuzzy match
	IsSystem           *bool        `json:"is_system" binding:"omitempty"`            // Whether system role
	PermissionIDs      []int        `json:"permission_ids" binding:"omitempty"`       // Permission ID filter
	UserCount          *NumberRange `json:"user_count" binding:"omitempty"`           // User count range
}

// ConvertToSearchRequest converts RoleSearchReq to SearchRequest with role-specific filters
func (rsr *SearchRoleReq) ConvertToSearchRequest() *SearchReq {
	sr := rsr.ConvertAdvancedToSearch()

	if rsr.NamePattern != "" {
		sr.AddFilter("name", OpLike, rsr.NamePattern)
	}
	if rsr.DisplayNamePattern != "" {
		sr.AddFilter("display_name", OpLike, rsr.DisplayNamePattern)
	}
	if rsr.DescriptionPattern != "" {
		sr.AddFilter("description", OpLike, rsr.DescriptionPattern)
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

	return sr
}

// UpdateRoleReq represents role update request
type UpdateRoleReq struct {
	DisplayName *string            `json:"display_name" binding:"omitempty"`
	Description *string            `json:"description" binding:"omitempty"`
	Status      *consts.StatusType `json:"status" binding:"omitempty"`
}

func (req *UpdateRoleReq) Validate() error {
	if req.DisplayName != nil {
		if *req.DisplayName != "" {
			*req.DisplayName = strings.TrimSpace(*req.DisplayName)
		}
	}
	return validateStatusField(req.Status, true)
}

func (req *UpdateRoleReq) PatchRoleModel(target *database.Role) {
	if req.DisplayName != nil {
		target.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		target.Description = *req.Description
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
}

// AssignRolePermissionReq represents request to assign permissions to a role
type AssignRolePermissionReq struct {
	PermissionIDs []int `json:"permission_ids" binding:"required,min=1,non_zero_int_slice"`
}

// RemoveRolePermissionReq represents request to remove permissions from a role
type RemoveRolePermissionReq struct {
	PermissionIDs []int `json:"permission_ids" binding:"required,min=1,non_zero_int_slice"`
}

// RoleResp represents role response
type RoleResp struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Type        string    `json:"type"`
	IsSystem    bool      `json:"is_system"`
	Status      string    `json:"status"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewRoleResp converts database Role to RoleResp DTO
func NewRoleResp(role *database.Role) *RoleResp {
	return &RoleResp{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		IsSystem:    role.IsSystem,
		Status:      consts.GetStatusTypeName(role.Status),
		UpdatedAt:   role.UpdatedAt,
	}
}

type RoleDetailResp struct {
	RoleResp

	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UserCount   int64     `json:"user_count"`

	Permissions []PermissionResp `json:"permissions"`
}

func NewRoleDetailResp(role *database.Role) *RoleDetailResp {
	resp := &RoleDetailResp{
		RoleResp:    *NewRoleResp(role),
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
	}
	return resp
}

// ListRoleResp represents paginated list of roles
type ListRoleResp struct {
	Items      []RoleResp     `json:"items"`
	Pagination PaginationInfo `json:"pagination"`
}
