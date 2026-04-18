package iamclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	"aegis/httpx"
	"aegis/middleware"
	authmodule "aegis/module/auth"
	rbacmodule "aegis/module/rbac"
	teammodule "aegis/module/team"
	usermodule "aegis/module/user"
	iamv1 "aegis/proto/iam/v1"
	"aegis/utils"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type Client struct {
	target string
	conn   *grpc.ClientConn
	rpc    iamv1.IAMServiceClient
}

func NewClient(lc fx.Lifecycle) (*Client, error) {
	target := config.GetString("clients.iam.target")
	if target == "" {
		target = config.GetString("iam.grpc.target")
	}
	if target == "" {
		return &Client{}, nil
	}

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(httpx.UnaryClientRequestIDInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("create iam grpc client: %w", err)
	}

	client := &Client{
		target: target,
		conn:   conn,
		rpc:    iamv1.NewIAMServiceClient(conn),
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return conn.Close()
		},
	})

	return client, nil
}

func (c *Client) Enabled() bool {
	return c != nil && c.rpc != nil
}

func (c *Client) VerifyToken(ctx context.Context, token string) (*utils.Claims, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}

	resp, err := c.rpc.VerifyToken(ctx, &iamv1.VerifyTokenRequest{Token: token})
	if err != nil {
		return nil, mapRPCError(err)
	}
	if resp.GetTokenType() != "user" {
		return nil, fmt.Errorf("token is not a user token")
	}
	return &utils.Claims{
		UserID:      int(resp.GetUserId()),
		Username:    resp.GetUsername(),
		Email:       resp.GetEmail(),
		IsActive:    resp.GetIsActive(),
		IsAdmin:     resp.GetIsAdmin(),
		Roles:       resp.GetRoles(),
		AuthType:    resp.GetAuthType(),
		AccessKeyID: int(resp.GetAccessKeyId()),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(resp.GetExpiresAtUnix(), 0)),
		},
	}, nil
}

func (c *Client) VerifyServiceToken(ctx context.Context, token string) (*utils.ServiceClaims, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}

	resp, err := c.rpc.VerifyToken(ctx, &iamv1.VerifyTokenRequest{Token: token})
	if err != nil {
		return nil, mapRPCError(err)
	}
	if resp.GetTokenType() != "service" {
		return nil, fmt.Errorf("token is not a service token")
	}
	return &utils.ServiceClaims{
		TaskID: resp.GetTaskId(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(resp.GetExpiresAtUnix(), 0)),
		},
	}, nil
}

func (c *Client) CheckUserPermission(ctx context.Context, params *dto.CheckPermissionParams) (bool, error) {
	if !c.Enabled() {
		return false, fmt.Errorf("iam grpc client is not configured")
	}
	if params == nil {
		return false, fmt.Errorf("permission params are nil")
	}

	req := &iamv1.CheckPermissionRequest{
		UserId:       int64(params.UserID),
		Action:       string(params.Action),
		Scope:        string(params.Scope),
		ResourceName: string(params.ResourceName),
	}
	if params.TeamID != nil {
		req.TeamId = int64(*params.TeamID)
	}
	if params.ProjectID != nil {
		req.ProjectId = int64(*params.ProjectID)
	}
	if params.ContainerID != nil {
		req.ContainerId = int64(*params.ContainerID)
	}
	if params.DatasetID != nil {
		req.DatasetId = int64(*params.DatasetID)
	}

	resp, err := c.rpc.CheckPermission(ctx, req)
	if err != nil {
		return false, mapRPCError(err)
	}
	return resp.GetAllowed(), nil
}

func (c *Client) IsUserTeamAdmin(ctx context.Context, userID, teamID int) (bool, error) {
	if !c.Enabled() {
		return false, fmt.Errorf("iam grpc client is not configured")
	}

	resp, err := c.rpc.IsUserTeamAdmin(ctx, &iamv1.UserTeamRequest{
		UserId: int64(userID),
		TeamId: int64(teamID),
	})
	if err != nil {
		return false, mapRPCError(err)
	}
	return resp.GetValue(), nil
}

func (c *Client) IsUserInTeam(ctx context.Context, userID, teamID int) (bool, error) {
	if !c.Enabled() {
		return false, fmt.Errorf("iam grpc client is not configured")
	}

	resp, err := c.rpc.IsUserInTeam(ctx, &iamv1.UserTeamRequest{
		UserId: int64(userID),
		TeamId: int64(teamID),
	})
	if err != nil {
		return false, mapRPCError(err)
	}
	return resp.GetValue(), nil
}

