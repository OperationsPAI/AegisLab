package database

import (
	"time"

	"aegis/config"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ExecutionResultProject struct {
	ID          int
	Algorithm   string `gorm:"column:algorithm"`
	Registry    string
	Repository  string
	Tag         string
	Dataset     string `gorm:"column:dataset"`
	Status      int
	CreatedAt   time.Time
	ProjectName string `gorm:"column:project_name"`
}

func (ExecutionResultProject) TableName() string {
	return "execution_result_project"
}

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

// FaultInjectionNoIssues view model
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

// FaultInjectionWithIssues view model
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
		MAX(CASE WHEN l.label_key = 'env' THEN l.label_value END) AS env,
		MAX(CASE WHEN l.label_key = 'batch' THEN l.label_value END) AS batch,
		MAX(CASE WHEN l.label_key = 'tag' THEN l.label_value END) AS tag,
		fis.injection_name, 
		fis.created_at`).
		Joins("LEFT JOIN fault_injection_labels fil ON fis.id = fil.fault_injection_id").
		Joins("LEFT JOIN labels l ON fil.label_id = l.id").
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
		MAX(CASE WHEN l.label_key = 'env' THEN l.label_value END) AS env,
		MAX(CASE WHEN l.label_key = 'batch' THEN l.label_value END) AS batch,
		MAX(CASE WHEN l.label_key = 'tag' THEN l.label_value END) AS tag,
		fis.injection_name, 
		fis.created_at, 
		d.issues, 
		d.abnormal_avg_duration, 
		d.normal_avg_duration, 
		d.abnormal_succ_rate, 
		d.normal_succ_rate, 
		d.abnormal_p99, 
		d.normal_p99`).
		Joins("LEFT JOIN fault_injection_labels fil ON fis.id = fil.fault_injection_id").
		Joins("LEFT JOIN labels l ON fil.label_id = l.id").
		Group("fis.id, fis.engine_config, fis.injection_name, fis.created_at, d.issues, d.abnormal_avg_duration, d.normal_avg_duration, d.abnormal_succ_rate, d.normal_succ_rate, d.abnormal_p99, d.normal_p99"),
	).Where("d.issues != '{}' AND d.issues IS NOT NULL")
	if err = DB.Migrator().CreateView("fault_injection_with_issues", gorm.ViewOption{Query: withIssuesQuery}); err != nil {
		logrus.Errorf("failed to create fault_injection_with_issues view: %v", err)
	}
}

func createExecutionResultViews() {
	var err error

	_ = DB.Migrator().DropView("execution_result_project")

	projectQuery := DB.Table("execution_results er").
		Select(`
		er.id,
		er.status,
		er.created_at,
		c.name AS algorithm,
		cv.registry,
		cv.namespace,
		cv.repository,
		cv.tag,
		fis.injection_name AS dataset,
		COALESCE(p.name, 'No Project') AS project_name`).
		Joins("JOIN container_versions cv ON er.algorithm_version_id = cv.id").
		Joins("JOIN containers c ON c.id = cv.container_id").
		Joins("JOIN fault_injection_schedules fis ON fis.id = er.datapack_id").
		Joins(`JOIN (
        	SELECT id AS task_id, project_id
        	FROM tasks
    	) t ON er.task_id = t.task_id`).
		Joins("LEFT JOIN projects p ON p.id = t.project_id")
	if err = DB.Migrator().CreateView("execution_result_project", gorm.ViewOption{Query: projectQuery}); err != nil {
		logrus.Errorf("failed to create execution_result_project view: %v", err)
	}
}

func createFaultInjectionViews() {
	var err error

	_ = DB.Migrator().DropView("fault_injection_project")

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
			MAX(CASE WHEN l.label_key = 'env' THEN l.label_value END) AS env,
			MAX(CASE WHEN l.label_key = 'batch' THEN l.label_value END) AS batch,
			MAX(CASE WHEN l.label_key = 'tag' THEN l.label_value END) AS tag,
			fis.injection_name, 
			fis.created_at,
			COALESCE(p.name, 'No Project') AS project_name`).
		Joins("LEFT JOIN fault_injection_labels fil ON fis.id = fil.fault_injection_id").
		Joins("LEFT JOIN labels l ON fil.label_id = l.id").
		Joins(`JOIN (
        	SELECT id AS task_id, project_id
        	FROM tasks
    	) t ON fis.task_id = t.task_id`).
		Joins("LEFT JOIN projects p ON p.id = t.project_id").
		Group("fis.id, fis.fault_type, fis.display_config, fis.engine_config, fis.pre_duration, fis.start_time, fis.end_time, fis.status, fis.benchmark, fis.injection_name, fis.created_at, p.name")
	if err = DB.Migrator().CreateView("fault_injection_project", gorm.ViewOption{Query: projectQuery}); err != nil {
		logrus.Errorf("failed to create fault_injection_project view: %v", err)
	}
}
