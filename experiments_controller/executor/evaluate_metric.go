package executor

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"

	"maps"

	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/utils"
)

type conclusionACatK struct {
	Level  string `json:"level"`  // 例如 service level
	Metric string `json:"metric"` // 例如 topk
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

var (
	metrics   = make(map[string]dto.EvaluateMetric)
	metricsMu sync.RWMutex
)

func init() {
	utils.Must(RegisterMetric("A@k", accuracyk))
	utils.Must(RegisterMetric("PR@k", precisionk))
	utils.Must(RegisterMetric("Avg@k", avgk))
	utils.Must(RegisterMetric("MAP@k", mapk))
	utils.Must(RegisterMetric("MRR", mrr))
}

// 注册新的评估指标
func RegisterMetric(name string, metric dto.EvaluateMetric) error {
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
func GetMetrics() map[string]dto.EvaluateMetric {
	metricsMu.RLock()
	defer metricsMu.RUnlock()

	copiedMetrics := make(map[string]dto.EvaluateMetric, len(metrics))
	maps.Copy(copiedMetrics, metrics)
	return copiedMetrics
}

// 解析配置并获取 ground truth 的公共函数
func parseConfigAndGetGroundTruth(execution dto.Execution) ([]handler.Groudtruth, error) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(execution.Dataset.Config), &payload); err != nil {
		return nil, err
	}

	conf, err := getInjectionMetaFromPayload(payload)
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

	return groundtruth, nil
}

