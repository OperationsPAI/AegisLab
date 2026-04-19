package injection

import (
	"aegis/consts"
	"aegis/dto"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
	"github.com/OperationsPAI/chaos-experiment/pkg/guidedcli"
	"github.com/sirupsen/logrus"
)

type injectionProcessItem struct {
	index         int
	faultDuration int
	nodes         []chaos.Node // legacy Node-DSL batch
	// guidedConfigs is populated when the submission came through the guided
	// CLI path. Mutually exclusive with nodes: exactly one is set per item.
	guidedConfigs []guidedcli.GuidedConfig
	executeTime   time.Time
}

func parseBatchInjectionSpecs(pedestal string, batchIndex int, specs []chaos.Node) (*injectionProcessItem, string, error) {
	if len(specs) == 0 {
		return nil, "", fmt.Errorf("empty fault injection batch at index %d", batchIndex)
	}

	maxDuration := 0
	nodes := make([]chaos.Node, 0, len(specs))
	for idx, spec := range specs {
		childNode, exists := spec.Children[strconv.Itoa(spec.Value)]
		if !exists {
			return nil, "", fmt.Errorf("failed to find key %d in the children at index %d", spec.Value, idx)
		}
		if len(childNode.Children) < 3 {
			return nil, "", fmt.Errorf("no child nodes found for fault spec at index %d", idx)
		}

		faultDuration := childNode.Children[consts.DurationNodeKey].Value
		if faultDuration > maxDuration {
			maxDuration = faultDuration
		}

		systemIdx := childNode.Children[consts.SystemNodeKey].Value
		system := chaos.GetAllSystemTypes()[systemIdx]
		if pedestal != system.String() {
			return nil, "", fmt.Errorf("mismatched system type %s for pedestal %s at index %d", system.String(), pedestal, idx)
		}

		nodes = append(nodes, spec)
	}

	uniqueServices := make(map[string]int, len(nodes))
	var duplicateServiceWarnings []string
	ctx := context.Background()
	for idx, node := range nodes {
		conf, err := chaos.NodeToStruct[chaos.InjectionConf](ctx, &node)
		if err != nil {
			return nil, "", fmt.Errorf("failed to convert node to InjectionConf at index %d: %w", idx, err)
		}

		groundtruth, err := conf.GetGroundtruth(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get groundtruth from InjectionConf at index %d: %w", idx, err)
		}

		for _, service := range groundtruth.Service {
			if service == "" {
				continue
			}
			if oldIdx, exists := uniqueServices[service]; exists {
				duplicateServiceWarnings = append(duplicateServiceWarnings, fmt.Sprintf("service '%s' at positions %d and %d", service, oldIdx, idx))
				continue
			}
			uniqueServices[service] = idx
		}
	}

	nodes = sortNodes(nodes)

	var warning string
	if len(duplicateServiceWarnings) > 0 {
		warning = fmt.Sprintf("Batch %d contains duplicate service injections: %s", batchIndex, strings.Join(duplicateServiceWarnings, "; "))
	}

	return &injectionProcessItem{
		index:         batchIndex,
		faultDuration: maxDuration,
		nodes:         nodes,
	}, warning, nil
}

func flattenYAMLToParameters(data map[string]any, prefix string) []dto.ParameterSpec {
	var params []dto.ParameterSpec
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]any:
			params = append(params, flattenYAMLToParameters(v, fullKey)...)
		case []any:
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				logrus.Warnf("Failed to marshal array for key %s: %v", fullKey, err)
				continue
			}
			params = append(params, dto.ParameterSpec{Key: fullKey, Value: string(jsonBytes)})
		default:
			params = append(params, dto.ParameterSpec{Key: fullKey, Value: v})
		}
	}
	return params
}

