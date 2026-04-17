package dto

import (
	"time"

	"aegis/model"
)

// InjectionItem is the shared runtime datapack payload carried across tasks/consumers.
type InjectionItem struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	PreDuration int       `json:"pre_duration"`
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
}

func NewInjectionItem(injection *model.FaultInjection) InjectionItem {
	item := InjectionItem{
		ID:          injection.ID,
		Name:        injection.Name,
		PreDuration: injection.PreDuration,
	}

	if injection.StartTime != nil {
		item.StartTime = *injection.StartTime
	}
	if injection.EndTime != nil {
		item.EndTime = *injection.EndTime
	}

	return item
}
