package executionmodule

import (
	"fmt"
	"strings"
	"time"

	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	"aegis/model"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

// ExecutionRef represents execution granularity results for evaluation.
type ExecutionRef struct {
	ExecutionID       int                     `json:"execution_id"`
	ExecutionDuration float64                 `json:"execution_duration"`
	DetectorResults   []DetectorResultItem    `json:"detector_results"`
	Predictions       []GranularityResultItem `json:"predictions"`
	ExecutedAt        time.Time               `json:"executed_at"`
}

func NewExecutionGranularityRef(execution *model.Execution) ExecutionRef {
	ref := &ExecutionRef{
		ExecutionID:       execution.ID,
		ExecutionDuration: execution.Duration,
		ExecutedAt:        execution.CreatedAt,
	}

	if len(execution.DetectorResults) > 0 {
		detectorItems := make([]DetectorResultItem, 0, len(execution.DetectorResults))
		for _, dr := range execution.DetectorResults {
			detectorItems = append(detectorItems, NewDetectorResultItem(&dr))
		}
		ref.DetectorResults = detectorItems
	}

	if len(execution.GranularityResults) > 0 {
		items := make([]GranularityResultItem, 0, len(execution.GranularityResults))
		for _, gr := range execution.GranularityResults {
			items = append(items, NewGranularityResultItem(&gr))
		}
		ref.Predictions = items
	}

	return *ref
}

// EvaluationExecutionsByDatapackReq resolves execution results for one algorithm/datapack pair.
type EvaluationExecutionsByDatapackReq struct {
	AlgorithmVersionID int             `json:"algorithm_version_id"`
	DatapackName       string          `json:"datapack_name"`
	FilterLabels       []dto.LabelItem `json:"filter_labels,omitempty"`
}

// EvaluationExecutionsByDatasetReq resolves execution results for one algorithm/dataset pair.
type EvaluationExecutionsByDatasetReq struct {
	AlgorithmVersionID int             `json:"algorithm_version_id"`
	DatasetVersionID   int             `json:"dataset_version_id"`
	FilterLabels       []dto.LabelItem `json:"filter_labels,omitempty"`
}

// EvaluationExecutionItem is the orchestrator-owned execution payload used by evaluation queries.
type EvaluationExecutionItem struct {
	Datapack     string              `json:"datapack"`
	Groundtruths []chaos.Groundtruth `json:"groundtruths,omitempty"`
	ExecutionRef
}

// BatchDeleteExecutionReq represents the request to batch delete executions.
type BatchDeleteExecutionReq struct {
	IDs    []int           `json:"ids" binding:"omitempty"`
	Labels []dto.LabelItem `json:"labels" binding:"omitempty"`
}

func (req *BatchDeleteExecutionReq) Validate() error {
	hasIDs := len(req.IDs) > 0
	hasLabels := len(req.Labels) > 0

	criteriaCount := 0
	if hasIDs {
		criteriaCount++
	}
	if hasLabels {
		criteriaCount++
	}

	if criteriaCount == 0 {
		return fmt.Errorf("must provide one of: ids, labels, or tags")
	}
	if criteriaCount > 1 {
		return fmt.Errorf("can only specify one deletion criteria (ids, labels, or tags)")
	}

	if hasIDs {
		for i, id := range req.IDs {
			if id <= 0 {
				return fmt.Errorf("invalid id at index %d: %d", i, id)
			}
		}
	}

	if hasLabels {
		for i, label := range req.Labels {
			if strings.TrimSpace(label.Key) == "" {
				return fmt.Errorf("empty label key at index %d", i)
			}
			if strings.TrimSpace(label.Value) == "" {
				return fmt.Errorf("empty label value at index %d", i)
			}
		}
	}

	return nil
}

// ListExecutionReq represents execution list query parameters.
type ListExecutionReq struct {
	dto.PaginationReq
	State      *consts.ExecutionState `form:"state" binding:"omitempty"`
	Status     *consts.StatusType     `form:"status" binding:"omitempty"`
	Labels     []string               `form:"labels" binding:"omitempty"`
	DatapackID *int                   `form:"datapack_id" binding:"omitempty"`
}

func (req *ListExecutionReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if err := validateExecutionStates(req.State); err != nil {
		return err
	}
	if err := validateExecutionStatus(req.Status); err != nil {
		return err
	}
	if err := validateExecutionLabels(req.Labels); err != nil {
		return err
	}
	return nil
}

// ManageExecutionLabelReq represents the request to manage labels for an execution.
type ManageExecutionLabelReq struct {
	AddLabels    []dto.LabelItem `json:"add_labels"`
	RemoveLabels []string        `json:"remove_labels"`
}

func (req *ManageExecutionLabelReq) Validate() error {
	if len(req.AddLabels) == 0 && len(req.RemoveLabels) == 0 {
		return fmt.Errorf("at least one of add_labels or remove_labels must be provided")
	}

	for i, label := range req.AddLabels {
		if strings.TrimSpace(label.Key) == "" {
			return fmt.Errorf("empty label key at index %d", i)
		}
		if strings.TrimSpace(label.Value) == "" {
			return fmt.Errorf("empty label value at index %d", i)
		}
	}

	for i, key := range req.RemoveLabels {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("empty label key at index %d in remove_labels", i)
		}
	}

	return nil
}

