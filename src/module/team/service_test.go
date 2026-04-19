package team

import (
	"context"
	"regexp"
	"testing"
	"time"

	"aegis/consts"
	"aegis/dto"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type stubProjectReader struct{}

func (stubProjectReader) CountProjects(context.Context, int) (int, error) {
	return 0, nil
}

func (stubProjectReader) ListProjects(context.Context, *TeamProjectListReq, int) (*dto.ListResp[TeamProjectItem], error) {
	return &dto.ListResp[TeamProjectItem]{Items: []TeamProjectItem{}}, nil
}

func newTeamService(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return NewService(NewRepository(db), stubProjectReader{}), mock, func() {
		_ = sqlDB.Close()
	}
}

func TestTeamServiceListTeamsSuccess(t *testing.T) {
	service, mock, cleanup := newTeamService(t)
	defer cleanup()

	now := time.Now()
	status := consts.CommonEnabled

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `teams` WHERE status = ?")).
		WithArgs(status).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `teams` WHERE status = ? LIMIT ?")).
		WithArgs(status, 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "is_public", "status", "created_at", "updated_at",
		}).AddRow(1, "platform", "platform team", true, consts.CommonEnabled, now, now))

	resp, err := service.ListTeams(t.Context(), &ListTeamReq{Status: &status}, 1, true)

	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	require.Equal(t, "platform", resp.Items[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTeamServiceRemoveMemberSelfRejected(t *testing.T) {
	service := NewService(nil, stubProjectReader{})

	err := service.RemoveMember(t.Context(), 1, 7, 7)

	require.Error(t, err)
	require.ErrorContains(t, err, "cannot remove yourself from the team")
}
