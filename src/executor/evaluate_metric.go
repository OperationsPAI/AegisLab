package executor

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"maps"

	"aegis/consts"
	"aegis/dto"
	"aegis/repository"
	"aegis/utils"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
)

type conclusionACatK struct {
	Level  string `json:"level"`  // For example service level
	Metric string `json:"metric"` // For example topk
	Hit    []int  `json:"hit"`
}

type conclusionPrecisionk struct {
	Metric string
	Level  string
	Sum    float64
	Count  int
}

type conclusionAvgk struct {
	Metric string
	Level  string
	Sum    float64
	Count  int
}

type conclusionMAPk struct {
	Metric string
	Level  string
	Sum    float64
	Count  int
}

type conclustionMRR struct {
	Metric string
	Level  string
	Sum    float64
	Count  int
}

type EvaluationPayload struct {
	Level string
	Label string
}

var (
	metrics   = make(map[string]dto.EvaluateMetric)
	metricsMu sync.RWMutex
)

func init() {
	utils.Must(registerMetric("AC@k", accuracyk))
	utils.Must(registerMetric("PR@k", precisionk))
	utils.Must(registerMetric("Avg@k", avgk))
	utils.Must(registerMetric("MAP@k", mapk))
	utils.Must(registerMetric("MRR", mrr))
}

// Register a new evaluation metric
func registerMetric(name string, metric dto.EvaluateMetric) error {
	if name == "" {
		return errors.New("metric name cannot be empty")
	}
	if metric == nil {
		return errors.New("metric cannot be nil")
	}

	metricsMu.Lock()
	defer metricsMu.Unlock()

	if _, exists := metrics[name]; exists {
		return errors.New("metric already registered")
	}

	metrics[name] = metric
	return nil
}

// Get all registered evaluation metrics
func GetMetrics() map[string]dto.EvaluateMetric {
	metricsMu.RLock()
	defer metricsMu.RUnlock()

	copiedMetrics := make(map[string]dto.EvaluateMetric, len(metrics))
	maps.Copy(copiedMetrics, metrics)
	return copiedMetrics
}

// AC@k evaluation logic
func accuracyk(executions []dto.Execution) ([]dto.Conclusion, error) {
	if len(executions) == 0 {
		return nil, errors.New("execution history is empty")
	}

	ks := []int{1, 3, 5}
	levelGran := make(map[string]map[string]*conclusionACatK)

	// Initialize all possible granularity levels and corresponding AC@k
	for _, execution := range executions {
		for _, g := range execution.GranularityRecords {
			if _, exists := levelGran[g.Level]; !exists {
				levelGran[g.Level] = make(map[string]*conclusionACatK)
				for _, k := range ks {
					metric := fmt.Sprintf("AC@%d", k)
					levelGran[g.Level][metric] = &conclusionACatK{
						Metric: metric,
						Level:  g.Level,
					}
				}
			}
		}
	}

	// Process each execution record
	for _, execution := range executions {
		payload, err := parseEvaluationPayload(execution.Dataset.Param)
		if err != nil {
			return nil, fmt.Errorf("failed to get accuracy: %v", err)
		}

		for _, g := range execution.GranularityRecords {
			if g.Result == payload.Label {
				hitLevels := []string{"AC@1", "AC@3", "AC@5"}
				hits := []int{0, 0, 0}

				switch {
				case g.Rank == 1:
					hits = []int{1, 1, 1} // AC@1, AC@3, AC@5 all hit
				case g.Rank <= 3:
					hits = []int{0, 1, 1} // AC@3 and AC@5 hit
				case g.Rank <= 5:
					hits = []int{0, 0, 1} // Only AC@5 hit
				}

				for i, level := range hitLevels {
					levelGran[g.Level][level].Hit = append(levelGran[g.Level][level].Hit, hits[i])
				}
			}
		}
	}

	// Generate evaluation results
	var results []dto.Conclusion
	for _, acMap := range levelGran {
		for _, value := range acMap {
			hitCount := 0
			for _, v := range value.Hit {
				if v == 1 {
					hitCount++
				}
			}

			// If not hit, hit rate is 0
			rate := 0.0
			if len(value.Hit) != 0 {
				rate = float64(hitCount) / float64(len(value.Hit))
			}

			results = append(results, dto.Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   rate,
			})
		}
	}

	return results, nil
}