func (c *Client) IsTeamPublic(ctx context.Context, teamID int) (bool, error) {
	if !c.Enabled() {
		return false, fmt.Errorf("iam grpc client is not configured")
	}

	resp, err := c.rpc.IsTeamPublic(ctx, &iamv1.TeamRequest{TeamId: int64(teamID)})
	if err != nil {
		return false, mapRPCError(err)
	}
	return resp.GetValue(), nil
}

func (c *Client) IsUserProjectAdmin(ctx context.Context, userID, projectID int) (bool, error) {
	if !c.Enabled() {
		return false, fmt.Errorf("iam grpc client is not configured")
	}

	resp, err := c.rpc.IsUserProjectAdmin(ctx, &iamv1.UserProjectRequest{
		UserId:    int64(userID),
		ProjectId: int64(projectID),
	})
	if err != nil {
		return false, mapRPCError(err)
	}
	return resp.GetValue(), nil
}

func (c *Client) IsUserInProject(ctx context.Context, userID, projectID int) (bool, error) {
	if !c.Enabled() {
		return false, fmt.Errorf("iam grpc client is not configured")
	}

	resp, err := c.rpc.IsUserInProject(ctx, &iamv1.UserProjectRequest{
		UserId:    int64(userID),
		ProjectId: int64(projectID),
	})
	if err != nil {
		return false, mapRPCError(err)
	}
	return resp.GetValue(), nil
}

func (c *Client) Login(ctx context.Context, req *authmodule.LoginReq) (*authmodule.LoginResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode login request: %w", err)
	}
	resp, err := c.rpc.Login(ctx, &iamv1.MutationRequest{Body: body})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[authmodule.LoginResp](resp.GetData())
}

func (c *Client) Register(ctx context.Context, req *authmodule.RegisterReq) (*authmodule.UserInfo, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode register request: %w", err)
	}
	resp, err := c.rpc.Register(ctx, &iamv1.MutationRequest{Body: body})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[authmodule.UserInfo](resp.GetData())
}

func (c *Client) RefreshToken(ctx context.Context, req *authmodule.TokenRefreshReq) (*authmodule.TokenRefreshResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode refresh token request: %w", err)
	}
	resp, err := c.rpc.RefreshToken(ctx, &iamv1.MutationRequest{Body: body})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[authmodule.TokenRefreshResp](resp.GetData())
}

func (c *Client) Logout(ctx context.Context, claims *utils.Claims) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	if claims == nil || claims.ExpiresAt == nil || claims.ID == "" {
		return fmt.Errorf("logout claims are incomplete")
	}
	_, err := c.rpc.Logout(ctx, &iamv1.LogoutRequest{
		UserId:        int64(claims.UserID),
		TokenId:       claims.ID,
		ExpiresAtUnix: claims.ExpiresAt.Unix(),
	})
	return mapRPCError(err)
}

func (c *Client) ChangePassword(ctx context.Context, req *authmodule.ChangePasswordReq, userID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return fmt.Errorf("encode change password request: %w", err)
	}
	_, err = c.rpc.ChangePassword(ctx, &iamv1.UserBodyRequest{
		UserId: int64(userID),
		Body:   body,
	})
	return mapRPCError(err)
}

