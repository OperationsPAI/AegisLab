package evaluationmodule

import (
	"testing"

	executionmodule "aegis/module/execution"
)

func TestListEvaluationExecutionsRequiresQuerySource(t *testing.T) {
	service := &Service{}

	_, err := service.listEvaluationExecutionsByDatapack(t.Context(), &executionmodule.EvaluationExecutionsByDatapackReq{})
	if err == nil {
		t.Fatalf("expected datapack query to fail without orchestrator or execution service")
	}

	_, err = service.listEvaluationExecutionsByDataset(t.Context(), &executionmodule.EvaluationExecutionsByDatasetReq{})
	if err == nil {
		t.Fatalf("expected dataset query to fail without orchestrator or execution service")
	}
}
