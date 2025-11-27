package dto

import (
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"fmt"
	"strings"
	"time"
)

// ExecutionRef represents execution granularity results for evaluation
type ExecutionRef struct {
	ExecutionID       int                     `json:"execution_id"`       // Execution ID
	ExecutionDuration float64                 `json:"execution_duration"` // Execution duration in seconds
	DetectorResults   []DetectorResultItem    `json:"detector_results"`   // Detector results
	Predictions       []GranularityResultItem `json:"predictions"`        // Algorithm predictions
	ExecutedAt        time.Time               `json:"executed_at"`        // Execution time
}

func NewExecutionGranularityRef(execution *database.Execution) ExecutionRef {
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

// BatchDeleteExecutionReq represents the request to batch delete executions
type BatchDeleteExecutionReq struct {
	IDs    []int       `json:"ids" binding:"omitempty"`    // List of injection IDs for deletion
	Labels []LabelItem `json:"labels" binding:"omitempty"` // List of label keys to match for deletion
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

type ListExecutionReq struct {
	PaginationReq
	State  *consts.ExecutionState `form:"state" binding:"omitempty"`
	Status *consts.StatusType     `form:"status" binding:"omitempty"`
	Labels []string               `form:"labels" binding:"omitempty"`
}

func (req *ListExecutionReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if err := validateExecutionStates(req.State); err != nil {
		return err
	}
	if err := validateStatusField(req.Status, false); err != nil {
		return err
	}
	if err := validateLabelsField(req.Labels); err != nil {
		return err
	}
	return nil
}

// ManageExecutionLabelReq Represents the request to manage labels for an execution
type ManageExecutionLabelReq struct {
	AddLabels    []LabelItem `json:"add_labels"`    // List of labels to add
	RemoveLabels []string    `json:"remove_labels"` // List of label keys to remove
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

type ExecutionSpec struct {
	Algorithm ContainerSpec `json:"algorithm" binding:"required"`
	Datapack  *string       `json:"datapack" binding:"omitempty"`
	Dataset   *DatasetRef   `json:"dataset" binding:"omitempty"`
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

	if hasDatapack {
		if *spec.Datapack == "" {
			return fmt.Errorf("datapack name cannot be empty")
		}
	}

	if hasDataset {
		if err := spec.Dataset.Validate(); err != nil {
			return fmt.Errorf("invalid dataset: %w", err)
		}
	}

	if err := spec.Algorithm.Validate(); err != nil {
		return fmt.Errorf("invalid algorithm: %w", err)
	}
	if spec.Algorithm.Name == config.GetString("algo.detector") {
		return fmt.Errorf("detector algorithm cannot be used for execution")
	}

	return nil
}

// SubmitExecutionReq represents the request to submit execution tasks
type SubmitExecutionReq struct {
	ProjectName string          `json:"project_name" binding:"required"`
	Specs       []ExecutionSpec `json:"specs" binding:"required"`
	Labels      []LabelItem     `json:"labels" binding:"omitempty"`
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

	return validateLabelItemsFiled(req.Labels)
}

type ExecutionResp struct {
	ID                 int       `json:"id"`
	Duration           float64   `json:"duration"`
	State              string    `json:"state"`
	Status             string    `json:"status"`
	TaskID             string    `json:"task_id"`
	AlgorithmID        int       `json:"algorithm_id"`
	AlgorithmName      string    `json:"algorithm_name"`
	AlgorithmVersionID int       `json:"algorithm_version_id"`
	AlgorithmVersion   string    `json:"algorithm_version"`
	DatapackID         int       `json:"datapack_id,omitempty"`
	DatapackName       string    `json:"datapack_name,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	Labels []LabelItem `json:"labels,omitempty"`
}

func NewExecutionResp(execution *database.Execution, labels []database.Label) *ExecutionResp {
	resp := &ExecutionResp{
		ID:                 execution.ID,
		Duration:           execution.Duration,
		State:              consts.GetExecuteStateName(execution.State),
		Status:             consts.GetStatusTypeName(execution.Status),
		TaskID:             execution.TaskID,
		AlgorithmID:        execution.AlgorithmVersion.ContainerID,
		AlgorithmName:      execution.AlgorithmVersion.Container.Name,
		AlgorithmVersionID: execution.AlgorithmVersionID,
		AlgorithmVersion:   execution.AlgorithmVersion.Name,
		DatapackID:         execution.DatapackID,
		DatapackName:       execution.Datapack.Name,
		CreatedAt:          execution.CreatedAt,
		UpdatedAt:          execution.UpdatedAt,
	}

	if len(labels) > 0 {
		resp.Labels = make([]LabelItem, 0, len(execution.Task.Labels))
		for _, l := range execution.Task.Labels {
			resp.Labels = append(resp.Labels, LabelItem{
				Key:   l.Key,
				Value: l.Value,
			})
		}
	}
	return resp
}

type ExecutionDetailResp struct {
	ExecutionResp

	DetectorResults    []DetectorResultItem    `json:"detector_results,omitempty"`
	GranularityResults []GranularityResultItem `json:"granularity_results,omitempty"`
}

func NewExecutionDetailResp(execution *database.Execution, labels []database.Label) *ExecutionDetailResp {
	return &ExecutionDetailResp{
		ExecutionResp: *NewExecutionResp(execution, labels),
	}
}

type SubmitExecutionItem struct {
	Index              int    `json:"index"`
	TraceID            string `json:"trace_id"`
	TaskID             string `json:"task_id"`
	AlgorithmID        int    `json:"algorithm_id"`
	AlgorithmVersionID int    `json:"algorithm_version_id"`
	DatapackID         *int   `json:"datapack_id,omitempty"`
	DatasetID          *int   `json:"dataset_id,omitempty"`
}

// SubmitExecutionResp represents the response for submitting execution tasks
type SubmitExecutionResp struct {
	GroupID string                `json:"group_id"`
	Items   []SubmitExecutionItem `json:"items"`
}

func validateExecutionStates(state *consts.ExecutionState) error {
	if state != nil {
		if *state < 0 {
			return fmt.Errorf("state must be a non-negative integer")
		}
		if _, exists := consts.ValidExecutionStates[*state]; !exists {
			return fmt.Errorf("invalid state: %d", *state)
		}
	}
	return nil
}
