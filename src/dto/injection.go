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

var validInjectionSortFields = map[string]struct{}{
	"id":         {},
	"name":       {},
	"start_time": {},
	"end_time":   {},
	"created_at": {},
	"updated_at": {},
}

type InjectionItem struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	PreDuration int       `json:"pre_duration"`
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
}

func NewInjectionItem(injection *database.FaultInjection) InjectionItem {
	item := InjectionItem{
		ID:          injection.ID,
		Name:        injection.Name,
		PreDuration: injection.PreDuration,
		StartTime:   *injection.StartTime,
		EndTime:     *injection.EndTime,
	}

	if injection.StartTime != nil {
		item.StartTime = *injection.StartTime
	}
	if injection.EndTime != nil {
		item.EndTime = *injection.EndTime
	}

	return item
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
	StatusMap        map[int]string                        `json:"status" swaggertype:"object"`
	FaultTypeMap     map[chaos.ChaosType]string            `json:"fault_type" swaggertype:"object"`
	FaultResourceMap map[string]chaos.ChaosResourceMapping `json:"fault_resource" swaggertype:"object"`
}

type ListInjectionFilters struct {
	FaultType       *chaos.ChaosType
	Benchmark       string
	State           *consts.DatapackState
	Status          *consts.StatusType
	LabelConditions []map[string]string
}

// ListInjectionReq represents the request to list injections with various filters
type ListInjectionReq struct {
	PaginationReq
	Type      *chaos.ChaosType      `form:"fault_type" binding:"omitempty"`
	Benchmark string                `form:"benchmark" binding:"omitempty"`
	State     *consts.DatapackState `form:"state" binding:"omitempty"`
	Status    *consts.StatusType    `form:"status" binding:"omitempty"`
	Labels    []string              `form:"labels" binding:"omitempty"`
}

func (req *ListInjectionReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if err := validateChaosType(req.Type); err != nil {
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
	labelConditions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		parts := strings.SplitN(item, ":", 2)
		labelConditions = append(labelConditions, map[string]string{
			"key":   parts[0],
			"value": parts[1],
		})
	}

	return &ListInjectionFilters{
		FaultType:       req.Type,
		Benchmark:       req.Benchmark,
		State:           req.State,
		Status:          req.Status,
		LabelConditions: labelConditions,
	}
}

type SearchInjectionFilters struct {
	TaskIDs         []string
	NamePattern     string
	FaultTypes      []chaos.ChaosType
	Benchmarks      []string
	States          []consts.DatapackState
	Statuses        []consts.StatusType
	LabelConditions []map[string]string
	StartTimeGte    *time.Time
	StartTimeLte    *time.Time
	EndTimeGte      *time.Time
	EndTimeLte      *time.Time
}

// InjectionV2SearchReq represents the request to search injections with various filters
type SearchInjectionReq struct {
	AdvancedSearchReq
	TaskIDs       []string               `json:"task_ids" binding:"omitempty"`
	Names         []string               `json:"names" binding:"omitempty"`
	NamePattern   string                 `json:"name_pattern" binding:"omitempty"`
	FaultTypes    []chaos.ChaosType      `json:"fault_types" binding:"omitempty"`
	States        []consts.DatapackState `json:"states" binding:"omitempty"`
	Benchmarks    []string               `json:"benchmarks" binding:"omitempty"`
	Labels        []LabelItem            `json:"labels" binding:"omitempty"` // Custom labels to filter by
	StartTime     *DateRange             `json:"start_time" binding:"omitempty"`
	EndTime       *DateRange             `json:"end_time" binding:"omitempty"`
	IncludeLabels bool                   `json:"include_labels" binding:"omitempty"` // Whether to include labels in the response
	IncludeTask   bool                   `json:"include_task" binding:"omitempty"`   // Whether to include task details in the response
}

