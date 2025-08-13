package dto

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/utils"
)

type InjectionConfig struct {
	Index         int
	FaultType     int
	FaultDuration int
	DisplayData   string
	Conf          *chaos.InjectionConf
	Node          *chaos.Node
	ExecuteTime   time.Time
	Labels        []LabelItem
}

type InjectionItem struct {
	ID            int       `json:"id"`
	FaultType     int       `json:"fault_type"`
	DisplayConfig string    `json:"display_config"`
	EngineConfig  string    `json:"engine_config"`
	PreDuration   int       `json:"pre_duration"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        int       `json:"status"`
	Benchmark     string    `json:"benchmark"`
	Env           string    `json:"env"`
	Batch         string    `json:"batch"`
	Tag           string    `json:"tag"`
	InjectionName string    `json:"injection_name"`
	CreatedAt     time.Time `json:"created_at"`
}

type InjectCancelResp struct{}

type InjectionConfReq struct {
	Namespace string `form:"namespace" binding:"required"`
	Mode      string `form:"mode" binding:"omitempty,oneof=display engine"`
}

func (req *InjectionConfReq) setDefaults() {
	if req.Mode == "" {
		req.Mode = "engine"
	}
}

func (req *InjectionConfReq) Validate() error {
	req.setDefaults()
	return nil
}

type ListDisplayConfigsReq struct {
	TraceIDs []string `form:"trace_ids" binding:"omitempty"`
}

func (req *ListDisplayConfigsReq) Validate() error {
	req.TraceIDs = utils.FilterEmptyStrings(req.TraceIDs)
	for _, traceID := range req.TraceIDs {
		if !utils.IsValidUUID(traceID) {
			return fmt.Errorf("invalid trace_id format: %s", traceID)
		}
	}

	return nil
}

type InjectionFilterOptions struct {
	ProjectName string `form:"project_name" binding:"omitempty"`

	Env       string `form:"env" binding:"omitempty"`
	Batch     string `form:"batch" binding:"omitempty"`
	Tag       string `form:"tag" binding:"omitempty"`
	Benchmark string `form:"benchmark" binding:"omitempty"`
	Status    *int   `form:"status" binding:"omitempty"`
	FaultType *int   `form:"fault_type" binding:"omitempty"`
}

func (opts *InjectionFilterOptions) Validate() error {
	if opts.Benchmark != "" {
		if _, exists := config.GetValidBenchmarkMap()[opts.Benchmark]; !exists {
			return fmt.Errorf("invalid benchmark: %s", opts.Benchmark)
		}
	}

	if opts.Status != nil {
		status := *opts.Status
		if status < 0 {
			return fmt.Errorf("status must be a non-negative integer")
		}

		if _, exists := DatasetStatusMap[status]; !exists {
			return fmt.Errorf("invalid status: %d", opts.Status)
		}
	}

	if opts.FaultType != nil {
		if _, exists := chaos.ChaosTypeMap[chaos.ChaosType(*opts.FaultType)]; !exists {
			return fmt.Errorf("invalid fault type: %d", opts.FaultType)
		}
	}

	return nil
}

type ListInjectionsReq struct {
	InjectionFilterOptions
	ListOptionsQuery
	PaginationQuery
	TimeRangeQuery
}

func (req *ListInjectionsReq) Validate() error {
	if err := req.InjectionFilterOptions.Validate(); err != nil {
		return err
	}

	if err := req.ListOptionsQuery.Validate(); err != nil {
		return err
	}

	if err := req.PaginationQuery.Validate(); err != nil {
		return err
	}

	hasLimit := req.ListOptionsQuery.Limit > 0
	hasPagination := req.PaginationQuery.PageNum > 0 && req.PaginationQuery.PageSize > 0

	if hasLimit && hasPagination {
		return fmt.Errorf("cannot use both limit and pagination (page_num/page_size) at the same time")
	}

	if err := req.TimeRangeQuery.Validate(); err != nil {
		return err
	}

	return nil
}

type QueryInjectionReq struct {
	Name   string `form:"name" binding:"omitempty"`
	TaskID string `form:"task_id" binding:"omitempty"`
}

func (req *QueryInjectionReq) Validate() error {
	if req.Name == "" && req.TaskID == "" {
		return fmt.Errorf("either name or task_id must be provided")
	}

	if req.Name != "" && req.TaskID != "" {
		return fmt.Errorf("only one of name or task_id should be provided")
	}

	if req.TaskID != "" {
		if !utils.IsValidUUID(req.TaskID) {
			return fmt.Errorf("invalid task_id format: %s", req.TaskID)
		}
	}

	return nil
}

type LabelItem struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type SubmitInjectionReq struct {
	ProjectName string `json:"project_name" binding:"required"`

	Interval    int             `json:"interval" binding:"required,min=1"`
	PreDuration int             `json:"pre_duration" binding:"required,min=1"`
	Specs       []chaos.Node    `json:"specs" binding:"required"`
	Benchmark   string          `json:"benchmark" binding:"required"`
	Algorithms  []AlgorithmItem `json:"algorithms" binding:"omitempty"`
	Labels      []LabelItem     `json:"labels" binding:"omitempty"`
}

func (req *SubmitInjectionReq) ParseInjectionSpecs() ([]InjectionConfig, error) {
	configs := make([]InjectionConfig, 0, len(req.Specs))
	for idx, spec := range req.Specs {
		childNode, exists := spec.Children[strconv.Itoa(spec.Value)]
		if !exists {
			return nil, fmt.Errorf("failed to find key %d in the children", spec.Value)
		}

		faultDuration := childNode.Children[consts.DurationNodeKey].Value

		conf, err := chaos.NodeToStruct[chaos.InjectionConf](&spec)
		if err != nil {
			return nil, fmt.Errorf("failed to convert node to injecton conf: %v", err)
		}

		displayConfig, err := conf.GetDisplayConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get display config: %v", err)
		}

		displayData, err := json.Marshal(displayConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal injection spec to display config: %v", err)
		}

		configs = append(configs, InjectionConfig{
			Index:         idx,
			FaultType:     spec.Value,
			FaultDuration: faultDuration,
			DisplayData:   string(displayData),
			Conf:          conf,
			Node:          &spec,
			Labels:        req.Labels,
		})
	}

	return configs, nil
}

func (req *SubmitInjectionReq) Validate() error {
	if req.ProjectName == "" {
		return fmt.Errorf("project name must not be blank")
	}

	if req.Interval <= req.PreDuration {
		return fmt.Errorf("interval must be greater than pre_duration")
	}

	if len(req.Specs) == 0 {
		return fmt.Errorf("specs must not be empty")
	}

	if req.Benchmark == "" {
		return fmt.Errorf("benchmark must not be blank")
	} else {
		if _, exists := config.GetValidBenchmarkMap()[req.Benchmark]; !exists {
			return fmt.Errorf("invalid benchmark: %s", req.Benchmark)
		}
	}

	if req.Algorithms != nil {
		for _, algorithm := range req.Algorithms {
			if algorithm.Name == "" {
				return fmt.Errorf("algorithm must not be empty")
			}

			detector := config.GetString("algo.detector")
			if algorithm.Name == detector {
				return fmt.Errorf("algorithm %s is not allowed for fault injection", detector)
			}
		}
	}

	if req.Labels == nil {
		req.Labels = make([]LabelItem, 0)
	}

	return nil
}

// analysis

type FaultInjectionNoIssuesReq struct {
	Env   string `form:"env" binding:"omitempty"`
	Batch string `form:"batch" binding:"omitempty"`

	TimeRangeQuery
}

func (req *FaultInjectionNoIssuesReq) Validate() error {
	return req.TimeRangeQuery.Validate()
}

type FaultInjectionWithIssuesReq struct {
	Env   string `form:"env" binding:"omitempty"`
	Batch string `form:"batch" binding:"omitempty"`

	TimeRangeQuery
}

func (req *FaultInjectionWithIssuesReq) Validate() error {
	return req.TimeRangeQuery.Validate()
}

type KeyResourceResp map[string]string

type NsResourcesResp map[string]chaos.Resources

type InjectionFieldMappingResp struct {
	StatusMap        map[int]string                 `json:"status" swaggertype:"object"`
	FaultTypeMap     map[chaos.ChaosType]string     `json:"fault_type" swaggertype:"object"`
	FaultResourceMap map[string]chaos.ResourceField `json:"fault_resource" swaggertype:"object"`
}

type ListInjectionsResp ListResp[InjectionItem]

type QueryInjectionResp struct {
	database.FaultInjectionSchedule
	GroundTruth chaos.Groundtruth `json:"ground_truth,omitempty"`
}

type SubmitInjectionResp struct {
	SubmitResp
	DuplicatedCount int `json:"duplicated_count"`
	OriginalCount   int `json:"original_count"`
}

// FaultInjectionNoIssuesResp Fault injection response without issues
type FaultInjectionNoIssuesResp struct {
	DatasetID     int        `json:"dataset_id"`
	EngineConfig  chaos.Node `json:"engine_config"`
	InjectionName string     `json:"injection_name"`
}

// FaultInjectionWithIssuesResp Fault injection response with issues
type FaultInjectionWithIssuesResp struct {
	DatasetID           int        `json:"dataset_id"`
	EngineConfig        chaos.Node `json:"engine_config"`
	InjectionName       string     `json:"injection_name"`
	Issues              string     `json:"issues"`
	AbnormalAvgDuration float64    `json:"abnormal_avg_duration"`
	NormalAvgDuration   float64    `json:"normal_avg_duration"`
	AbnormalSuccRate    float64    `json:"abnormal_succ_rate"`
	NormalSuccRate      float64    `json:"normal_succ_rate"`
	AbnormalP99         float64    `json:"abnormal_p99"`
	NormalP99           float64    `json:"normal_p99"`
}

// InjectionStatsResp Fault injection statistics response
type InjectionStatsResp struct {
	NoIssuesRecords      int64 `json:"no_issues_records"`
	WithIssuesRecords    int64 `json:"with_issues_records"`
	NoIssuesInjections   int64 `json:"no_issues_injections"`
	WithIssuesInjections int64 `json:"with_issues_injections"`
}

// V2 API DTOs for FaultInjectionSchedule

// InjectionV2Response represents the response structure for injection
type InjectionV2Response struct {
	ID            int       `json:"id"`
	TaskID        string    `json:"task_id"`
	FaultType     int       `json:"fault_type"`
	DisplayConfig string    `json:"display_config"`
	EngineConfig  string    `json:"engine_config"`
	PreDuration   int       `json:"pre_duration"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        int       `json:"status"`
	Description   string    `json:"description"`
	Benchmark     string    `json:"benchmark"`
	InjectionName string    `json:"injection_name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Optional relations
	Task   *TaskV2Response  `json:"task,omitempty"`
	Labels []database.Label `json:"labels,omitempty"` // Associated labels
}

// InjectionV2ListReq represents the request for listing injections
type InjectionV2ListReq struct {
	Page      int      `form:"page" binding:"omitempty,min=1"`
	Size      int      `form:"size" binding:"omitempty,min=1"`
	TaskID    string   `form:"task_id" binding:"omitempty"`
	FaultType *int     `form:"fault_type" binding:"omitempty"`
	Status    *int     `form:"status" binding:"omitempty"`
	Benchmark string   `form:"benchmark" binding:"omitempty"`
	Search    string   `form:"search" binding:"omitempty"`
	Tags      []string `form:"tags" binding:"omitempty"` // Tag values to filter by
	SortBy    string   `form:"sort_by" binding:"omitempty,oneof=id task_id fault_type status benchmark injection_name created_at updated_at"`
	SortOrder string   `form:"sort_order" binding:"omitempty,oneof=asc desc"`
	Include   string   `form:"include" binding:"omitempty"`
}

// InjectionV2UpdateReq represents the request for updating injection
type InjectionV2UpdateReq struct {
	TaskID        *string    `json:"task_id" binding:"omitempty"`
	FaultType     *int       `json:"fault_type" binding:"omitempty"`
	DisplayConfig *string    `json:"display_config" binding:"omitempty"`
	EngineConfig  *string    `json:"engine_config" binding:"omitempty"`
	PreDuration   *int       `json:"pre_duration" binding:"omitempty"`
	StartTime     *time.Time `json:"start_time" binding:"omitempty"`
	EndTime       *time.Time `json:"end_time" binding:"omitempty"`
	Status        *int       `json:"status" binding:"omitempty"`
	Description   *string    `json:"description" binding:"omitempty"`
	Benchmark     *string    `json:"benchmark" binding:"omitempty"`
	InjectionName *string    `json:"injection_name" binding:"omitempty"`
}

// InjectionV2SearchReq represents the request for advanced search
type InjectionV2SearchReq struct {
	Page         int         `json:"page" binding:"omitempty,min=1"`
	Size         int         `json:"size" binding:"omitempty,min=1,max=100"`
	TaskIDs      []string    `json:"task_ids" binding:"omitempty"`
	FaultTypes   []int       `json:"fault_types" binding:"omitempty"`
	Statuses     []int       `json:"statuses" binding:"omitempty"`
	Benchmarks   []string    `json:"benchmarks" binding:"omitempty"`
	Search       string      `json:"search" binding:"omitempty"`
	Tags         []string    `json:"tags" binding:"omitempty"`   // Tag values to filter by
	Labels       []LabelItem `json:"labels" binding:"omitempty"` // Custom labels to filter by
	StartTimeGte *time.Time  `json:"start_time_gte" binding:"omitempty"`
	StartTimeLte *time.Time  `json:"start_time_lte" binding:"omitempty"`
	EndTimeGte   *time.Time  `json:"end_time_gte" binding:"omitempty"`
	EndTimeLte   *time.Time  `json:"end_time_lte" binding:"omitempty"`
	CreatedAtGte *time.Time  `json:"created_at_gte" binding:"omitempty"`
	CreatedAtLte *time.Time  `json:"created_at_lte" binding:"omitempty"`
	SortBy       string      `json:"sort_by" binding:"omitempty,oneof=id task_id fault_type status benchmark injection_name created_at updated_at"`
	SortOrder    string      `json:"sort_order" binding:"omitempty,oneof=asc desc"`
	Include      string      `json:"include" binding:"omitempty"`
}

// InjectionSearchResponse represents the search response
type InjectionSearchResponse struct {
	Items      []InjectionV2Response `json:"items"`
	Pagination PaginationInfo        `json:"pagination"`
}

// ToInjectionV2Response converts database model to response DTO
func ToInjectionV2Response(injection *database.FaultInjectionSchedule, includeTask bool) *InjectionV2Response {
	response := &InjectionV2Response{
		ID:            injection.ID,
		TaskID:        injection.TaskID,
		FaultType:     injection.FaultType,
		DisplayConfig: injection.DisplayConfig,
		EngineConfig:  injection.EngineConfig,
		PreDuration:   injection.PreDuration,
		StartTime:     utils.GetTimeValue(injection.StartTime, time.Time{}),
		EndTime:       utils.GetTimeValue(injection.EndTime, time.Time{}),
		Status:        injection.Status,
		Description:   injection.Description,
		Benchmark:     injection.Benchmark,
		InjectionName: injection.InjectionName,
		CreatedAt:     injection.CreatedAt,
		UpdatedAt:     injection.UpdatedAt,
	}

	if includeTask && injection.Task != nil {
		response.Task = ToTaskV2Response(injection.Task)
	}

	return response
}

// ToInjectionV2ResponseWithLabels converts database model to response DTO with labels
func ToInjectionV2ResponseWithLabels(injection *database.FaultInjectionSchedule, includeTask bool, labels []database.Label) *InjectionV2Response {
	response := ToInjectionV2Response(injection, includeTask)
	response.Labels = labels
	return response
}

// TaskV2Response represents a simplified task response for injection
type TaskV2Response struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToTaskV2Response converts task model to response DTO
func ToTaskV2Response(task *database.Task) *TaskV2Response {
	return &TaskV2Response{
		ID:        task.ID,
		Status:    task.Status,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
	}
}

// InjectionV2CreateItem represents a single injection creation request
type InjectionV2CreateItem struct {
	TaskID        *string    `json:"task_id" binding:"omitempty"`
	FaultType     int        `json:"fault_type" binding:"required"`
	DisplayConfig string     `json:"display_config" binding:"required"`
	EngineConfig  string     `json:"engine_config" binding:"required"`
	PreDuration   int        `json:"pre_duration" binding:"required,min=0"`
	StartTime     *time.Time `json:"start_time" binding:"omitempty"`
	EndTime       *time.Time `json:"end_time" binding:"omitempty"`
	Status        int        `json:"status" binding:"omitempty"`
	Description   string     `json:"description" binding:"omitempty"`
	Benchmark     string     `json:"benchmark" binding:"required"`
	InjectionName string     `json:"injection_name" binding:"required"`
}

func (item *InjectionV2CreateItem) ToEntity() database.FaultInjectionSchedule {
	return database.FaultInjectionSchedule{
		TaskID:        utils.GetStringValue(item.TaskID, ""),
		FaultType:     item.FaultType,
		DisplayConfig: item.DisplayConfig,
		EngineConfig:  item.EngineConfig,
		PreDuration:   item.PreDuration,
		StartTime:     utils.GetTimePtr(item.StartTime, time.Now()),
		EndTime:       utils.GetTimePtr(item.EndTime, time.Now().Add(time.Hour)),
		Status:        utils.GetIntValue(&item.Status, 0),
		Description:   item.Description,
		Benchmark:     item.Benchmark,
		InjectionName: item.InjectionName,
	}
}

// InjectionV2CreateReq represents the batch creation request for injections
type InjectionV2CreateReq struct {
	Injections []InjectionV2CreateItem `json:"injections" binding:"required,min=1,max=100"`
}

// Validate validates the injection creation request
func (req *InjectionV2CreateReq) Validate() error {
	if len(req.Injections) == 0 {
		return fmt.Errorf("at least one injection must be provided")
	}

	if len(req.Injections) > 100 {
		return fmt.Errorf("cannot create more than 100 injections at once")
	}

	for i, injection := range req.Injections {
		if injection.InjectionName == "" {
			return fmt.Errorf("injection_name is required for injection %d", i+1)
		}

		if injection.Benchmark == "" {
			return fmt.Errorf("benchmark is required for injection %d", i+1)
		}

		if injection.DisplayConfig == "" {
			return fmt.Errorf("display_config is required for injection %d", i+1)
		}

		if injection.EngineConfig == "" {
			return fmt.Errorf("engine_config is required for injection %d", i+1)
		}

		if injection.PreDuration < 0 {
			return fmt.Errorf("pre_duration must be non-negative for injection %d", i+1)
		}

		// Validate time range if both start and end time are provided
		if injection.StartTime != nil && injection.EndTime != nil {
			if injection.EndTime.Before(*injection.StartTime) {
				return fmt.Errorf("end_time must be after start_time for injection %d", i+1)
			}
		}
	}

	return nil
}

// InjectionV2CreateResponse represents the response for batch injection creation
type InjectionV2CreateResponse struct {
	CreatedCount int                    `json:"created_count"`
	CreatedItems []InjectionV2Response  `json:"created_items"`
	FailedCount  int                    `json:"failed_count,omitempty"`
	FailedItems  []InjectionCreateError `json:"failed_items,omitempty"`
	Message      string                 `json:"message"`
}

// InjectionCreateError represents an error during injection creation
type InjectionCreateError struct {
	Index int                   `json:"index"`
	Error string                `json:"error"`
	Item  InjectionV2CreateItem `json:"item"`
}

// InjectionV2LabelManageReq Manage labels in injection
type InjectionV2LabelManageReq struct {
	AddTags    []string `json:"add_tags"`    // List of tag values to add
	RemoveTags []string `json:"remove_tags"` // List of tag values to remove
}

// InjectionV2CustomLabelManageReq Manage custom labels (key-value pairs) in injection
type InjectionV2CustomLabelManageReq struct {
	AddLabels    []LabelItem `json:"add_labels"`    // List of labels to add
	RemoveLabels []string    `json:"remove_labels"` // List of label keys to remove
}