func (c *Client) GetProfile(ctx context.Context, userID int) (*authmodule.UserProfileResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	resp, err := c.rpc.GetProfile(ctx, &iamv1.UserIDRequest{UserId: int64(userID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[authmodule.UserProfileResp](resp.GetData())
}

func (c *Client) CreateAccessKey(ctx context.Context, userID int, req *authmodule.CreateAccessKeyReq) (*authmodule.AccessKeyWithSecretResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode create access key request: %w", err)
	}
	resp, err := c.rpc.CreateAccessKey(ctx, &iamv1.UserBodyRequest{
		UserId: int64(userID),
		Body:   body,
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[authmodule.AccessKeyWithSecretResp](resp.GetData())
}

func (c *Client) ListAccessKeys(ctx context.Context, userID int, req *authmodule.ListAccessKeyReq) (*authmodule.ListAccessKeyResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode list access keys request: %w", err)
	}
	resp, err := c.rpc.ListAccessKeys(ctx, &iamv1.UserQueryRequest{
		UserId: int64(userID),
		Query:  query,
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[authmodule.ListAccessKeyResp](resp.GetData())
}

func (c *Client) GetAccessKey(ctx context.Context, userID, accessKeyID int) (*authmodule.AccessKeyInfo, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	resp, err := c.rpc.GetAccessKey(ctx, &iamv1.UserScopedIDRequest{
		UserId: int64(userID),
		Id:     int64(accessKeyID),
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[authmodule.AccessKeyInfo](resp.GetData())
}

func (c *Client) DeleteAccessKey(ctx context.Context, userID, accessKeyID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.DeleteAccessKey(ctx, &iamv1.UserScopedIDRequest{
		UserId: int64(userID),
		Id:     int64(accessKeyID),
	})
	return mapRPCError(err)
}

func (c *Client) DisableAccessKey(ctx context.Context, userID, accessKeyID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.DisableAccessKey(ctx, &iamv1.UserScopedIDRequest{
		UserId: int64(userID),
		Id:     int64(accessKeyID),
	})
	return mapRPCError(err)
}

func (c *Client) EnableAccessKey(ctx context.Context, userID, accessKeyID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.EnableAccessKey(ctx, &iamv1.UserScopedIDRequest{
		UserId: int64(userID),
		Id:     int64(accessKeyID),
	})
	return mapRPCError(err)
}

func (c *Client) RotateAccessKey(ctx context.Context, userID, accessKeyID int) (*authmodule.AccessKeyWithSecretResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	resp, err := c.rpc.RotateAccessKey(ctx, &iamv1.UserScopedIDRequest{
		UserId: int64(userID),
		Id:     int64(accessKeyID),
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[authmodule.AccessKeyWithSecretResp](resp.GetData())
}

func (c *Client) ExchangeAccessKeyToken(ctx context.Context, req *authmodule.AccessKeyTokenReq, method, path string) (*authmodule.AccessKeyTokenResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	resp, err := c.rpc.ExchangeAccessKeyToken(ctx, &iamv1.ExchangeAccessKeyTokenRequest{
		AccessKey: req.AccessKey,
		Timestamp: req.Timestamp,
		Nonce:     req.Nonce,
		Signature: req.Signature,
		Method:    method,
		Path:      path,
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return &authmodule.AccessKeyTokenResp{
		Token:     resp.GetToken(),
		TokenType: resp.GetTokenType(),
		ExpiresAt: time.Unix(resp.GetExpiresAtUnix(), 0),
		AuthType:  resp.GetAuthType(),
		AccessKey: resp.GetAccessKey(),
	}, nil
}

func (c *Client) CreateUser(ctx context.Context, req *usermodule.CreateUserReq) (*usermodule.UserResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode create user request: %w", err)
	}
	resp, err := c.rpc.CreateUser(ctx, &iamv1.MutationRequest{Body: body})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[usermodule.UserResp](resp.GetData())
}

func (c *Client) DeleteUser(ctx context.Context, userID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.DeleteUser(ctx, &iamv1.IDRequest{Id: int64(userID)})
	return mapRPCError(err)
}

func (c *Client) GetUser(ctx context.Context, userID int) (*usermodule.UserDetailResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	resp, err := c.rpc.GetUser(ctx, &iamv1.IDRequest{Id: int64(userID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[usermodule.UserDetailResp](resp.GetData())
}

func (c *Client) ListUsers(ctx context.Context, req *usermodule.ListUserReq) (*dto.ListResp[usermodule.UserResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode list users request: %w", err)
	}
	resp, err := c.rpc.ListUsers(ctx, &iamv1.QueryRequest{Query: query})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[usermodule.UserResp]](resp.GetData())
}

func (c *Client) UpdateUser(ctx context.Context, req *usermodule.UpdateUserReq, userID int) (*usermodule.UserResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode update user request: %w", err)
	}
	resp, err := c.rpc.UpdateUser(ctx, &iamv1.UpdateByIDRequest{
		Id:   int64(userID),
		Body: body,
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[usermodule.UserResp](resp.GetData())
}

func (c *Client) AssignUserRole(ctx context.Context, userID, roleID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.AssignUserRole(ctx, &iamv1.UserRoleBindingRequest{
		UserId: int64(userID),
		RoleId: int64(roleID),
	})
	return mapRPCError(err)
}

func (c *Client) RemoveUserRole(ctx context.Context, userID, roleID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.RemoveUserRole(ctx, &iamv1.UserRoleBindingRequest{
		UserId: int64(userID),
		RoleId: int64(roleID),
	})
	return mapRPCError(err)
}

func (c *Client) AssignUserPermissions(ctx context.Context, userID int, req *usermodule.AssignUserPermissionReq) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return fmt.Errorf("encode assign user permissions request: %w", err)
	}
	_, err = c.rpc.AssignUserPermissions(ctx, &iamv1.UserBodyRequest{
		UserId: int64(userID),
		Body:   body,
	})
	return mapRPCError(err)
}

func (c *Client) RemoveUserPermissions(ctx context.Context, userID int, req *usermodule.RemoveUserPermissionReq) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return fmt.Errorf("encode remove user permissions request: %w", err)
	}
	_, err = c.rpc.RemoveUserPermissions(ctx, &iamv1.UserBodyRequest{
		UserId: int64(userID),
		Body:   body,
	})
	return mapRPCError(err)
}

func (c *Client) AssignUserContainer(ctx context.Context, userID, containerID, roleID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.AssignUserContainer(ctx, &iamv1.UserResourceBindingRequest{
		UserId:     int64(userID),
		ResourceId: int64(containerID),
		RoleId:     int64(roleID),
	})
	return mapRPCError(err)
}

func (c *Client) RemoveUserContainer(ctx context.Context, userID, containerID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.RemoveUserContainer(ctx, &iamv1.UserScopedIDRequest{
		UserId: int64(userID),
		Id:     int64(containerID),
	})
	return mapRPCError(err)
}

func (c *Client) AssignUserDataset(ctx context.Context, userID, datasetID, roleID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.AssignUserDataset(ctx, &iamv1.UserResourceBindingRequest{
		UserId:     int64(userID),
		ResourceId: int64(datasetID),
		RoleId:     int64(roleID),
	})
	return mapRPCError(err)
}

func (c *Client) RemoveUserDataset(ctx context.Context, userID, datasetID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.RemoveUserDataset(ctx, &iamv1.UserScopedIDRequest{
		UserId: int64(userID),
		Id:     int64(datasetID),
	})
	return mapRPCError(err)
}

func (c *Client) AssignUserProject(ctx context.Context, userID, projectID, roleID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.AssignUserProject(ctx, &iamv1.UserResourceBindingRequest{
		UserId:     int64(userID),
		ResourceId: int64(projectID),
		RoleId:     int64(roleID),
	})
	return mapRPCError(err)
}

func (c *Client) RemoveUserProject(ctx context.Context, userID, projectID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.RemoveUserProject(ctx, &iamv1.UserScopedIDRequest{
		UserId: int64(userID),
		Id:     int64(projectID),
	})
	return mapRPCError(err)
}

func (c *Client) CreateRole(ctx context.Context, req *rbacmodule.CreateRoleReq) (*rbacmodule.RoleResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode create role request: %w", err)
	}
	resp, err := c.rpc.CreateRole(ctx, &iamv1.MutationRequest{Body: body})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[rbacmodule.RoleResp](resp.GetData())
}

func (c *Client) DeleteRole(ctx context.Context, roleID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.DeleteRole(ctx, &iamv1.IDRequest{Id: int64(roleID)})
	return mapRPCError(err)
}

func (c *Client) GetRole(ctx context.Context, roleID int) (*rbacmodule.RoleDetailResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	resp, err := c.rpc.GetRole(ctx, &iamv1.IDRequest{Id: int64(roleID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[rbacmodule.RoleDetailResp](resp.GetData())
}

func (c *Client) ListRoles(ctx context.Context, req *rbacmodule.ListRoleReq) (*dto.ListResp[rbacmodule.RoleResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode list roles request: %w", err)
	}
	resp, err := c.rpc.ListRoles(ctx, &iamv1.QueryRequest{Query: query})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[rbacmodule.RoleResp]](resp.GetData())
}

func (c *Client) UpdateRole(ctx context.Context, req *rbacmodule.UpdateRoleReq, roleID int) (*rbacmodule.RoleResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode update role request: %w", err)
	}
	resp, err := c.rpc.UpdateRole(ctx, &iamv1.UpdateByIDRequest{
		Id:   int64(roleID),
		Body: body,
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[rbacmodule.RoleResp](resp.GetData())
}

func (c *Client) AssignRolePermissions(ctx context.Context, roleID int, permissionIDs []int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.AssignRolePermissions(ctx, &iamv1.RolePermissionsRequest{
		RoleId:        int64(roleID),
		PermissionIds: intsToInt64s(permissionIDs),
	})
	return mapRPCError(err)
}

func (c *Client) RemoveRolePermissions(ctx context.Context, roleID int, permissionIDs []int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.RemoveRolePermissions(ctx, &iamv1.RolePermissionsRequest{
		RoleId:        int64(roleID),
		PermissionIds: intsToInt64s(permissionIDs),
	})
	return mapRPCError(err)
}

func (c *Client) ListUsersFromRole(ctx context.Context, roleID int) ([]rbacmodule.UserListItem, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	resp, err := c.rpc.ListUsersFromRole(ctx, &iamv1.IDRequest{Id: int64(roleID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	data, err := decodeStruct[[]rbacmodule.UserListItem](resp.GetData())
	if err != nil {
		return nil, err
	}
	return *data, nil
}

func (c *Client) GetPermission(ctx context.Context, permissionID int) (*rbacmodule.PermissionDetailResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	resp, err := c.rpc.GetPermission(ctx, &iamv1.IDRequest{Id: int64(permissionID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[rbacmodule.PermissionDetailResp](resp.GetData())
}

func (c *Client) ListPermissions(ctx context.Context, req *rbacmodule.ListPermissionReq) (*dto.ListResp[rbacmodule.PermissionResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode list permissions request: %w", err)
	}
	resp, err := c.rpc.ListPermissions(ctx, &iamv1.QueryRequest{Query: query})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[rbacmodule.PermissionResp]](resp.GetData())
}

func (c *Client) ListRolesFromPermission(ctx context.Context, permissionID int) ([]rbacmodule.RoleResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	resp, err := c.rpc.ListRolesFromPermission(ctx, &iamv1.IDRequest{Id: int64(permissionID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	data, err := decodeStruct[[]rbacmodule.RoleResp](resp.GetData())
	if err != nil {
		return nil, err
	}
	return *data, nil
}

func (c *Client) GetResource(ctx context.Context, resourceID int) (*rbacmodule.ResourceResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	resp, err := c.rpc.GetResource(ctx, &iamv1.IDRequest{Id: int64(resourceID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[rbacmodule.ResourceResp](resp.GetData())
}

func (c *Client) ListResources(ctx context.Context, req *rbacmodule.ListResourceReq) (*dto.ListResp[rbacmodule.ResourceResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode list resources request: %w", err)
	}
	resp, err := c.rpc.ListResources(ctx, &iamv1.QueryRequest{Query: query})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[rbacmodule.ResourceResp]](resp.GetData())
}

func (c *Client) ListResourcePermissions(ctx context.Context, resourceID int) ([]rbacmodule.PermissionResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	resp, err := c.rpc.ListResourcePermissions(ctx, &iamv1.IDRequest{Id: int64(resourceID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	data, err := decodeStruct[[]rbacmodule.PermissionResp](resp.GetData())
	if err != nil {
		return nil, err
	}
	return *data, nil
}

func (c *Client) CreateTeam(ctx context.Context, req *teammodule.CreateTeamReq, userID int) (*teammodule.TeamResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode create team request: %w", err)
	}
	resp, err := c.rpc.CreateTeam(ctx, &iamv1.CreateTeamRequest{
		UserId: int64(userID),
		Body:   body,
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[teammodule.TeamResp](resp.GetData())
}

func (c *Client) DeleteTeam(ctx context.Context, teamID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.DeleteTeam(ctx, &iamv1.TeamRequest{TeamId: int64(teamID)})
	return mapRPCError(err)
}

func (c *Client) GetTeam(ctx context.Context, teamID int) (*teammodule.TeamDetailResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	resp, err := c.rpc.GetTeam(ctx, &iamv1.TeamRequest{TeamId: int64(teamID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[teammodule.TeamDetailResp](resp.GetData())
}

func (c *Client) ListTeams(ctx context.Context, req *teammodule.ListTeamReq, userID int, isAdmin bool) (*dto.ListResp[teammodule.TeamResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode list teams request: %w", err)
	}
	resp, err := c.rpc.ListTeams(ctx, &iamv1.ListTeamsRequest{
		UserId:  int64(userID),
		IsAdmin: isAdmin,
		Query:   query,
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[teammodule.TeamResp]](resp.GetData())
}

func (c *Client) UpdateTeam(ctx context.Context, req *teammodule.UpdateTeamReq, teamID int) (*teammodule.TeamResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode update team request: %w", err)
	}
	resp, err := c.rpc.UpdateTeam(ctx, &iamv1.UpdateTeamRequest{
		TeamId: int64(teamID),
		Body:   body,
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[teammodule.TeamResp](resp.GetData())
}

func (c *Client) ListTeamProjects(ctx context.Context, req *teammodule.TeamProjectListReq, teamID int) (*dto.ListResp[teammodule.TeamProjectItem], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode list team projects request: %w", err)
	}
	resp, err := c.rpc.ListTeamProjects(ctx, &iamv1.ListTeamProjectsRequest{
		TeamId: int64(teamID),
		Query:  query,
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[teammodule.TeamProjectItem]](resp.GetData())
}

func (c *Client) AddTeamMember(ctx context.Context, req *teammodule.AddTeamMemberReq, teamID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return fmt.Errorf("encode add team member request: %w", err)
	}
	_, err = c.rpc.AddTeamMember(ctx, &iamv1.AddTeamMemberRequest{
		TeamId: int64(teamID),
		Body:   body,
	})
	return mapRPCError(err)
}

func (c *Client) RemoveTeamMember(ctx context.Context, teamID, currentUserID, targetUserID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	_, err := c.rpc.RemoveTeamMember(ctx, &iamv1.RemoveTeamMemberRequest{
		TeamId:        int64(teamID),
		CurrentUserId: int64(currentUserID),
		TargetUserId:  int64(targetUserID),
	})
	return mapRPCError(err)
}

func (c *Client) UpdateTeamMemberRole(ctx context.Context, req *teammodule.UpdateTeamMemberRoleReq, teamID, targetUserID, currentUserID int) error {
	if !c.Enabled() {
		return fmt.Errorf("iam grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return fmt.Errorf("encode update team member role request: %w", err)
	}
	_, err = c.rpc.UpdateTeamMemberRole(ctx, &iamv1.UpdateTeamMemberRoleRequest{
		TeamId:        int64(teamID),
		TargetUserId:  int64(targetUserID),
		CurrentUserId: int64(currentUserID),
		Body:          body,
	})
	return mapRPCError(err)
}

func (c *Client) ListTeamMembers(ctx context.Context, req *teammodule.ListTeamMemberReq, teamID int) (*dto.ListResp[teammodule.TeamMemberResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("iam grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode list team members request: %w", err)
	}
	resp, err := c.rpc.ListTeamMembers(ctx, &iamv1.ListTeamMembersRequest{
		TeamId: int64(teamID),
		Query:  query,
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[teammodule.TeamMemberResp]](resp.GetData())
}

var _ middleware.TokenVerifier = (*Client)(nil)

func toStructPB(value any) (*structpb.Struct, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return structpb.NewStruct(payload)
}

func decodeStruct[T any](payload *structpb.Struct) (*T, error) {
	if payload == nil {
		return nil, fmt.Errorf("iam payload is nil")
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

func intsToInt64s(items []int) []int64 {
	if len(items) == 0 {
		return nil
	}
	result := make([]int64, 0, len(items))
	for _, item := range items {
		result = append(result, int64(item))
	}
	return result
}

func mapRPCError(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.Unauthenticated:
		return fmt.Errorf("%w: %s", consts.ErrAuthenticationFailed, st.Message())
	case codes.PermissionDenied:
		return fmt.Errorf("%w: %s", consts.ErrPermissionDenied, st.Message())
	case codes.InvalidArgument:
		return fmt.Errorf("%w: %s", consts.ErrBadRequest, st.Message())
	case codes.NotFound:
		return fmt.Errorf("%w: %s", consts.ErrNotFound, st.Message())
	case codes.AlreadyExists:
		return fmt.Errorf("%w: %s", consts.ErrAlreadyExists, st.Message())
	default:
		return fmt.Errorf("iam rpc failed: %w", err)
	}
}
