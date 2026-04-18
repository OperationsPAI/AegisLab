package grpciaminterface

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/middleware"
	authmodule "aegis/module/auth"
	rbacmodule "aegis/module/rbac"
	teammodule "aegis/module/team"
	usermodule "aegis/module/user"
	iamv1 "aegis/proto/iam/v1"
	"aegis/utils"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type iamServer struct {
	iamv1.UnimplementedIAMServiceServer
	auth       *authmodule.Service
	authAPI    authmodule.HandlerService
	team       teammodule.HandlerService
	user       usermodule.HandlerService
	rbac       rbacmodule.HandlerService
	middleware middleware.Service
}

func newIAMServer(
	auth *authmodule.Service,
	authAPI authmodule.HandlerService,
	team teammodule.HandlerService,
	user usermodule.HandlerService,
	rbac rbacmodule.HandlerService,
	middlewareService middleware.Service,
) *iamServer {
	return &iamServer{
		auth:       auth,
		authAPI:    authAPI,
		team:       team,
		user:       user,
		rbac:       rbac,
		middleware: middlewareService,
	}
}

func (s *iamServer) VerifyToken(ctx context.Context, req *iamv1.VerifyTokenRequest) (*iamv1.VerifyTokenResponse, error) {
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	claims, err := s.auth.VerifyToken(ctx, req.GetToken())
	if err == nil {
		return &iamv1.VerifyTokenResponse{
			Valid:         true,
			TokenType:     "user",
			UserId:        int64(claims.UserID),
			Username:      claims.Username,
			Email:         claims.Email,
			IsActive:      claims.IsActive,
			IsAdmin:       claims.IsAdmin,
			Roles:         claims.Roles,
			ExpiresAtUnix: claims.ExpiresAt.Unix(),
			AuthType:      claims.AuthType,
			AccessKeyId:   int64(claims.AccessKeyID),
		}, nil
	}

	serviceClaims, serviceErr := s.auth.VerifyServiceToken(ctx, req.GetToken())
	if serviceErr == nil {
		return &iamv1.VerifyTokenResponse{
			Valid:         true,
			TokenType:     "service",
			TaskId:        serviceClaims.TaskID,
			ExpiresAtUnix: serviceClaims.ExpiresAt.Unix(),
		}, nil
	}

	return nil, status.Error(codes.Unauthenticated, err.Error())
}

func (s *iamServer) CheckPermission(ctx context.Context, req *iamv1.CheckPermissionRequest) (*iamv1.CheckPermissionResponse, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	params := &dto.CheckPermissionParams{
		UserID:       int(req.GetUserId()),
		Action:       consts.ActionName(req.GetAction()),
		Scope:        consts.ResourceScope(req.GetScope()),
		ResourceName: consts.ResourceName(req.GetResourceName()),
		TeamID:       optionalID(req.GetTeamId()),
		ProjectID:    optionalID(req.GetProjectId()),
		ContainerID:  optionalID(req.GetContainerId()),
		DatasetID:    optionalID(req.GetDatasetId()),
	}

	allowed, err := s.middleware.CheckUserPermission(ctx, params)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &iamv1.CheckPermissionResponse{Allowed: allowed}, nil
}