// AC@k 评估逻辑
func accuracyk(executions []dto.Execution) ([]*dto.Conclusion, error) {
	if len(executions) == 0 {
		return nil, errors.New("execution history is empty")
	}

	levelGran := make(map[string]map[string]*conclusionACatK)

	// 初始化级别
	for _, g := range executions[0].GranularityResults {
		levelGran[g.Level] = map[string]*conclusionACatK{
			"AC@1": {Metric: "AC@1", Level: g.Level},
			"AC@3": {Metric: "AC@3", Level: g.Level},
			"AC@5": {Metric: "AC@5", Level: g.Level},
		}
	}

	// 处理每个执行记录
	for _, execution := range executions {
		groundtruth, err := parseConfigAndGetGroundTruth(execution)
		if err != nil {
			return nil, err
		}

		for _, g := range execution.GranularityResults {
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
	var results []*dto.Conclusion
	for _, acMap := range levelGran {
		for _, value := range acMap {
			hitCount := 0
			for _, v := range value.Hit {
				if v == 1 {
					hitCount++
				}
			}

			// 未能命中的情况下，命中率为 0
			rate := 0.0
			if len(value.Hit) != 0 {
				rate = float64(hitCount) / float64(len(value.Hit))
			}

			results = append(results, &dto.Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   rate,
			})
		}
	}

	return results, nil
}

// Precision@k 评估逻辑
func precisionk(executions []dto.Execution) ([]*dto.Conclusion, error) {
	if len(executions) == 0 {
		return nil, errors.New("execution history is empty")
	}

	ks := []int{1, 3, 5}
	levelGran := make(map[string]map[string]*conclusionPrecisionk)

	// 初始化所有可能的粒度级别和对应的 PR@k
	for _, execution := range executions {
		for _, g := range execution.GranularityResults {
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

	// 处理每个执行记录
	for _, execution := range executions {
		groundtruth, err := parseConfigAndGetGroundTruth(execution)
		if err != nil {
			return nil, err
		}

		// 创建按级别分类的 groundtruth 集合
		groundtruthMap := make(map[string]map[string]struct{})
		for _, gt := range groundtruth {
			if _, exists := groundtruthMap[string(gt.Level)]; !exists {
				groundtruthMap[string(gt.Level)] = make(map[string]struct{})
			}
			groundtruthMap[string(gt.Level)][gt.Name] = struct{}{}
		}

		// 按级别收集所有预测结果，并按 Rank 排序
		levelPredictions := make(map[string][]database.GranularityResult)
		for _, g := range execution.GranularityResults {
			levelPredictions[g.Level] = append(levelPredictions[g.Level], g)
		}

		for level, preds := range levelPredictions {
			gtSet, exists := groundtruthMap[level]
			if !exists || len(gtSet) == 0 {
				continue // 无相关 ground truth，跳过
			}

			// 假设 preds 已经按 Rank 排序，如果没有排序需要排序
			sort.Slice(preds, func(i, j int) bool {
				return preds[i].Rank < preds[j].Rank
			})

			for _, k := range ks {
				minK := k
				if len(gtSet) < k {
					minK = len(gtSet)
				}

				if minK == 0 {
					continue // 避免除以零
				}

				// 获取前 k 个预测
				topK := preds
				if len(preds) > k {
					topK = preds[:k]
				}

				// 计算命中数
				hits := 0
				for _, pred := range topK {
					if _, hit := gtSet[pred.Result]; hit {
						hits++
					}
				}

				// 计算 PR@k
				precision := float64(hits) / float64(minK)

				// 累加到对应的结论
				metric := fmt.Sprintf("PR@%d", k)
				if conclusion, exists := levelGran[level][metric]; exists {
					conclusion.Sum += precision
					conclusion.Count++
				}
			}
		}
	}

	// 生成评估结果
	var results []*dto.Conclusion
	for _, prMap := range levelGran {
		for _, value := range prMap {
			if value.Count == 0 {
				results = append(results, &dto.Conclusion{
					Level:  value.Level,
					Metric: value.Metric,
					Rate:   0.0, // 或者根据需求设定为其他默认值
				})
			} else {
				results = append(results, &dto.Conclusion{
					Level:  value.Level,
					Metric: value.Metric,
					Rate:   value.Sum / float64(value.Count),
				})
			}
		}
	}

	return results, nil
}

// Avg@k 评估逻辑
func avgk(executions []dto.Execution) ([]*dto.Conclusion, error) {
	if len(executions) == 0 {
		return nil, errors.New("execution history is empty")
	}

	ks := []int{1, 3, 5}
	K := len(ks)
	levelGran := make(map[string]*conclusionAvgk)

	// 初始化所有可能的粒度级别和对应的 Avg@k
	for _, execution := range executions {
		for _, g := range execution.GranularityResults {
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

	// 处理每个执行记录
	for _, execution := range executions {
		groundtruth, err := parseConfigAndGetGroundTruth(execution)
		if err != nil {
			return nil, err
		}

		// 创建按级别分类的 ground truth 集合
		groundtruthMap := make(map[string]map[string]struct{})
		for _, gt := range groundtruth {
			if _, exists := groundtruthMap[string(gt.Level)]; !exists {
				groundtruthMap[string(gt.Level)] = make(map[string]struct{})
			}
			groundtruthMap[string(gt.Level)][gt.Name] = struct{}{}
		}

		// 按级别收集所有预测结果，并按 Rank 排序
		levelPredictions := make(map[string][]database.GranularityResult)
		for _, g := range execution.GranularityResults {
			levelPredictions[g.Level] = append(levelPredictions[g.Level], g)
		}

		for level, preds := range levelPredictions {
			gtSet, exists := groundtruthMap[level]
			if !exists || len(gtSet) == 0 {
				continue // 无相关 ground truth，跳过
			}

			// 假设 preds 已经按 Rank 排序，如果没有排序需要排序
			sort.Slice(preds, func(i, j int) bool {
				return preds[i].Rank < preds[j].Rank
			})

			// 计算 A@k 并累加
			var sumAvgk float64
			for _, k := range ks {
				minK := k
				if len(preds) < k {
					minK = len(preds)
				}

				if minK == 0 {
					continue // 避免除以零
				}

				// 计算 A@k
				totalRank := 0
				count := 0
				for i := 0; i < minK && i < len(preds); i++ {
					pred := preds[i]
					if _, hit := gtSet[pred.Result]; hit {
						totalRank += pred.Rank
						count++
					}
				}

				if count == 0 {
					continue // 避免除以零
				}

				avgk := float64(totalRank) / float64(count)
				sumAvgk += avgk
			}

			// 计算 Avg@K = sum(A@k) / K
			avgK := sumAvgk / float64(K)
			levelGran[level].Sum += avgK
			levelGran[level].Count++
		}
	}

	// 生成评估结果
	var results []*dto.Conclusion
	for _, value := range levelGran {
		if value.Count == 0 {
			results = append(results, &dto.Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   0.0, // 或者根据需求设定为其他默认值
			})
		} else {
			avgK := value.Sum / float64(value.Count)
			results = append(results, &dto.Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   avgK,
			})
		}
	}

	return results, nil
}

// MAP@k 评估逻辑
func mapk(executions []dto.Execution) ([]*dto.Conclusion, error) {
	if len(executions) == 0 {
		return nil, errors.New("execution history is empty")
	}

	ks := []int{1, 3, 5}
	K := len(ks)
	levelGran := make(map[string]*conclusionMAPk)

	// 初始化所有可能的粒度级别和对应的 MAP@k
	for _, execution := range executions {
		for _, g := range execution.GranularityResults {
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

	// 处理每个执行记录
	for _, execution := range executions {
		groundtruth, err := parseConfigAndGetGroundTruth(execution)
		if err != nil {
			return nil, err
		}

		// 创建按级别分类的 ground truth 集合
		groundtruthMap := make(map[string]map[string]struct{})
		for _, gt := range groundtruth {
			if _, exists := groundtruthMap[string(gt.Level)]; !exists {
				groundtruthMap[string(gt.Level)] = make(map[string]struct{})
			}
			groundtruthMap[string(gt.Level)][gt.Name] = struct{}{}
		}

		// 按级别收集所有预测结果，并按 Rank 排序
		levelPredictions := make(map[string][]database.GranularityResult)
		for _, g := range execution.GranularityResults {
			levelPredictions[g.Level] = append(levelPredictions[g.Level], g)
		}

		for level, preds := range levelPredictions {
			gtSet, exists := groundtruthMap[level]
			if !exists || len(gtSet) == 0 {
				continue // 无相关 ground truth，跳过
			}

			// 假设 preds 已经按 Rank 排序，如果没有排序需要排序
			sort.Slice(preds, func(i, j int) bool {
				return preds[i].Rank < preds[j].Rank
			})

			// 计算 PR@k 并累加
			for _, k := range ks {
				minK := k
				if len(preds) < k {
					minK = len(preds)
				}

				if minK == 0 {
					continue // 避免除以零
				}

				hits := 0
				sumPrecision := 0.0

				for i := 0; i < minK && i < len(preds); i++ {
					pred := preds[i]
					if _, hit := gtSet[pred.Result]; hit {
						hits++
						sumPrecision += float64(hits) / float64(i+1)
					}
				}

				// 计算 PR@k 并累加
				PRk := sumPrecision / float64(K*len(groundtruthMap[level]))
				levelGran[level].Sum += PRk
				levelGran[level].Count++
			}
		}
	}

	// 生成评估结果
	var results []*dto.Conclusion
	for _, value := range levelGran {
		if value.Count == 0 {
			results = append(results, &dto.Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   0.0, // 或者根据需求设定为其他默认值
			})
		} else {
			mapkValue := value.Sum / float64(value.Count)
			results = append(results, &dto.Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   mapkValue,
			})
		}
	}

	return results, nil
}

func mrr(executions []dto.Execution) ([]*dto.Conclusion, error) {
	if len(executions) == 0 {
		return nil, errors.New("execution history is empty")
	}

	levelGran := make(map[string]*conclustionMRR)

	for _, exexecution := range executions {
		for _, g := range exexecution.GranularityResults {
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

	for _, execution := range executions {
		groundtruth, err := parseConfigAndGetGroundTruth(execution)
		if err != nil {
			return nil, err
		}

		// 创建按级别分类的 groundtruth 集合
		groundtruthMap := make(map[string]map[string]struct{})
		for _, gt := range groundtruth {
			if _, exists := groundtruthMap[string(gt.Level)]; !exists {
				groundtruthMap[string(gt.Level)] = make(map[string]struct{})
			}
			groundtruthMap[string(gt.Level)][gt.Name] = struct{}{}
		}

		// 按级别收集所有预测结果，并按 Rank 排序
		levelPredictions := make(map[string][]database.GranularityResult)
		for _, g := range execution.GranularityResults {
			levelPredictions[g.Level] = append(levelPredictions[g.Level], g)
		}

		for level, preds := range levelPredictions {
			gtSet, exists := groundtruthMap[level]
			if !exists || len(gtSet) == 0 {
				continue
			}

			// 假设 preds 已经按 Rank 排序，如果没有排序需要排序
			sort.Slice(preds, func(i, j int) bool {
				return preds[i].Rank < preds[j].Rank
			})

			for _, pred := range preds {
				if _, hit := gtSet[pred.Result]; hit {
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
	}

	// 生成评估结果
	var results []*dto.Conclusion
	for _, value := range levelGran {
		if value.Count == 0 {
			results = append(results, &dto.Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   0.0, // 或者根据需求设定为其他默认值
			})
		} else {
			results = append(results, &dto.Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   value.Sum / float64(value.Count),
			})
		}
	}

	return results, nil
}
