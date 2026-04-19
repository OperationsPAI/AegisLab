package dto

import (
	"fmt"

	"aegis/consts"
)

// CheckPermissionParams represents permission check parameters
type CheckPermissionParams struct {
	UserID       int                  `json:"user_id"`
	Action       consts.ActionName    `json:"action"`
	Scope        consts.ResourceScope `json:"scope"`
	ResourceName consts.ResourceName  `json:"resource_name"`
	TeamID       *int                 `json:"team_id"`
	ProjectID    *int                 `json:"project_id"`
	ContainerID  *int                 `json:"container_id"`
	DatasetID    *int                 `json:"dataset_id"`
}

func (req *CheckPermissionParams) Validate() error {
	if req.UserID <= 0 {
		return fmt.Errorf("user_id must be positive")
	}
	if req.Action == "" {
		return fmt.Errorf("action cannot be empty")
	}
	if req.ResourceName == "" {
		return fmt.Errorf("resource_name cannot be empty")
	}
	return nil
}
