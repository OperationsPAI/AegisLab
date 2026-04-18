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
	CreateAPIKey(context.Context, int, *authmodule.CreateAPIKeyReq) (*authmodule.APIKeyWithSecretResp, error)
	ListAPIKeys(context.Context, int, *authmodule.ListAPIKeyReq) (*authmodule.ListAPIKeyResp, error)
	GetAPIKey(context.Context, int, int) (*authmodule.APIKeyInfo, error)
	DeleteAPIKey(context.Context, int, int) error
	DisableAPIKey(context.Context, int, int) error
	EnableAPIKey(context.Context, int, int) error
	RevokeAPIKey(context.Context, int, int) error
	RotateAPIKey(context.Context, int, int) (*authmodule.APIKeyWithSecretResp, error)
	ExchangeAPIKeyToken(context.Context, *authmodule.APIKeyTokenReq, string, string) (*authmodule.APIKeyTokenResp, error)
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

func (s remoteAwareAuthService) CreateAPIKey(ctx context.Context, userID int, req *authmodule.CreateAPIKeyReq) (*authmodule.APIKeyWithSecretResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.CreateAPIKey(ctx, userID, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) ListAPIKeys(ctx context.Context, userID int, req *authmodule.ListAPIKeyReq) (*authmodule.ListAPIKeyResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListAPIKeys(ctx, userID, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) GetAPIKey(ctx context.Context, userID, accessKeyID int) (*authmodule.APIKeyInfo, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetAPIKey(ctx, userID, accessKeyID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) DeleteAPIKey(ctx context.Context, userID, accessKeyID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.DeleteAPIKey(ctx, userID, accessKeyID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) DisableAPIKey(ctx context.Context, userID, accessKeyID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.DisableAPIKey(ctx, userID, accessKeyID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) EnableAPIKey(ctx context.Context, userID, accessKeyID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.EnableAPIKey(ctx, userID, accessKeyID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) RevokeAPIKey(ctx context.Context, userID, accessKeyID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.RevokeAPIKey(ctx, userID, accessKeyID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) RotateAPIKey(ctx context.Context, userID, accessKeyID int) (*authmodule.APIKeyWithSecretResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.RotateAPIKey(ctx, userID, accessKeyID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) ExchangeAPIKeyToken(ctx context.Context, req *authmodule.APIKeyTokenReq, method, path string) (*authmodule.APIKeyTokenResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ExchangeAPIKeyToken(ctx, req, method, path)
	}
	return nil, missingRemoteDependency("iam-service")
}
