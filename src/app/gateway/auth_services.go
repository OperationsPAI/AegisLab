package gatewayapp

import (
	"context"

	authmodule "aegis/module/auth"
	"aegis/utils"
)

type authIAMClient interface {
	Enabled() bool
	Login(context.Context, *authmodule.LoginReq) (*authmodule.LoginResp, error)
	Register(context.Context, *authmodule.RegisterReq) (*authmodule.UserInfo, error)
	RefreshToken(context.Context, *authmodule.TokenRefreshReq) (*authmodule.TokenRefreshResp, error)
	Logout(context.Context, *utils.Claims) error
	ChangePassword(context.Context, *authmodule.ChangePasswordReq, int) error
	GetProfile(context.Context, int) (*authmodule.UserProfileResp, error)
	CreateAccessKey(context.Context, int, *authmodule.CreateAccessKeyReq) (*authmodule.AccessKeyWithSecretResp, error)
	ListAccessKeys(context.Context, int, *authmodule.ListAccessKeyReq) (*authmodule.ListAccessKeyResp, error)
	GetAccessKey(context.Context, int, int) (*authmodule.AccessKeyInfo, error)
	DeleteAccessKey(context.Context, int, int) error
	DisableAccessKey(context.Context, int, int) error
	EnableAccessKey(context.Context, int, int) error
	RotateAccessKey(context.Context, int, int) (*authmodule.AccessKeyWithSecretResp, error)
	ExchangeAccessKeyToken(context.Context, *authmodule.AccessKeyTokenReq, string, string) (*authmodule.AccessKeyTokenResp, error)
}

type remoteAwareAuthService struct {
	authmodule.HandlerService
	iam authIAMClient
}

func (s remoteAwareAuthService) Login(ctx context.Context, req *authmodule.LoginReq) (*authmodule.LoginResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.Login(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) Register(ctx context.Context, req *authmodule.RegisterReq) (*authmodule.UserInfo, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.Register(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) RefreshToken(ctx context.Context, req *authmodule.TokenRefreshReq) (*authmodule.TokenRefreshResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.RefreshToken(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) Logout(ctx context.Context, claims *utils.Claims) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.Logout(ctx, claims)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) ChangePassword(ctx context.Context, req *authmodule.ChangePasswordReq, userID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ChangePassword(ctx, req, userID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) GetProfile(ctx context.Context, userID int) (*authmodule.UserProfileResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetProfile(ctx, userID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) CreateAccessKey(ctx context.Context, userID int, req *authmodule.CreateAccessKeyReq) (*authmodule.AccessKeyWithSecretResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.CreateAccessKey(ctx, userID, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) ListAccessKeys(ctx context.Context, userID int, req *authmodule.ListAccessKeyReq) (*authmodule.ListAccessKeyResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListAccessKeys(ctx, userID, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) GetAccessKey(ctx context.Context, userID, accessKeyID int) (*authmodule.AccessKeyInfo, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetAccessKey(ctx, userID, accessKeyID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) DeleteAccessKey(ctx context.Context, userID, accessKeyID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.DeleteAccessKey(ctx, userID, accessKeyID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) DisableAccessKey(ctx context.Context, userID, accessKeyID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.DisableAccessKey(ctx, userID, accessKeyID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) EnableAccessKey(ctx context.Context, userID, accessKeyID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.EnableAccessKey(ctx, userID, accessKeyID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) RotateAccessKey(ctx context.Context, userID, accessKeyID int) (*authmodule.AccessKeyWithSecretResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.RotateAccessKey(ctx, userID, accessKeyID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) ExchangeAccessKeyToken(ctx context.Context, req *authmodule.AccessKeyTokenReq, method, path string) (*authmodule.AccessKeyTokenResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ExchangeAccessKeyToken(ctx, req, method, path)
	}
	return nil, missingRemoteDependency("iam-service")
}