func (s *Service) removeDuplicated(items []injectionProcessItem) ([]injectionProcessItem, []int, []int, error) {
	engineConfigStrs := make([]string, len(items))
	for i, item := range items {
		var payload any
		switch {
		case len(item.guidedConfigs) > 0:
			payload = item.guidedConfigs
		case len(item.nodes) > 0:
			payload = item.nodes
		default:
			continue
		}

		b, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to marshal engine config at batch index %d: %w", i, err)
		}
		engineConfigStrs[i] = string(b)
	}

	orderedUniqueIdx := make([]int, 0, len(engineConfigStrs))
	seen := make(map[string]struct{}, len(engineConfigStrs))
	duplicatedInRequest := make([]int, 0)
	for i, key := range engineConfigStrs {
		if key == "" {
			orderedUniqueIdx = append(orderedUniqueIdx, i)
			continue
		}
		if _, ok := seen[key]; ok {
			duplicatedInRequest = append(duplicatedInRequest, items[i].index)
			continue
		}
		seen[key] = struct{}{}
		orderedUniqueIdx = append(orderedUniqueIdx, i)
	}

	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}

	existed := make(map[string]struct{})
	for start := 0; start < len(keys); start += 100 {
		end := min(start+100, len(keys))
		existing, err := s.repo.listExistingEngineConfigs(keys[start:end])
		if err != nil {
			return nil, nil, nil, err
		}
		for _, v := range existing {
			existed[v] = struct{}{}
		}
	}

	out := make([]injectionProcessItem, 0, len(orderedUniqueIdx))
	alreadyExisted := make([]int, 0)
	for _, idx := range orderedUniqueIdx {
		key := engineConfigStrs[idx]
		if key != "" {
			if _, ok := existed[key]; ok {
				alreadyExisted = append(alreadyExisted, items[idx].index)
				continue
			}
		}

		items[idx].executeTime = time.Now().Add(time.Duration(idx*2) * time.Second)
		out = append(out, items[idx])
	}

	return out, duplicatedInRequest, alreadyExisted, nil
}

func sortNodes(nodes []chaos.Node) []chaos.Node {
	if len(nodes) <= 1 {
		return nodes
	}

	sortedNodes := make([]chaos.Node, len(nodes))
	copy(sortedNodes, nodes)
	for i := 0; i < len(sortedNodes)-1; i++ {
		for j := i + 1; j < len(sortedNodes); j++ {
			if sortedNodes[i].Value > sortedNodes[j].Value {
				sortedNodes[i], sortedNodes[j] = sortedNodes[j], sortedNodes[i]
				continue
			}
			if sortedNodes[i].Value == sortedNodes[j].Value {
				iJSON, _ := json.Marshal(sortedNodes[i])
				jJSON, _ := json.Marshal(sortedNodes[j])
				if string(iJSON) > string(jJSON) {
					sortedNodes[i], sortedNodes[j] = sortedNodes[j], sortedNodes[i]
				}
			}
		}
	}
	return sortedNodes
}

// parseBatchGuidedSpecs parses a single batch of GuidedConfig specs for
// parallel execution. Each GuidedConfig is resolved to an InjectionConf via
// guidedcli.BuildInjection solely to compute duration, system-type sanity
// check, and groundtruth-service dedup warnings. The returned item carries
// the original GuidedConfigs; the actual BuildInjection call at execute-time
// lives in the consumer.
func parseBatchGuidedSpecs(ctx context.Context, pedestal string, batchIndex int, configs []guidedcli.GuidedConfig) (*injectionProcessItem, string, error) {
	if len(configs) == 0 {
		return nil, "", fmt.Errorf("empty guided fault batch at index %d", batchIndex)
	}

	maxDuration := 0
	uniqueServices := make(map[string]int, len(configs))
	var duplicateServiceWarnings []string

	for idx, cfg := range configs {
		conf, systemType, err := guidedcli.BuildInjection(ctx, cfg)
		if err != nil {
			return nil, "", fmt.Errorf("failed to build injection from guided config at index %d: %w", idx, err)
		}
		if pedestal != systemType.String() {
			return nil, "", fmt.Errorf("mismatched system type %s for pedestal %s at index %d", systemType.String(), pedestal, idx)
		}

		duration := 0
		if cfg.Duration != nil {
			duration = *cfg.Duration
		}
		if duration > maxDuration {
			maxDuration = duration
		}

		groundtruth, err := conf.GetGroundtruth(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get groundtruth from guided config at index %d: %w", idx, err)
		}
		for _, service := range groundtruth.Service {
			if service == "" {
				continue
			}
			if oldIdx, exists := uniqueServices[service]; exists {
				duplicateServiceWarnings = append(duplicateServiceWarnings,
					fmt.Sprintf("service '%s' at positions %d and %d", service, oldIdx, idx))
				continue
			}
			uniqueServices[service] = idx
		}
	}

	var warning string
	if len(duplicateServiceWarnings) > 0 {
		warning = fmt.Sprintf("Batch %d contains duplicate service injections: %s",
			batchIndex, strings.Join(duplicateServiceWarnings, "; "))
	}

	return &injectionProcessItem{
		index:         batchIndex,
		faultDuration: maxDuration,
		guidedConfigs: configs,
	}, warning, nil
}
