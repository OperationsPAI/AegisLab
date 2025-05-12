package dto

import "github.com/CUHK-SE-Group/rcabench/database"

type AlgorithmListResp struct {
	Algorithms []string `json:"algorithms"`
}

type AlgorithmExecutionPayload struct {
	Image   string            `json:"image"`
	Tag     string            `json:"tag"`
	Dataset string            `json:"dataset"`
	EnvVars map[string]string `json:"env_vars"`
}

type DetectorRecord struct {
	SpanName            string   `json:"span_name"`
	Issues              string   `json:"issue"`
	AbnormalAvgDuration *float64 `json:"abnormal_avg_duration"`
	NormalAvgDuration   *float64 `json:"normal_avg_duration"`
	AbnormalSuccRate    *float64 `json:"abnormal_succ_rate"`
	NormalSuccRate      *float64 `json:"normal_succ_rate"`
	AbnormalP90         *float64 `json:"abnormal_p90"`
	NormalP90           *float64 `json:"normal_p90"`
	AbnormalP95         *float64 `json:"abnormal_p95"`
	NormalP95           *float64 `json:"normal_p95"`
	AbnormalP99         *float64 `json:"abnormal_p99"`
	NormalP99           *float64 `json:"normal_p99"`
}

type ExecutionRecord struct {
	Algorithm          string              `json:"algorithm"`
	GranularityRecords []GranularityRecord `json:"granularity_records"`
}

type ExecutionRecordWithDatasetID struct {
	DatasetID int
	ExecutionRecord
}

type GranularityRecord struct {
	Level      string  `json:"level"`
	Result     string  `json:"result"`
	Rank       int     `json:"rank"`
	Confidence float64 `json:"confidence"`
}

func (g *GranularityRecord) Convert(result database.GranularityResult) {
	g.Level = result.Level
	g.Result = result.Result
	g.Rank = result.Rank
	g.Confidence = result.Confidence
}

var ExecuteEnvVarNameMap = map[string]struct{}{
	"ALGORITHM": {},
	"SERVICE":   {},
	"VENV":      {},
}
