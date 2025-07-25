package database

import (
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type FaultInjectionProject struct {
	ID            int
	FaultType     int `gorm:"index"`
	DisplayConfig string
	EngineConfig  string
	PreDuration   int
	StartTime     time.Time
	EndTime       time.Time
	Status        int
	Benchmark     string
	Env           string `gorm:"column:env"`
	Batch         string `gorm:"column:batch"`
	Tag           string `gorm:"column:tag"`
	InjectionName string
	CreatedAt     time.Time
	ProjectName   string `gorm:"column:project_name"`
}

func (FaultInjectionProject) TableName() string {
	return "fault_injection_project"
}

// FaultInjectionNoIssues 视图模型
type FaultInjectionNoIssues struct {
	DatasetID     int       `gorm:"column:dataset_id" json:"dataset_id"`
	EngineConfig  string    `gorm:"column:engine_config" json:"engine_config"`
	Env           string    `gorm:"column:env" json:"env"`
	Batch         string    `gorm:"column:batch" json:"batch"`
	Tag           string    `gorm:"column:tag" json:"tag"`
	InjectionName string    `gorm:"column:injection_name" json:"injection_name"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

func (FaultInjectionNoIssues) TableName() string {
	return "fault_injection_no_issues"
}

// FaultInjectionWithIssues 视图模型
type FaultInjectionWithIssues struct {
	DatasetID           int       `gorm:"column:dataset_id" json:"dataset_id"`
	EngineConfig        string    `gorm:"column:engine_config" json:"engine_config"`
	Env                 string    `gorm:"column:env" json:"env"`
	Batch               string    `gorm:"column:batch" json:"batch"`
	Tag                 string    `gorm:"column:tag" json:"tag"`
	InjectionName       string    `gorm:"column:injection_name" json:"injection_name"`
	CreatedAt           time.Time `gorm:"column:created_at" json:"created_at"`
	Issues              string    `gorm:"column:issues" json:"issues"`
	AbnormalAvgDuration float64   `gorm:"column:abnormal_avg_duration" json:"abnormal_avg_duration"`
	NormalAvgDuration   float64   `gorm:"column:normal_avg_duration" json:"normal_avg_duration"`
	AbnormalSuccRate    float64   `gorm:"column:abnormal_succ_rate" json:"abnormal_succ_rate"`
	NormalSuccRate      float64   `gorm:"column:normal_succ_rate" json:"normal_succ_rate"`
	AbnormalP99         float64   `gorm:"column:abnormal_p99" json:"abnormal_p99"`
	NormalP99           float64   `gorm:"column:normal_p99" json:"normal_p99"`
}

func (FaultInjectionWithIssues) TableName() string {
	return "fault_injection_with_issues"
}

func createFaultInjectionViews() {
	var err error

	// Drop existing views
	DB.Migrator().DropView("fault_injection_project")
	DB.Migrator().DropView("fault_injection_no_issues")
	DB.Migrator().DropView("fault_injection_with_issues")

	projectQuery := DB.Table("fault_injection_schedules fis").
		Select(`
			fis.id,
			fis.fault_type,
			fis.display_config, 
			fis.engine_config,
			fis.pre_duration,
			fis.start_time,
			fis.end_time,
			fis.status,
			fis.benchmark, 
			fis.labels ->> 'env' AS env,
			fis.labels ->> 'batch' AS batch,
			fis.labels ->> 'tag' AS tag,
			fis.injection_name, 
			fis.created_at,
			p.name AS project_name`).
		Joins(`JOIN (
        	SELECT id AS task_id, project_id
        	FROM tasks
    	) t ON fis.task_id = t.task_id`).
		Joins("JOIN projects p ON p.id = t.project_id")
	if err = DB.Migrator().CreateView("fault_injection_project", gorm.ViewOption{Query: projectQuery}); err != nil {
		logrus.Errorf("failed to create fault_injection_project view: %v", err)
	}

	// Create view for fault injections with no issues
	noIssuesQuery := DB.Table("fault_injection_schedules fis").
		Select(`DISTINCT 
			fis.id AS dataset_id, 
			fis.engine_config, 
			fis.labels ->> 'env' AS env,
			fis.labels ->> 'batch' AS batch,
			fis.labels ->> 'tag' AS tag,
			fis.injection_name, 
			fis.created_at`).
		Joins(`JOIN (
        	SELECT id, dataset, algorithm,
               ROW_NUMBER() OVER (PARTITION BY dataset, algorithm ORDER BY created_at DESC, id DESC) as rn
        	FROM execution_results
    	) er_ranked ON fis.injection_name = er_ranked.dataset AND er_ranked.rn = 1`).
		Joins("JOIN detectors d ON er_ranked.id = d.execution_id").
		Where("d.issues = '{}' OR d.issues IS NULL")
	if err = DB.Migrator().CreateView("fault_injection_no_issues", gorm.ViewOption{Query: noIssuesQuery}); err != nil {
		logrus.Errorf("failed to create fault_injection_no_issues view: %v", err)
	}

	// Create view for fault injections with issues
	withIssuesQuery := DB.Table("fault_injection_schedules fis").
		Select(`DISTINCT 
			fis.id AS dataset_id, 
			fis.engine_config, 
			fis.labels ->> 'env' AS env,
			fis.labels ->> 'batch' AS batch,
			fis.labels ->> 'tag' AS tag,
			fis.injection_name, 
			fis.created_at, 
			d.issues, 
			d.abnormal_avg_duration, 
			d.normal_avg_duration, 
			d.abnormal_succ_rate, 
			d.normal_succ_rate, 
			d.abnormal_p99, 
			d.normal_p99`).
		Joins(`JOIN (
        	SELECT id, dataset, algorithm,
        	    ROW_NUMBER() OVER (PARTITION BY dataset, algorithm ORDER BY created_at DESC, id DESC) as rn
        	FROM execution_results
    	) er_ranked ON fis.injection_name = er_ranked.dataset AND er_ranked.rn = 1`).
		Joins("JOIN detectors d ON er_ranked.id = d.execution_id").
		Where("d.issues != '{}' AND d.issues IS NOT NULL")
	if err = DB.Migrator().CreateView("fault_injection_with_issues", gorm.ViewOption{Query: withIssuesQuery}); err != nil {
		logrus.Errorf("failed to create fault_injection_with_issues view: %v", err)
	}
}
