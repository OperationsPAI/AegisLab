package handlers

import (
	"dagger/rcabench/executor"
	"encoding/json"
	"errors"
	"sync"

	"github.com/CUHK-SE-Group/chaos-experiment/handler"
)

type EvaluationMetric func([]Execution) ([]*Conclusion, error)

var (
	metrics = map[string]EvaluationMetric{
		"AC@k": accuracyk,
	}
	metricsMu sync.RWMutex
)

func RegisterMetric(name string, metric EvaluationMetric) error {
	metricsMu.Lock()
	defer metricsMu.Unlock()

	if name == "" {
		return errors.New("metric name cannot be empty")
	}
	if metric == nil {
		return errors.New("metric cannot be nil")
	}
	if _, exists := metrics[name]; exists {
		return errors.New("metric already registered")
	}

	metrics[name] = metric
	return nil
}

func GetMetrics() map[string]EvaluationMetric {
	metricsMu.RLock()
	defer metricsMu.RUnlock()

	copy := make(map[string]EvaluationMetric, len(metrics))
	for k, v := range metrics {
		copy[k] = v
	}
	return copy
}

func accuracyk(exe []Execution) ([]*Conclusion, error) {
	levelGran := make(map[string]map[string]*ConclusionACatK, 0)

	if len(exe) == 0 {
		return nil, errors.New("there is no execution history")
	}

	// initialize the level. We suppose that the algorithm will output the same level
	for _, g := range exe[0].GranularityResult {
		levelGran[g.Level] = make(map[string]*ConclusionACatK, 0)
		levelGran[g.Level]["AC@1"].Metric = "AC@1"
		levelGran[g.Level]["AC@1"].Level = g.Level
		levelGran[g.Level]["AC@3"].Metric = "AC@3"
		levelGran[g.Level]["AC@3"].Level = g.Level
		levelGran[g.Level]["AC@5"].Metric = "AC@5"
		levelGran[g.Level]["AC@5"].Level = g.Level
	}

	for _, e := range exe {
		for _, g := range e.GranularityResult {
			var payload map[string]interface{}
			err := json.Unmarshal([]byte(e.Dataset.Config), &payload)
			if err != nil {
				return nil, err
			}
			conf, err := executor.ParseFaultInjectionPayload(payload)
			if err != nil {
				return nil, err
			}
			groundtruth := []handler.Groudtruth{
				{
					Level: handler.Service,
					Name:  conf.Pod,
				},
				{
					Level: handler.Pod,
					Name:  conf.Pod,
				},
			}
			if additionalGroundtruth := handler.ChaosHandlers[handler.ChaosType(conf.FaultType)].GetGroudtruth(); additionalGroundtruth != nil {
				groundtruth = append(groundtruth, additionalGroundtruth...)
			}

			for _, gt := range groundtruth {
				if gt.Level == handler.Level(g.Level) {
					// 在 ground truth 中找到属于这个 granularity level 的
					if g.Result == gt.Name {
						acLevels := []string{"AC@1", "AC@3", "AC@5"}
						hits := []int{0, 0, 0} // 默认情况下，一个都没有命中
						switch {
						case g.Rank == 1:
							hits = []int{1, 1, 1} // a@1 a@3 和 a@5都命中
						case g.Rank <= 3:
							hits = []int{0, 1, 1}
						case g.Rank <= 5:
							hits = []int{0, 0, 1} // 只命中a@5
						}

						for i, ac := range acLevels {
							levelGran[g.Level][ac].Hit = append(levelGran[g.Level][ac].Hit, hits[i])
						}
					}
				}
			}
		}
	}
	result := make([]*Conclusion, 0)
	for _, acmap := range levelGran {
		for _, value := range acmap {
			count := 0
			for _, v := range value.Hit {
				if v == 1 {
					count += 1
				}
			}
			result = append(result, &Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   float64(count) / float64(len(value.Hit)),
			})
		}
	}
	return result, nil
}
