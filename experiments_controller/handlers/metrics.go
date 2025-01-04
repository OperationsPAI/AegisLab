package handlers

import (
	"dagger/rcabench/executor"
	"encoding/json"
	"errors"
	"sync"

	"github.com/CUHK-SE-Group/chaos-experiment/handler"
)

// 定义评估函数类型
type EvaluationMetric func([]Execution) ([]*Conclusion, error)

var (
	metrics   = make(map[string]EvaluationMetric)
	metricsMu sync.RWMutex
)

func init() {
	RegisterMetric("AC@k", accuracyk)
}

// 注册新的评估指标
func RegisterMetric(name string, metric EvaluationMetric) error {
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

// 获取所有已注册的评估指标
func GetMetrics() map[string]EvaluationMetric {
	metricsMu.RLock()
	defer metricsMu.RUnlock()

	copiedMetrics := make(map[string]EvaluationMetric, len(metrics))
	for k, v := range metrics {
		copiedMetrics[k] = v
	}
	return copiedMetrics
}

// AC@k 评估逻辑
func accuracyk(executions []Execution) ([]*Conclusion, error) {
	if len(executions) == 0 {
		return nil, errors.New("execution history is empty")
	}

	levelGran := make(map[string]map[string]*ConclusionACatK)

	// 初始化级别
	for _, g := range executions[0].GranularityResult {
		levelGran[g.Level] = map[string]*ConclusionACatK{
			"AC@1": {Metric: "AC@1", Level: g.Level},
			"AC@3": {Metric: "AC@3", Level: g.Level},
			"AC@5": {Metric: "AC@5", Level: g.Level},
		}
	}

	// 处理每个执行记录
	for _, execution := range executions {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(execution.Dataset.Config), &payload); err != nil {
			return nil, err
		}

		conf, err := executor.ParseFaultInjectionPayload(payload)
		if err != nil {
			return nil, err
		}

		groundtruth := []handler.Groudtruth{
			{Level: handler.Service, Name: conf.Pod},
			{Level: handler.Pod, Name: conf.Pod},
		}

		if additional := handler.ChaosHandlers[handler.ChaosType(conf.FaultType)].GetGroudtruth(); additional != nil {
			groundtruth = append(groundtruth, additional...)
		}

		for _, g := range execution.GranularityResult {
			for _, gt := range groundtruth {
				if gt.Level == handler.Level(g.Level) && g.Result == gt.Name {
					hitLevels := []string{"AC@1", "AC@3", "AC@5"}
					hits := []int{0, 0, 0}

					switch {
					case g.Rank == 1:
						hits = []int{1, 1, 1} // AC@1, AC@3, AC@5 全部命中
					case g.Rank <= 3:
						hits = []int{0, 1, 1} // AC@3 和 AC@5 命中
					case g.Rank <= 5:
						hits = []int{0, 0, 1} // 仅命中 AC@5
					}

					for i, level := range hitLevels {
						levelGran[g.Level][level].Hit = append(levelGran[g.Level][level].Hit, hits[i])
					}
				}
			}
		}
	}

	// 生成评估结果
	var results []*Conclusion
	for _, acMap := range levelGran {
		for _, value := range acMap {
			hitCount := 0
			for _, v := range value.Hit {
				if v == 1 {
					hitCount++
				}
			}
			results = append(results, &Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   float64(hitCount) / float64(len(value.Hit)),
			})
		}
	}
	return results, nil
}
