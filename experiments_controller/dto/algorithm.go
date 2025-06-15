package dto

import "github.com/LGU-SE-Internal/rcabench/database"

type AlgorithmListResp struct {
	Algorithms []string `json:"algorithms"`
}

type AlgorithmExecutionPayload struct {
	Algorithm string            `json:"algorithm"`
	Dataset   string            `json:"dataset"`
	EnvVars   map[string]string `json:"env_vars" swaggertype:"object"`
}

type DetectorRecord struct {
	SpanName            string   `json:"span_name"`
	Issues              string   `json:"issue"`
	AbnormalAvgDuration *float64 `json:"abnormal_avg_duration" swaggertype:"number" example:"0.5"`
	NormalAvgDuration   *float64 `json:"normal_avg_duration" swaggertype:"number" example:"0.3"`
	AbnormalSuccRate    *float64 `json:"abnormal_succ_rate" swaggertype:"number" example:"0.8"`
	NormalSuccRate      *float64 `json:"normal_succ_rate" swaggertype:"number" example:"0.95"`
	AbnormalP90         *float64 `json:"abnormal_p90" swaggertype:"number" example:"1.2"`
	NormalP90           *float64 `json:"normal_p90" swaggertype:"number" example:"0.8"`
	AbnormalP95         *float64 `json:"abnormal_p95" swaggertype:"number" example:"1.5"`
	NormalP95           *float64 `json:"normal_p95" swaggertype:"number" example:"1.0"`
	AbnormalP99         *float64 `json:"abnormal_p99" swaggertype:"number" example:"2.0"`
	NormalP99           *float64 `json:"normal_p99" swaggertype:"number" example:"1.3"`
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
