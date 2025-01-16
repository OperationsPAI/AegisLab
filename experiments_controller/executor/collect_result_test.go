package executor

import (
	"context"
	"testing"
)

func TestExecuteCollectResult(t *testing.T) {
	datasetname := "ts-ts-preserve-service-cpu-exhaustion-hs5lgx"
	err := executeCollectResult(context.Background(), "1", map[string]interface{}{"dataset": datasetname, "execution_id": 1})
	if err != nil {
		t.Error(err)
	}
}
