package task

import (
	"aegis/consts"
	"aegis/dto"
	"time"
)

// WSLogMessage is the WebSocket payload for task log streaming.
type WSLogMessage struct {
	Type    consts.WSLogType `json:"type"`
	Logs    []dto.LogEntry   `json:"logs,omitempty"`
	Message string           `json:"message,omitempty"`
	Total   int              `json:"total,omitempty"`
}

// TaskLogPollResp represents one task log poll batch for remote websocket forwarding.
type TaskLogPollResp struct {
	Logs      []dto.LogEntry `json:"logs"`
	Terminal  bool           `json:"terminal"`
	State     string         `json:"state"`
	CreatedAt time.Time      `json:"created_at"`
}
