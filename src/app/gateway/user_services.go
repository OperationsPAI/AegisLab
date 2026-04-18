package gatewayapp

import (
	"context"

	"aegis/dto"
	usermodule "aegis/module/user"
)

type userIAMClient interface {
	Enabled() bool
	CreateUser(context.Context, *usermodule.CreateUserReq) (*usermodule.UserResp, error)
	DeleteUser(context.Context, int) error
	GetUser(context.Context, int) (*usermodule.UserDetailResp, error)
	ListUsers(context.Context, *usermodule.ListUserReq) (*dto.ListResp[usermodule.UserResp], error)
	UpdateUser(context.Context, *usermodule.UpdateUserReq, int) (*usermodule.UserResp, error)
	AssignUserRole(context.Context, int, int) error
	RemoveUserRole(context.Context, int, int) error
	AssignUserPermissions(context.Context, int, *usermodule.AssignUserPermissionReq) error
	RemoveUserPermissions(context.Context, int, *usermodule.RemoveUserPermissionReq) error
	AssignUserContainer(context.Context, int, int, int) error
	RemoveUserContainer(context.Context, int, int) error
	AssignUserDataset(context.Context, int, int, int) error
	RemoveUserDataset(context.Context, int, int) error
	AssignUserProject(context.Context, int, int, int) error
	RemoveUserProject(context.Context, int, int) error
}

type remoteAwareUserService struct {
	usermodule.HandlerService
	iam userIAMClient
}

func (s remoteAwareUserService) CreateUser(ctx context.Context, req *usermodule.CreateUserReq) (*usermodule.UserResp, error) {
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

func (s remoteAwareUserService) GetUserDetail(ctx context.Context, userID int) (*usermodule.UserDetailResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetUser(ctx, userID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) ListUsers(ctx context.Context, req *usermodule.ListUserReq) (*dto.ListResp[usermodule.UserResp], error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListUsers(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) UpdateUser(ctx context.Context, req *usermodule.UpdateUserReq, userID int) (*usermodule.UserResp, error) {
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

func (s remoteAwareUserService) AssignPermissions(ctx context.Context, req *usermodule.AssignUserPermissionReq, userID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.AssignUserPermissions(ctx, userID, req)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareUserService) RemovePermissions(ctx context.Context, req *usermodule.RemoveUserPermissionReq, userID int) error {
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
