package rbac

import (
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
)

// CreateRoleReq represents role creation request.
type CreateRoleReq struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description,omitempty" binding:"omitempty"`
}

func (req *CreateRoleReq) ConvertToRole() *model.Role {
	return &model.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsSystem:    false,
		Status:      consts.CommonEnabled,
	}
}

// ListRoleReq represents role list query parameters.
type ListRoleReq struct {
	dto.PaginationReq
	IsSystem *bool              `form:"is_system" binding:"omitempty"`
	Status   *consts.StatusType `form:"status" binding:"omitempty"`
}

func (req *ListRoleReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	return validateStatus(req.Status, false)
}

// UpdateRoleReq represents role update request.
type UpdateRoleReq struct {
	DisplayName *string            `json:"display_name" binding:"omitempty"`
	Description *string            `json:"description" binding:"omitempty"`
	Status      *consts.StatusType `json:"status" binding:"omitempty"`
}

func (req *UpdateRoleReq) Validate() error {
	if req.DisplayName != nil && *req.DisplayName != "" {
		*req.DisplayName = strings.TrimSpace(*req.DisplayName)
	}
	return validateStatus(req.Status, true)
}

func (req *UpdateRoleReq) PatchRoleModel(target *model.Role) {
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

// AssignRolePermissionReq represents request to assign permissions to a role.
type AssignRolePermissionReq struct {
	PermissionIDs []int `json:"permission_ids" binding:"required,min=1,non_zero_int_slice"`
}

// RemoveRolePermissionReq represents request to remove permissions from a role.
type RemoveRolePermissionReq struct {
	PermissionIDs []int `json:"permission_ids" binding:"required,min=1,non_zero_int_slice"`
}

// ListResourceReq represents request for listing resources.
type ListResourceReq struct {
	dto.PaginationReq

	Type     *consts.ResourceType     `form:"type" binding:"omitempty"`
	Category *consts.ResourceCategory `form:"category" binding:"omitempty"`
}

func (req *ListResourceReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if req.Type != nil {
		if _, exists := consts.ValidResourceTypes[*req.Type]; !exists {
			return fmt.Errorf("invalid resource type: %d", *req.Type)
		}
	}
	if req.Category != nil {
		if _, exists := consts.ValidResourceCategories[*req.Category]; !exists {
			return fmt.Errorf("invalid resource category: %d", *req.Category)
		}
	}
	return nil
}

// ResourceResp represents an RBAC resource response.
type ResourceResp struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	ParentID    *int   `json:"parent_id,omitempty"`
}

func NewResourceResp(resource *model.Resource) *ResourceResp {
	return &ResourceResp{
		ID:          resource.ID,
		Name:        resource.Name.String(),
		DisplayName: resource.DisplayName,
		Type:        consts.GetResourceTypeName(resource.Type),
		Category:    consts.GetResourceCategoryName(resource.Category),
		ParentID:    resource.ParentID,
	}
}

// ResourceDetailResp represents a detailed RBAC resource response.
type ResourceDetailResp struct {
	ResourceResp

	Description string `json:"description,omitempty"`
}

func NewResourceDetailResp(resource *model.Resource) *ResourceDetailResp {
	return &ResourceDetailResp{
		ResourceResp: *NewResourceResp(resource),
		Description:  resource.Description,
	}
}

func validateStatus(statusPtr *consts.StatusType, isMutation bool) error {
	if statusPtr == nil {
		return nil
	}

	status := *statusPtr
	if _, exists := consts.ValidStatuses[status]; !exists {
		return fmt.Errorf("invalid status value: %d", status)
	}
	if isMutation && status == consts.CommonDeleted {
		return fmt.Errorf("status value cannot be set to deleted (%d) directly through this update/create operation", consts.CommonDeleted)
	}
	return nil
}

