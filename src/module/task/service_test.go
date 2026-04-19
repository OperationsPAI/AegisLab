package task

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"aegis/consts"
	lokiinfra "aegis/infra/loki"
	"aegis/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newTaskService(t *testing.T, gateway *LokiGateway) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	if gateway == nil {
		gateway = NewLokiGateway(&lokiinfra.Client{})
	}

	service := NewService(NewRepository(db), NewTaskLogService(NewRepository(db), nil, gateway), gateway)
	return service, mock, func() {
		_ = sqlDB.Close()
	}
}

func TestTaskServiceListSuccess(t *testing.T) {
	service, mock, cleanup := newTaskService(t, nil)
	defer cleanup()

	now := time.Now()
	state := consts.TaskPending
	req := &ListTaskReq{State: &state}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `tasks` WHERE state = ?")).
		WithArgs(state).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `tasks` WHERE state = ? ORDER BY created_at DESC LIMIT ?")).
		WithArgs(state, 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "type", "immediate", "execute_time", "cron_expr", "payload", "trace_id", "parent_task_id",
			"level", "sequence", "state", "status", "created_at", "updated_at",
		}).AddRow("task-1", consts.TaskTypeRunAlgorithm, true, 0, "", "{}", "trace-1", nil, 0, 0, consts.TaskPending, consts.CommonEnabled, now, now))

	resp, err := service.List(t.Context(), req)

	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	require.Equal(t, "task-1", resp.Items[0].ID)
	require.Equal(t, consts.GetTaskStateName(consts.TaskPending), resp.Items[0].State)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskServiceQueryHistoricalLogsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/loki/api/v1/query_range", r.URL.Path)
		require.True(t, strings.Contains(r.URL.Query().Get("query"), `task_id="task-1"`))
		_, _ = w.Write([]byte(`{
			"status":"success",
			"data":{
				"resultType":"streams",
				"result":[{
					"stream":{"trace_id":"trace-1","job_id":"job-1"},
					"values":[
						["1710000000000000000","first log"],
						["1710000001000000000","second log"]
					]
				}]
			}
		}`))
	}))
	defer server.Close()

	viper.Set("loki.address", server.URL)
	viper.Set("loki.max_entries", 100)

	gateway := NewLokiGateway(lokiinfra.NewClient())
	service, _, cleanup := newTaskService(t, gateway)
	defer cleanup()

	logs := service.queryHistoricalLogs(context.Background(), &model.Task{
		ID:        "task-1",
		CreatedAt: time.Unix(1710000000, 0),
	})

	require.Equal(t, []string{"first log", "second log"}, logs)
}
