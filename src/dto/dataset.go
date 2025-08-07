package dto

import (
	"encoding/json"
	"fmt"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/utils"
)

type DatasetDeleteReq struct {
	Names []string `form:"names" binding:"required,min=1,dive,required,max=64"`
}

type DatasetDeleteResp struct {
	SuccessCount int64    `json:"success_count"`
	FailedNames  []string `json:"failed_names"`
}

type DatasetDownloadReq struct {
	GroupIDs []string `form:"group_ids"`
	Names    []string `form:"names"`
}

func (r *DatasetDownloadReq) Validate() error {
	hasGroupIDs := len(r.GroupIDs) > 0
	hasNames := len(r.Names) > 0
	if !hasGroupIDs && !hasNames {
		return fmt.Errorf("one of group_ids or names must be provided")
	}

	if hasGroupIDs && hasNames {
		return fmt.Errorf("only one of group_ids or names must be provided")
	}

	return nil
}

type DatasetItem struct {
	Name      string         `json:"name"`
	Param     map[string]any `json:"param" swaggertype:"array,object"`
	StartTime time.Time      `json:"start_time"`
	EndTime   time.Time      `json:"end_time"`
}

func (d *DatasetItem) Convert(record database.FaultInjectionSchedule) error {
	var param map[string]any
	if err := json.Unmarshal([]byte(record.DisplayConfig), &param); err != nil {
		return fmt.Errorf("faild to unmarshal display config: %v", err)
	}

	param["fault_type"] = chaos.ChaosTypeMap[chaos.ChaosType(record.FaultType)]
	param["pre_duration"] = record.PreDuration

	d.Name = record.InjectionName
	d.Param = param
	d.StartTime = utils.GetTimeValue(record.StartTime, time.Time{})
	d.EndTime = utils.GetTimeValue(record.EndTime, time.Time{})

	return nil
}

type DatasetItemWithID struct {
	ID int
	DatasetItem
}

func (d *DatasetItemWithID) Convert(record database.FaultInjectionSchedule) error {
	var item DatasetItem
	err := item.Convert(record)
	if err != nil {
		return err
	}

	d.ID = record.ID
	d.DatasetItem = item
	return nil
}

type DatasetBuildPayload struct {
	Benchmark   string            `json:"benchmark" binding:"omitempty"`
	Name        string            `json:"name" binding:"required"`
	PreDuration *int              `json:"pre_duration" binding:"omitempty"`
	EnvVars     map[string]string `json:"env_vars" binding:"omitempty" swaggertype:"object"`
}

func (p *DatasetBuildPayload) Validate() error {
	if p.Benchmark == "" {
		p.Benchmark = "clickhouse"
	}

	if p.PreDuration != nil && *p.PreDuration <= 0 {
		return fmt.Errorf("pre_duration must be greater than 0")
	}

	for key := range p.EnvVars {
		if err := utils.IsValidEnvVar(key); err != nil {
			return fmt.Errorf("invalid environment variable key %s: %v", key, err)
		}
	}

	return nil
}

type SubmitDatasetBuildingReq struct {
	ProjectName string                `json:"project_name" binding:"required"`
	Payloads    []DatasetBuildPayload `json:"payloads" binding:"required,dive,required"`
}

func (req *SubmitDatasetBuildingReq) Validate() error {
	if req.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}
	if len(req.Payloads) == 0 {
		return fmt.Errorf("at least one dataset build payload is required")
	}
	for _, payload := range req.Payloads {
		if err := payload.Validate(); err != nil {
			return fmt.Errorf("invalid dataset build payload: %v", err)
		}
	}
	return nil
}

type DatasetJoinedResult struct {
	GroupID string
	Name    string
}

func (d *DatasetJoinedResult) Convert(groupID, name string) {
	d.GroupID = groupID
	d.Name = name
}

var DatasetStatusMap = map[int]string{
	consts.DatapackInitial:       "initial",
	consts.DatapackInjectSuccess: "inject_success",
	consts.DatapackInjectFailed:  "inject_failed",
	consts.DatapackBuildSuccess:  "build_success",
	consts.DatapackBuildFailed:   "build_failed",
	consts.DatapackDeleted:       "deleted",
}

