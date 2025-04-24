package dto

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	chaos "github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
)

type InjectCancelResp struct {
}

type InjectionConfReq struct {
	Mode string `form:"mode" binding:"oneof=display engine"`
}

type InjectionItem struct {
	ID        int            `json:"id"`
	TaskID    string         `json:"task_id"`
	FaultType string         `json:"fault_type"`
	Spec      map[string]any `json:"spec"`
	Status    string         `json:"status"`
	StartTime time.Time      `json:"start_time"`
	EndTime   time.Time      `json:"end_time"`
}

func (i *InjectionItem) Convert(record database.FaultInjectionSchedule) error {
	var config map[string]any
	if err := json.Unmarshal([]byte(record.DisplayConfig), &config); err != nil {
		return err
	}

	i.ID = record.ID
	i.TaskID = record.TaskID
	i.FaultType = chaos.ChaosTypeMap[chaos.ChaosType(record.FaultType)]
	i.Spec = config
	i.Status = DatasetStatusMap[record.Status]
	i.StartTime = record.StartTime
	i.EndTime = record.EndTime

	return nil
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

type InjectionSubmitReq struct {
	Interval    int              `json:"interval"`
	PreDuration int              `json:"pre_duration"`
	Specs       []map[string]any `json:"specs"`
	Benchmark   string           `json:"benchmark"`
}

type InjectionConfig struct {
	Index       int
	FaultType   int
	Conf        *chaos.InjectionConf
	RawConf     string
	ExecuteTime time.Time
}

func (r *InjectionSubmitReq) ParseInjectionSpecs() ([]*InjectionConfig, error) {
	if len(r.Specs) == 0 {
		return nil, fmt.Errorf("spec must not be blank")
	}

	intervalDuration := time.Duration(r.Interval) * consts.DefaultTimeUnit
	preDuration := time.Duration(r.PreDuration) * consts.DefaultTimeUnit

	currentTime := time.Now()
	prevEnd := currentTime
	configs := make([]*InjectionConfig, 0, len(r.Specs))
	for idx, spec := range r.Specs {
		node, err := chaos.MapToNode(spec)
		if err != nil {
			return nil, fmt.Errorf("failed to convert spec[%d] to node: %v", idx, err)
		}

		newSpec := chaos.NodeToMap(node, true)
		newSpecBytes, err := json.Marshal(newSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal spec[%d]: %v", idx, err)
		}

		execTime := currentTime.Add(intervalDuration * time.Duration(idx))
		if !config.GetBool("debugging.enable") {
			childNode, exists := node.Children[strconv.Itoa(node.Value)]
			if !exists {
				return nil, fmt.Errorf("failed to find key %d in the children", node.Value)
			}

			faultDuration := childNode.Children[strconv.Itoa(0)].Value
			start := execTime.Add(-preDuration)
			end := execTime.Add(time.Duration(faultDuration) * consts.DefaultTimeUnit)

			if idx > 0 && !start.After(prevEnd) {
				return nil, fmt.Errorf("spec[%d] time conflict", idx)
			}

			prevEnd = end
		}

		conf, err := chaos.NodeToStruct[chaos.InjectionConf](node)
		if err != nil {
			return nil, err
		}

		configs = append(configs, &InjectionConfig{
			Index:       idx,
			FaultType:   node.Value,
			Conf:        conf,
			RawConf:     string(newSpecBytes),
			ExecuteTime: execTime,
		})
	}

	return configs, nil
}