// Precision@k evaluation logic
func precisionk(executions []dto.Execution) ([]dto.Conclusion, error) {
	if len(executions) == 0 {
		return nil, errors.New("execution history is empty")
	}

	ks := []int{1, 3, 5}
	levelGran := make(map[string]map[string]*conclusionPrecisionk)

	// Initialize all possible granularity levels and corresponding PR@k
	indexLevel := make(map[int]string)
	for idx, execution := range executions {
		for _, g := range execution.GranularityRecords {
			if _, exists := indexLevel[idx]; !exists {
				indexLevel[idx] = g.Level
			}

			if _, exists := levelGran[g.Level]; !exists {
				levelGran[g.Level] = make(map[string]*conclusionPrecisionk)
				for _, k := range ks {
					metric := fmt.Sprintf("PR@%d", k)
					levelGran[g.Level][metric] = &conclusionPrecisionk{
						Metric: metric,
						Level:  g.Level,
						Sum:    0.0,
						Count:  0,
					}
				}
			}
		}
	}

	// Process each execution record
	for idx, execution := range executions {
		payload, err := parseEvaluationPayload(execution.Dataset.Param)
		if err != nil {
			return nil, fmt.Errorf("failed to get precesion: %v", err)
		}

		for _, k := range ks {
			minK := k
			if minK == 0 {
				continue // Avoid division by zero
			}

			// Get top k predictions
			topK := execution.GranularityRecords
			if len(execution.GranularityRecords) > k {
				topK = execution.GranularityRecords[:k]
			}

			// Count hits
			hits := 0
			for _, pred := range topK {
				if pred.Result == payload.Label {
					hits++
				}
			}

			// Calculate PR@k
			precision := float64(hits) / float64(minK)

			// Accumulate to corresponding conclusion
			level := indexLevel[idx]
			metric := fmt.Sprintf("PR@%d", k)
			if conclusion, exists := levelGran[level][metric]; exists {
				conclusion.Sum += precision
				conclusion.Count++
			}
		}
	}

	// Generate evaluation results
	var results []dto.Conclusion
	for _, prMap := range levelGran {
		for _, value := range prMap {
			if value.Count == 0 {
				results = append(results, dto.Conclusion{
					Level:  value.Level,
					Metric: value.Metric,
					Rate:   0.0, // Or set to other default value as needed
				})
			} else {
				results = append(results, dto.Conclusion{
					Level:  value.Level,
					Metric: value.Metric,
					Rate:   value.Sum / float64(value.Count),
				})
			}
		}
	}

	return results, nil
}

