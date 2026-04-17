package authmodule

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"
	"time"

	"aegis/consts"
	"aegis/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type passwordHashMatcher struct {
	plain string
}

func (m passwordHashMatcher) Match(v driver.Value) bool {
	hash, ok := v.(string)
	if !ok {
		return false
	}
	return utils.VerifyPassword(m.plain, hash)
}

func newAuthService(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	service := NewService(NewUserRepository(db), NewRoleRepository(db), NewAccessKeyRepository(db), &TokenStore{})
	return service, mock, func() {
		_ = sqlDB.Close()
	}
}

func TestAuthServiceRegisterSuccess(t *testing.T) {
	service, mock, cleanup := newAuthService(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE username = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("new_user", 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE email = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("new@example.com", 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users` (`username`,`email`,`password`,`full_name`,`avatar`,`phone`,`last_login_at`,`is_active`,`status`,`created_at`,`updated_at`) VALUES (?,?,?,?,?,?,?,?,?,?,?)")).
		WithArgs("new_user", "new@example.com", passwordHashMatcher{plain: "password123"}, "", "", "", nil, true, consts.CommonEnabled, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(9, 1))
	mock.ExpectCommit()

	resp, err := service.Register(t.Context(), &RegisterReq{
		Username: "new_user",
		Email:    "new@example.com",
		Password: "password123",
	})

	require.NoError(t, err)
	require.Equal(t, "new_user", resp.Username)
	require.Equal(t, 9, resp.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthServiceLoginSuccess(t *testing.T) {
	service, mock, cleanup := newAuthService(t)
	defer cleanup()

	now := time.Now()
	hashedPassword, err := utils.HashPassword("password123")
	require.NoError(t, err)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE username = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("demo_user", 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "full_name", "avatar", "phone", "last_login_at",
			"is_active", "status", "created_at", "updated_at",
		}).AddRow(7, "demo_user", "demo@example.com", hashedPassword, "Demo User", "", "", nil, true, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `roles`.`id`,`roles`.`name`,`roles`.`display_name`,`roles`.`description`,`roles`.`is_system`,`roles`.`status`,`roles`.`created_at`,`roles`.`updated_at`,`roles`.`active_name` FROM `roles` JOIN user_roles ur ON ur.role_id = roles.id WHERE ur.user_id = ? AND roles.status = ?")).
		WithArgs(7, consts.CommonEnabled).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "display_name", "description", "is_system", "status", "created_at", "updated_at", "active_name",
		}).AddRow(1, consts.RoleAdmin.String(), "Admin", "", true, consts.CommonEnabled, now, now, consts.RoleAdmin.String()))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `last_login_at`=?,`updated_at`=? WHERE id = ? AND status != ?")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 7, consts.CommonDeleted).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `roles`.`id`,`roles`.`name`,`roles`.`display_name`,`roles`.`description`,`roles`.`is_system`,`roles`.`status`,`roles`.`created_at`,`roles`.`updated_at`,`roles`.`active_name` FROM `roles` JOIN user_roles ur ON ur.role_id = roles.id WHERE ur.user_id = ? AND roles.status = ?")).
		WithArgs(7, consts.CommonEnabled).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "display_name", "description", "is_system", "status", "created_at", "updated_at", "active_name",
		}).AddRow(1, consts.RoleAdmin.String(), "Admin", "", true, consts.CommonEnabled, now, now, consts.RoleAdmin.String()))

	resp, err := service.Login(t.Context(), &LoginReq{
		Username: "demo_user",
		Password: "password123",
	})

	require.NoError(t, err)
	require.Equal(t, "demo_user", resp.User.Username)
	require.Equal(t, consts.RoleAdmin.String(), resp.User.Role)

	claims, err := utils.ValidateToken(resp.Token)
	require.NoError(t, err)
	require.Equal(t, 7, claims.UserID)
	require.True(t, claims.IsAdmin)
	require.Contains(t, claims.Roles, consts.RoleAdmin.String())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthServiceRefreshTokenSuccess(t *testing.T) {
	service, mock, cleanup := newAuthService(t)
	defer cleanup()

	now := time.Now()
	token, _, err := utils.GenerateToken(7, "demo_user", "demo@example.com", true, false, []string{consts.RoleUser.String()})
	require.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(7, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "full_name", "avatar", "phone", "last_login_at",
			"is_active", "status", "created_at", "updated_at",
		}).AddRow(7, "demo_user", "demo@example.com", "ignored", "Demo User", "", "", nil, true, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `roles`.`id`,`roles`.`name`,`roles`.`display_name`,`roles`.`description`,`roles`.`is_system`,`roles`.`status`,`roles`.`created_at`,`roles`.`updated_at`,`roles`.`active_name` FROM `roles` JOIN user_roles ur ON ur.role_id = roles.id WHERE ur.user_id = ? AND roles.status = ?")).
		WithArgs(7, consts.CommonEnabled).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "display_name", "description", "is_system", "status", "created_at", "updated_at", "active_name",
		}).AddRow(2, consts.RoleUser.String(), "User", "", true, consts.CommonEnabled, now, now, consts.RoleUser.String()))

	resp, err := service.RefreshToken(t.Context(), &TokenRefreshReq{Token: token})

	require.NoError(t, err)
	require.NotEmpty(t, resp.Token)

	claims, err := utils.ValidateToken(resp.Token)
	require.NoError(t, err)
	require.Equal(t, 7, claims.UserID)
	require.Contains(t, claims.Roles, consts.RoleUser.String())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthServiceCreateAccessKeySuccess(t *testing.T) {
	service, mock, cleanup := newAuthService(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `user_access_keys` (`user_id`,`name`,`description`,`access_key`,`secret_hash`,`secret_ciphertext`,`last_used_at`,`expires_at`,`status`,`created_at`,`updated_at`,`active_access_key`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)")).
		WithArgs(7, "ci-bot", "SDK credential", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil, nil, consts.CommonEnabled, sqlmock.AnyArg(), sqlmock.AnyArg(), "").
		WillReturnResult(sqlmock.NewResult(11, 1))
	mock.ExpectCommit()

	resp, err := service.CreateAccessKey(t.Context(), 7, &CreateAccessKeyReq{
		Name:        "ci-bot",
		Description: "SDK credential",
	})

	require.NoError(t, err)
	require.Equal(t, 11, resp.ID)
	require.Equal(t, "ci-bot", resp.Name)
	require.NotEmpty(t, resp.AccessKey)
	require.NotEmpty(t, resp.SecretKey)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthServiceExchangeAccessKeyTokenSuccess(t *testing.T) {
	service, mock, cleanup := newAuthService(t)
	defer cleanup()

	now := time.Now()
	secret := "sk_test_secret_123456"
	secretHash, err := utils.HashPassword(secret)
	require.NoError(t, err)
	secretCiphertext, err := utils.EncryptAccessKeySecret(secret)
	require.NoError(t, err)
	req := &AccessKeyTokenReq{
		AccessKey: "ak_test_credential",
		Timestamp: fmt.Sprintf("%d", now.Unix()),
		Nonce:     "nonce_123",
	}
	req.Signature = utils.SignAccessKeyRequest(secret, req.CanonicalString("POST", "/api/v2/auth/access-key/token"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `user_access_keys` WHERE access_key = ? AND status != ? ORDER BY `user_access_keys`.`id` LIMIT ?")).
		WithArgs("ak_test_credential", consts.CommonDeleted, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "name", "description", "access_key", "secret_hash", "secret_ciphertext", "last_used_at", "expires_at", "status", "created_at", "updated_at",
		}).AddRow(5, 7, "ci-bot", "SDK credential", "ak_test_credential", secretHash, secretCiphertext, nil, nil, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(7, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "full_name", "avatar", "phone", "last_login_at",
			"is_active", "status", "created_at", "updated_at",
		}).AddRow(7, "demo_user", "demo@example.com", "ignored", "Demo User", "", "", nil, true, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `roles`.`id`,`roles`.`name`,`roles`.`display_name`,`roles`.`description`,`roles`.`is_system`,`roles`.`status`,`roles`.`created_at`,`roles`.`updated_at`,`roles`.`active_name` FROM `roles` JOIN user_roles ur ON ur.role_id = roles.id WHERE ur.user_id = ? AND roles.status = ?")).
		WithArgs(7, consts.CommonEnabled).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "display_name", "description", "is_system", "status", "created_at", "updated_at", "active_name",
		}).AddRow(2, consts.RoleUser.String(), "User", "", true, consts.CommonEnabled, now, now, consts.RoleUser.String()))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `user_access_keys` SET `last_used_at`=?,`updated_at`=? WHERE id = ?")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 5).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	resp, err := service.ExchangeAccessKeyToken(t.Context(), req, "POST", "/api/v2/auth/access-key/token")

	require.NoError(t, err)
	require.Equal(t, "Bearer", resp.TokenType)
	require.Equal(t, "access_key", resp.AuthType)

	claims, err := utils.ValidateToken(resp.Token)
	require.NoError(t, err)
	require.Equal(t, 7, claims.UserID)
	require.Equal(t, "access_key", claims.AuthType)
	require.Equal(t, 5, claims.AccessKeyID)
	require.NoError(t, mock.ExpectationsWereMet())
}
