package executor

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"maps"

	"github.com/CUHK-SE-Group/chaos-experiment/handler"
)

type GranularityResult struct {
	Level      string  `json:"level"`      // 粒度类型 (e.g., "service", "pod", "span", "metric")
	Result     string  `json:"result"`     // 定位结果，以逗号分隔
	Rank       int     `json:"rank"`       // 排序，表示top1, top2等
	Confidence float64 `json:"confidence"` // 可信度（可选）
}

type EvaluationItem struct {
	Algorithm   string       `json:"algorithm"`
	Conclusions []Conclusion `json:"conclusions"`
}

type Conclusion struct {
	Level  string  `json:"level"`  // 例如 service level
	Metric string  `json:"metric"` // 例如 topk
	Rate   float64 `json:"rate"`
}

type EvaluateMetric func([]GranularityResult, string) ([]*Conclusion, error)

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
	metrics   = make(map[string]EvaluateMetric)
	metricsMu sync.RWMutex
)

func init() {
	RegisterMetric("AC@k", accuracyk)
	RegisterMetric("PR@k", precisionk)
	RegisterMetric("Avg@k", avgk)
	RegisterMetric("MAP@k", mapk)
	RegisterMetric("MRR", mrr)
}

// 注册新的评估指标
func RegisterMetric(name string, metric EvaluateMetric) error {
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
func GetMetrics() map[string]EvaluateMetric {
	metricsMu.RLock()
	defer metricsMu.RUnlock()

	copiedMetrics := make(map[string]EvaluateMetric, len(metrics))
	maps.Copy(copiedMetrics, metrics)
	return copiedMetrics
}

func GetGranularityResults(records [][]string) ([]GranularityResult, error) {
	var granularityResults []GranularityResult
	for i, record := range records {
		if i == 0 {
			continue
		}

		rank, err := strconv.Atoi(record[2])
		if err != nil {
			return []GranularityResult{}, fmt.Errorf("failed to convert the element in (row: %d, column: %d) string to int: %v", i, 3, err)
		}

		confidence, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			return []GranularityResult{}, fmt.Errorf("failed to convert the element in (row: %d, column: %d) string to float: %v", i, 4, err)
		}

		granularityResults = append(granularityResults, GranularityResult{
			Level:      record[0],
			Result:     record[1],
			Rank:       rank,
			Confidence: confidence,
		})
	}

	return granularityResults, nil
}

// AC@k 评估逻辑
func accuracyk(granularityResults []GranularityResult, service string) ([]*Conclusion, error) {
	levelGran := make(map[string]map[string]*conclusionACatK)

	// 初始化级别
	for _, g := range granularityResults {
		levelGran[g.Level] = map[string]*conclusionACatK{
			"AC@1": {Metric: "AC@1", Level: g.Level},
			"AC@3": {Metric: "AC@3", Level: g.Level},
			"AC@5": {Metric: "AC@5", Level: g.Level},
		}
	}

	groundtruth := []handler.Groudtruth{
		{Level: handler.Service, Name: service},
		{Level: handler.Pod, Name: service},
	}

	for _, g := range granularityResults {
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

			// 未能命中的情况下，命中率为 0
			rate := 0.0
			if len(value.Hit) != 0 {
				rate = float64(hitCount) / float64(len(value.Hit))
			}

			results = append(results, &Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   rate,
			})
		}
	}

	return results, nil
}

