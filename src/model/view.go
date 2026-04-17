package model

import (
	"time"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

// FaultInjectionNoIssues view model
type FaultInjectionNoIssues struct {
	ID           int              `gorm:"column:datapack_id"`
	Name         string           `gorm:"column:datapack_name"`
	FaultType    chaos.ChaosType  `gorm:"column:fault_type"`
	Category     chaos.SystemType `gorm:"column:category"`
	EngineConfig string           `gorm:"column:engine_config"`
	LabelKey     string           `gorm:"column:label_key"`
	LabelValue   string           `gorm:"column:value_key"`
	CreatedAt    time.Time        `gorm:"column:created_at"`
}

func (FaultInjectionNoIssues) TableName() string {
	return "fault_injection_no_issues"
}

// FaultInjectionWithIssues view model
type FaultInjectionWithIssues struct {
	ID                  int              `gorm:"column:datapack_id"`
	Name                string           `gorm:"column:datapack_name"`
	FaultType           chaos.ChaosType  `gorm:"column:fault_type"`
	Category            chaos.SystemType `gorm:"column:category"`
	EngineConfig        string           `gorm:"column:engine_config"`
	LabelKey            string           `gorm:"column:label_key"`
	LabelValue          string           `gorm:"column:value_key"`
	CreatedAt           time.Time        `gorm:"column:created_at"`
	Issues              string           `gorm:"column:issues"`
	AbnormalAvgDuration float64          `gorm:"column:abnormal_avg_duration"`
	NormalAvgDuration   float64          `gorm:"column:normal_avg_duration"`
	AbnormalSuccRate    float64          `gorm:"column:abnormal_succ_rate"`
	NormalSuccRate      float64          `gorm:"column:normal_succ_rate"`
	AbnormalP99         float64          `gorm:"column:abnormal_p99"`
	NormalP99           float64          `gorm:"column:normal_p99"`
}

func (FaultInjectionWithIssues) TableName() string {
	return "fault_injection_with_issues"
}
