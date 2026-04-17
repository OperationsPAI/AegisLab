package dto

import (
	"fmt"

	"aegis/utils"
)

// DatasetRef is the shared dataset reference used across modules and tasks.
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
