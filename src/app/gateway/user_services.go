package gateway

import (
	"context"

	"aegis/dto"
	user "aegis/module/user"
)

type userIAMClient interface {
	Enabled() bool
	CreateUser(context.Context, *user.CreateUserReq) (*user.UserResp, error)
	DeleteUser(context.Context, int) error
	GetUser(context.Context, int) (*user.UserDetailResp, error)
	ListUsers(context.Context, *user.ListUserReq) (*dto.ListResp[user.UserResp], error)
	UpdateUser(context.Context, *user.UpdateUserReq, int) (*user.UserResp, error)
	AssignUserRole(context.Context, int, int) error
	RemoveUserRole(context.Context, int, int) error
	AssignUserPermissions(context.Context, int, *user.AssignUserPermissionReq) error
	RemoveUserPermissions(context.Context, int, *user.RemoveUserPermissionReq) error
	AssignUserContainer(context.Context, int, int, int) error
	RemoveUserContainer(context.Context, int, int) error
	AssignUserDataset(context.Context, int, int, int) error
	RemoveUserDataset(context.Context, int, int) error
	AssignUserProject(context.Context, int, int, int) error
	RemoveUserProject(context.Context, int, int) error
}

type remoteAwareUserService struct {
	user.HandlerService
	iam userIAMClient
}

func (s remoteAwareUserService) CreateUser(ctx context.Context, req *user.CreateUserReq) (*user.UserResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.CreateUser(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) DeleteUser(ctx context.Context, userID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.DeleteUser(ctx, userID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) GetUserDetail(ctx context.Context, userID int) (*user.UserDetailResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetUser(ctx, userID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) ListUsers(ctx context.Context, req *user.ListUserReq) (*dto.ListResp[user.UserResp], error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListUsers(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) UpdateUser(ctx context.Context, req *user.UpdateUserReq, userID int) (*user.UserResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.UpdateUser(ctx, req, userID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) AssignRole(ctx context.Context, userID, roleID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.AssignUserRole(ctx, userID, roleID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) RemoveRole(ctx context.Context, userID, roleID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.RemoveUserRole(ctx, userID, roleID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) AssignPermissions(ctx context.Context, req *user.AssignUserPermissionReq, userID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.AssignUserPermissions(ctx, userID, req)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) RemovePermissions(ctx context.Context, req *user.RemoveUserPermissionReq, userID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.RemoveUserPermissions(ctx, userID, req)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) AssignContainer(ctx context.Context, userID, containerID, roleID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.AssignUserContainer(ctx, userID, containerID, roleID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) RemoveContainer(ctx context.Context, userID, containerID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.RemoveUserContainer(ctx, userID, containerID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) AssignDataset(ctx context.Context, userID, datasetID, roleID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.AssignUserDataset(ctx, userID, datasetID, roleID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) RemoveDataset(ctx context.Context, userID, datasetID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.RemoveUserDataset(ctx, userID, datasetID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) AssignProject(ctx context.Context, userID, projectID, roleID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.AssignUserProject(ctx, userID, projectID, roleID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) RemoveProject(ctx context.Context, userID, projectID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.RemoveUserProject(ctx, userID, projectID)
	}
	return missingRemoteDependency("iam-service")
}
