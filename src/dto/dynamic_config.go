package dto

import (
	"encoding/json"
	"fmt"
	"time"

	"aegis/utils"
)

// ConfigUpdateResponse represents the response to a configuration update event
type ConfigUpdateResponse struct {
	ID          string    `json:"id"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
	ProcessedAt time.Time `json:"processed_at"`
	Payload     any       `json:"payload,omitempty"`
}

func NewConfigUpdateResponse() *ConfigUpdateResponse {
	return &ConfigUpdateResponse{
		ID:          utils.GenerateULID(nil),
		Success:     false,
		ProcessedAt: time.Now(),
	}
}

func (r *ConfigUpdateResponse) ToMap() (map[string]any, error) {
	m := map[string]any{
		"id":           r.ID,
		"success":      r.Success,
		"processed_at": r.ProcessedAt.Format(time.RFC3339),
	}

	if r.Error != "" {
		m["error"] = r.Error
	}
	if r.Payload != nil {
		payloadStr, err := json.Marshal(r.Payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload to JSON: %w", err)
		}
		m["payload"] = string(payloadStr)
	}

	return m, nil
}