func (s *iamServer) Login(ctx context.Context, req *iamv1.MutationRequest) (*iamv1.StructResponse, error) {
	body, err := decodeBody[authmodule.LoginReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.authAPI.Login(ctx, body)
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) Register(ctx context.Context, req *iamv1.MutationRequest) (*iamv1.StructResponse, error) {
	body, err := decodeBody[authmodule.RegisterReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.authAPI.Register(ctx, body)
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) RefreshToken(ctx context.Context, req *iamv1.MutationRequest) (*iamv1.StructResponse, error) {
	body, err := decodeBody[authmodule.TokenRefreshReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.authAPI.RefreshToken(ctx, body)
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) Logout(ctx context.Context, req *iamv1.LogoutRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 || req.GetTokenId() == "" || req.GetExpiresAtUnix() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id, token_id, and expires_at_unix are required")
	}
	claims := &utils.Claims{
		UserID: int(req.GetUserId()),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        req.GetTokenId(),
			ExpiresAt: jwt.NewNumericDate(time.Unix(req.GetExpiresAtUnix(), 0)),
		},
	}
	if err := s.authAPI.Logout(ctx, claims); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) ChangePassword(ctx context.Context, req *iamv1.UserBodyRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	body, err := decodeBody[authmodule.ChangePasswordReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.authAPI.ChangePassword(ctx, body, int(req.GetUserId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) GetProfile(ctx context.Context, req *iamv1.UserIDRequest) (*iamv1.StructResponse, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	resp, err := s.authAPI.GetProfile(ctx, int(req.GetUserId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) CreateAccessKey(ctx context.Context, req *iamv1.UserBodyRequest) (*iamv1.StructResponse, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	body, err := decodeBody[authmodule.CreateAccessKeyReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.authAPI.CreateAccessKey(ctx, int(req.GetUserId()), body)
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) ListAccessKeys(ctx context.Context, req *iamv1.UserQueryRequest) (*iamv1.StructResponse, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	query, err := decodeQuery[authmodule.ListAccessKeyReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.authAPI.ListAccessKeys(ctx, int(req.GetUserId()), query)
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) GetAccessKey(ctx context.Context, req *iamv1.UserScopedIDRequest) (*iamv1.StructResponse, error) {
	if req.GetUserId() <= 0 || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and id are required")
	}
	resp, err := s.authAPI.GetAccessKey(ctx, int(req.GetUserId()), int(req.GetId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) DeleteAccessKey(ctx context.Context, req *iamv1.UserScopedIDRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and id are required")
	}
	if err := s.authAPI.DeleteAccessKey(ctx, int(req.GetUserId()), int(req.GetId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) DisableAccessKey(ctx context.Context, req *iamv1.UserScopedIDRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and id are required")
	}
	if err := s.authAPI.DisableAccessKey(ctx, int(req.GetUserId()), int(req.GetId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) EnableAccessKey(ctx context.Context, req *iamv1.UserScopedIDRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and id are required")
	}
	if err := s.authAPI.EnableAccessKey(ctx, int(req.GetUserId()), int(req.GetId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) RotateAccessKey(ctx context.Context, req *iamv1.UserScopedIDRequest) (*iamv1.StructResponse, error) {
	if req.GetUserId() <= 0 || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and id are required")
	}
	resp, err := s.authAPI.RotateAccessKey(ctx, int(req.GetUserId()), int(req.GetId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) CreateUser(ctx context.Context, req *iamv1.MutationRequest) (*iamv1.StructResponse, error) {
	body, err := decodeBody[usermodule.CreateUserReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.user.CreateUser(ctx, body)
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) DeleteUser(ctx context.Context, req *iamv1.IDRequest) (*emptypb.Empty, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if err := s.user.DeleteUser(ctx, int(req.GetId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) GetUser(ctx context.Context, req *iamv1.IDRequest) (*iamv1.StructResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	resp, err := s.user.GetUserDetail(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) ListUsers(ctx context.Context, req *iamv1.QueryRequest) (*iamv1.StructResponse, error) {
	query, err := decodeQuery[usermodule.ListUserReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.user.ListUsers(ctx, query)
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) UpdateUser(ctx context.Context, req *iamv1.UpdateByIDRequest) (*iamv1.StructResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	body, err := decodeBody[usermodule.UpdateUserReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.user.UpdateUser(ctx, body, int(req.GetId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) AssignUserRole(ctx context.Context, req *iamv1.UserRoleBindingRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 || req.GetRoleId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and role_id are required")
	}
	if err := s.user.AssignRole(ctx, int(req.GetUserId()), int(req.GetRoleId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) RemoveUserRole(ctx context.Context, req *iamv1.UserRoleBindingRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 || req.GetRoleId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and role_id are required")
	}
	if err := s.user.RemoveRole(ctx, int(req.GetUserId()), int(req.GetRoleId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) AssignUserPermissions(ctx context.Context, req *iamv1.UserBodyRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	body, err := decodeBody[usermodule.AssignUserPermissionReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.user.AssignPermissions(ctx, body, int(req.GetUserId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) RemoveUserPermissions(ctx context.Context, req *iamv1.UserBodyRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	body, err := decodeBody[usermodule.RemoveUserPermissionReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.user.RemovePermissions(ctx, body, int(req.GetUserId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) AssignUserContainer(ctx context.Context, req *iamv1.UserResourceBindingRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 || req.GetResourceId() <= 0 || req.GetRoleId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id, resource_id, and role_id are required")
	}
	if err := s.user.AssignContainer(ctx, int(req.GetUserId()), int(req.GetResourceId()), int(req.GetRoleId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) RemoveUserContainer(ctx context.Context, req *iamv1.UserScopedIDRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and id are required")
	}
	if err := s.user.RemoveContainer(ctx, int(req.GetUserId()), int(req.GetId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) AssignUserDataset(ctx context.Context, req *iamv1.UserResourceBindingRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 || req.GetResourceId() <= 0 || req.GetRoleId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id, resource_id, and role_id are required")
	}
	if err := s.user.AssignDataset(ctx, int(req.GetUserId()), int(req.GetResourceId()), int(req.GetRoleId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) RemoveUserDataset(ctx context.Context, req *iamv1.UserScopedIDRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and id are required")
	}
	if err := s.user.RemoveDataset(ctx, int(req.GetUserId()), int(req.GetId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) AssignUserProject(ctx context.Context, req *iamv1.UserResourceBindingRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 || req.GetResourceId() <= 0 || req.GetRoleId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id, resource_id, and role_id are required")
	}
	if err := s.user.AssignProject(ctx, int(req.GetUserId()), int(req.GetResourceId()), int(req.GetRoleId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) RemoveUserProject(ctx context.Context, req *iamv1.UserScopedIDRequest) (*emptypb.Empty, error) {
	if req.GetUserId() <= 0 || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and id are required")
	}
	if err := s.user.RemoveProject(ctx, int(req.GetUserId()), int(req.GetId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) CreateRole(ctx context.Context, req *iamv1.MutationRequest) (*iamv1.StructResponse, error) {
	body, err := decodeBody[rbacmodule.CreateRoleReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.rbac.CreateRole(ctx, body)
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) DeleteRole(ctx context.Context, req *iamv1.IDRequest) (*emptypb.Empty, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if err := s.rbac.DeleteRole(ctx, int(req.GetId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) GetRole(ctx context.Context, req *iamv1.IDRequest) (*iamv1.StructResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	resp, err := s.rbac.GetRole(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) ListRoles(ctx context.Context, req *iamv1.QueryRequest) (*iamv1.StructResponse, error) {
	query, err := decodeQuery[rbacmodule.ListRoleReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.rbac.ListRoles(ctx, query)
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) UpdateRole(ctx context.Context, req *iamv1.UpdateByIDRequest) (*iamv1.StructResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	body, err := decodeBody[rbacmodule.UpdateRoleReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.rbac.UpdateRole(ctx, body, int(req.GetId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) AssignRolePermissions(ctx context.Context, req *iamv1.RolePermissionsRequest) (*emptypb.Empty, error) {
	if req.GetRoleId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "role_id is required")
	}
	if err := validatePositiveInt64s(req.GetPermissionIds(), "permission_ids"); err != nil {
		return nil, err
	}
	if err := s.rbac.AssignRolePermissions(ctx, int64sToInts(req.GetPermissionIds()), int(req.GetRoleId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) RemoveRolePermissions(ctx context.Context, req *iamv1.RolePermissionsRequest) (*emptypb.Empty, error) {
	if req.GetRoleId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "role_id is required")
	}
	if err := validatePositiveInt64s(req.GetPermissionIds(), "permission_ids"); err != nil {
		return nil, err
	}
	if err := s.rbac.RemoveRolePermissions(ctx, int64sToInts(req.GetPermissionIds()), int(req.GetRoleId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) ListUsersFromRole(ctx context.Context, req *iamv1.IDRequest) (*iamv1.StructResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	resp, err := s.rbac.ListUsersFromRole(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) GetPermission(ctx context.Context, req *iamv1.IDRequest) (*iamv1.StructResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	resp, err := s.rbac.GetPermission(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) ListPermissions(ctx context.Context, req *iamv1.QueryRequest) (*iamv1.StructResponse, error) {
	query, err := decodeQuery[rbacmodule.ListPermissionReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.rbac.ListPermissions(ctx, query)
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) ListRolesFromPermission(ctx context.Context, req *iamv1.IDRequest) (*iamv1.StructResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	resp, err := s.rbac.ListRolesFromPermission(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) GetResource(ctx context.Context, req *iamv1.IDRequest) (*iamv1.StructResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	resp, err := s.rbac.GetResource(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) ListResources(ctx context.Context, req *iamv1.QueryRequest) (*iamv1.StructResponse, error) {
	query, err := decodeQuery[rbacmodule.ListResourceReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.rbac.ListResources(ctx, query)
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) ListResourcePermissions(ctx context.Context, req *iamv1.IDRequest) (*iamv1.StructResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	resp, err := s.rbac.ListResourcePermissions(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) IsUserTeamAdmin(ctx context.Context, req *iamv1.UserTeamRequest) (*iamv1.BoolResponse, error) {
	if req.GetUserId() <= 0 || req.GetTeamId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and team_id are required")
	}

	allowed, err := s.middleware.IsUserTeamAdmin(ctx, int(req.GetUserId()), int(req.GetTeamId()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &iamv1.BoolResponse{Value: allowed}, nil
}

func (s *iamServer) IsUserInTeam(ctx context.Context, req *iamv1.UserTeamRequest) (*iamv1.BoolResponse, error) {
	if req.GetUserId() <= 0 || req.GetTeamId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and team_id are required")
	}

	allowed, err := s.middleware.IsUserInTeam(ctx, int(req.GetUserId()), int(req.GetTeamId()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &iamv1.BoolResponse{Value: allowed}, nil
}

func (s *iamServer) IsTeamPublic(ctx context.Context, req *iamv1.TeamRequest) (*iamv1.BoolResponse, error) {
	if req.GetTeamId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "team_id is required")
	}

	allowed, err := s.middleware.IsTeamPublic(ctx, int(req.GetTeamId()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &iamv1.BoolResponse{Value: allowed}, nil
}

func (s *iamServer) IsUserProjectAdmin(ctx context.Context, req *iamv1.UserProjectRequest) (*iamv1.BoolResponse, error) {
	if req.GetUserId() <= 0 || req.GetProjectId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and project_id are required")
	}

	allowed, err := s.middleware.IsUserProjectAdmin(ctx, int(req.GetUserId()), int(req.GetProjectId()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &iamv1.BoolResponse{Value: allowed}, nil
}

func (s *iamServer) IsUserInProject(ctx context.Context, req *iamv1.UserProjectRequest) (*iamv1.BoolResponse, error) {
	if req.GetUserId() <= 0 || req.GetProjectId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and project_id are required")
	}

	allowed, err := s.middleware.IsUserInProject(ctx, int(req.GetUserId()), int(req.GetProjectId()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &iamv1.BoolResponse{Value: allowed}, nil
}

func (s *iamServer) ExchangeAccessKeyToken(ctx context.Context, req *iamv1.ExchangeAccessKeyTokenRequest) (*iamv1.ExchangeAccessKeyTokenResponse, error) {
	authReq := &authmodule.AccessKeyTokenReq{
		AccessKey: req.GetAccessKey(),
		Timestamp: req.GetTimestamp(),
		Nonce:     req.GetNonce(),
		Signature: req.GetSignature(),
	}
	if err := authReq.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if req.GetMethod() == "" || req.GetPath() == "" {
		return nil, status.Error(codes.InvalidArgument, "method and path are required")
	}

	resp, err := s.auth.ExchangeAccessKeyToken(ctx, authReq, req.GetMethod(), req.GetPath())
	if err != nil {
		return nil, mapIAMError(err)
	}
	return &iamv1.ExchangeAccessKeyTokenResponse{
		Token:         resp.Token,
		TokenType:     resp.TokenType,
		ExpiresAtUnix: resp.ExpiresAt.Unix(),
		AuthType:      resp.AuthType,
		AccessKey:     resp.AccessKey,
	}, nil
}

func (s *iamServer) CreateTeam(ctx context.Context, req *iamv1.CreateTeamRequest) (*iamv1.StructResponse, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	body, err := decodeBody[teammodule.CreateTeamReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.team.CreateTeam(ctx, body, int(req.GetUserId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) DeleteTeam(ctx context.Context, req *iamv1.TeamRequest) (*emptypb.Empty, error) {
	if req.GetTeamId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "team_id is required")
	}
	if err := s.team.DeleteTeam(ctx, int(req.GetTeamId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) GetTeam(ctx context.Context, req *iamv1.TeamRequest) (*iamv1.StructResponse, error) {
	if req.GetTeamId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "team_id is required")
	}
	resp, err := s.team.GetTeamDetail(ctx, int(req.GetTeamId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) ListTeams(ctx context.Context, req *iamv1.ListTeamsRequest) (*iamv1.StructResponse, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	query, err := decodeQuery[teammodule.ListTeamReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.team.ListTeams(ctx, query, int(req.GetUserId()), req.GetIsAdmin())
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) UpdateTeam(ctx context.Context, req *iamv1.UpdateTeamRequest) (*iamv1.StructResponse, error) {
	if req.GetTeamId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "team_id is required")
	}
	body, err := decodeBody[teammodule.UpdateTeamReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.team.UpdateTeam(ctx, body, int(req.GetTeamId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) ListTeamProjects(ctx context.Context, req *iamv1.ListTeamProjectsRequest) (*iamv1.StructResponse, error) {
	if req.GetTeamId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "team_id is required")
	}
	query, err := decodeQuery[teammodule.TeamProjectListReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.team.ListTeamProjects(ctx, query, int(req.GetTeamId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func (s *iamServer) AddTeamMember(ctx context.Context, req *iamv1.AddTeamMemberRequest) (*emptypb.Empty, error) {
	if req.GetTeamId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "team_id is required")
	}
	body, err := decodeBody[teammodule.AddTeamMemberReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.team.AddMember(ctx, body, int(req.GetTeamId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) RemoveTeamMember(ctx context.Context, req *iamv1.RemoveTeamMemberRequest) (*emptypb.Empty, error) {
	if req.GetTeamId() <= 0 || req.GetCurrentUserId() <= 0 || req.GetTargetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "team_id, current_user_id, and target_user_id are required")
	}
	if err := s.team.RemoveMember(ctx, int(req.GetTeamId()), int(req.GetCurrentUserId()), int(req.GetTargetUserId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) UpdateTeamMemberRole(ctx context.Context, req *iamv1.UpdateTeamMemberRoleRequest) (*emptypb.Empty, error) {
	if req.GetTeamId() <= 0 || req.GetTargetUserId() <= 0 || req.GetCurrentUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "team_id, target_user_id, and current_user_id are required")
	}
	body, err := decodeBody[teammodule.UpdateTeamMemberRoleReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.team.UpdateMemberRole(ctx, body, int(req.GetTeamId()), int(req.GetTargetUserId()), int(req.GetCurrentUserId())); err != nil {
		return nil, mapIAMError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *iamServer) ListTeamMembers(ctx context.Context, req *iamv1.ListTeamMembersRequest) (*iamv1.StructResponse, error) {
	if req.GetTeamId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "team_id is required")
	}
	query, err := decodeQuery[teammodule.ListTeamMemberReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.team.ListMembers(ctx, query, int(req.GetTeamId()))
	if err != nil {
		return nil, mapIAMError(err)
	}
	return encodeStruct(resp)
}

func optionalID(value int64) *int {
	if value <= 0 {
		return nil
	}
	id := int(value)
	return &id
}

func mapIAMError(err error) error {
	switch {
	case errors.Is(err, consts.ErrBadRequest):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, consts.ErrAuthenticationFailed):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, consts.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, consts.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, consts.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case err != nil:
		return status.Error(codes.Internal, err.Error())
	default:
		return nil
	}
}

func encodeStruct(value any) (*iamv1.StructResponse, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	body, err := structpb.NewStruct(payload)
	if err != nil {
		return nil, err
	}
	return &iamv1.StructResponse{Data: body}, nil
}

func decodeBody[T any](payload *structpb.Struct) (*T, error) {
	return decodeQuery[T](payload)
}

func decodeQuery[T any](payload *structpb.Struct) (*T, error) {
	if payload == nil {
		var zero T
		return &zero, nil
	}
	data, err := json.Marshal(payload.AsMap())
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func validatePositiveInt64s(items []int64, field string) error {
	if len(items) == 0 {
		return status.Errorf(codes.InvalidArgument, "%s is required", field)
	}
	for _, item := range items {
		if item <= 0 {
			return status.Errorf(codes.InvalidArgument, "%s must contain positive integers", field)
		}
	}
	return nil
}

func int64sToInts(items []int64) []int {
	if len(items) == 0 {
		return nil
	}
	result := make([]int, 0, len(items))
	for _, item := range items {
		result = append(result, int(item))
	}
	return result
}
