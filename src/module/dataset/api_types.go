package dataset

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"aegis/utils"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

// CreateDatasetReq represents dataset creation request.
type CreateDatasetReq struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Description string `json:"description" binding:"omitempty"`
	IsPublic    *bool  `json:"is_public" binding:"omitempty"`

	VersionReq *CreateDatasetVersionReq `json:"version" binding:"omitempty"`
}

func (req *CreateDatasetReq) Validate() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Type = strings.TrimSpace(req.Type)

	if req.Name == "" {
		return fmt.Errorf("dataset name cannot be empty")
	}
	if req.Type == "" {
		return fmt.Errorf("dataset type cannot be empty")
	}
	if req.IsPublic == nil {
		req.IsPublic = utils.BoolPtr(true)
	}
	if req.VersionReq != nil {
		if err := req.VersionReq.Validate(); err != nil {
			return fmt.Errorf("invalid dataset version request: %v", err)
		}
	}
	return nil
}

func (req *CreateDatasetReq) ConvertToDataset() *model.Dataset {
	return &model.Dataset{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		IsPublic:    *req.IsPublic,
		Status:      consts.CommonEnabled,
	}
}

// ListDatasetReq represents dataset list query parameters.
type ListDatasetReq struct {
	dto.PaginationReq
	Type     string             `form:"type" binding:"omitempty"`
	IsPublic *bool              `form:"is_public" binding:"omitempty"`
	Status   *consts.StatusType `form:"status" binding:"omitempty"`
}

func (req *ListDatasetReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	return validateStatus(req.Status, false)
}

// UpdateDatasetReq represents dataset update request.
type UpdateDatasetReq struct {
	Description *string            `json:"description" binding:"omitempty"`
	IsPublic    *bool              `json:"is_public" binding:"omitempty"`
	Status      *consts.StatusType `json:"status" binding:"omitempty"`
}

func (req *UpdateDatasetReq) Validate() error {
	return validateStatus(req.Status, true)
}

func (req *UpdateDatasetReq) PatchDatasetModel(target *model.Dataset) {
	if req.Description != nil {
		target.Description = *req.Description
	}
	if req.IsPublic != nil {
		target.IsPublic = *req.IsPublic
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
}

// ManageDatasetLabelReq represents dataset label management request.
type ManageDatasetLabelReq struct {
	AddLabels    []dto.LabelItem `json:"add_labels" binding:"omitempty"`
	RemoveLabels []string        `json:"remove_labels" binding:"omitempty"`
}

func (req *ManageDatasetLabelReq) Validate() error {
	if len(req.AddLabels) == 0 && len(req.RemoveLabels) == 0 {
		return fmt.Errorf("at least one of add_labels or remove_labels must be provided")
	}
	if err := validateLabelItems(req.AddLabels); err != nil {
		return err
	}
	for i, key := range req.RemoveLabels {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("empty label key at index %d in remove_labels", i)
		}
	}
	return nil
}

