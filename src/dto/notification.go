package dto

import "time"

type NotificationEvent struct {
	Type      string    `json:"type"`
	EntityID  string    `json:"entity_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status,omitempty"`
}

type GetNotificationStreamReq struct {
	LastID string `form:"last_id" json:"last_id"`
}

func (r *GetNotificationStreamReq) Validate() error {
	return nil
}