var DatasetStatusReverseMap = map[string]int{
	"initial":        consts.DatapackInitial,
	"inject_success": consts.DatapackInjectSuccess,
	"inject_failed":  consts.DatapackInjectFailed,
	"build_success":  consts.DatapackBuildSuccess,
	"build_failed":   consts.DatapackBuildFailed,
	"deleted":        consts.DatapackDeleted,
}

// ===================== V2 API DTOs =====================

// InjectionRef represents an injection reference by ID or name
type InjectionRef struct {
	ID   *int    `json:"id,omitempty"`   // Injection ID
	Name *string `json:"name,omitempty"` // Injection name
}

// DatasetV2CreateReq Create dataset request
type DatasetV2CreateReq struct {
	Name          string                    `json:"name" binding:"required,max=255"` // Dataset name
	Version       string                    `json:"version" binding:"max=50"`        // Dataset version, optional, defaults to v1.0
	Description   string                    `json:"description" binding:"max=1000"`  // Dataset description
	Type          string                    `json:"type" binding:"required,max=50"`  // Dataset type
	DataSource    string                    `json:"data_source" binding:"max=500"`   // Data source description
	Format        string                    `json:"format" binding:"max=50"`         // Data format
	IsPublic      *bool                     `json:"is_public"`                       // Whether public, optional, defaults to false
	InjectionRefs []InjectionRef            `json:"injection_refs"`                  // Associated fault injection references (ID or name)
	LabelIDs      []int                     `json:"label_ids"`                       // Associated label ID list
	NewLabels     []DatasetV2LabelCreateReq `json:"new_labels"`                      // New label list
}

type DatasetV2GetReq struct {
	IncludeInjections bool `form:"include_injections"` // Include related fault injections
	IncludeLabels     bool `form:"include_labels"`     // Include related labels
}

// DatasetV2LabelCreateReq Create label request
type DatasetV2LabelCreateReq struct {
	Key         string `json:"key" binding:"required,max=100"`   // Label key
	Value       string `json:"value" binding:"required,max=255"` // Label value
	Category    string `json:"category" binding:"max=50"`        // Label category
	Description string `json:"description" binding:"max=500"`    // Label description
	Color       string `json:"color" binding:"max=7"`            // Label color (hex format)
}

// DatasetV2UpdateReq Update dataset request
type DatasetV2UpdateReq struct {
	Name          *string                   `json:"name" binding:"omitempty,max=255"`         // Dataset name
	Version       *string                   `json:"version" binding:"omitempty,max=50"`       // Dataset version
	Description   *string                   `json:"description" binding:"omitempty,max=1000"` // Dataset description
	Type          *string                   `json:"type" binding:"omitempty,max=50"`          // Dataset type
	DataSource    *string                   `json:"data_source" binding:"omitempty,max=500"`  // Data source description
	Format        *string                   `json:"format" binding:"omitempty,max=50"`        // Data format
	IsPublic      *bool                     `json:"is_public"`                                // Whether public
	InjectionRefs []InjectionRef            `json:"injection_refs"`                           // Update associated fault injection references (complete replacement)
	LabelIDs      []int                     `json:"label_ids"`                                // Update associated label ID list (complete replacement)
	NewLabels     []DatasetV2LabelCreateReq `json:"new_labels"`                               // New label list
}

// DatasetV2InjectionManageReq Manage fault injections in dataset
type DatasetV2InjectionManageReq struct {
	AddInjections    []int `json:"add_injections"`    // List of fault injection IDs to add
	RemoveInjections []int `json:"remove_injections"` // List of fault injection IDs to remove
}

// DatasetV2LabelManageReq Manage labels in dataset
type DatasetV2LabelManageReq struct {
	AddLabels    []int                     `json:"add_labels"`    // List of label IDs to add
	RemoveLabels []int                     `json:"remove_labels"` // List of label IDs to remove
	NewLabels    []DatasetV2LabelCreateReq `json:"new_labels"`    // New label list
}

// DatasetV2Response Dataset response
type DatasetV2Response struct {
	ID          int    `json:"id"`          // Unique identifier
	Name        string `json:"name"`        // Dataset name
	Version     string `json:"version"`     // Dataset version
	Description string `json:"description"` // Dataset description
	Type        string `json:"type"`        // Dataset type

	FileCount   int       `json:"file_count"`             // File count
	DataSource  string    `json:"data_source"`            // Data source description
	Format      string    `json:"format"`                 // Data format
	Status      int       `json:"status"`                 // Status
	IsPublic    bool      `json:"is_public"`              // Whether public
	DownloadURL string    `json:"download_url,omitempty"` // Download URL
	Checksum    string    `json:"checksum,omitempty"`     // File checksum
	CreatedAt   time.Time `json:"created_at"`             // Creation time
	UpdatedAt   time.Time `json:"updated_at"`             // Update time

	Injections []InjectionV2Response `json:"injections,omitempty"` // Associated fault injections
	Labels     []database.Label      `json:"labels,omitempty"`     // Associated labels
}