// RoleResp represents role response.
type RoleResp struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Type        string    `json:"type"`
	IsSystem    bool      `json:"is_system"`
	Status      string    `json:"status"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewRoleResp(role *model.Role) *RoleResp {
	return &RoleResp{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		IsSystem:    role.IsSystem,
		Status:      consts.GetStatusTypeName(role.Status),
		UpdatedAt:   role.UpdatedAt,
	}
}

// RoleDetailResp represents role detail response.
type RoleDetailResp struct {
	RoleResp

	Description string           `json:"description"`
	CreatedAt   time.Time        `json:"created_at"`
	UserCount   int64            `json:"user_count"`
	Permissions []PermissionResp `json:"permissions"`
}

func NewRoleDetailResp(role *model.Role) *RoleDetailResp {
	return &RoleDetailResp{
		RoleResp:    *NewRoleResp(role),
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
	}
}

// ListPermissionReq represents permission list query parameters.
type ListPermissionReq struct {
	dto.PaginationReq
	Action   consts.ActionName  `form:"action" binding:"omitempty"`
	IsSystem *bool              `form:"is_system" binding:"omitempty"`
	Status   *consts.StatusType `form:"status" binding:"omitempty"`
}

func (req *ListPermissionReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if req.Action != "" {
		if _, exists := consts.ValidActions[req.Action]; !exists {
			return fmt.Errorf("invalid action: %s", req.Action)
		}
	}
	return validateStatus(req.Status, false)
}

// PermissionBaseResp contains common fields for permission responses.
type PermissionBaseResp struct {
	ID          int                  `json:"id"`
	Name        string               `json:"name"`
	DisplayName string               `json:"display_name"`
	Action      consts.ActionName    `json:"action"`
	Scope       consts.ResourceScope `json:"scope"`
	IsSystem    bool                 `json:"is_system"`
	Status      string               `json:"status"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

func NewPermissionBaseResp(perm *model.Permission) *PermissionBaseResp {
	return &PermissionBaseResp{
		ID:          perm.ID,
		Name:        perm.Name,
		DisplayName: perm.DisplayName,
		Action:      perm.Action,
		Scope:       perm.Scope,
		IsSystem:    perm.IsSystem,
		Status:      consts.GetStatusTypeName(perm.Status),
		UpdatedAt:   perm.UpdatedAt,
	}
}

// PermissionResp represents permission summary information.
type PermissionResp struct {
	PermissionBaseResp
	Resource string `json:"resource_name"`
}

func NewPermissionResp(perm *model.Permission) *PermissionResp {
	resp := &PermissionResp{
		PermissionBaseResp: *NewPermissionBaseResp(perm),
	}
	if perm.Resource != nil {
		resp.Resource = perm.Resource.Name.String()
	}
	return resp
}

// PermissionDetailResp represents permission detail information.
type PermissionDetailResp struct {
	PermissionBaseResp
	Description string                  `json:"description"`
	Resource    *PermissionResourceResp `json:"resource,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
}

func NewPermissionDetailResp(perm *model.Permission) *PermissionDetailResp {
	resp := &PermissionDetailResp{
		PermissionBaseResp: *NewPermissionBaseResp(perm),
		Description:        perm.Description,
		CreatedAt:          perm.CreatedAt,
	}
	if perm.Resource != nil {
		resp.Resource = NewPermissionResourceResp(perm.Resource)
	}
	return resp
}

// PermissionResourceResp keeps the resource snapshot embedded in permission detail responses.
type PermissionResourceResp struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	ParentID    *int   `json:"parent_id,omitempty"`
}

func NewPermissionResourceResp(resource *model.Resource) *PermissionResourceResp {
	return &PermissionResourceResp{
		ID:          resource.ID,
		Name:        resource.Name.String(),
		DisplayName: resource.DisplayName,
		Type:        consts.GetResourceTypeName(resource.Type),
		Category:    consts.GetResourceCategoryName(resource.Category),
		ParentID:    resource.ParentID,
	}
}

// UserListItem is the RBAC-facing user summary contract for role membership queries.
type UserListItem struct {
	ID          int        `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	FullName    string     `json:"full_name"`
	Avatar      string     `json:"avatar,omitempty"`
	Phone       string     `json:"phone,omitempty"`
	IsActive    bool       `json:"is_active"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func NewUserListItem(user *model.User) *UserListItem {
	return &UserListItem{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FullName:    user.FullName,
		Avatar:      user.Avatar,
		Phone:       user.Phone,
		IsActive:    user.IsActive,
		Status:      consts.GetStatusTypeName(user.Status),
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
