package execution

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

func newExecutionService(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
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

	return NewService(NewRepository(db), redis.NewGateway(nil)), mock, func() {
		cleanupRedis()
		_ = sqlDB.Close()
	}
}

func TestServiceListAvailableLabelsSuccess(t *testing.T) {
	service, mock, cleanup := newExecutionService(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `labels` WHERE status != ? ORDER BY usage_count DESC, created_at DESC")).
		WithArgs(consts.CommonDeleted).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "label_key", "label_value", "category", "description", "color", "usage_count", "is_system", "status", "created_at", "updated_at",
		}).AddRow(1, "source", "manual", consts.ExecutionCategory, "manual source", "#1890ff", 2, false, consts.CommonEnabled, now, now))

	labels, err := service.ListAvailableLabels(t.Context())

	require.NoError(t, err)
	require.Len(t, labels, 1)
	require.Equal(t, "source", labels[0].Key)
	require.Equal(t, "manual", labels[0].Value)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceBatchDeleteEmptyRequestSucceeds(t *testing.T) {
	service := NewService(nil, nil)

	err := service.BatchDelete(t.Context(), &BatchDeleteExecutionReq{})

	require.NoError(t, err)
}

func TestServiceListExecutionsSuccessWithLabelFilter(t *testing.T) {
	service, mock, cleanup := newExecutionService(t)
	defer cleanup()

	status := consts.CommonEnabled
	req := &ListExecutionReq{
		Status: &status,
		Labels: []string{"source:manual"},
	}

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `executions` WHERE status = \\? AND executions\\.id IN \\(SELECT eil\\.execution_id FROM execution_injection_labels eil JOIN labels ON labels\\.id = eil\\.label_id WHERE labels\\.label_key = \\? AND labels\\.label_value = \\?\\)").
		WithArgs(status, "source", "manual").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT \\* FROM `executions` WHERE status = \\? AND executions\\.id IN \\(SELECT eil\\.execution_id FROM execution_injection_labels eil JOIN labels ON labels\\.id = eil\\.label_id WHERE labels\\.label_key = \\? AND labels\\.label_value = \\?\\) ORDER BY updated_at DESC LIMIT \\?").
		WithArgs(status, "source", "manual", 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "duration", "task_id", "algorithm_version_id", "datapack_id", "dataset_version_id", "state", "status", "created_at", "updated_at",
		}))

	resp, err := service.ListExecutions(t.Context(), req)

	require.NoError(t, err)
	require.Empty(t, resp.Items)
	require.Equal(t, int64(0), resp.Pagination.Total)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceUploadDetectorResultsSuccess(t *testing.T) {
	service, mock, cleanup := newExecutionService(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `executions` WHERE id = ? AND status != ? ORDER BY `executions`.`id` LIMIT ?")).
		WithArgs(12, consts.CommonDeleted, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "duration", "task_id", "algorithm_version_id", "datapack_id", "dataset_version_id", "state", "status", "created_at", "updated_at",
		}).AddRow(12, 0, nil, 5, 7, nil, consts.ExecutionInitial, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `container_versions` WHERE `container_versions`.`id` = ?")).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "container_id", "registry", "namespace", "repository", "tag", "status", "created_at", "updated_at",
		}).AddRow(5, "1.0.0", 8, "docker.io", "", "algo", "latest", consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `containers` WHERE `containers`.`id` = ?")).
		WithArgs(8).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "type", "readme", "is_public", "status", "created_at", "updated_at",
		}).AddRow(8, "algo", consts.ContainerTypeAlgorithm, "", true, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `fault_injections` WHERE `fault_injections`.`id` = ?")).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "source", "fault_type", "category", "description", "engine_config", "groundtruth_source", "pre_duration", "benchmark_id", "pedestal_id", "task_id", "state", "status", "created_at", "updated_at",
		}).AddRow(7, "dp-1", consts.DatapackSourceInjection, 0, "train-ticket", "", "{}", "auto", 0, nil, nil, nil, consts.DatapackInitial, consts.CommonEnabled, now, now))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `executions` SET `duration`=?,`updated_at`=? WHERE id = ? AND status != ?")).
		WithArgs(12.5, sqlmock.AnyArg(), 12, consts.CommonDeleted).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `detector_results`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	duration := 12.5
	resp, err := service.UploadDetectorResults(t.Context(), &UploadDetectorResultReq{
		Duration: duration,
		Results: []DetectorResultItem{
			{SpanName: "checkout", Issues: `{"latency":true}`},
		},
	}, 12)

	require.NoError(t, err)
	require.Equal(t, 1, resp.ResultCount)
	require.True(t, resp.HasAnomalies)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceUploadGranularityResultsSuccess(t *testing.T) {
	service, mock, cleanup := newExecutionService(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `executions` WHERE id = ? AND status != ? ORDER BY `executions`.`id` LIMIT ?")).
		WithArgs(15, consts.CommonDeleted, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "duration", "task_id", "algorithm_version_id", "datapack_id", "dataset_version_id", "state", "status", "created_at", "updated_at",
		}).AddRow(15, 0, nil, 6, 9, nil, consts.ExecutionInitial, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `container_versions` WHERE `container_versions`.`id` = ?")).
		WithArgs(6).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "container_id", "registry", "namespace", "repository", "tag", "status", "created_at", "updated_at",
		}).AddRow(6, "1.0.0", 10, "docker.io", "", "algo", "latest", consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `containers` WHERE `containers`.`id` = ?")).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "type", "readme", "is_public", "status", "created_at", "updated_at",
		}).AddRow(10, "locator", consts.ContainerTypeAlgorithm, "", true, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `fault_injections` WHERE `fault_injections`.`id` = ?")).
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "source", "fault_type", "category", "description", "engine_config", "groundtruth_source", "pre_duration", "benchmark_id", "pedestal_id", "task_id", "state", "status", "created_at", "updated_at",
		}).AddRow(9, "dp-2", consts.DatapackSourceInjection, 0, "train-ticket", "", "{}", "auto", 0, nil, nil, nil, consts.DatapackInitial, consts.CommonEnabled, now, now))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `executions` SET `duration`=?,`updated_at`=? WHERE id = ? AND status != ?")).
		WithArgs(8.8, sqlmock.AnyArg(), 15, consts.CommonDeleted).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `granularity_results`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	resp, err := service.UploadGranularityResults(t.Context(), &UploadGranularityResultReq{
		Duration: 8.8,
		Results: []GranularityResultItem{
			{Level: "service", Result: "checkout", Rank: 1, Confidence: 0.91},
		},
	}, 15)

	require.NoError(t, err)
	require.Equal(t, 1, resp.ResultCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceSubmitAlgorithmExecutionSuccess(t *testing.T) {
	addr, cleanupRedis := testutil.StartRedisStub(t)
	defer cleanupRedis()
	viper.Set("redis.host", addr)

	service, mock, cleanup := newExecutionService(t)
	defer cleanup()

	mock.MatchExpectationsInOrder(false)

	now := time.Now()
	start := now.Add(-5 * time.Minute)
	end := now.Add(-1 * time.Minute)
	datapackName := "dp-1"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `projects` WHERE name = ? AND status != ? ORDER BY `projects`.`id` LIMIT ?")).
		WithArgs("demo-project", consts.CommonDeleted, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "team_id", "is_public", "status", "created_at", "updated_at",
		}).AddRow(3, "demo-project", "demo", nil, true, consts.CommonEnabled, now, now))
	mock.ExpectQuery("SELECT .* FROM container_versions cv .*").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "name_major", "name_minor", "name_patch", "github_link", "registry", "namespace", "repository", "tag", "command", "usage_count", "container_id", "user_id", "status", "created_at", "updated_at",
		}).AddRow(5, "1.0.0", 1, 0, 0, "", "docker.io", "", "algo", "latest", "", 0, 8, 1, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `containers` WHERE `containers`.`id` = ?")).
		WithArgs(8).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "type", "readme", "is_public", "status", "created_at", "updated_at",
		}).AddRow(8, "algo", consts.ContainerTypeAlgorithm, "", true, consts.CommonEnabled, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `fault_injections` WHERE name = ? AND status != ? ORDER BY `fault_injections`.`id` LIMIT ?")).
		WithArgs(datapackName, consts.CommonDeleted, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "source", "fault_type", "category", "description", "display_config", "engine_config", "groundtruths", "groundtruth_source", "pre_duration", "start_time", "end_time", "benchmark_id", "pedestal_id", "task_id", "state", "status", "created_at", "updated_at",
		}).AddRow(7, datapackName, consts.DatapackSourceInjection, 0, "ts", "", nil, "{}", "[]", "auto", 5, start, end, nil, nil, nil, consts.DatapackDetectorSuccess, consts.CommonEnabled, now, now))
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

	resp, err := service.SubmitAlgorithmExecution(t.Context(), &SubmitExecutionReq{
		ProjectName: "demo-project",
		Specs: []ExecutionSpec{
			{
				Algorithm: dto.ContainerSpec{
					ContainerRef: dto.ContainerRef{Name: "algo", Version: "1.0.0"},
				},
				Datapack: &datapackName,
			},
		},
	}, "group-1", 1)

	require.NoError(t, err)
	require.Equal(t, "group-1", resp.GroupID)
	require.Len(t, resp.Items, 1)
	require.Equal(t, 5, resp.Items[0].AlgorithmVersionID)
	require.Equal(t, 7, *resp.Items[0].DatapackID)
	require.NotEmpty(t, resp.Items[0].TaskID)
	require.NotEmpty(t, resp.Items[0].TraceID)
	require.NoError(t, mock.ExpectationsWereMet())
}