// Precision@k 评估逻辑
func precisionk(granularityResults []GranularityResult, service string) ([]*Conclusion, error) {
	ks := []int{1, 3, 5}
	levelGran := make(map[string]map[string]*conclusionPrecisionk)

	// 初始化所有可能的粒度级别和对应的 PR@k
	for _, g := range granularityResults {
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

	groundtruth := []handler.Groudtruth{
		{Level: handler.Service, Name: service},
		{Level: handler.Pod, Name: service},
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
	levelPredictions := make(map[string][]GranularityResult)
	for _, g := range granularityResults {
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

	// 生成评估结果
	var results []*Conclusion
	for _, prMap := range levelGran {
		for _, value := range prMap {
			if value.Count == 0 {
				results = append(results, &Conclusion{
					Level:  value.Level,
					Metric: value.Metric,
					Rate:   0.0, // 或者根据需求设定为其他默认值
				})
			} else {
				results = append(results, &Conclusion{
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
func avgk(granularityResults []GranularityResult, service string) ([]*Conclusion, error) {
	ks := []int{1, 3, 5}
	K := len(ks)
	levelGran := make(map[string]*conclusionAvgk)

	// 初始化所有可能的粒度级别和对应的 Avg@k
	for _, g := range granularityResults {
		if _, exists := levelGran[g.Level]; !exists {
			levelGran[g.Level] = &conclusionAvgk{
				Metric: "Avg@k",
				Level:  g.Level,
				Sum:    0.0,
				Count:  0,
			}
		}
	}

	groundtruth := []handler.Groudtruth{
		{Level: handler.Service, Name: service},
		{Level: handler.Pod, Name: service},
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
	levelPredictions := make(map[string][]GranularityResult)
	for _, g := range granularityResults {
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

		// 计算 AC@k 并累加
		var sumAvgk float64
		for _, k := range ks {
			minK := k
			if len(preds) < k {
				minK = len(preds)
			}

			if minK == 0 {
				continue // 避免除以零
			}

			// 计算 AC@k
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

		// 计算 Avg@K = sum(AC@k) / K
		avgK := sumAvgk / float64(K)
		levelGran[level].Sum += avgK
		levelGran[level].Count++
	}

	// 生成评估结果
	var results []*Conclusion
	for _, value := range levelGran {
		if value.Count == 0 {
			results = append(results, &Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   0.0, // 或者根据需求设定为其他默认值
			})
		} else {
			avgK := value.Sum / float64(value.Count)
			results = append(results, &Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   avgK,
			})
		}
	}

	return results, nil
}

// MAP@k 评估逻辑
func mapk(granularityResults []GranularityResult, service string) ([]*Conclusion, error) {
	ks := []int{1, 3, 5}
	K := len(ks)
	levelGran := make(map[string]*conclusionMAPk)

	// 初始化所有可能的粒度级别和对应的 MAP@k
	for _, g := range granularityResults {
		if _, exists := levelGran[g.Level]; !exists {
			levelGran[g.Level] = &conclusionMAPk{
				Metric: "MAP@k",
				Level:  g.Level,
				Sum:    0.0,
				Count:  0,
			}
		}
	}

	groundtruth := []handler.Groudtruth{
		{Level: handler.Service, Name: service},
		{Level: handler.Pod, Name: service},
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
	levelPredictions := make(map[string][]GranularityResult)
	for _, g := range granularityResults {
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

	// 生成评估结果
	var results []*Conclusion
	for _, value := range levelGran {
		if value.Count == 0 {
			results = append(results, &Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   0.0, // 或者根据需求设定为其他默认值
			})
		} else {
			mapkValue := value.Sum / float64(value.Count)
			results = append(results, &Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   mapkValue,
			})
		}
	}

	return results, nil
}

func mrr(granularityResults []GranularityResult, service string) ([]*Conclusion, error) {
	levelGran := make(map[string]*conclustionMRR)

	for _, g := range granularityResults {
		if _, exists := levelGran[g.Level]; !exists {
			levelGran[g.Level] = &conclustionMRR{
				Metric: "MRR",
				Level:  g.Level,
				Sum:    0.0,
				Count:  0,
			}
		}
	}

	groundtruth := []handler.Groudtruth{
		{Level: handler.Service, Name: service},
		{Level: handler.Pod, Name: service},
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
	levelPredictions := make(map[string][]GranularityResult)
	for _, g := range granularityResults {
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

	// 生成评估结果
	var results []*Conclusion
	for _, value := range levelGran {
		if value.Count == 0 {
			results = append(results, &Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   0.0, // 或者根据需求设定为其他默认值
			})
		} else {
			results = append(results, &Conclusion{
				Level:  value.Level,
				Metric: value.Metric,
				Rate:   value.Sum / float64(value.Count),
			})
		}
	}

	return results, nil
}
