package database

import (
	"fmt"
	"time"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/sirupsen/logrus"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	InjectionName string    `gorm:"unique,index" json:"injection_name"` // 在k8s资源里注入的名字
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`   // 创建时间
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`   // 更新时间
}

type ExecutionResult struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"` // 唯一标识
	TaskID    string    `gorm:"index" json:"task_id"`               // 从属什么 taskid
	Dataset   int       `gorm:"index,unique" json:"dataset"`        // 数据集标识
	Algorithm string    `json:"algorithm"`                          // 使用的算法
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
	ID          int       `gorm:"primaryKey"`
	ExecutionID int       `gorm:"index,unique" json:"execution_id"` // ExecutionID 是主键
	SpanName    string    `gorm:"type:varchar(255)"`                // SpanName 数据库字段类型
	Issues      string    `gorm:"type:text"`                        // Issues 字段类型为文本
	AvgDuration *float64  `gorm:"type:float"`                       // AvgDuration 是浮点类型
	SuccRate    *float64  `gorm:"type:float"`                       // SuccRate 是浮点类型
	P90         *float64  `gorm:"type:float"`                       // P90 是浮点类型
	P95         *float64  `gorm:"type:float"`                       // P95 是浮点类型
	P99         *float64  `gorm:"type:float"`                       // P99 是浮点类型
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"` // CreatedAt 自动设置为当前时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"` // UpdatedAt 自动更新时间
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
	err = DB.AutoMigrate(&Task{}, &FaultInjectionSchedule{}, &ExecutionResult{}, &GranularityResult{}, &Detector{})
	if err != nil {
		logrus.Fatalf("Failed to migrate database: %v", err)
	}
}
