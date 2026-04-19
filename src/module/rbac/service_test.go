package rbac

import (
	"regexp"
	"testing"
	"time"

	"aegis/consts"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newRBACService(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
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

func TestServiceListRolesSuccess(t *testing.T) {
	service, mock, cleanup := newRBACService(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `roles`")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `roles` ORDER BY updated_at DESC LIMIT ?")).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "display_name", "description", "is_system", "status", "created_at", "updated_at",
		}).AddRow(1, "admin", "Admin", "system admin", false, consts.CommonEnabled, now, now))

	resp, err := service.ListRoles(t.Context(), &ListRoleReq{})

	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	require.Equal(t, "admin", resp.Items[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceGetRoleNotFound(t *testing.T) {
	service, mock, cleanup := newRBACService(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `roles` WHERE id = ? and status != ? ORDER BY `roles`.`id` LIMIT ?")).
		WithArgs(99, consts.CommonDeleted, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := service.GetRole(t.Context(), 99)

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to get role")
	require.NoError(t, mock.ExpectationsWereMet())
}