// DatasetV2InjectionRelationResponse Dataset fault injection relation response
type DatasetV2InjectionRelationResponse struct {
	ID               int                              `json:"id"`                        // Relation ID
	FaultInjectionID int                              `json:"fault_injection_id"`        // Fault injection ID
	CreatedAt        time.Time                        `json:"created_at"`                // Creation time
	UpdatedAt        time.Time                        `json:"updated_at"`                // Update time
	FaultInjection   *database.FaultInjectionSchedule `json:"fault_injection,omitempty"` // Fault injection details
}

// DatasetV2ListReq Dataset list query request
type DatasetV2ListReq struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`         // Page number, defaults to 1
	Size      int    `form:"size" binding:"omitempty,min=1,max=100"` // Page size, defaults to 20
	Type      string `form:"type"`                                   // Dataset type filter
	Status    *int   `form:"status"`                                 // Status filter
	IsPublic  *bool  `form:"is_public"`                              // Public filter
	Search    string `form:"search"`                                 // Search keywords (name, description)
	SortBy    string `form:"sort_by"`                                // Sort field (id, name, created_at, updated_at)
	SortOrder string `form:"sort_order"`                             // Sort direction (asc, desc)
	Include   string `form:"include"`                                // Included related data (injections, labels)
}

// DatasetV2SearchReq Dataset search request (POST method, supports complex conditions)
type DatasetV2SearchReq struct {
	Page        int              `json:"page" binding:"omitempty,min=1"`         // Page number
	Size        int              `json:"size" binding:"omitempty,min=1,max=100"` // Page size
	Types       []string         `json:"types"`                                  // Dataset type list
	Statuses    []int            `json:"statuses"`                               // Status list
	IsPublic    *bool            `json:"is_public"`                              // Whether public
	Search      string           `json:"search"`                                 // Search keywords
	DateRange   *DateRangeFilter `json:"date_range"`                             // Date range filter
	SizeRange   *SizeRangeFilter `json:"size_range"`                             // Size range filter
	Include     []string         `json:"include"`                                // Included related data
	SortBy      string           `json:"sort_by"`                                // Sort field
	SortOrder   string           `json:"sort_order"`                             // Sort direction
	LabelKeys   []string         `json:"label_keys"`                             // Filter by label key
	LabelValues []string         `json:"label_values"`                           // Filter by label value
}

// DateRangeFilter Date range filter
type DateRangeFilter struct {
	StartTime *time.Time `json:"start_time"` // Start time
	EndTime   *time.Time `json:"end_time"`   // End time
}

// SizeRangeFilter Size range filter
type SizeRangeFilter struct {
	MinSize *int64 `json:"min_size"` // Minimum size (bytes)
	MaxSize *int64 `json:"max_size"` // Maximum size (bytes)
}

// DatasetSearchResponse Dataset search response structure
type DatasetSearchResponse struct {
	Items      []DatasetV2Response `json:"items"`      // Result list
	Pagination PaginationInfo      `json:"pagination"` // Pagination info
}

// ToDatasetV2Response converts Database.Dataset to DatasetV2Response
func ToDatasetV2Response(dataset *database.Dataset, includeRelations bool) *DatasetV2Response {
	resp := &DatasetV2Response{
		ID:          dataset.ID,
		Name:        dataset.Name,
		Version:     dataset.Version,
		Description: dataset.Description,
		Type:        dataset.Type,
		FileCount:   dataset.FileCount,
		DataSource:  dataset.DataSource,
		Format:      dataset.Format,
		Status:      dataset.Status,
		IsPublic:    dataset.IsPublic,
		DownloadURL: dataset.DownloadURL,
		Checksum:    dataset.Checksum,
		CreatedAt:   dataset.CreatedAt,
		UpdatedAt:   dataset.UpdatedAt,
	}

	return resp
}
