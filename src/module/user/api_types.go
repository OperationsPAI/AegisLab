package user

import (
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	rbac "aegis/module/rbac"
)

// CreateUserReq represents user creation request.
type CreateUserReq struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
	Phone    string `json:"phone" binding:"omitempty"`
	Avatar   string `json:"avatar" binding:"omitempty"`
}

func (req *CreateUserReq) Validate() error {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if req.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if req.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

// ListUserReq represents user list query parameters.
type ListUserReq struct {
	dto.PaginationReq
	IsActive *bool              `form:"is_active"`
	Status   *consts.StatusType `form:"status"`
}

func (req *ListUserReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	return validateStatus(req.Status, false)
}

// UpdateUserReq represents user update request.
type UpdateUserReq struct {
	Email    *string            `json:"email,omitempty" binding:"omitempty,email"`
	FullName *string            `json:"full_name,omitempty" binding:"omitempty"`
	Phone    *string            `json:"phone,omitempty" binding:"omitempty"`
	Avatar   *string            `json:"avatar,omitempty" binding:"omitempty"`
	IsActive *bool              `json:"is_active,omitempty" binding:"omitempty"`
	Status   *consts.StatusType `json:"status,omitempty" binding:"omitempty"`
}

func (req *UpdateUserReq) Validate() error {
	return validateStatus(req.Status, true)
}

func (req *UpdateUserReq) PatchUserModel(target *model.User) {
	if req.Email != nil {
		target.Email = *req.Email
	}
	if req.FullName != nil {
		target.FullName = *req.FullName
	}
	if req.Phone != nil {
		target.Phone = *req.Phone
	}
	if req.Avatar != nil {
		target.Avatar = *req.Avatar
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
	if req.IsActive != nil {
		target.IsActive = *req.IsActive
	}
}

// UserResp represents basic user response.
type UserResp struct {
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

func NewUserResp(user *model.User) *UserResp {
	return &UserResp{
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

// UserDetailResp represents detailed user response with roles and permissions.
type UserDetailResp struct {
	UserResp

	GlobalRoles    []rbac.RoleResp       `json:"global_roles,omitempty"`
	Permissions    []rbac.PermissionResp `json:"permissions,omitempty"`
	ContainerRoles []UserContainerInfo   `json:"container_roles,omitempty"`
	DatasetRoles   []UserDatasetInfo     `json:"dataset_roles,omitempty"`
	ProjectRoles   []UserProjectInfo     `json:"project_roles,omitempty"`
}

func NewUserDetailResp(user *model.User) *UserDetailResp {
	return &UserDetailResp{
		UserResp: *NewUserResp(user),
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

// AssignUserPermissionItem represents a single user-permission assignment item.
type AssignUserPermissionItem struct {
	PermissionID int               `json:"permission_id" binding:"required,min=1"`
	GrantType    *consts.GrantType `json:"grant_type" binding:"required"`
	ExpiresAt    *time.Time        `json:"expires_at" binding:"omitempty"`
	ContainerID  *int              `json:"container_id" binding:"omitempty,min_ptr=1"`
	DatasetID    *int              `json:"dataset_id" binding:"omitempty,min_ptr=1"`
	ProjectID    *int              `json:"project_id" binding:"omitempty,min_ptr=1"`
}

func (item *AssignUserPermissionItem) Validate() error {
	if item.GrantType == nil {
		return fmt.Errorf("grant_type is required")
	}
	if _, valid := consts.ValidGrantTypes[*item.GrantType]; !valid {
		return fmt.Errorf("invalid grant_type: %d", item.GrantType)
	}
	if item.ExpiresAt != nil && item.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("expires_at cannot be in the past")
	}
	return nil
}

func (item *AssignUserPermissionItem) ConvertToUserPermission() *model.UserPermission {
	return &model.UserPermission{
		PermissionID: item.PermissionID,
		GrantType:    *item.GrantType,
		ExpiresAt:    item.ExpiresAt,
		ContainerID:  item.ContainerID,
		DatasetID:    item.DatasetID,
		ProjectID:    item.ProjectID,
	}
}

// AssignUserPermissionReq represents direct user-permission assignment request.
type AssignUserPermissionReq struct {
	Items []AssignUserPermissionItem `json:"items" binding:"required"`
}

func (req *AssignUserPermissionReq) Validate() error {
	if len(req.Items) == 0 {
		return fmt.Errorf("items cannot be empty")
	}
	for idx, item := range req.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("invalid item %d: %v", idx, err)
		}
	}
	return nil
}

// RemoveUserPermissionReq represents direct user-permission removal request.
type RemoveUserPermissionReq struct {
	PermissionIDs []int `json:"permission_ids" binding:"required"`
}

func (req *RemoveUserPermissionReq) Validate() error {
	if len(req.PermissionIDs) == 0 {
		return fmt.Errorf("permission_ids cannot be empty")
	}
	for _, id := range req.PermissionIDs {
		if id <= 0 {
			return fmt.Errorf("invalid permission ID: %d", id)
		}
	}
	return nil
}

// UserContainerInfo represents a user's role binding on a container.
type UserContainerInfo struct {
	ContainerID   int       `json:"container_id"`
	ContainerName string    `json:"container_name"`
	RoleName      string    `json:"role_name"`
	JoinedAt      time.Time `json:"joined_at"`
}

func NewUserContainerInfo(userContainer *model.UserContainer) *UserContainerInfo {
	resp := &UserContainerInfo{
		ContainerID: userContainer.ContainerID,
		JoinedAt:    userContainer.CreatedAt,
	}
	if userContainer.Container != nil {
		resp.ContainerName = userContainer.Container.Name
	}
	if userContainer.Role != nil {
		resp.RoleName = userContainer.Role.Name
	}
	return resp
}

// UserDatasetInfo represents a user's role binding on a dataset.
type UserDatasetInfo struct {
	DatasetID   int       `json:"dataset_id"`
	DatasetName string    `json:"dataset_name"`
	RoleName    string    `json:"role_name"`
	JoinedAt    time.Time `json:"joined_at"`
}

func NewUserDatasetInfo(userDataset *model.UserDataset) *UserDatasetInfo {
	resp := &UserDatasetInfo{
		DatasetID: userDataset.DatasetID,
		JoinedAt:  userDataset.CreatedAt,
	}
	if userDataset.Dataset != nil {
		resp.DatasetName = userDataset.Dataset.Name
	}
	if userDataset.Role != nil {
		resp.RoleName = userDataset.Role.Name
	}
	return resp
}

// UserProjectInfo represents a user's role binding on a project.
type UserProjectInfo struct {
	ProjectID   int       `json:"project_id"`
	ProjectName string    `json:"project_name"`
	RoleName    string    `json:"role_name"`
	JoinedAt    time.Time `json:"joined_at"`
}

func NewUserProjectInfo(userProject *model.UserProject) *UserProjectInfo {
	resp := &UserProjectInfo{
		ProjectID: userProject.ProjectID,
		JoinedAt:  userProject.CreatedAt,
	}
	if userProject.Project != nil {
		resp.ProjectName = userProject.Project.Name
	}
	if userProject.Role != nil {
		resp.RoleName = userProject.Role.Name
	}
	return resp
}