// Avg@k evaluation logic
func avgk(executions []dto.Execution) ([]dto.Conclusion, error) {
	if len(executions) == 0 {
		return nil, errors.New("execution history is empty")
	}

	ks := []int{1, 3, 5}
	K := len(ks)
	levelGran := make(map[string]*conclusionAvgk)

	// Initialize all possible granularity levels and corresponding Avg@k
	indexLevel := make(map[int]string)
	for idx, execution := range executions {
		for _, g := range execution.GranularityRecords {
			if _, exists := indexLevel[idx]; !exists {
				indexLevel[idx] = g.Level
			}

			if _, exists := levelGran[g.Level]; !exists {
				levelGran[g.Level] = &conclusionAvgk{
					Metric: "Avg@k",
					Level:  g.Level,
					Sum:    0.0,
					Count:  0,
				}
			}
		}
	}

	// Process each execution record
	for idx, execution := range executions {
		payload, err := parseEvaluationPayload(execution.Dataset.Param)
		if err != nil {
			return nil, fmt.Errorf("failed to get average: %v", err)
		}

		// Calculate A@k and accumulate
		var sumAvgk float64
		for _, k := range ks {
			minK := min(len(execution.GranularityRecords), k)
			if minK == 0 {
				continue // Avoid division by zero
			}

			// Calculate A@k
			totalRank := 0
			count := 0
			for i := range minK {
				pred := execution.GranularityRecords[i]
				if pred.Result == payload.Label {
					totalRank += pred.Rank
					count++
				}
			}

			if count == 0 {
				continue // Avoid division by zero
			}

			avgk := float64(totalRank) / float64(count)
			sumAvgk += avgk
		}

		// Calculate Avg@K = sum(A@k) / K
		avgK := sumAvgk / float64(K)
		level := indexLevel[idx]
		levelGran[level].Sum += avgK
		levelGran[level].Count++
	}

	// Generate evaluation results
	var results []dto.Conclusion
	for _, value := range levelGran {
		if value.Count == 0 {
			results = append(results, dto.Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   0.0, // Or set to other default value as needed
			})
		} else {
			avgK := value.Sum / float64(value.Count)
			results = append(results, dto.Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   avgK,
			})
		}
	}

	return results, nil
}

// MAP@k evaluation logic
func mapk(executions []dto.Execution) ([]dto.Conclusion, error) {
	if len(executions) == 0 {
		return nil, errors.New("execution history is empty")
	}

	ks := []int{1, 3, 5}
	K := len(ks)
	levelGran := make(map[string]*conclusionMAPk)

	// Initialize all possible granularity levels and corresponding MAP@k
	indexLevel := make(map[int]string)
	for idx, execution := range executions {
		for _, g := range execution.GranularityRecords {
			if _, exists := indexLevel[idx]; !exists {
				indexLevel[idx] = g.Level
			}

			if _, exists := levelGran[g.Level]; !exists {
				levelGran[g.Level] = &conclusionMAPk{
					Metric: "MAP@k",
					Level:  g.Level,
					Sum:    0.0,
					Count:  0,
				}
			}
		}
	}

	// Process each execution record
	for idx, execution := range executions {
		payload, err := parseEvaluationPayload(execution.Dataset.Param)
		if err != nil {
			return nil, fmt.Errorf("failed to get precesion: %v", err)
		}

		// Calculate PR@k and accumulate
		for _, k := range ks {
			minK := min(len(execution.GranularityRecords), k)
			if minK == 0 {
				continue // Avoid division by zero
			}

			hits := 0
			sumPrecision := 0.0

			for i := range minK {
				pred := execution.GranularityRecords[i]
				if pred.Result == payload.Label {
					hits++
					sumPrecision += float64(hits) / float64(i+1)
				}
			}

			// Calculate PR@k and accumulate
			PRk := sumPrecision / float64(K)
			level := indexLevel[idx]
			levelGran[level].Sum += PRk
			levelGran[level].Count++
		}
	}

	// Generate evaluation results
	var results []dto.Conclusion
	for _, value := range levelGran {
		if value.Count == 0 {
			results = append(results, dto.Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   0.0, // Or set to other default value as needed
			})
		} else {
			mapkValue := value.Sum / float64(value.Count)
			results = append(results, dto.Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   mapkValue,
			})
		}
	}

	return results, nil
}