// ExecutionSpec represents a single execution request item.
type ExecutionSpec struct {
	Algorithm dto.ContainerSpec `json:"algorithm" binding:"required"`
	Datapack  *string           `json:"datapack" binding:"omitempty"`
	Dataset   *dto.DatasetRef   `json:"dataset" binding:"omitempty"`
}

func (spec *ExecutionSpec) Validate() error {
	hasDatapack := spec.Datapack != nil
	hasDataset := spec.Dataset != nil

	if !hasDatapack && !hasDataset {
		return fmt.Errorf("either datapack or dataset must be specified")
	}
	if hasDatapack && hasDataset {
		return fmt.Errorf("cannot specify both datapack and dataset")
	}
	if hasDatapack && *spec.Datapack == "" {
		return fmt.Errorf("datapack name cannot be empty")
	}

	if hasDataset {
		if err := spec.Dataset.Validate(); err != nil {
			return fmt.Errorf("invalid dataset: %w", err)
		}
	}

	if err := spec.Algorithm.Validate(); err != nil {
		return fmt.Errorf("invalid algorithm: %w", err)
	}
	if spec.Algorithm.Name == config.GetDetectorName() {
		return fmt.Errorf("detector algorithm cannot be used for execution")
	}

	return nil
}

// SubmitExecutionReq represents the request to submit execution tasks.
type SubmitExecutionReq struct {
	ProjectName string          `json:"project_name" binding:"required"`
	Specs       []ExecutionSpec `json:"specs" binding:"required"`
	Labels      []dto.LabelItem `json:"labels" binding:"omitempty"`
}

func (req *SubmitExecutionReq) Validate() error {
	if req.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}
	if len(req.Specs) == 0 {
		return fmt.Errorf("at least one execution spec is required")
	}
	for i, spec := range req.Specs {
		if err := spec.Validate(); err != nil {
			return fmt.Errorf("invalid execution spec at index %d: %w", i, err)
		}
	}
	return validateExecutionLabelItems(req.Labels)
}

// ExecutionResp represents execution summary information.
type ExecutionResp struct {
	ID                 int                   `json:"id"`
	Duration           float64               `json:"duration"`
	State              consts.ExecutionState `json:"state" swaggertype:"string" enums:"Initial,Failed,Success"`
	Status             string                `json:"status"`
	TaskID             string                `json:"task_id"`
	AlgorithmID        int                   `json:"algorithm_id"`
	AlgorithmName      string                `json:"algorithm_name"`
	AlgorithmVersionID int                   `json:"algorithm_version_id"`
	AlgorithmVersion   string                `json:"algorithm_version"`
	DatapackID         int                   `json:"datapack_id,omitempty"`
	DatapackName       string                `json:"datapack_name,omitempty"`
	CreatedAt          time.Time             `json:"created_at"`
	UpdatedAt          time.Time             `json:"updated_at"`
	Labels             []dto.LabelItem       `json:"labels,omitempty"`
}

