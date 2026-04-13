package dto

import (
	"aegis/database"
	"encoding/json"
	"time"
)

// CreateChaosSystemReq represents the request to create a new chaos system
type CreateChaosSystemReq struct {
	Name           string `json:"name" binding:"required"`
	DisplayName    string `json:"display_name" binding:"required"`
	NsPattern      string `json:"ns_pattern" binding:"required"`
	ExtractPattern string `json:"extract_pattern" binding:"required"`
	Count          int    `json:"count" binding:"required,min=1"`
	Description    string `json:"description"`
}

// UpdateChaosSystemReq represents the request to update a chaos system
type UpdateChaosSystemReq struct {
	DisplayName    *string `json:"display_name"`
	NsPattern      *string `json:"ns_pattern"`
	ExtractPattern *string `json:"extract_pattern"`
	Count          *int    `json:"count" binding:"omitempty,min=1"`
	Description    *string `json:"description"`
}

// ChaosSystemResp represents a chaos system in API responses
type ChaosSystemResp struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	DisplayName    string    `json:"display_name"`
	NsPattern      string    `json:"ns_pattern"`
	ExtractPattern string    `json:"extract_pattern"`
	Count          int       `json:"count"`
	Description    string    `json:"description"`
	IsBuiltin      bool      `json:"is_builtin"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ListChaosSystemReq represents the request to list chaos systems
type ListChaosSystemReq struct {
	PaginationReq
}

// NewChaosSystemResp creates a ChaosSystemResp from a database System
func NewChaosSystemResp(s *database.System) *ChaosSystemResp {
	return &ChaosSystemResp{
		ID:             s.ID,
		Name:           s.Name,
		DisplayName:    s.DisplayName,
		NsPattern:      s.NsPattern,
		ExtractPattern: s.ExtractPattern,
		Count:          s.Count,
		Description:    s.Description,
		IsBuiltin:      s.IsBuiltin,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

// UpsertSystemMetadataReq represents a single metadata upsert request
type UpsertSystemMetadataReq struct {
	MetadataType string          `json:"metadata_type" binding:"required"` // "service_endpoint", "java_class_method", "database_operation", "grpc_operation", "network_dependency"
	ServiceName  string          `json:"service_name" binding:"required"`
	Data         json.RawMessage `json:"data" binding:"required"`
}

// BulkUpsertSystemMetadataReq represents a bulk metadata upsert request
type BulkUpsertSystemMetadataReq struct {
	Items []UpsertSystemMetadataReq `json:"items" binding:"required,dive"`
}

// SystemMetadataResp represents system metadata in API responses
type SystemMetadataResp struct {
	ID           int             `json:"id"`
	SystemName   string          `json:"system_name"`
	MetadataType string          `json:"metadata_type"`
	ServiceName  string          `json:"service_name"`
	Data         json.RawMessage `json:"data"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// NewSystemMetadataResp creates a SystemMetadataResp from a database SystemMetadata
func NewSystemMetadataResp(m *database.SystemMetadata) *SystemMetadataResp {
	return &SystemMetadataResp{
		ID:           m.ID,
		SystemName:   m.SystemName,
		MetadataType: m.MetadataType,
		ServiceName:  m.ServiceName,
		Data:         json.RawMessage(m.Data),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
