package dto

import (
	"time"

	chaos "github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/executor"
)

type InjectCancelResp struct {
}

type InjectionDetailResp struct {
	Task InjectionTask `json:"task"`
	Logs []string      `json:"logs"`
}

type InjectionItem struct {
	ID              int                            `json:"id"`
	TaskID          string                         `json:"task_id"`
	FaultType       string                         `json:"fault_type"`
	Name            string                         `gorm:"column:injection_name" json:"name"`
	Status          string                         `json:"status"`
	InjectTime      time.Time                      `gorm:"column:start_time" json:"inject_time"`
	ProposedEndTime time.Time                      `json:"proposed_end_time"`
	Duration        int                            `json:"duration"`
	Payload         executor.FaultInjectionPayload `gorm:"-" json:"payload"`
}

type InjectionListReq struct {
	PaginationReq
}

type InjectionNamespaceInfoResp struct {
	NamespaceInfo map[string][]string `json:"namespace_info"`
}

type InjectionParaResp struct {
	Specification map[string][]chaos.ActionSpace `json:"specification"`
	KeyMap        map[chaos.ChaosType]string     `json:"keymap"`
}

type InjectionTask struct {
	ID        string                         `json:"id"`
	Type      string                         `json:"type"`
	Payload   executor.FaultInjectionPayload `json:"payload"`
	Status    string                         `json:"status"`
	CreatedAt time.Time                      `json:"created_at"`
}