// DatasetResp represents dataset summary information.
type DatasetResp struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	IsPublic  bool            `json:"is_public"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Labels    []dto.LabelItem `json:"labels,omitempty"`
}

func NewDatasetResp(dataset *model.Dataset) *DatasetResp {
	resp := &DatasetResp{
		ID:        dataset.ID,
		Name:      dataset.Name,
		Type:      dataset.Type,
		IsPublic:  dataset.IsPublic,
		Status:    consts.GetStatusTypeName(dataset.Status),
		CreatedAt: dataset.CreatedAt,
		UpdatedAt: dataset.UpdatedAt,
	}

	if len(dataset.Labels) > 0 {
		resp.Labels = make([]dto.LabelItem, 0, len(dataset.Labels))
		for _, l := range dataset.Labels {
			resp.Labels = append(resp.Labels, dto.LabelItem{Key: l.Key, Value: l.Value})
		}
	}
	return resp
}

// DatasetDetailResp represents detailed dataset information.
type DatasetDetailResp struct {
	DatasetResp
	Description string               `json:"description"`
	Versions    []DatasetVersionResp `json:"versions"`
}

func NewDatasetDetailResp(dataset *model.Dataset) *DatasetDetailResp {
	return &DatasetDetailResp{
		DatasetResp: *NewDatasetResp(dataset),
		Description: dataset.Description,
	}
}

func validateStatus(statusPtr *consts.StatusType, isMutation bool) error {
	if statusPtr == nil {
		return nil
	}
	status := *statusPtr
	if _, exists := consts.ValidStatuses[status]; !exists {
		return fmt.Errorf("invalid status value: %d", status)
	}
	if isMutation && status == consts.CommonDeleted {
		return fmt.Errorf("status value cannot be set to deleted (%d) directly through this update/create operation", consts.CommonDeleted)
	}
	return nil
}

func validateLabelItems(items []dto.LabelItem) error {
	for i, label := range items {
		if strings.TrimSpace(label.Key) == "" {
			return fmt.Errorf("empty label key at index %d in add_labels", i)
		}
		if strings.TrimSpace(label.Value) == "" {
			return fmt.Errorf("empty label value at index %d in add_labels", i)
		}
	}
	return nil
}

// SearchDatasetReq represents advanced dataset search.
type SearchDatasetReq struct {
	dto.AdvancedSearchReq[consts.DatasetField]

	NamePattern     string `json:"name_pattern" binding:"omitempty"`
	IncludeVersions bool   `json:"include_versions" binding:"omitempty"`
}

func (req *SearchDatasetReq) Validate() error {
	if err := req.AdvancedSearchReq.Validate(); err != nil {
		return err
	}
	for i, sortField := range req.Sort {
		if _, valid := consts.DatasetAllowedFields[sortField.Field]; !valid {
			return fmt.Errorf("invalid sort_by field at index %d: %s", i, sortField.Field)
		}
	}
	for i, field := range req.GroupBy {
		if _, valid := consts.DatasetAllowedFields[field]; !valid {
			return fmt.Errorf("invalid group_by field at index %d: %s", i, field)
		}
	}
	return nil
}

func (req *SearchDatasetReq) ConvertToSearchReq() *dto.SearchReq[consts.DatasetField] {
	sr := req.ConvertAdvancedToSearch()

	if req.NamePattern != "" {
		sr.AddFilter("name", dto.OpLike, req.NamePattern)
	}
	if req.IncludeVersions {
		sr.AddInclude("Versions")
	}

	return sr
}

// ManageDatasetVersionInjectionReq represents datapack membership changes for a dataset version.
type ManageDatasetVersionInjectionReq struct {
	AddDatapacks    []string `json:"add_datapacks" binding:"omitempty"`
	RemoveDatapacks []string `json:"remove_datapacks" binding:"omitempty"`
}

func (req *ManageDatasetVersionInjectionReq) Validate() error {
	if len(req.AddDatapacks) == 0 && len(req.RemoveDatapacks) == 0 {
		return fmt.Errorf("at least one of add_injections or remove_injections must be provided")
	}

	for i, datapack := range req.AddDatapacks {
		if strings.TrimSpace(datapack) == "" {
			return fmt.Errorf("empty datapack name at index %d in add_datapacks", i)
		}
	}
	for i, datapack := range req.RemoveDatapacks {
		if strings.TrimSpace(datapack) == "" {
			return fmt.Errorf("empty datapack name at index %d in add_datapacks", i)
		}
	}

	return nil
}

// CreateDatasetVersionReq represents dataset version creation.
type CreateDatasetVersionReq struct {
	Name      string   `json:"name" binding:"required"`
	Datapacks []string `json:"datapacks" binding:"omitempty"`
}

func (req *CreateDatasetVersionReq) Validate() error {
	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if _, _, _, err := utils.ParseSemanticVersion(req.Name); err != nil {
		return fmt.Errorf("invalid semantic version: %s, %v", req.Name, err)
	}
	for i, datapack := range req.Datapacks {
		if strings.TrimSpace(datapack) == "" {
			return fmt.Errorf("empty datapack name at index %d", i)
		}
	}

	return nil
}

func (req *CreateDatasetVersionReq) ConvertToDatasetVersion() *model.DatasetVersion {
	return &model.DatasetVersion{
		Name:   req.Name,
		Status: consts.CommonEnabled,
	}
}

// ListDatasetVersionReq represents dataset version list query parameters.
type ListDatasetVersionReq struct {
	dto.PaginationReq
	Status *consts.StatusType `json:"status" binding:"omitempty"`
}

func (req *ListDatasetVersionReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	return validateStatus(req.Status, false)
}

// UpdateDatasetVersionReq represents mutable dataset version fields.
type UpdateDatasetVersionReq struct {
	Status *consts.StatusType `json:"status" binding:"omitempty"`
}

func (req *UpdateDatasetVersionReq) Validate() error {
	return validateStatus(req.Status, true)
}

func (req *UpdateDatasetVersionReq) PatchDatasetVersionModel(target *model.DatasetVersion) {
	if req.Status != nil {
		target.Status = *req.Status
	}
}

// DatasetVersionResp represents dataset version summary information.
type DatasetVersionResp struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Checksum  string    `json:"checksum"`
	FileCount int       `json:"file_count"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewDatasetVersionResp(version *model.DatasetVersion) *DatasetVersionResp {
	return &DatasetVersionResp{
		ID:        version.ID,
		Name:      version.Name,
		Checksum:  version.Checksum,
		FileCount: version.FileCount,
		UpdatedAt: version.UpdatedAt,
	}
}