func (req *SearchInjectionReq) Validate() error {
	if err := req.AdvancedSearchReq.Validate(); err != nil {
		return err
	}

	for i, id := range req.TaskIDs {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("empty task ID at index %d", i)
		}
		if !utils.IsValidUUID(id) {
			return fmt.Errorf("invalid task ID format at index %d: %s", i, id)
		}
	}

	if len(req.Names) > 0 && req.NamePattern != "" {
		return fmt.Errorf("can only specify one of names or name_pattern for filtering")
	}

	for i, name := range req.Names {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("empty injection name at index %d", i)
		}
	}

	if err := validateLabelItemsFiled(req.Labels); err != nil {
		return err
	}

	if req.StartTime != nil {
		if err := req.StartTime.Validate(); err != nil {
			return fmt.Errorf("invalid start_time: %w", err)
		}
	}
	if req.EndTime != nil {
		if err := req.EndTime.Validate(); err != nil {
			return fmt.Errorf("invalid end_time: %w", err)
		}
	}

	for i, sortField := range req.Sort {
		if _, valid := validInjectionSortFields[sortField.Field]; !valid {
			return fmt.Errorf("invalid sort_by field at index %d: %s", i, sortField.Field)
		}
	}

	return nil
}

func (req *SearchInjectionReq) ConvertToSearchReq() *SearchReq {
	sr := req.AdvancedSearchReq.ConvertAdvancedToSearch()

	if len(req.TaskIDs) > 0 {
		sr.AddFilter("task_id", OpIn, req.TaskIDs)
	}
	if len(req.Names) > 0 {
		sr.AddFilter("name", OpIn, req.Names)
	}
	if req.NamePattern != "" {
		sr.AddFilter("name", OpLike, req.NamePattern)
	}
	if len(req.Benchmarks) > 0 {
		sr.AddFilter("benchmark", OpIn, req.Benchmarks)
	}

	if len(req.FaultTypes) > 0 {
		faultTypeValues := make([]string, len(req.FaultTypes))
		for i, ft := range req.FaultTypes {
			faultTypeValues[i] = fmt.Sprintf("%d", ft)
		}
		sr.AddFilter("fault_type", OpIn, faultTypeValues)
	}
	if len(req.States) > 0 {
		stateValues := make([]string, len(req.States))
		for i, st := range req.States {
			stateValues[i] = fmt.Sprintf("%d", st)
		}
		sr.AddFilter("state", OpIn, stateValues)
	}
	if len(req.Statuses) > 0 {
		statusValues := make([]string, len(req.Statuses))
		for i, st := range req.Statuses {
			statusValues[i] = fmt.Sprintf("%d", st)
		}
		sr.AddFilter("status", OpIn, statusValues)
	}

	if req.StartTime != nil {
		if req.StartTime.From != nil && req.StartTime.To != nil {
			sr.AddFilter("created_at", OpDateBetween, []any{req.StartTime.From, req.StartTime.To})
		} else if req.StartTime.From != nil {
			sr.AddFilter("created_at", OpDateAfter, req.StartTime.From)
		} else if req.StartTime.To != nil {
			sr.AddFilter("created_at", OpDateBefore, req.StartTime.To)
		}
	}
	if req.EndTime != nil {
		if req.EndTime.From != nil && req.EndTime.To != nil {
			sr.AddFilter("created_at", OpDateBetween, []any{req.EndTime.From, req.EndTime.To})
		} else if req.EndTime.From != nil {
			sr.AddFilter("created_at", OpDateAfter, req.EndTime.From)
		} else if req.EndTime.To != nil {
			sr.AddFilter("created_at", OpDateBefore, req.EndTime.To)
		}
	}

	if req.IncludeLabels {
		sr.AddInclude("Labels")
	}
	if req.IncludeTask {
		sr.AddInclude("Task")
	}

	return sr
}

