package dto

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/utils"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
)

type InjectionItem struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	PreDuration int       `json:"pre_duration"`
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
}

func NewInjectionItem(injection *database.FaultInjection) InjectionItem {
	return InjectionItem{
		ID:          injection.ID,
		Name:        injection.Name,
		PreDuration: injection.PreDuration,
		StartTime:   *injection.StartTime,
		EndTime:     *injection.EndTime,
	}
}

// BatchDeleteInjectionReq represents the request to batch delete injections
type BatchDeleteInjectionReq struct {
	IDs    []int       `json:"ids,omitempty"`    // List of injection IDs for deletion
	Labels []LabelItem `json:"labels,omitempty"` // List of label keys to match for deletion
}

func (req *BatchDeleteInjectionReq) Validate() error {
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

// TriggerDatasetBuildItemResponse represents the response for a single injection in batch trigger
type TriggerDatasetBuildItemResponse struct {
	TaskID        string `json:"task_id"`
	TraceID       string `json:"trace_id"`
	InjectionName string `json:"injection_name"`
	Benchmark     string `json:"benchmark"`
	Namespace     string `json:"namespace"`
	Message       string `json:"message"`
}

// TriggerDatasetBuildError represents an error during dataset build trigger
type TriggerDatasetBuildError struct {
	InjectionName string `json:"injection_name"`
	Error         string `json:"error"`
}

// TriggerFailedDatapackRebuildRequest represents the request for triggering rebuild of failed datapacks
type TriggerFailedDatapackRebuildRequest struct {
	Namespace string `json:"namespace,omitempty"` // Optional namespace, defaults to "ts"
	Days      *int   `json:"days,omitempty"`      // Number of days to look back, defaults to 3
}

// TriggerFailedDatapackRebuildResponse represents the response for triggering rebuild of failed datapacks
type TriggerFailedDatapackRebuildResponse struct {
	SuccessCount int                               `json:"success_count"`
	SuccessItems []TriggerDatasetBuildItemResponse `json:"success_items"`
	FailedCount  int                               `json:"failed_count"`
	FailedItems  []TriggerDatasetBuildError        `json:"failed_items,omitempty"`
	TotalFound   int                               `json:"total_found"`   // Total number of failed datapacks found
	DaysSearched int                               `json:"days_searched"` // Number of days searched
	SearchCutoff string                            `json:"search_cutoff"` // ISO timestamp of search cutoff
	Message      string                            `json:"message"`
}

// TriggerFailedDatapackRebuildProgressEvent represents a single progress event for SSE
type TriggerFailedDatapackRebuildProgressEvent struct {
	Type          string                                `json:"type"`                     // "start", "progress", "item_success", "item_error", "complete", "error"
	Message       string                                `json:"message"`                  // Human readable message
	TotalFound    int                                   `json:"total_found"`              // Total number of failed datapacks found
	CurrentIndex  int                                   `json:"current_index"`            // Current processing index (0-based)
	Progress      float64                               `json:"progress"`                 // Progress percentage (0-100)
	SuccessCount  int                                   `json:"success_count"`            // Number of successful triggers so far
	FailedCount   int                                   `json:"failed_count"`             // Number of failed triggers so far
	CurrentItem   *TriggerDatasetBuildItemResponse      `json:"current_item,omitempty"`   // Current successful item
	CurrentError  *TriggerDatasetBuildError             `json:"current_error,omitempty"`  // Current error item
	EstimatedTime *time.Duration                        `json:"estimated_time,omitempty"` // Estimated remaining time
	FinalResponse *TriggerFailedDatapackRebuildResponse `json:"final_response,omitempty"` // Final response (only for "complete" type)
}

type InjectionFieldMappingResp struct {
	StatusMap        map[int]string                 `json:"status" swaggertype:"object"`
	FaultTypeMap     map[chaos.ChaosType]string     `json:"fault_type" swaggertype:"object"`
	FaultResourceMap map[string]chaos.ResourceField `json:"fault_resource" swaggertype:"object"`
}

type ListInjectionFilters struct {
	FaultType      *chaos.ChaosType
	Benchmark      string
	State          *consts.DatapackState
	Status         *consts.StatusType
	LabelConditons []map[string]string
}

// ListInjectionReq represents the request to list injections with various filters
type ListInjectionReq struct {
	PaginationReq
	FaultType *chaos.ChaosType      `form:"fault_type" binding:"omitempty"`
	Benchmark string                `form:"benchmark" binding:"omitempty"`
	State     *consts.DatapackState `form:"state" binding:"omitempty"`
	Status    *consts.StatusType    `form:"status" binding:"omitempty"`
	Labels    []string              `form:"labels" binding:"omitempty"`
}

func (req *ListInjectionReq) Validate() error {
	if err := validateBenchmarkName(req.Benchmark); err != nil {
		return err
	}
	if err := validateChaosType(req.FaultType); err != nil {
		return err
	}
	if err := validateDatapackState(req.State); err != nil {
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

func (req *ListInjectionReq) ToFilterOptions() *ListInjectionFilters {
	labelCondtions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		parts := strings.SplitN(item, ":", 2)
		labelCondtions = append(labelCondtions, map[string]string{
			"key":   parts[0],
			"value": parts[1],
		})
	}

	return &ListInjectionFilters{
		FaultType:      req.FaultType,
		Benchmark:      req.Benchmark,
		State:          req.State,
		Status:         req.Status,
		LabelConditons: labelCondtions,
	}
}

// InjectionV2SearchReq represents the request to search injections with various filters
type SearchInjectionReq struct {
	Page          *int        `json:"page" binding:"omitempty"`
	Size          *int        `json:"size" binding:"omitempty"`
	TaskIDs       []string    `json:"task_ids" binding:"omitempty"`
	FaultTypes    []int       `json:"fault_types" binding:"omitempty"`
	Statuses      []int       `json:"statuses" binding:"omitempty"`
	Benchmarks    []string    `json:"benchmarks" binding:"omitempty"`
	Search        string      `json:"search" binding:"omitempty"`
	Tags          []string    `json:"tags" binding:"omitempty"`   // Tag values to filter by
	Labels        []LabelItem `json:"labels" binding:"omitempty"` // Custom labels to filter by
	StartTimeGte  *time.Time  `json:"start_time_gte" binding:"omitempty"`
	StartTimeLte  *time.Time  `json:"start_time_lte" binding:"omitempty"`
	EndTimeGte    *time.Time  `json:"end_time_gte" binding:"omitempty"`
	EndTimeLte    *time.Time  `json:"end_time_lte" binding:"omitempty"`
	CreatedAtGte  *time.Time  `json:"created_at_gte" binding:"omitempty"`
	CreatedAtLte  *time.Time  `json:"created_at_lte" binding:"omitempty"`
	SortBy        string      `json:"sort_by" binding:"omitempty,oneof=id task_id fault_type status benchmark injection_name created_at updated_at"`
	SortOrder     string      `json:"sort_order" binding:"omitempty,oneof=asc desc"`
	IncludeLabels bool        `json:"include_labels" binding:"omitempty"` // Whether to include labels in the response
	IncludeTask   bool        `json:"include_task" binding:"omitempty"`   // Whether to include task details in the response
}

func (req *SearchInjectionReq) Validate() error {
	if req.Page != nil && *req.Page < 1 {
		return fmt.Errorf("page must be greater than 0")
	}
	if req.Size != nil && *req.Size < 1 {
		return fmt.Errorf("size must be greater than 0")
	}

	if req.StartTimeGte != nil && req.StartTimeLte != nil && req.StartTimeGte.After(*req.StartTimeLte) {
		return fmt.Errorf("start_time_gte must be before start_time_lte")
	}
	if req.EndTimeGte != nil && req.EndTimeLte != nil && req.EndTimeGte.After(*req.EndTimeLte) {
		return fmt.Errorf("end_time_gte must be before end_time_lte")
	}
	if req.CreatedAtGte != nil && req.CreatedAtLte != nil && req.CreatedAtGte.After(*req.CreatedAtLte) {
		return fmt.Errorf("created_at_gte must be before created_at_lte")
	}

	return nil
}

type SubmitInjectionReq struct {
	ProjectName string          `json:"project_name" binding:"required"`
	Pedestal    *ContainerSpec  `json:"pedestal" binding:"required"`
	Benchmark   *ContainerSpec  `json:"benchmark" binding:"required"`
	Interval    int             `json:"interval" binding:"required,min=1"`
	PreDuration int             `json:"pre_duration" binding:"required,min=1"`
	Specs       []chaos.Node    `json:"specs" binding:"required"`
	Algorithms  []ContainerSpec `json:"algorithms" binding:"omitempty"`
	Labels      []LabelItem     `json:"labels" binding:"omitempty"`
}

func (req *SubmitInjectionReq) Validate() error {
	if req.Pedestal == nil {
		return fmt.Errorf("pedestal must not be nil")
	} else {
		if err := req.Pedestal.Validate(); err != nil {
			return fmt.Errorf("invalid pedestal: %w", err)
		}
	}

	if req.Benchmark == nil {
		return fmt.Errorf("benchmark must not be nil")
	} else {
		if err := req.Benchmark.Validate(); err != nil {
			return fmt.Errorf("invalid benchmark: %w", err)
		}
		if err := validateBenchmarkName(req.Benchmark.Name); err != nil {
			return err
		}
	}

	if req.ProjectName == "" {
		return fmt.Errorf("project name must not be blank")
	}
	if req.Interval <= req.PreDuration {
		return fmt.Errorf("interval must be greater than pre_duration")
	}
	if len(req.Specs) == 0 {
		return fmt.Errorf("specs must not be empty")
	}

	if req.Algorithms != nil {
		for idx, algorithm := range req.Algorithms {
			if err := algorithm.Validate(); err != nil {
				return fmt.Errorf("invalid algorithm at index %d: %w", idx, err)
			}
			if algorithm.Name == config.GetString("algo.detector") {
				return fmt.Errorf("algorithm name %s is reserved and cannot be used", config.GetString("algo.detector"))
			}
		}
	}

	if req.Labels == nil {
		req.Labels = make([]LabelItem, 0)
	}

	return nil
}

type InjectionResp struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	FaultType     string     `json:"fault_type"`
	DisplayConfig *string    `json:"display_config,omitempty"`
	PreDuration   int        `json:"pre_duration"`
	StartTime     *time.Time `json:"start_time,omitempty"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	State         string     `json:"state"`
	Status        string     `json:"status"`
	TaskID        string     `json:"task_id"`
	BenchmarkID   int        `json:"benchmark_id"`
	BenchmarkName string     `json:"benchmark_name"`
	PedestalID    int        `json:"pedestal_id"`
	PedestalName  string     `json:"pedestal_name"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	Labels []LabelItem `json:"labels,omitempty"`
}

func NewInjectionResp(injection *database.FaultInjection) *InjectionResp {
	resp := &InjectionResp{
		ID:            injection.ID,
		Name:          injection.Name,
		FaultType:     chaos.ChaosTypeMap[injection.FaultType],
		DisplayConfig: injection.DisplayConfig,
		PreDuration:   injection.PreDuration,
		StartTime:     injection.StartTime,
		EndTime:       injection.EndTime,
		State:         consts.GetDatapackStateName(injection.State),
		Status:        consts.GetStatusTypeName(injection.Status),
		TaskID:        injection.TaskID,
		BenchmarkID:   injection.BenchmarkID,
		BenchmarkName: injection.Benchmark.Container.Name,
		PedestalID:    injection.PedestalID,
		PedestalName:  injection.Pedestal.Container.Name,
		CreatedAt:     injection.CreatedAt,
		UpdatedAt:     injection.UpdatedAt,
	}

	// Get labels from associated Task instead of directly from injection
	if injection.Task != nil && len(injection.Task.Labels) > 0 {
		resp.Labels = make([]LabelItem, 0, len(injection.Task.Labels))
		for _, l := range injection.Task.Labels {
			resp.Labels = append(resp.Labels, LabelItem{
				Key:   l.Key,
				Value: l.Value,
			})
		}
	}
	return resp
}

type InjectionDetailResp struct {
	InjectionResp

	Description  string             `json:"description,omitempty"`
	EngineConfig string             `json:"engine_config"`
	GroundTruth  *chaos.Groundtruth `json:"ground_truth,omitempty"`
}

func NewInjectionDetailResp(entity *database.FaultInjection) *InjectionDetailResp {
	injectionResp := NewInjectionResp(entity)
	resp := &InjectionDetailResp{
		InjectionResp: *injectionResp,
		Description:   entity.Description,
		EngineConfig:  entity.EngineConfig,
	}
	return resp
}

// InjectionMetadataResp represents the metadata response for injections
type InjectionMetadataResp struct {
	Config           *chaos.Node                    `json:"config"`
	FaultTypeMap     map[chaos.ChaosType]string     `json:"fault_type_map"`
	FaultResourceMap map[string]chaos.ResourceField `json:"fault_resource_map"`
	NsResources      chaos.Resources                `json:"ns_resources"`
}

type SubmitInjectionItem struct {
	TraceID string `json:"trace_id"`
	TaskID  string `json:"task_id"`
	Index   int    `json:"index"`
}

type SubmitInjectionResp struct {
	GroupID         string                `json:"group_id"`
	Items           []SubmitInjectionItem `json:"items"`
	DuplicatedCount int                   `json:"duplicated_count"`
	OriginalCount   int                   `json:"original_count"`
}

type SubmitDatapackBuildingReq struct {
	ProjectName string         `json:"project_name" binding:"required"`
	Specs       []BuildingSpec `json:"specs" binding:"required"`
	Labels      []LabelItem    `json:"labels" binding:"omitempty"`
}

func (req *SubmitDatapackBuildingReq) Validate() error {
	if req.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}

	if len(req.Specs) == 0 {
		return fmt.Errorf("at least one datapack spec is required")
	}

	for _, spec := range req.Specs {
		if err := spec.Validate(); err != nil {
			return fmt.Errorf("invalid datapack spec: %w", err)
		}
	}

	return validateLabelItemsFiled(req.Labels)
}

