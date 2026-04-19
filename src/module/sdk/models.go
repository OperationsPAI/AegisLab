package sdk

import "time"

// SDKDatasetSample maps to the Python SDK's `data` table (read-only from AegisLab).
// Do NOT add this to AutoMigrate - the SDK creates and manages this table.
type SDKDatasetSample struct {
	ID          int     `gorm:"primaryKey;column:id" json:"id"`
	Dataset     string  `gorm:"column:dataset" json:"dataset"`
	Index       *int    `gorm:"column:index" json:"index"`
	Source      string  `gorm:"column:source" json:"source"`
	SourceIndex *int    `gorm:"column:source_index" json:"source_index"`
	Question    string  `gorm:"column:question" json:"question"`
	Answer      *string `gorm:"column:answer" json:"answer"`
	Topic       *string `gorm:"column:topic" json:"topic"`
	Level       *int    `gorm:"column:level" json:"level"`
	FileName    *string `gorm:"column:file_name" json:"file_name"`
	Meta        *string `gorm:"column:meta;type:json" json:"meta"`
	Tags        *string `gorm:"column:tags;type:json" json:"tags"`
}

func (SDKDatasetSample) TableName() string { return "data" }

// SDKEvaluationSample maps to the Python SDK's `evaluation_data` table (read-only from AegisLab).
// Do NOT add this to AutoMigrate - the SDK creates and manages this table.
type SDKEvaluationSample struct {
	ID                   int        `gorm:"primaryKey;column:id" json:"id"`
	CreatedAt            *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt            *time.Time `gorm:"column:updated_at" json:"updated_at"`
	Dataset              string     `gorm:"column:dataset" json:"dataset"`
	DatasetIndex         *int       `gorm:"column:dataset_index" json:"dataset_index"`
	Source               string     `gorm:"column:source" json:"source"`
	RawQuestion          string     `gorm:"column:raw_question" json:"raw_question"`
	Level                *int       `gorm:"column:level" json:"level"`
	AugmentedQuestion    *string    `gorm:"column:augmented_question" json:"augmented_question"`
	CorrectAnswer        *string    `gorm:"column:correct_answer" json:"correct_answer"`
	FileName             *string    `gorm:"column:file_name" json:"file_name"`
	Meta                 *string    `gorm:"column:meta;type:json" json:"meta"`
	TraceID              *string    `gorm:"column:trace_id" json:"trace_id"`
	TraceURL             *string    `gorm:"column:trace_url" json:"trace_url"`
	Response             *string    `gorm:"column:response" json:"response"`
	TimeCost             *float64   `gorm:"column:time_cost" json:"time_cost"`
	Trajectories         *string    `gorm:"column:trajectories;type:json" json:"trajectories"`
	ExtractedFinalAnswer *string    `gorm:"column:extracted_final_answer" json:"extracted_final_answer"`
	JudgedResponse       *string    `gorm:"column:judged_response" json:"judged_response"`
	Reasoning            *string    `gorm:"column:reasoning" json:"reasoning"`
	Correct              *bool      `gorm:"column:correct" json:"correct"`
	Confidence           *float64   `gorm:"column:confidence" json:"confidence"`
	ExpID                string     `gorm:"column:exp_id" json:"exp_id"`
	AgentType            *string    `gorm:"column:agent_type" json:"agent_type"`
	ModelName            *string    `gorm:"column:model_name" json:"model_name"`
	Stage                string     `gorm:"column:stage" json:"stage"`
}

func (SDKEvaluationSample) TableName() string { return "evaluation_data" }