// SubmitInjectionReq represents a request to submit fault injection tasks with parallel fault support
// Each element in Specs represents a batch of faults to be injected in parallel within a single experiment
type SubmitInjectionReq struct {
	ProjectName string          `json:"project_name" binding:"required"`       // Project name
	Pedestal    *ContainerSpec  `json:"pedestal" binding:"required"`           // Pedestal (workload) configuration
	Benchmark   *ContainerSpec  `json:"benchmark" binding:"required"`          // Benchmark (detector) configuration
	Interval    int             `json:"interval" binding:"required,min=1"`     // Total experiment interval in minutes
	PreDuration int             `json:"pre_duration" binding:"required,min=1"` // Normal data collection duration before fault injection
	Specs       [][]chaos.Node  `json:"specs" binding:"required"`              // Fault injection specs - 2D array where each sub-array is a batch of parallel faults
	Algorithms  []ContainerSpec `json:"algorithms" binding:"omitempty"`        // RCA algorithms to execute (optional)
	Labels      []LabelItem     `json:"labels" binding:"omitempty"`            // Labels to attach to the injection
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
			if algorithm.Name == config.GetString(consts.DetectorKey) {
				return fmt.Errorf("algorithm name %s is reserved and cannot be used", config.GetString(consts.DetectorKey))
			}
		}
	}

	if req.Labels == nil {
		req.Labels = make([]LabelItem, 0)
	}

	return nil
}

type InjectionResp struct {
	ID            int            `json:"id"`
	Name          string         `json:"name"`
	FaultType     string         `json:"fault_type"`
	DisplayConfig map[string]any `json:"display_config,omitempty" swaggertype:"object"`
	PreDuration   int            `json:"pre_duration"`
	StartTime     *time.Time     `json:"start_time,omitempty"`
	EndTime       *time.Time     `json:"end_time,omitempty"`
	State         string         `json:"state"`
	Status        string         `json:"status"`
	TaskID        string         `json:"task_id"`
	BenchmarkID   int            `json:"benchmark_id"`
	BenchmarkName string         `json:"benchmark_name"`
	PedestalID    int            `json:"pedestal_id"`
	PedestalName  string         `json:"pedestal_name"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`

	Labels []LabelItem `json:"labels,omitempty"`
}

