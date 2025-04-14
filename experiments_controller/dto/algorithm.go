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
	SpanName    string   `json:"span_name"`
	Issues      string   `json:"issue"`
	AvgDuration *float64 `json:"avg_duration"`
	SuccRate    *float64 `json:"succ_rate"`
	P90         *float64 `json:"P90"`
	P95         *float64 `json:"P95"`
	P99         *float64 `json:"P99"`
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
