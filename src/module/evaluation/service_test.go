package evaluation

import (
	"testing"

	execution "aegis/module/execution"
)

func TestListEvaluationExecutionsRequiresQuerySource(t *testing.T) {
	service := &Service{}

	_, err := service.listEvaluationExecutionsByDatapack(t.Context(), &execution.EvaluationExecutionsByDatapackReq{})
	if err == nil {
		t.Fatalf("expected datapack query to fail without orchestrator or execution service")
	}

	_, err = service.listEvaluationExecutionsByDataset(t.Context(), &execution.EvaluationExecutionsByDatasetReq{})
	if err == nil {
		t.Fatalf("expected dataset query to fail without orchestrator or execution service")
	}
}
