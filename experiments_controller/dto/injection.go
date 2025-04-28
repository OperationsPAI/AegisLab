package dto

import (
	"encoding/json"
	"fmt"
	"sort"
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
	Namespace string `form:"namespace" binding:"required"`
	Mode      string `form:"mode" binding:"oneof=display engine"`
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
	Index         int
	FaultType     int
	FaultDuration int
	DisplayData   string
	Conf          *chaos.InjectionConf
	Node          *chaos.Node
	ExecuteTime   time.Time
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

		childNode, exists := node.Children[strconv.Itoa(node.Value)]
		if !exists {
			return nil, fmt.Errorf("failed to find key %d in the children", node.Value)
		}

		m := config.GetMap("injection.namespace_target_map")
		namespacePrefixs := make([]string, 0, len(m))
		namespaceTargetMap := make(map[string]int, len(m))
		for ns, value := range m {
			count, _ := value.(int64)
			namespaceTargetMap[ns] = int(count)
			namespacePrefixs = append(namespacePrefixs, ns)
		}

		sort.Strings(namespacePrefixs)
		index := childNode.Children[consts.NamespaceNodeKey].Value
		namespaceCount := namespaceTargetMap[namespacePrefixs[index]]

		var execTime time.Time
		if idx < namespaceCount {
			execTime = currentTime
		} else {
			execTime = currentTime.Add(intervalDuration * time.Duration(idx/namespaceCount)).Add(consts.DefaultTimeUnit)
		}

		faultDuration := childNode.Children[consts.DurationNodeKey].Value
		if !config.GetBool("debugging.enable") {
			if idx%namespaceCount == 0 {
				start := execTime.Add(-preDuration)
				end := execTime.Add(time.Duration(faultDuration) * consts.DefaultTimeUnit)

				if idx > 0 && !start.After(prevEnd) {
					return nil, fmt.Errorf("spec[%d] time conflict", idx)
				}

				prevEnd = end
			}
		}

		conf, err := chaos.NodeToStruct[chaos.InjectionConf](node)
		if err != nil {
			return nil, fmt.Errorf("failed to convert node to injecton conf: %v", err)
		}

		displayConfig, err := conf.GetDisplayConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get display config: %v", err)
		}

		displayData, err := json.Marshal(displayConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal injection spec to display config: %v", err)
		}

		configs = append(configs, &InjectionConfig{
			Index:         idx,
			FaultType:     node.Value,
			FaultDuration: faultDuration,
			DisplayData:   string(displayData),
			Conf:          conf,
			Node:          node,
			ExecuteTime:   execTime,
		})
	}

	return configs, nil
}