// ManageInjectionLabelReq Represents the request to manage labels for an injection
type ManageInjectionLabelReq struct {
	AddLabels    []LabelItem `json:"add_labels"`    // List of labels to add
	RemoveLabels []string    `json:"remove_labels"` // List of label keys to remove
}

func (req *ManageInjectionLabelReq) Validate() error {
	if len(req.AddLabels) == 0 && len(req.RemoveLabels) == 0 {
		return fmt.Errorf("at least one of add_labels or remove_labels must be provided")
	}

	if err := validateLabelItemsFiled(req.AddLabels); err != nil {
		return err
	}

	for i, key := range req.RemoveLabels {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("empty label key at index %d in remove_labels", i)
		}
	}

	return nil
}

// analysis
type ListInjectionNoIssuesReq struct {
	Labels []string `form:"labels" binding:"omitempty"`
	TimeRangeQuery
}

func (req *ListInjectionNoIssuesReq) Validate() error {
	if err := validateLabelsField(req.Labels); err != nil {
		return err
	}
	return req.TimeRangeQuery.Validate()
}

type ListInjectionWithIssuesReq struct {
	Labels []string `form:"labels" binding:"omitempty"`
	TimeRangeQuery
}

func (req *ListInjectionWithIssuesReq) Validate() error {
	if err := validateLabelsField(req.Labels); err != nil {
		return err
	}
	return req.TimeRangeQuery.Validate()
}