func NewExecutionResp(execution *model.Execution, labels []model.Label) *ExecutionResp {
	resp := &ExecutionResp{
		ID:                 execution.ID,
		Duration:           execution.Duration,
		State:              execution.State,
		Status:             consts.GetStatusTypeName(execution.Status),
		AlgorithmID:        execution.AlgorithmVersion.ContainerID,
		AlgorithmName:      execution.AlgorithmVersion.Container.Name,
		AlgorithmVersionID: execution.AlgorithmVersionID,
		AlgorithmVersion:   execution.AlgorithmVersion.Name,
		DatapackID:         execution.DatapackID,
		DatapackName:       execution.Datapack.Name,
		CreatedAt:          execution.CreatedAt,
		UpdatedAt:          execution.UpdatedAt,
	}

	if execution.TaskID != nil {
		resp.TaskID = *execution.TaskID
	}

	if len(labels) > 0 {
		resp.Labels = make([]dto.LabelItem, 0, len(labels))
		for _, label := range labels {
			resp.Labels = append(resp.Labels, dto.LabelItem{Key: label.Key, Value: label.Value})
		}
	}
	return resp
}

// ExecutionDetailResp represents execution detail information.
type ExecutionDetailResp struct {
	ExecutionResp

	DetectorResults    []DetectorResultItem    `json:"detector_results,omitempty"`
	GranularityResults []GranularityResultItem `json:"granularity_results,omitempty"`
}

func NewExecutionDetailResp(execution *model.Execution, labels []model.Label) *ExecutionDetailResp {
	return &ExecutionDetailResp{
		ExecutionResp: *NewExecutionResp(execution, labels),
	}
}

// SubmitExecutionItem describes a single submitted execution task.
type SubmitExecutionItem struct {
	Index              int    `json:"index"`
	TraceID            string `json:"trace_id"`
	TaskID             string `json:"task_id"`
	AlgorithmID        int    `json:"algorithm_id"`
	AlgorithmVersionID int    `json:"algorithm_version_id"`
	DatapackID         *int   `json:"datapack_id,omitempty"`
	DatasetID          *int   `json:"dataset_id,omitempty"`
}

// SubmitExecutionResp represents the response for submitting execution tasks.
type SubmitExecutionResp struct {
	GroupID string                `json:"group_id"`
	Items   []SubmitExecutionItem `json:"items"`
}

func validateExecutionStates(state *consts.ExecutionState) error {
	if state == nil {
		return nil
	}
	if *state < 0 {
		return fmt.Errorf("state must be a non-negative integer")
	}
	if _, exists := consts.ValidExecutionStates[*state]; !exists {
		return fmt.Errorf("invalid state: %d", *state)
	}
	return nil
}

func validateExecutionStatus(statusPtr *consts.StatusType) error {
	if statusPtr == nil {
		return nil
	}
	status := *statusPtr
	if _, exists := consts.ValidStatuses[status]; !exists {
		return fmt.Errorf("invalid status value: %d", status)
	}
	return nil
}

func validateExecutionLabels(labels []string) error {
	for i, label := range labels {
		parts := strings.SplitN(label, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid label format at index %d: %q, expected key:value", i, label)
		}
		if strings.TrimSpace(parts[0]) == "" {
			return fmt.Errorf("empty label key at index %d", i)
		}
		if strings.TrimSpace(parts[1]) == "" {
			return fmt.Errorf("empty label value at index %d", i)
		}
	}
	return nil
}

func validateExecutionLabelItems(items []dto.LabelItem) error {
	for i, label := range items {
		if strings.TrimSpace(label.Key) == "" {
			return fmt.Errorf("empty label key at index %d", i)
		}
		if strings.TrimSpace(label.Value) == "" {
			return fmt.Errorf("empty label value at index %d", i)
		}
	}
	return nil
}
