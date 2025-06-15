package database

import (
	"fmt"
	"time"

	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/sirupsen/logrus"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

// 全局 DB 对象
var DB *gorm.DB

// Task 模型
type Task struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Type        string    `json:"type"`
	Immediate   bool      `json:"immediate"`
	ExecuteTime int64     `json:"execute_time"`
	CronExpr    string    `json:"cron_expr,omitempty"`
	Payload     string    `json:"payload"`
	Status      string    `json:"status"`
	TraceID     string    `json:"trace_id"`
	GroupID     string    `json:"group_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// FaultInjectionSchedule 模型
type FaultInjectionSchedule struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"` // 唯一标识
	TaskID        string    `gorm:"index" json:"task_id"`               // 从属什么 taskid
	FaultType     int       `gorm:"index" json:"fault_type"`            // 故障类型
	DisplayConfig string    `json:"display_config"`                     // 面向用户的展示配置
	EngineConfig  string    `json:"engine_config"`                      // 面向系统的运行配置
	PreDuration   int       `json:"pre_duration"`                       // 正常数据时间
	StartTime     time.Time `gorm:"default:null" json:"start_time"`     // 预计故障开始时间
	EndTime       time.Time `gorm:"default:null" json:"end_time"`       // 预计故障结束时间
	Status        int       `json:"status"`                             // 0: 初始状态 1: 注入结束且成功 2: 注入结束且失败 3: 收集数据成功 4:收集数据失败
	Description   string    `json:"description"`                        // 描述（可选字段）
	Benchmark     string    `json:"benchmark"`                          // 基准数据库
	InjectionName string    `gorm:"unique,index" json:"injection_name"` // 在k8s资源里注入的名字
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`   // 创建时间
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`   // 更新时间
}

// TODO 添加数据的接口
type Algorithm struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"` // 唯一标识
	Name      string    `gorm:"index;not null" json:"name"`         // 算法名称
	Image     string    `gorm:"not null" json:"image"`              // 算法镜像
	Tag       string    `gorm:"not null" json:"tag"`                // 算法镜像标签
	Status    bool      `gorm:"default:true" json:"is_public"`      // 0: 已删除 1: 活跃
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`   // 创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`   // 更新时间
}

// TODO 算法执行状态
type ExecutionResult struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"` // 唯一标识
	TaskID    string    `gorm:"index" json:"task_id"`               // 从属什么 taskid
	Algorithm string    `gorm:"index" json:"algorithm"`             // 使用的算法
	Dataset   string    `gorm:"index" json:"dataset"`               // 数据集标识
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`   // 创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`   // 更新时间
}

type GranularityResult struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"` // 唯一标识
	ExecutionID int       `gorm:"index,unique" json:"execution_id"`   // 关联ExecutionResult的ID
	Level       string    `json:"level"`                              // 粒度类型 (e.g., "service", "pod", "span", "metric")
	Result      string    `json:"result"`                             // 定位结果，以逗号分隔
	Rank        int       `json:"rank"`                               // 排序，表示top1, top2等
	Confidence  float64   `json:"confidence"`                         // 可信度（可选）
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`   // 更新时间
}

type Detector struct {
	ID                  int       `gorm:"primaryKey"`
	ExecutionID         int       `gorm:"index,unique" json:"execution_id"` // ExecutionID 是主键
	SpanName            string    `gorm:"type:varchar(255)"`                // SpanName 数据库字段类型
	Issues              string    `gorm:"type:text"`                        // Issues 字段类型为文本
	AbnormalAvgDuration *float64  `gorm:"type:float"`                       // 异常时段的平均耗时
	NormalAvgDuration   *float64  `gorm:"type:float"`                       // 正常时段的平均耗时
	AbnormalSuccRate    *float64  `gorm:"type:float"`                       // 异常时段的成功率
	NormalSuccRate      *float64  `gorm:"type:float"`                       // 正常时段的成功率
	AbnormalP90         *float64  `gorm:"type:float"`                       // 异常时段的P90
	NormalP90           *float64  `gorm:"type:float"`                       // 正常时段的P90
	AbnormalP95         *float64  `gorm:"type:float"`                       // 异常时段的P95
	NormalP95           *float64  `gorm:"type:float"`                       // 正常时段的P95
	AbnormalP99         *float64  `gorm:"type:float"`                       // 异常时段的P99
	NormalP99           *float64  `gorm:"type:float"`                       // 正常时段的P99
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"` // CreatedAt 自动设置为当前时间
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updated_at"` // UpdatedAt 自动更新时间
}