type InjectionNoIssuesResp struct {
	ID           int         `json:"datapack_id"`
	Name         string      `json:"datapack_name"`
	EngineConfig *chaos.Node `json:"engine_config"`
}

func NewInjectionNoIssuesResp(entity database.FaultInjectionNoIssues) (*InjectionNoIssuesResp, error) {
	var engineConfig *chaos.Node
	err := json.Unmarshal([]byte(entity.EngineConfig), engineConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal engine config: %w", err)
	}

	return &InjectionNoIssuesResp{
		ID:           entity.ID,
		Name:         entity.Name,
		EngineConfig: engineConfig,
	}, nil
}

// InjectionWithIssuesResp represents the response for fault injections with issues
type InjectionWithIssuesResp struct {
	ID                  int        `json:"datapack_id"`
	Name                string     `json:"datapack_name"`
	EngineConfig        chaos.Node `json:"engine_config"`
	Issues              string     `json:"issues"`
	AbnormalAvgDuration float64    `json:"abnormal_avg_duration"`
	NormalAvgDuration   float64    `json:"normal_avg_duration"`
	AbnormalSuccRate    float64    `json:"abnormal_succ_rate"`
	NormalSuccRate      float64    `json:"normal_succ_rate"`
	AbnormalP99         float64    `json:"abnormal_p99"`
	NormalP99           float64    `json:"normal_p99"`
}

