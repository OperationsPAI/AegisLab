package dto

import (
	"time"

	"rcabench/database"
)

type ProjectV2GetReq struct {
	IncludeContainers bool `form:"include_containers"` // Include related containers
	IncludeDatasets   bool `form:"include_datasets"`   // Include related datasets
	IncludeInjections bool `form:"include_injections"` // Include related fault injections
	IncludeLabels     bool `form:"include_labels"`     // Include related labels
}

type ProjectV2Response struct {
	ID          int    `json:"id"`          // Unique identifier
	Name        string `json:"name"`        // Project name
	Description string `json:"description"` // Project description

	Status    int       `json:"status"`     // Status
	IsPublic  bool      `json:"is_public"`  // Whether public
	CreatedAt time.Time `json:"created_at"` // Creation time
	UpdatedAt time.Time `json:"updated_at"` // Update time

	Containers []database.Container  `json:"containers,omitempty"` // Associated containers
	Datasets   []database.Dataset    `json:"datasets,omitempty"`   // Associated datasets
	Injections []InjectionV2Response `json:"injections,omitempty"` // Associated fault injections
	Labels     []database.Label      `json:"labels,omitempty"`     // Associated labels
}

// ProjectV2ContainerRelationResponse Project container relation response
type ProjectV2ContainerRelationResponse struct {
	ID          int                 `json:"id"`                  // Relation ID
	ContainerID int                 `json:"container_id"`        // Container ID
	CreatedAt   time.Time           `json:"created_at"`          // Creation time
	UpdatedAt   time.Time           `json:"updated_at"`          // Update time
	Container   *database.Container `json:"container,omitempty"` // Container details
}

// ProjectV2InjectionRelationResponse Project fault injection relation response
type ProjectV2InjectionRelationResponse struct {
	ID               int                              `json:"id"`                        // Relation ID
	FaultInjectionID int                              `json:"fault_injection_id"`        // Fault injection ID
	CreatedAt        time.Time                        `json:"created_at"`                // Creation time
	UpdatedAt        time.Time                        `json:"updated_at"`                // Update time
	FaultInjection   *database.FaultInjectionSchedule `json:"fault_injection,omitempty"` // Fault injection details
}

func ToProjectV2Response(project *database.Project, includeRelations bool) *ProjectV2Response {
	resp := &ProjectV2Response{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Status:      project.Status,
		IsPublic:    project.IsPublic,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}

	return resp
}