// DatasetVersionDetailResp represents dataset version details including datapacks.
type DatasetVersionDetailResp struct {
	DatasetVersionResp

	Datapacks []DatasetDatapackResp `json:"datapacks,omitempty"`
}

func NewDatasetVersionDetailResp(version *model.DatasetVersion) *DatasetVersionDetailResp {
	resp := &DatasetVersionDetailResp{
		DatasetVersionResp: *NewDatasetVersionResp(version),
	}

	if len(version.Datapacks) > 0 {
		resp.Datapacks = make([]DatasetDatapackResp, 0, len(version.Datapacks))
		for _, datapack := range version.Datapacks {
			resp.Datapacks = append(resp.Datapacks, *NewDatasetDatapackResp(&datapack))
		}
	}

	return resp
}

type DatasetDatapackResp struct {
	ID                int                  `json:"id"`
	Name              string               `json:"name"`
	Source            string               `json:"source"`
	FaultType         string               `json:"fault_type"`
	Category          string               `json:"category"`
	DisplayConfig     map[string]any       `json:"display_config,omitempty" swaggertype:"object"`
	PreDuration       int                  `json:"pre_duration"`
	StartTime         *time.Time           `json:"start_time,omitempty"`
	EndTime           *time.Time           `json:"end_time,omitempty"`
	State             consts.DatapackState `json:"state" swaggertype:"string"`
	Status            string               `json:"status"`
	GroundtruthSource string               `json:"groundtruth_source"`
	BenchmarkID       *int                 `json:"benchmark_id"`
	BenchmarkName     string               `json:"benchmark_name"`
	PedestalID        *int                 `json:"pedestal_id"`
	PedestalName      string               `json:"pedestal_name"`
	CreatedAt         time.Time            `json:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at"`
	Labels            []dto.LabelItem      `json:"labels,omitempty"`
}

func NewDatasetDatapackResp(injection *model.FaultInjection) *DatasetDatapackResp {
	resp := &DatasetDatapackResp{
		ID:                injection.ID,
		Name:              injection.Name,
		Source:            string(injection.Source),
		Category:          injection.Category.String(),
		PreDuration:       injection.PreDuration,
		StartTime:         injection.StartTime,
		EndTime:           injection.EndTime,
		State:             injection.State,
		Status:            consts.GetStatusTypeName(injection.Status),
		GroundtruthSource: injection.GroundtruthSource,
		BenchmarkID:       injection.BenchmarkID,
		PedestalID:        injection.PedestalID,
		CreatedAt:         injection.CreatedAt,
		UpdatedAt:         injection.UpdatedAt,
	}

	if injection.FaultType == consts.Hybrid {
		resp.FaultType = "hybrid"
	} else {
		resp.FaultType = chaos.ChaosTypeMap[injection.FaultType]
	}

	if injection.DisplayConfig != nil {
		var displayConfigData map[string]any
		_ = json.Unmarshal([]byte(*injection.DisplayConfig), &displayConfigData)
		resp.DisplayConfig = displayConfigData
	}

	if injection.Benchmark != nil && injection.Benchmark.Container != nil {
		resp.BenchmarkName = injection.Benchmark.Container.Name
	}
	if injection.Pedestal != nil && injection.Pedestal.Container != nil {
		resp.PedestalName = injection.Pedestal.Container.Name
	}
	if len(injection.Labels) > 0 {
		resp.Labels = make([]dto.LabelItem, 0, len(injection.Labels))
		for _, l := range injection.Labels {
			resp.Labels = append(resp.Labels, dto.LabelItem{Key: l.Key, Value: l.Value, IsSystem: l.IsSystem})
		}
	}
	return resp
}