func NewInjectionResp(injection *database.FaultInjection) *InjectionResp {
	var faultTypeName string
	if injection.FaultType == consts.Hybrid {
		faultTypeName = "hybrid"
	} else {
		faultTypeName = chaos.ChaosTypeMap[injection.FaultType]
	}

	resp := &InjectionResp{
		ID:          injection.ID,
		Name:        injection.Name,
		FaultType:   faultTypeName,
		PreDuration: injection.PreDuration,
		StartTime:   injection.StartTime,
		EndTime:     injection.EndTime,
		State:       consts.GetDatapackStateName(injection.State),
		Status:      consts.GetStatusTypeName(injection.Status),
		BenchmarkID: injection.BenchmarkID,
		PedestalID:  injection.PedestalID,
		CreatedAt:   injection.CreatedAt,
		UpdatedAt:   injection.UpdatedAt,
	}

	if injection.DisplayConfig != nil {
		var displayConfigData map[string]any
		_ = json.Unmarshal([]byte(*injection.DisplayConfig), &displayConfigData)
		resp.DisplayConfig = displayConfigData
	}

	if injection.Benchmark != nil {
		if injection.Benchmark.Container != nil {
			resp.BenchmarkName = injection.Benchmark.Container.Name
		}
	}
	if injection.Pedestal != nil {
		if injection.Pedestal.Container != nil {
			resp.PedestalName = injection.Pedestal.Container.Name
		}
	}

	if injection.TaskID != nil {
		resp.TaskID = *injection.TaskID
	}

	// Get labels from associated Task instead of directly from injection
	if len(injection.Labels) > 0 {
		resp.Labels = make([]LabelItem, 0, len(injection.Labels))
		for _, l := range injection.Labels {
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

	Description  string              `json:"description,omitempty"`
	EngineConfig []map[string]any    `json:"engine_config" swaggertype:"array,object"`
	Groundtruths []chaos.Groundtruth `json:"ground_truth,omitempty"`
}

func NewInjectionDetailResp(entity *database.FaultInjection) *InjectionDetailResp {
	injectionResp := NewInjectionResp(entity)
	resp := &InjectionDetailResp{
		InjectionResp: *injectionResp,
		Description:   entity.Description,
	}

	if entity.EngineConfig != "" {
		var engineConfigData []map[string]any
		_ = json.Unmarshal([]byte(entity.EngineConfig), &engineConfigData)
		resp.EngineConfig = engineConfigData
	}

	resp.Groundtruths = make([]chaos.Groundtruth, 0, len(entity.Groundtruths))
	if len(entity.Groundtruths) > 0 {
		for _, gt := range entity.Groundtruths {
			resp.Groundtruths = append(resp.Groundtruths, *gt.ConvertToChaosGroundtruth())
		}
	}

	return resp
}

// InjectionMetadataResp represents the metadata response for injections
type InjectionMetadataResp struct {
	Config           *chaos.Node                           `json:"config"`
	FaultTypeMap     map[chaos.ChaosType]string            `json:"fault_type_map"`
	FaultResourceMap map[string]chaos.ChaosResourceMapping `json:"fault_resource_map"`
	SystemResource   chaos.SystemResource                  `json:"ns_resources"`
}

type SubmitInjectionItem struct {
	Index   int    `json:"index"` // Index of the batch this injection belongs to
	TraceID string `json:"trace_id"`
	TaskID  string `json:"task_id"`
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

// InjectionLabelOperation represents label operations for a single injection
type InjectionLabelOperation struct {
	InjectionID  int         `json:"injection_id" binding:"required"` // Injection ID to manage
	AddLabels    []LabelItem `json:"add_labels,omitempty"`            // Labels to add to this injection
	RemoveLabels []LabelItem `json:"remove_labels,omitempty"`         // Labels to remove from this injection
}

// BatchManageInjectionLabelReq represents the request to batch manage injection labels
// Each injection can have its own set of label operations
type BatchManageInjectionLabelReq struct {
	Items []InjectionLabelOperation `json:"items" binding:"required,min=1,dive"` // List of label operations per injection
}

func (req *BatchManageInjectionLabelReq) Validate() error {
	if len(req.Items) == 0 {
		return fmt.Errorf("items list cannot be empty")
	}

	seenIDs := make(map[int]struct{}, len(req.Items))
	for i, item := range req.Items {
		if _, exists := seenIDs[item.InjectionID]; exists {
			return fmt.Errorf("duplicate injection_id at index %d: %d", i, item.InjectionID)
		}
		seenIDs[item.InjectionID] = struct{}{}

		if item.InjectionID <= 0 {
			return fmt.Errorf("invalid injection_id at index %d: %d", i, item.InjectionID)
		}

		if len(item.AddLabels) == 0 && len(item.RemoveLabels) == 0 {
			return fmt.Errorf("at least one of add_labels or remove_labels must be provided for injection_id %d at index %d", item.InjectionID, i)
		}

		if err := validateLabelItemsFiled(item.AddLabels); err != nil {
			return fmt.Errorf("invalid add_labels for injection_id %d at index %d: %w", item.InjectionID, i, err)
		}
		if err := validateLabelItemsFiled(item.RemoveLabels); err != nil {
			return fmt.Errorf("invalid remove_labels for injection_id %d at index %d: %w", item.InjectionID, i, err)
		}
	}

	return nil
}

// BatchManageInjectionLabelResp represents the response for batch injection label management
type BatchManageInjectionLabelResp struct {
	FailedCount  int             `json:"failed_count"`
	FailedItems  []string        `json:"failed_items"`
	SuccessCount int             `json:"success_count"`
	SuccessItems []InjectionResp `json:"success_items"`
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
	FaultType    string      `json:"fault_type"`
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
		FaultType:    chaos.ChaosTypeMap[entity.FaultType],
		EngineConfig: engineConfig,
	}, nil
}

// InjectionWithIssuesResp represents the response for fault injections with issues
type InjectionWithIssuesResp struct {
	ID                  int        `json:"datapack_id"`
	Name                string     `json:"datapack_name"`
	FaultType           string     `json:"fault_type"`
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
		FaultType:           chaos.ChaosTypeMap[entity.FaultType],
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
	Benchmark   ContainerSpec `json:"benchmark" binding:"required"`
	Datapack    *string       `json:"datapack" binding:"omitempty"`
	Dataset     *DatasetRef   `json:"dataset" binding:"omitempty"`
	PreDuration *int          `json:"pre_duration" binding:"omitempty"`
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
