package dto

import (
	"time"

	"github.com/LGU-SE-Internal/rcabench/database"
)

// ExecutionRecord represents algorithm execution record
type ExecutionRecord struct {
	Algorithm          string              `json:"algorithm"`
	GranularityRecords []GranularityRecord `json:"granularity_records"`
}

// ExecutionRecordWithDatasetID represents execution record with dataset ID
type ExecutionRecordWithDatasetID struct {
	DatasetID int
	ExecutionRecord
}

// GranularityRecord represents granularity analysis result
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

// SuccessfulExecutionItem represents successful execution item
type SuccessfulExecutionItem struct {
	ID        int       `json:"id"`         // Execution ID
	Algorithm string    `json:"algorithm"`  // Algorithm name
	Dataset   string    `json:"dataset"`    // Dataset name
	CreatedAt time.Time `json:"created_at"` // Creation time
}

// SuccessfulExecutionsResp represents successful executions response
type SuccessfulExecutionsResp []SuccessfulExecutionItem

// SuccessfulExecutionsReq represents successful executions request
type SuccessfulExecutionsReq struct {
	StartTime *time.Time `json:"start_time" form:"start_time"`
	EndTime   *time.Time `json:"end_time" form:"end_time"`
	Limit     *int       `json:"limit" form:"limit"`
	Offset    *int       `json:"offset" form:"offset"`
}
