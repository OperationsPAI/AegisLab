package injection

import (
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

// RuntimeCreateInjectionReq captures fault injection writes initiated by runtime-worker-service.
type RuntimeCreateInjectionReq struct {
	Name              string               `json:"name"`
	FaultType         chaos.ChaosType      `json:"fault_type"`
	Category          chaos.SystemType     `json:"category"`
	Description       string               `json:"description"`
	DisplayConfig     string               `json:"display_config"`
	EngineConfig      string               `json:"engine_config"`
	Groundtruths      []model.Groundtruth  `json:"groundtruths"`
	GroundtruthSource string               `json:"groundtruth_source"`
	PreDuration       int                  `json:"pre_duration"`
	TaskID            string               `json:"task_id"`
	BenchmarkID       *int                 `json:"benchmark_id,omitempty"`
	PedestalID        *int                 `json:"pedestal_id,omitempty"`
	Labels            []dto.LabelItem      `json:"labels,omitempty"`
	State             consts.DatapackState `json:"state"`
}

// RuntimeUpdateInjectionStateReq captures datapack state mutations initiated by runtime-worker-service.
type RuntimeUpdateInjectionStateReq struct {
	Name  string               `json:"name"`
	State consts.DatapackState `json:"state"`
}

// RuntimeUpdateInjectionTimestampReq captures datapack timestamp updates initiated by runtime-worker-service.
type RuntimeUpdateInjectionTimestampReq struct {
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}
