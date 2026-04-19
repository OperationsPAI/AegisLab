package user

import (
	"database/sql/driver"
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

func newUserTestService(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return NewService(NewRepository(db)), mock, func() {
		_ = sqlDB.Close()
	}
}

func TestServiceCreateUserValidationError(t *testing.T) {
	service := NewService(nil)

	_, err := service.CreateUser(t.Context(), &CreateUserReq{
		Username: "demo",
		Email:    "demo@example.com",
		Password: "short",
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "validation failed")
	require.ErrorContains(t, err, "password must be at least 8 characters")
}

func TestServiceListUsersSuccess(t *testing.T) {
	service, mock, cleanup := newUserTestService(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users` WHERE status != ?")).
		WithArgs(consts.CommonDeleted).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE status != ? LIMIT ?")).
		WithArgs(consts.CommonDeleted, 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "full_name", "avatar", "phone", "last_login_at",
			"is_active", "status", "created_at", "updated_at",
		}).AddRow(1, "demo", "demo@example.com", "hashed", "Demo User", "", "", nil, true, consts.CommonEnabled, now, now))

	resp, err := service.ListUsers(t.Context(), &ListUserReq{})

	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	require.Equal(t, "demo", resp.Items[0].Username)
	require.Equal(t, 1, resp.Pagination.Page)
	require.Equal(t, int(consts.PageSizeMedium), resp.Pagination.Size)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceCreateUserSuccess(t *testing.T) {
	service, mock, cleanup := newUserTestService(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE username = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("demo", 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE email = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("demo@example.com", 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users` (`username`,`email`,`password`,`full_name`,`avatar`,`phone`,`last_login_at`,`is_active`,`status`,`created_at`,`updated_at`) VALUES (?,?,?,?,?,?,?,?,?,?,?)")).
		WithArgs("demo", "demo@example.com", passwordHashMatcher{plain: "password123"}, "Demo User", "", "", nil, true, consts.CommonEnabled, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(5, 1))
	mock.ExpectCommit()

	resp, err := service.CreateUser(t.Context(), &CreateUserReq{
		Username: "demo",
		Email:    "demo@example.com",
		Password: "password123",
		FullName: "Demo User",
	})

	require.NoError(t, err)
	require.Equal(t, 5, resp.ID)
	require.Equal(t, "demo", resp.Username)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceGetUserDetailSuccess(t *testing.T) {
	service, mock, cleanup := newUserTestService(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "full_name", "avatar", "phone", "last_login_at",
			"is_active", "status", "created_at", "updated_at",
		}).AddRow(1, "demo", "demo@example.com", "hashed", "Demo User", "", "", nil, true, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `roles`.`id`,`roles`.`name`,`roles`.`display_name`,`roles`.`description`,`roles`.`is_system`,`roles`.`status`,`roles`.`created_at`,`roles`.`updated_at`,`roles`.`active_name` FROM `roles` JOIN user_roles ur ON ur.role_id = roles.id WHERE ur.user_id = ? AND roles.status = ?")).
		WithArgs(1, consts.CommonEnabled).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "display_name", "description", "is_system", "status", "created_at", "updated_at", "active_name",
		}).AddRow(2, "admin", "Admin", "", true, consts.CommonEnabled, now, now, "admin"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `permissions`.`id`,`permissions`.`name`,`permissions`.`display_name`,`permissions`.`description`,`permissions`.`action`,`permissions`.`scope`,`permissions`.`resource_id`,`permissions`.`is_system`,`permissions`.`status`,`permissions`.`created_at`,`permissions`.`updated_at`,`permissions`.`active_name` FROM `permissions` JOIN user_permissions up ON up.permission_id = permissions.id WHERE up.user_id = ? AND permissions.status = ?")).
		WithArgs(1, consts.CommonEnabled).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "display_name", "description", "action", "scope", "resource_id", "is_system", "status", "created_at", "updated_at", "active_name",
		}).AddRow(3, "user.read", "User Read", "", consts.ActionRead, consts.ScopeAll, 1, true, consts.CommonEnabled, now, now, "user.read"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `user_containers` WHERE user_id = ? AND status != ?")).
		WithArgs(1, consts.CommonDeleted).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "container_id", "role_id", "status", "created_at", "updated_at"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `user_datasets` WHERE user_id = ? AND status != ?")).
		WithArgs(1, consts.CommonDeleted).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "dataset_id", "role_id", "status", "created_at", "updated_at"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `user_projects` WHERE user_id = ? AND status != ?")).
		WithArgs(1, consts.CommonDeleted).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "project_id", "role_id", "workspace_config", "status", "created_at", "updated_at", "active_user_project"}))

	resp, err := service.GetUserDetail(t.Context(), 1)

	require.NoError(t, err)
	require.Equal(t, "demo", resp.Username)
	require.Len(t, resp.GlobalRoles, 1)
	require.Len(t, resp.Permissions, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceDeleteUserSuccess(t *testing.T) {
	service, mock, cleanup := newUserTestService(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? AND status != ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(1, consts.CommonDeleted, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "full_name", "avatar", "phone", "last_login_at",
			"is_active", "status", "created_at", "updated_at",
		}).AddRow(1, "demo", "demo@example.com", "hashed", "Demo User", "", "", nil, true, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? AND status != ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(1, consts.CommonDeleted, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "full_name", "avatar", "phone", "last_login_at",
			"is_active", "status", "created_at", "updated_at",
		}).AddRow(1, "demo", "demo@example.com", "hashed", "Demo User", "", "", nil, true, consts.CommonEnabled, now, now))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `user_containers` SET `status`=?,`updated_at`=? WHERE user_id = ? AND status != ?")).
		WithArgs(consts.CommonDeleted, sqlmock.AnyArg(), 1, consts.CommonDeleted).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `user_datasets` SET `status`=?,`updated_at`=? WHERE user_id = ? AND status != ?")).
		WithArgs(consts.CommonDeleted, sqlmock.AnyArg(), 1, consts.CommonDeleted).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `user_projects` SET `status`=?,`updated_at`=? WHERE user_id = ? AND status != ?")).
		WithArgs(consts.CommonDeleted, sqlmock.AnyArg(), 1, consts.CommonDeleted).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `user_permissions` WHERE user_id = ?")).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `user_roles` WHERE user_id = ?")).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `status`=?,`updated_at`=? WHERE id = ? AND status != ?")).
		WithArgs(consts.CommonDeleted, sqlmock.AnyArg(), 1, consts.CommonDeleted).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := service.DeleteUser(t.Context(), 1)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
