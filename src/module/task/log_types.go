package taskmodule

import (
	"aegis/consts"
	"aegis/dto"
)

// WSLogMessage is the WebSocket payload for task log streaming.
type WSLogMessage struct {
	Type    consts.WSLogType `json:"type"`
	Logs    []dto.LogEntry   `json:"logs,omitempty"`
	Message string           `json:"message,omitempty"`
	Total   int              `json:"total,omitempty"`
}