// FaultInjectionNoIssues 视图模型
type FaultInjectionNoIssues struct {
	DatasetID     int    `gorm:"column:DatasetID" json:"dataset_id"`
	DisplayConfig string `gorm:"column:display_config" json:"display_config"`
	EngineConfig  string `gorm:"column:engine_config" json:"engine_config"`
	PreDuration   int    `gorm:"column:pre_duration" json:"pre_duration"`
	InjectionName string `gorm:"column:injection_name" json:"injection_name"`
}

func (FaultInjectionNoIssues) TableName() string {
	return "fault_injection_no_issues"
}

// FaultInjectionWithIssues 视图模型
type FaultInjectionWithIssues struct {
	DatasetID     int    `gorm:"column:DatasetID" json:"dataset_id"`
	DisplayConfig string `gorm:"column:display_config" json:"display_config"`
	EngineConfig  string `gorm:"column:engine_config" json:"engine_config"`
	PreDuration   int    `gorm:"column:pre_duration" json:"pre_duration"`
	InjectionName string `gorm:"column:injection_name" json:"injection_name"`
	Issues        string `gorm:"column:issues" json:"issues"`
}

func (FaultInjectionWithIssues) TableName() string {
	return "fault_injection_with_issues"
}

func InitDB() {
	var err error
	mysqlUser := config.GetString("database.mysql_user")
	mysqlPassWord := config.GetString("database.mysql_password")
	mysqlHost := config.GetString("database.mysql_host")
	mysqlPort := config.GetString("database.mysql_port")
	mysqlDBName := config.GetString("database.mysql_db")

	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", mysqlUser, mysqlPassWord, mysqlHost, mysqlPort, mysqlDBName)

	maxRetries := 3
	retryDelay := 10 * time.Second

	for i := 0; i <= maxRetries; i++ {
		DB, err = gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{})
		if err == nil {
			logrus.Info("Successfully connected to the database.")
			if err := DB.Use(tracing.NewPlugin()); err != nil {
				panic(err)
			}
			break // Connection successful, exit loop
		}

		logrus.Errorf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries+1, err)
		if i < maxRetries {
			logrus.Infof("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		logrus.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries+1, err)
	}

	err = DB.AutoMigrate(
		&Task{},
		&FaultInjectionSchedule{},
		&Algorithm{},
		&ExecutionResult{},
		&GranularityResult{},
		&Detector{},
	)
	if err != nil {
		logrus.Fatalf("Failed to migrate database: %v", err)
	}

	DB.Migrator().DropView("fault_injection_no_issues")
	DB.Migrator().DropView("fault_injection_with_issues")

	noIssuesQuery := DB.Table("fault_injection_schedules fis").
		Select("DISTINCT fis.id AS DatasetID, fis.fault_type, fis.display_config, fis.engine_config, fis.pre_duration, fis.injection_name").
		Joins("JOIN execution_results er ON fis.id = er.dataset").
		Joins("JOIN detectors d ON er.id = d.execution_id").
		Where("d.issues = '{}'")
	err = DB.Migrator().CreateView("fault_injection_no_issues", gorm.ViewOption{Query: noIssuesQuery})
	if err != nil {
		logrus.Errorf("failed to create fault_injection_no_issues view: %v", err)
	}

	withIssuesQuery := DB.Table("fault_injection_schedules fis").
		Select("fis.id AS DatasetID, fis.fault_type, fis.display_config, fis.engine_config, fis.pre_duration, fis.injection_name, d.issues").
		Joins("JOIN execution_results er ON fis.id = er.dataset").
		Joins("JOIN detectors d ON er.id = d.execution_id").
		Where("d.issues != '{}'")
	err = DB.Migrator().CreateView("fault_injection_with_issues", gorm.ViewOption{Query: withIssuesQuery})
	if err != nil {
		logrus.Errorf("failed to create fault_injection_with_issues view: %v", err)
	}
}