func NewInjectionWithIssuesResp(entity database.FaultInjectionWithIssues) (*InjectionWithIssuesResp, error) {
	var engineConfig chaos.Node
	err := json.Unmarshal([]byte(entity.EngineConfig), &engineConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal engine config: %w", err)
	}
	return &InjectionWithIssuesResp{
		ID:                  entity.ID,
		Name:                entity.Name,
		EngineConfig:        engineConfig,
		Issues:              entity.Issues,
		AbnormalAvgDuration: entity.AbnormalAvgDuration,
		NormalAvgDuration:   entity.NormalAvgDuration,
		AbnormalSuccRate:    entity.AbnormalSuccRate,
		NormalSuccRate:      entity.NormalSuccRate,
		AbnormalP99:         entity.AbnormalP99,
		NormalP99:           entity.NormalP99,
	}, nil
}

// datapack
type BuildingSpec struct {
	Benchmark   *ContainerSpec `json:"benchmark" binding:"required"`
	Datapack    *string        `json:"datapack" binding:"omitempty"`
	Dataset     *DatasetRef    `json:"dataset" binding:"omitempty"`
	PreDuration *int           `json:"pre_duration" binding:"omitempty"`
}

func (spec *BuildingSpec) Validate() error {
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

	if spec.Benchmark != nil {
		return fmt.Errorf("benchmark must not be nil")
	} else {
		if err := spec.Benchmark.Validate(); err != nil {
			return fmt.Errorf("invalid benchmark: %w", err)
		}
		if err := validateBenchmarkName(spec.Benchmark.Name); err != nil {
			return err
		}
	}

	if spec.PreDuration != nil && *spec.PreDuration <= 0 {
		return fmt.Errorf("pre_duration must be greater than 0")
	}

	return nil
}

