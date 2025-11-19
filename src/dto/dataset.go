package dto

import (
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/database"
	"aegis/utils"
)

// =====================================================================
// Dataset Service DTOs
// =====================================================================

type DatasetRef struct {
	Name    string `json:"name" binding:"required"`
	Version string `json:"version" binding:"omitempty"`
}

func (ref *DatasetRef) Validate() error {
	if ref.Name == "" {
		return fmt.Errorf("dataset name is required")
	}
	if ref.Version != "" {
		if _, _, _, err := utils.ParseSemanticVersion(ref.Version); err != nil {
			return fmt.Errorf("invalid semantic version: %s, %v", ref.Version, err)
		}
	}
	return nil
}

// ===================== Dataset CRUD DTOs =====================

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

func (req *CreateDatasetReq) ConvertToDataset() *database.Dataset {
	return &database.Dataset{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		IsPublic:    *req.IsPublic,
		Status:      consts.CommonEnabled,
	}
}

type ListDatasetReq struct {
	PaginationReq
	Type     string             `form:"type" binding:"omitempty"`
	IsPublic *bool              `form:"is_public" binding:"omitempty"`
	Status   *consts.StatusType `form:"status" binding:"omitempty"`
}

func (req *ListDatasetReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	return validateStatusField(req.Status, false)
}

type SearchDatasetReq struct {
	AdvancedSearchReq

	Name        *string `json:"name,omitempty"`
	Type        *string `json:"type,omitempty"`
	Description *string `json:"description,omitempty"`
	DataSource  *string `json:"data_source,omitempty"`
	Format      *string `json:"format,omitempty"`
	Status      *int    `json:"status,omitempty"`
}

func (dsr *SearchDatasetReq) ConvertToSearchRequest() *SearchReq {
	sr := dsr.AdvancedSearchReq.ConvertAdvancedToSearch()

	if dsr.Name != nil {
		sr.AddFilter("name", OpLike, *dsr.Name)
	}
	if dsr.Type != nil {
		sr.AddFilter("type", OpEqual, *dsr.Type)
	}
	if dsr.Description != nil {
		sr.AddFilter("description", OpLike, *dsr.Description)
	}
	if dsr.DataSource != nil {
		sr.AddFilter("data_source", OpLike, *dsr.DataSource)
	}
	if dsr.Format != nil {
		sr.AddFilter("format", OpEqual, *dsr.Format)
	}

	return sr
}

type UpdateDatasetReq struct {
	Description *string            `json:"description" binding:"omitempty"`
	IsPublic    *bool              `json:"is_public" binding:"omitempty"`
	Status      *consts.StatusType `json:"status" binding:"omitempty"`
}

func (req *UpdateDatasetReq) Validate() error {
	return validateStatusField(req.Status, true)
}

func (req *UpdateDatasetReq) PatchDatasetModel(target *database.Dataset) {
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

type ManageDatasetLabelReq struct {
	AddLabels    []LabelItem `json:"add_labels" binding:"omitempty"`    // List of labels to add
	RemoveLabels []string    `json:"remove_labels" binding:"omitempty"` // List of label keys to remove
}

func (req *ManageDatasetLabelReq) Validate() error {
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

type ManageDatasetVersionInjectionReq struct {
	AddInjections    []int `json:"add_injections" binding:"omitempty"`    // List of injection IDs to add
	RemoveInjections []int `json:"remove_injections" binding:"omitempty"` // List of injection IDs to remove
}

func (req *ManageDatasetVersionInjectionReq) Validate() error {
	if len(req.AddInjections) == 0 && len(req.RemoveInjections) == 0 {
		return fmt.Errorf("at least one of add_injections or remove_injections must be provided")
	}

	for i, id := range req.AddInjections {
		if id <= 0 {
			return fmt.Errorf("invalid injection id at index %d in add_injections: %d", i, id)
		}
	}
	for i, id := range req.RemoveInjections {
		if id <= 0 {
			return fmt.Errorf("invalid injection id at index %d in remove_injections: %d", i, id)
		}
	}

	return nil
}

type DatasetResp struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	IsPublic  bool      `json:"is_public"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Labels []LabelItem `json:"labels,omitempty"`
}

func NewDatasetResp(dataset *database.Dataset) *DatasetResp {
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
		resp.Labels = make([]LabelItem, 0, len(dataset.Labels))
		for _, l := range dataset.Labels {
			resp.Labels = append(resp.Labels, LabelItem{
				Key:   l.Key,
				Value: l.Value,
			})
		}
	}
	return resp
}

type DatasetDetailResp struct {
	DatasetResp

	Description string `json:"description"`

	Versions []DatasetVersionResp `json:"versions"`
}

func NewDatasetDetailResp(dataset *database.Dataset) *DatasetDetailResp {
	return &DatasetDetailResp{
		DatasetResp: *NewDatasetResp(dataset),
		Description: dataset.Description,
	}
}

// ===================== Dataset Version CRUD DTOs =====================

type CreateDatasetVersionReq struct {
	Name        string `json:"name" binding:"required"`
	DataSource  string `json:"data_source" binding:"omitempty"`
	Format      string `json:"format" binding:"omitempty"`
	DownloadURL string `json:"download_url" binding:"omitempty"`
	Checksum    string `json:"checksum" binding:"omitempty"`
}

func (req *CreateDatasetVersionReq) Validate() error {
	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if _, _, _, err := utils.ParseSemanticVersion(req.Name); err != nil {
		return fmt.Errorf("invalid semantic version: %s, %v", req.Name, err)
	}

	if req.Format != "" {
		req.Format = strings.TrimSpace(req.Format)
	}

	return nil
}

func (req *CreateDatasetVersionReq) ConvertToDatasetVersion() *database.DatasetVersion {
	version := &database.DatasetVersion{
		Name:     req.Name,
		Format:   req.Format,
		Checksum: req.Checksum,
		Status:   consts.CommonEnabled,
	}

	return version
}

// ListDatasetVersionReq represents dataset version list query parameters
type ListDatasetVersionReq struct {
	PaginationReq
	Status *consts.StatusType `json:"status" binding:"omitempty"`
}

func (req *ListDatasetVersionReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	return validateStatusField(req.Status, false)
}

type UpdateDatasetVersionReq struct {
	Checksum *string            `json:"checksum" binding:"omitempty"`
	Format   *string            `json:"format" binding:"omitempty"`
	Status   *consts.StatusType `json:"status" binding:"omitempty"`
}

func (req *UpdateDatasetVersionReq) Validate() error {
	return validateStatusField(req.Status, true)
}

func (req *UpdateDatasetVersionReq) PatchDatasetVersionModel(target *database.DatasetVersion) {
	if req.Format != nil {
		target.Format = *req.Format
	}
	if req.Checksum != nil {
		target.Checksum = *req.Checksum
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
}

type DatasetVersionResp struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewDatasetVersionResp(version *database.DatasetVersion) *DatasetVersionResp {
	return &DatasetVersionResp{
		ID:        version.ID,
		Name:      version.Name,
		UpdatedAt: version.UpdatedAt,
	}
}

type DatasetVersionDetailResp struct {
	DatasetVersionResp

	Format    string `json:"format"`
	Checksum  string `json:"checksum"`
	FileCount int    `json:"file_count"`
}

func NewDatasetVersionDetailResp(version *database.DatasetVersion) *DatasetVersionDetailResp {
	return &DatasetVersionDetailResp{
		DatasetVersionResp: *NewDatasetVersionResp(version),
		Format:             version.Format,
		Checksum:           version.Checksum,
		FileCount:          version.FileCount,
	}
}
