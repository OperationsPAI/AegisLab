package database

import (
	"time"

	"aegis/config"
	"aegis/consts"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// FaultInjectionNoIssues view model
type FaultInjectionNoIssues struct {
	ID           int             `gorm:"column:datapack_id"`
	Name         string          `gorm:"column:datapack_name"`
	FaultType    chaos.ChaosType `gorm:"column:fault_type"`
	EngineConfig string          `gorm:"column:engine_config"`
	LabelKey     string          `gorm:"column:label_key"`
	LabelValue   string          `gorm:"column:value_key"`
	CreatedAt    time.Time       `gorm:"column:created_at"`
}

func (FaultInjectionNoIssues) TableName() string {
	return "fault_injection_no_issues"
}

// FaultInjectionWithIssues view model
type FaultInjectionWithIssues struct {
	ID                  int             `gorm:"column:datapack_id"`
	Name                string          `gorm:"column:datapack_name"`
	FaultType           chaos.ChaosType `gorm:"column:fault_type"`
	EngineConfig        string          `gorm:"column:engine_config"`
	LabelKey            string          `gorm:"column:label_key"`
	LabelValue          string          `gorm:"column:value_key"`
	CreatedAt           time.Time       `gorm:"column:created_at"`
	Issues              string          `gorm:"column:issues"`
	AbnormalAvgDuration float64         `gorm:"column:abnormal_avg_duration"`
	NormalAvgDuration   float64         `gorm:"column:normal_avg_duration"`
	AbnormalSuccRate    float64         `gorm:"column:abnormal_succ_rate"`
	NormalSuccRate      float64         `gorm:"column:normal_succ_rate"`
	AbnormalP99         float64         `gorm:"column:abnormal_p99"`
	NormalP99           float64         `gorm:"column:normal_p99"`
}

func (FaultInjectionWithIssues) TableName() string {
	return "fault_injection_with_issues"
}

func addDetectorJoins(query *gorm.DB) *gorm.DB {
	return query.
		Joins(`JOIN (
            SELECT 
                e.id,
                c.id AS algorithm_id,
                e.datapack_id,
                ROW_NUMBER() OVER (
                    PARTITION BY c.id, e.datapack_id 
                    ORDER BY e.created_at DESC, e.id DESC
                ) as rn
            FROM executions e
            JOIN container_versions cv ON e.algorithm_version_id = cv.id
            JOIN containers c ON c.id = cv.container_id
            WHERE e.state = 2 AND e.status = 1 AND c.name = ?
        ) er_ranked ON fi.id = er_ranked.datapack_id AND er_ranked.rn = 1`, config.GetString(consts.DetectorKey)).
		Joins("JOIN detector_results dr ON er_ranked.id = dr.execution_id")
}

func createDetectorViews() {
	var err error

	_ = DB.Migrator().DropView("fault_injection_no_issues")
	_ = DB.Migrator().DropView("fault_injection_with_issues")

	// Create view for fault injections with no issues
	noIssuesQuery := addDetectorJoins(DB.Table("fault_injections fi").
		Select(`DISTINCT 
		fi.id AS datapack_id, 
		fi.name AS name, 
		fi.fault_type AS fault_type, 
		fi.engine_config AS engine_config, 
		l.label_key as label_key,
		l.label_value as label_value,
		fi.created_at`).
		Joins("LEFT JOIN fault_injection_labels fil ON fil.fault_injection_id = fi.id").
		Joins("LEFT JOIN labels l ON fil.label_id = l.id").
		Group("fi.id, fi.name, fi.fault_type, fi.engine_config, fi.created_at, l.label_key, l.label_value"),
	).Where("dr.issues = '{}' OR dr.issues IS NULL")
	if err = DB.Migrator().CreateView("fault_injection_no_issues", gorm.ViewOption{Query: noIssuesQuery}); err != nil {
		logrus.Errorf("failed to create fault_injection_no_issues view: %v", err)
	}

	// Create view for fault injections with issues
	withIssuesQuery := addDetectorJoins(DB.Table("fault_injections fi").
		Select(`DISTINCT 
		fi.id AS datapack_id, 
		fi.name AS name,
		fi.fault_type AS fault_type, 
		fi.engine_config AS engine_config, 
		l.label_key as label_key,
		l.label_value as label_value,
		fi.created_at, 
		dr.issues, 
		dr.abnormal_avg_duration, 
		dr.normal_avg_duration, 
		dr.abnormal_succ_rate, 
		dr.normal_succ_rate, 
		dr.abnormal_p99, 
		dr.normal_p99`).
		Joins("LEFT JOIN tasks t ON t.id = fi.task_id").
		Joins("LEFT JOIN fault_injection_labels fil ON fil.fault_injection_id = fi.id").
		Joins("LEFT JOIN labels l ON fil.label_id = l.id").
		Group("fi.id, fi.name, fi.fault_type, fi.engine_config, fi.created_at, l.label_key, l.label_value, dr.issues, dr.abnormal_avg_duration, dr.normal_avg_duration, dr.abnormal_succ_rate, dr.normal_succ_rate, dr.abnormal_p99, dr.normal_p99"),
	).Where("dr.issues != '{}' AND dr.issues IS NOT NULL")
	if err = DB.Migrator().CreateView("fault_injection_with_issues", gorm.ViewOption{Query: withIssuesQuery}); err != nil {
		logrus.Errorf("failed to create fault_injection_with_issues view: %v", err)
	}
}
