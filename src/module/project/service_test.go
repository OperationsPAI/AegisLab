package projectmodule

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

func newProjectService(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
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

func TestProjectServiceListProjectsSuccess(t *testing.T) {
	service, mock, cleanup := newProjectService(t)
	defer cleanup()

	now := time.Now()
	isPublic := true
	status := consts.CommonEnabled

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `projects` WHERE is_public = ? AND status = ?")).
		WithArgs(isPublic, status).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `projects` WHERE is_public = ? AND status = ? LIMIT ?")).
		WithArgs(isPublic, status, 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "team_id", "is_public", "status", "created_at", "updated_at",
		}).AddRow(1, "demo-project", "demo", nil, true, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT labels.*, pl.project_id FROM `labels` JOIN project_labels pl ON pl.label_id = labels.id WHERE pl.project_id IN (?)")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "label_key", "label_value", "category", "description", "color", "usage_count", "is_system", "status", "created_at", "updated_at", "project_id",
		}).AddRow(10, "env", "prod", consts.ProjectCategory, "", "#1890ff", 1, false, consts.CommonEnabled, now, now, 1))
	mock.ExpectQuery("SELECT tr\\.project_id, COUNT\\(\\*\\) as count, MAX\\(fi\\.updated_at\\) as last_at FROM fault_injections fi .* WHERE tr\\.project_id IN \\(\\?\\) GROUP BY `tr`\\.`project_id`").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"project_id", "count", "last_at"}).AddRow(1, 2, now))
	mock.ExpectQuery("SELECT tr\\.project_id, COUNT\\(\\*\\) as count, MAX\\(e\\.updated_at\\) as last_at FROM executions e .* WHERE tr\\.project_id IN \\(\\?\\) GROUP BY `tr`\\.`project_id`").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"project_id", "count", "last_at"}).AddRow(1, 3, now))

	resp, err := service.ListProjects(t.Context(), &ListProjectReq{
		IsPublic: &isPublic,
		Status:   &status,
	})

	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	require.Equal(t, "demo-project", resp.Items[0].Name)
	require.Len(t, resp.Items[0].Labels, 1)
	require.Equal(t, 2, resp.Items[0].InjectionCount)
	require.Equal(t, 3, resp.Items[0].ExecutionCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectServiceCreateProjectSuccess(t *testing.T) {
	service, mock, cleanup := newProjectService(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `roles` WHERE name = ? AND status != ? ORDER BY `roles`.`id` LIMIT ?")).
		WithArgs(consts.RoleProjectAdmin.String(), consts.CommonDeleted, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "display_name", "description", "is_system", "status", "created_at", "updated_at",
		}).AddRow(3, consts.RoleProjectAdmin.String(), "Project Admin", "", true, consts.CommonEnabled, time.Now(), time.Now()))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `projects` (`name`,`description`,`team_id`,`is_public`,`status`,`created_at`,`updated_at`) VALUES (?,?,?,?,?,?,?)")).
		WithArgs("demo-project", "demo", nil, true, consts.CommonEnabled, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(11, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `user_projects` (`user_id`,`project_id`,`role_id`,`workspace_config`,`status`,`created_at`,`updated_at`,`active_user_project`) VALUES (?,?,?,?,?,?,?,?)")).
		WithArgs(7, 11, 3, "", consts.CommonEnabled, sqlmock.AnyArg(), sqlmock.AnyArg(), "").
		WillReturnResult(sqlmock.NewResult(21, 1))
	mock.ExpectCommit()

	isPublic := true
	resp, err := service.CreateProject(t.Context(), &CreateProjectReq{
		Name:        "demo-project",
		Description: "demo",
		IsPublic:    &isPublic,
	}, 7)

	require.NoError(t, err)
	require.Equal(t, 11, resp.ID)
	require.Equal(t, "demo-project", resp.Name)
	require.True(t, resp.IsPublic)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectServiceGetProjectDetailSuccess(t *testing.T) {
	service, mock, cleanup := newProjectService(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `projects` WHERE id = ? ORDER BY `projects`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "team_id", "is_public", "status", "created_at", "updated_at",
		}).AddRow(1, "demo-project", "demo", nil, true, consts.CommonEnabled, now, now))
	mock.ExpectQuery("SELECT tr\\.project_id, COUNT\\(\\*\\) as count, MAX\\(fi\\.updated_at\\) as last_at FROM fault_injections fi .* WHERE tr\\.project_id IN \\(\\?\\) GROUP BY `tr`\\.`project_id`").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"project_id", "count", "last_at"}).AddRow(1, 2, now))
	mock.ExpectQuery("SELECT tr\\.project_id, COUNT\\(\\*\\) as count, MAX\\(e\\.updated_at\\) as last_at FROM executions e .* WHERE tr\\.project_id IN \\(\\?\\) GROUP BY `tr`\\.`project_id`").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"project_id", "count", "last_at"}).AddRow(1, 3, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `user_projects` WHERE project_id = ? AND status = ?")).
		WithArgs(1, consts.CommonEnabled).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))

	resp, err := service.GetProjectDetail(t.Context(), 1)

	require.NoError(t, err)
	require.Equal(t, "demo-project", resp.Name)
	require.Equal(t, 4, resp.UserCount)
	require.Equal(t, 2, resp.InjectionCount)
	require.Equal(t, 3, resp.ExecutionCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectServiceUpdateProjectSuccess(t *testing.T) {
	service, mock, cleanup := newProjectService(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `projects` WHERE id = ? ORDER BY `projects`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "team_id", "is_public", "status", "created_at", "updated_at",
		}).AddRow(1, "demo-project", "old-desc", nil, true, consts.CommonEnabled, now, now))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `projects` SET `name`=?,`description`=?,`team_id`=?,`is_public`=?,`status`=?,`created_at`=?,`updated_at`=? WHERE `id` = ?")).
		WithArgs("demo-project", "new-desc", nil, false, consts.CommonDisabled, sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	description := "new-desc"
	isPublic := false
	status := consts.CommonDisabled
	resp, err := service.UpdateProject(t.Context(), &UpdateProjectReq{
		Description: &description,
		IsPublic:    &isPublic,
		Status:      &status,
	}, 1)

	require.NoError(t, err)
	require.Equal(t, "demo-project", resp.Name)
	require.False(t, resp.IsPublic)
	require.Equal(t, consts.GetStatusTypeName(consts.CommonDisabled), resp.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectServiceDeleteProjectSuccess(t *testing.T) {
	service, mock, cleanup := newProjectService(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `user_projects` SET `status`=?,`updated_at`=? WHERE project_id = ? AND status != ?")).
		WithArgs(consts.CommonDeleted, sqlmock.AnyArg(), 1, consts.CommonDeleted).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `projects` SET `status`=?,`updated_at`=? WHERE id = ? AND status != ?")).
		WithArgs(consts.CommonDeleted, sqlmock.AnyArg(), 1, consts.CommonDeleted).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := service.DeleteProject(t.Context(), 1)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectServiceManageLabelsNilRequest(t *testing.T) {
	service := NewService(nil)

	_, err := service.ManageProjectLabels(t.Context(), nil, 1)

	require.Error(t, err)
	require.ErrorContains(t, err, "manage project labels request is nil")
}