func mrr(executions []dto.Execution) ([]dto.Conclusion, error) {
	if len(executions) == 0 {
		return nil, errors.New("execution history is empty")
	}

	levelGran := make(map[string]*conclustionMRR)

	indexLevel := make(map[int]string)
	for idx, exexecution := range executions {
		for _, g := range exexecution.GranularityRecords {
			if _, exists := indexLevel[idx]; !exists {
				indexLevel[idx] = g.Level
			}

			if _, exists := levelGran[g.Level]; !exists {
				levelGran[g.Level] = &conclustionMRR{
					Metric: "MRR",
					Level:  g.Level,
					Sum:    0.0,
					Count:  0,
				}
			}
		}
	}

	for idx, execution := range executions {
		payload, err := parseEvaluationPayload(execution.Dataset.Param)
		if err != nil {
			return nil, fmt.Errorf("failed to get precesion: %v", err)
		}

		for _, pred := range execution.GranularityRecords {
			if pred.Result == payload.Label {
				level := indexLevel[idx]
				if conclusion, exists := levelGran[level]; exists {
					precision := 0.0
					if pred.Rank != 0 {
						precision = 1.0 / float64(pred.Rank)
					}

					conclusion.Sum += precision
					conclusion.Count++
				}
			}
		}
	}

	// Generate evaluation results
	var results []dto.Conclusion
	for _, value := range levelGran {
		if value.Count == 0 {
			results = append(results, dto.Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   0.0,
			})
		} else {
			results = append(results, dto.Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   value.Sum / float64(value.Count),
			})
		}
	}

	return results, nil
}

func parseEvaluationPayload(payload map[string]any) (*EvaluationPayload, error) {
	message := "missing or invalid '%s' key in payload"

	label, ok := payload[consts.EvaluateLabel].(string)
	if !ok || label == "" {
		return nil, fmt.Errorf(message, consts.EvaluateLabel)
	}

	return &EvaluationPayload{
		Label: label,
	}, nil
}

func ParseConfigAndGetGroundTruthMap(executions []dto.Execution) ([]map[string]map[string]struct{}, error) {
	names := make([]string, 0, len(executions))
	for _, execution := range executions {
		names = append(names, execution.Dataset.Name)
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no dataset names found in executions")
	}
	return GetGTByDatasetName(names)
}

func GetGTByDatasetName(names []string) ([]map[string]map[string]struct{}, error) {
	configs, err := repository.ListEngineConfigsByNames(names)
	if err != nil {
		return nil, err
	}

	groundtruths := make([]chaos.Groundtruth, 0, len(configs))
	for _, config := range configs {
		var m map[string]any
		if err := json.Unmarshal([]byte(config), &m); err != nil {
			return nil, err
		}

		node, err := chaos.MapToNode(m)
		if err != nil {
			return nil, err
		}

		injectionConf, err := chaos.NodeToStruct[chaos.InjectionConf](node)
		if err != nil {
			return nil, err
		}

		gt, err := injectionConf.GetGroundtruth()
		if err != nil {
			return nil, err
		}

		groundtruths = append(groundtruths, gt)
	}

	groundtruthMaps := make([]map[string]map[string]struct{}, 0, len(groundtruths))
	for idx, gt := range groundtruths {
		t := reflect.TypeOf(gt)
		v := reflect.ValueOf(gt)

		groundtruthMap := make(map[string]map[string]struct{}, t.NumField())
		for i := range t.NumField() {
			field := t.Field(i)
			value := v.Field(i)
			if value.Kind() == reflect.Slice && value.Type().Elem().Kind() == reflect.String {
				if value.IsNil() || value.Len() == 0 {
					return nil, fmt.Errorf("the value of field %s in groundtruth %d is not valid slice of string",
						field.Name, idx)
				}
			}

			for j := range value.Len() {
				elem, ok := value.Index(j).Interface().(string)
				if !ok {
					return nil, fmt.Errorf("failed to read string[%d] in field %s in groundtruth %d", j, field.Name, idx)
				}

				if _, exists := groundtruthMap[field.Name]; !exists {
					groundtruthMap[field.Name] = make(map[string]struct{})
				}

				groundtruthMap[field.Name][elem] = struct{}{}
			}
		}

		groundtruthMaps = append(groundtruthMaps, groundtruthMap)
	}

	return groundtruthMaps, nil
}
