package database

import (
	"time"

	"aegis/config"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// FaultInjectionNoIssues view model
type FaultInjectionNoIssues struct {
	ID           int       `gorm:"column:datapack_id"`
	Name         string    `gorm:"column:datapack_name"`
	EngineConfig string    `gorm:"column:engine_config"`
	LabelKey     string    `gorm:"column:label_key"`
	LabelValue   string    `gorm:"column:value_key"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (FaultInjectionNoIssues) TableName() string {
	return "fault_injection_no_issues"
}

// FaultInjectionWithIssues view model
type FaultInjectionWithIssues struct {
	ID                  int       `gorm:"column:datapack_id"`
	Name                string    `gorm:"column:datapack_name"`
	EngineConfig        string    `gorm:"column:engine_config"`
	LabelKey            string    `gorm:"column:label_key"`
	LabelValue          string    `gorm:"column:value_key"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	Issues              string    `gorm:"column:issues"`
	AbnormalAvgDuration float64   `gorm:"column:abnormal_avg_duration"`
	NormalAvgDuration   float64   `gorm:"column:normal_avg_duration"`
	AbnormalSuccRate    float64   `gorm:"column:abnormal_succ_rate"`
	NormalSuccRate      float64   `gorm:"column:normal_succ_rate"`
	AbnormalP99         float64   `gorm:"column:abnormal_p99"`
	NormalP99           float64   `gorm:"column:normal_p99"`
}

func (FaultInjectionWithIssues) TableName() string {
	return "fault_injection_with_issues"
}

func addDetectorJoins(query *gorm.DB) *gorm.DB {
	return query.
		Joins(`JOIN (
            SELECT 
                er.id,
                c.id AS algorithm_id,
                er.datapack_id,
                ROW_NUMBER() OVER (
                    PARTITION BY c.id, er.datapack_id 
                    ORDER BY er.created_at DESC, er.id DESC
                ) as rn
            FROM execution_results er
            JOIN container_versions cv ON er.algorithm_version_id = cv.id
            JOIN containers c ON c.id = cv.container_id
            WHERE er.status = 2 AND c.name = ?
        ) er_ranked ON fis.id = er_ranked.datapack_id AND er_ranked.rn = 1`, config.GetString("algo.detector")).
		Joins("JOIN detectors d ON er_ranked.id = d.execution_id")
}

func createDetectorViews() {
	var err error

	_ = DB.Migrator().DropView("fault_injection_no_issues")
	_ = DB.Migrator().DropView("fault_injection_with_issues")

	// Create view for fault injections with no issues
	noIssuesQuery := addDetectorJoins(DB.Table("fault_injection_schedules fis").
		Select(`DISTINCT 
		fis.id AS dataset_id, 
		fis.engine_config, 
		l.label_key as label_key,
		l.value_key as value_key,
		fis.injection_name, 
		fis.created_at`).
		Joins("LEFT JOIN tasks t ON t.id = fis.task_id").
		Joins("LEFT JOIN task_labels tl ON tl.task_id = t.id").
		Joins("LEFT JOIN labels l ON tl.label_id = l.id").
		Group("fis.id, fis.engine_config, fis.injection_name, fis.created_at"),
	).Where("d.issues = '{}' OR d.issues IS NULL")
	if err = DB.Migrator().CreateView("fault_injection_no_issues", gorm.ViewOption{Query: noIssuesQuery}); err != nil {
		logrus.Errorf("failed to create fault_injection_no_issues view: %v", err)
	}

	// Create view for fault injections with issues
	withIssuesQuery := addDetectorJoins(DB.Table("fault_injection_schedules fis").
		Select(`DISTINCT 
		fis.id AS dataset_id, 
		fis.engine_config, 
		l.label_key as label_key,
		l.value_key as value_key,
		fis.injection_name, 
		fis.created_at, 
		d.issues, 
		d.abnormal_avg_duration, 
		d.normal_avg_duration, 
		d.abnormal_succ_rate, 
		d.normal_succ_rate, 
		d.abnormal_p99, 
		d.normal_p99`).
		Joins("LEFT JOIN tasks t ON t.id = fis.task_id").
		Joins("LEFT JOIN task_labels tl ON tl.task_id = t.id").
		Joins("LEFT JOIN labels l ON tl.label_id = l.id").
		Group("fis.id, fis.engine_config, fis.injection_name, fis.created_at, d.issues, d.abnormal_avg_duration, d.normal_avg_duration, d.abnormal_succ_rate, d.normal_succ_rate, d.abnormal_p99, d.normal_p99"),
	).Where("d.issues != '{}' AND d.issues IS NOT NULL")
	if err = DB.Migrator().CreateView("fault_injection_with_issues", gorm.ViewOption{Query: withIssuesQuery}); err != nil {
		logrus.Errorf("failed to create fault_injection_with_issues view: %v", err)
	}
}
