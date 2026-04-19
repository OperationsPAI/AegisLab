package sdk

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newSDKService(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
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

func TestSDKServiceListEvaluationsSuccess(t *testing.T) {
	service, mock, cleanup := newSDKService(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `evaluation_data` WHERE exp_id = ? AND stage = ?")).
		WithArgs("exp-1", "judged").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `evaluation_data` WHERE exp_id = ? AND stage = ? ORDER BY id DESC LIMIT ?")).
		WithArgs("exp-1", "judged", 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "dataset", "dataset_index", "source", "raw_question", "level",
			"augmented_question", "correct_answer", "file_name", "meta", "trace_id", "trace_url", "response",
			"time_cost", "trajectories", "extracted_final_answer", "judged_response", "reasoning", "correct",
			"confidence", "exp_id", "agent_type", "model_name", "stage",
		}).AddRow(1, now, now, "demo", 0, "manual", "q", 1, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, true, 0.9, "exp-1", nil, nil, "judged"))

	resp, err := service.ListEvaluations(t.Context(), &ListSDKEvaluationReq{
		ExpID: "exp-1",
		Stage: "judged",
	})

	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	require.Equal(t, "exp-1", resp.Items[0].ExpID)
	require.Equal(t, "judged", resp.Items[0].Stage)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSDKServiceGetEvaluationSuccess(t *testing.T) {
	service, mock, cleanup := newSDKService(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `evaluation_data` WHERE id = ? ORDER BY `evaluation_data`.`id` LIMIT ?")).
		WithArgs(3, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "dataset", "dataset_index", "source", "raw_question", "level",
			"augmented_question", "correct_answer", "file_name", "meta", "trace_id", "trace_url", "response",
			"time_cost", "trajectories", "extracted_final_answer", "judged_response", "reasoning", "correct",
			"confidence", "exp_id", "agent_type", "model_name", "stage",
		}).AddRow(3, now, now, "demo", 0, "manual", "q", 1, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, false, 0.2, "exp-2", nil, nil, "rollout"))

	item, err := service.GetEvaluation(t.Context(), 3)

	require.NoError(t, err)
	require.Equal(t, 3, item.ID)
	require.Equal(t, "exp-2", item.ExpID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSDKServiceListExperimentsSuccess(t *testing.T) {
	service, mock, cleanup := newSDKService(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT `exp_id` FROM `evaluation_data`")).
		WillReturnRows(sqlmock.NewRows([]string{"exp_id"}).AddRow("exp-1").AddRow("exp-2"))

	resp, err := service.ListExperiments(t.Context())

	require.NoError(t, err)
	require.Equal(t, []string{"exp-1", "exp-2"}, resp.Experiments)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSDKServiceListDatasetSamplesSuccess(t *testing.T) {
	service, mock, cleanup := newSDKService(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `data` WHERE dataset = ?")).
		WithArgs("gsm8k").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `data` WHERE dataset = ? ORDER BY id DESC LIMIT ?")).
		WithArgs("gsm8k", 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "dataset", "index", "source", "source_index", "question", "answer", "topic", "level", "file_name", "meta", "tags",
		}).AddRow(2, "gsm8k", 1, "manual", 0, "question", "answer", "math", 2, "sample.json", nil, nil))

	resp, err := service.ListDatasetSamples(t.Context(), &ListSDKDatasetSampleReq{
		Dataset: "gsm8k",
	})

	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	require.Equal(t, "gsm8k", resp.Items[0].Dataset)
	require.NoError(t, mock.ExpectationsWereMet())
}