type SubmitBuildingItem struct {
	Index   int    `json:"index"`
	TraceID string `json:"trace_id"`
	TaskID  string `json:"task_id"`
}

// SubmitDatapackResp represents the response for submitting datapack building tasks
type SubmitDatapackBuildingResp struct {
	GroupID string               `json:"group_id"`
	Items   []SubmitBuildingItem `json:"items"`
}

// validateBenchmark checks if the benchmark name is valid
func validateBenchmarkName(benchmark string) error {
	if benchmark == "" {
		return fmt.Errorf("benchmark must not be blank")
	} else {
		if _, exists := utils.GetValidBenchmarkMap()[benchmark]; !exists {
			return fmt.Errorf("invalid benchmark: %s", benchmark)
		}
	}

	return nil
}

// validateChaosType checks if the provided chaos type is valid
func validateChaosType(faultType *chaos.ChaosType) error {
	if faultType != nil {
		if _, exists := chaos.ChaosTypeMap[*faultType]; !exists {
			return fmt.Errorf("invalid fault type: %d", faultType)
		}
	}
	return nil
}

// validateDatapackState checks if the provided datapack state is valid
func validateDatapackState(state *consts.DatapackState) error {
	if state != nil {
		if *state < 0 {
			return fmt.Errorf("state must be a non-negative integer")
		}
		if _, exists := consts.ValidDatapackStates[consts.DatapackState(*state)]; !exists {
			return fmt.Errorf("invalid state: %d", *state)
		}
	}
	return nil
}
