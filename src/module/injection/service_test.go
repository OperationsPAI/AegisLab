package injection

import (
	"regexp"
	"testing"
	"time"

	"aegis/consts"
	"aegis/dto"
	redis "aegis/infra/redis"
	"aegis/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newInjectionService(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()

	addr, cleanupRedis := testutil.StartRedisStub(t)
	viper.Set("redis.host", addr)

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return NewService(NewRepository(db), nil, nil, redis.NewGateway(nil)), mock, func() {
		cleanupRedis()
		_ = sqlDB.Close()
	}
}

func TestServiceSearchNilRequest(t *testing.T) {
	service := NewService(nil, nil, nil, nil)

	_, err := service.Search(t.Context(), nil, nil)

	require.Error(t, err)
	require.ErrorContains(t, err, "search injection request is nil")
}

func TestServiceListNoIssuesEmptyLabelsSucceeds(t *testing.T) {
	service := NewService(nil, nil, nil, nil)

	resp, err := service.ListNoIssues(t.Context(), &ListInjectionNoIssuesReq{}, nil)

	require.NoError(t, err)
	require.Nil(t, resp)
}

func TestServiceListProjectInjectionsSuccess(t *testing.T) {
	service, mock, cleanup := newInjectionService(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `projects` WHERE id = ? ORDER BY `projects`.`id` LIMIT ?")).
		WithArgs(7, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "team_id", "is_public", "status", "created_at", "updated_at",
		}).AddRow(7, "demo-project", "demo", nil, true, consts.CommonEnabled, now, now))
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `fault_injections` JOIN tasks ON tasks\\.id = fault_injections\\.task_id JOIN traces on traces\\.id = tasks\\.trace_id WHERE traces\\.project_id = \\? AND fault_injections\\.status != \\?").
		WithArgs(7, consts.CommonDeleted).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT `fault_injections`\\.`id`,`fault_injections`\\.`name`.*FROM `fault_injections` JOIN tasks ON tasks\\.id = fault_injections\\.task_id JOIN traces on traces\\.id = tasks\\.trace_id WHERE traces\\.project_id = \\? AND fault_injections\\.status != \\? ORDER BY fault_injections\\.updated_at DESC LIMIT \\?").
		WithArgs(7, consts.CommonDeleted, 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "source", "fault_type", "category", "description", "display_config", "engine_config", "groundtruths", "groundtruth_source", "pre_duration", "start_time", "end_time", "benchmark_id", "pedestal_id", "task_id", "state", "status", "created_at", "updated_at",
		}))

	resp, err := service.ListProjectInjections(t.Context(), &ListInjectionReq{}, 7)

	require.NoError(t, err)
	require.Empty(t, resp.Items)
	require.Equal(t, int64(0), resp.Pagination.Total)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceSubmitDatapackBuildingSuccess(t *testing.T) {
	addr, cleanupRedis := testutil.StartRedisStub(t)
	defer cleanupRedis()
	viper.Set("redis.host", addr)

	service, mock, cleanup := newInjectionService(t)
	defer cleanup()

	mock.MatchExpectationsInOrder(false)

	now := time.Now()
	start := now.Add(-10 * time.Minute)
	end := now.Add(-2 * time.Minute)
	projectID := 9
	datapackName := "dp-build"

	mock.ExpectQuery("SELECT .* FROM container_versions cv .*").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "name_major", "name_minor", "name_patch", "github_link", "registry", "namespace", "repository", "tag", "command", "usage_count", "container_id", "user_id", "status", "created_at", "updated_at",
		}).AddRow(4, "1.0.0", 1, 0, 0, "", "docker.io", "", "bench", "latest", "", 0, 6, 1, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `containers` WHERE `containers`.`id` = ?")).
		WithArgs(6).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "type", "readme", "is_public", "status", "created_at", "updated_at",
		}).AddRow(6, "bench", consts.ContainerTypeBenchmark, "", true, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `fault_injections` WHERE name = ? AND status != ? ORDER BY `fault_injections`.`id` LIMIT ?")).
		WithArgs(datapackName, consts.CommonDeleted, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "source", "fault_type", "category", "description", "display_config", "engine_config", "groundtruths", "groundtruth_source", "pre_duration", "start_time", "end_time", "benchmark_id", "pedestal_id", "task_id", "state", "status", "created_at", "updated_at",
		}).AddRow(11, datapackName, consts.DatapackSourceInjection, 0, "ts", "", nil, "{}", "[]", "auto", 5, start, end, nil, nil, nil, consts.DatapackInjectSuccess, consts.CommonEnabled, now, now))
	mock.ExpectQuery("SELECT .* FROM `fault_injection_labels` .*").
		WillReturnRows(sqlmock.NewRows([]string{"fault_injection_id", "label_id"}))
	mock.ExpectQuery("SELECT .* FROM `parameter_configs` JOIN container_version_env_vars .*").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "config_key", "type", "category", "value_type", "description", "default_value", "template_string", "required", "overridable",
		}))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `traces`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `tasks`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	resp, err := service.SubmitDatapackBuilding(t.Context(), &SubmitDatapackBuildingReq{
		Specs: []BuildingSpec{
			{
				Benchmark: dto.ContainerSpec{
					ContainerRef: dto.ContainerRef{Name: "bench", Version: "1.0.0"},
				},
				Datapack: &datapackName,
			},
		},
	}, "group-build", 1, &projectID)

	require.NoError(t, err)
	require.Equal(t, "group-build", resp.GroupID)
	require.Len(t, resp.Items, 1)
	require.NotEmpty(t, resp.Items[0].TaskID)
	require.NotEmpty(t, resp.Items[0].TraceID)
	require.NoError(t, mock.ExpectationsWereMet())
}
