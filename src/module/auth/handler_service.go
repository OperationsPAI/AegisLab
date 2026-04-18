package authmodule

import (
	"context"

	"aegis/utils"
)

// HandlerService captures the auth operations consumed by the HTTP handler.
type HandlerService interface {
	Login(context.Context, *LoginReq) (*LoginResp, error)
	Register(context.Context, *RegisterReq) (*UserInfo, error)
	RefreshToken(context.Context, *TokenRefreshReq) (*TokenRefreshResp, error)
	Logout(context.Context, *utils.Claims) error
	ChangePassword(context.Context, *ChangePasswordReq, int) error
	GetProfile(context.Context, int) (*UserProfileResp, error)
	CreateAccessKey(context.Context, int, *CreateAccessKeyReq) (*AccessKeyWithSecretResp, error)
	ListAccessKeys(context.Context, int, *ListAccessKeyReq) (*ListAccessKeyResp, error)
	GetAccessKey(context.Context, int, int) (*AccessKeyInfo, error)
	DeleteAccessKey(context.Context, int, int) error
	DisableAccessKey(context.Context, int, int) error
	EnableAccessKey(context.Context, int, int) error
	RotateAccessKey(context.Context, int, int) (*AccessKeyWithSecretResp, error)
	ExchangeAccessKeyToken(context.Context, *AccessKeyTokenReq, string, string) (*AccessKeyTokenResp, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
