package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/sirupsen/logrus"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 全局 DB 对象
var DB *gorm.DB

// Task 模型
type Task struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Type      string    `json:"type"`
	Payload   string    `json:"payload"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// FaultInjectionSchedule 模型
type FaultInjectionSchedule struct {
	ID              int       `gorm:"primaryKey;autoIncrement" json:"id"` // 唯一标识
	TaskID          string    `gorm:"index" json:"task_id"`               // 从属什么 taskid
	FaultType       int       `json:"fault_type" gorm:"index"`            // 故障类型
	Config          string    `json:"config"`                             // 配置 JSON 格式
	Duration        int       `json:"duration"`                           // 故障持续时间
	StartTime       time.Time `gorm:"default:null" json:"start_time"`     // 预计故障开始时间
	EndTime         time.Time `gorm:"default:null" json:"end_time"`       // 预计故障结束时间
	Status          int       `json:"status"`                             // 0: 初始状态，没有检查 1: 检查了，注入结束且成功 2: 检查了，注入结束且失败; 如果状态是 1，则可以用于数据集查询
	Description     string    `json:"description"`                        // 描述（可选字段）
	InjectionName   string    `gorm:"unique,index" json:"injection_name"` // 在k8s资源里注入的名字
	ProposedEndTime time.Time `json:"proposed_end_time"`                  // 预计结束时间
	CreatedAt       time.Time `json:"created_at"`                         // 创建时间
	UpdatedAt       time.Time `json:"updated_at"`                         // 更新时间
}

type ExecutionResult struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"` // 唯一标识
	TaskID    string    `gorm:"index" json:"task_id"`               // 从属什么 taskid
	Dataset   int       `json:"dataset" gorm:"index,unique"`        // 数据集标识
	Algo      string    `json:"algo"`                               // 使用的算法
	CreatedAt time.Time `json:"created_at"`                         // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                         // 更新时间
}

type GranularityResult struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"` // 唯一标识
	ExecutionID int       `gorm:"index,unique" json:"execution_id"`   // 关联ExecutionResult的ID
	Level       string    `json:"level"`                              // 粒度类型 (e.g., "service", "pod", "span", "metric")
	Result      string    `json:"result"`                             // 定位结果，以逗号分隔
	Rank        int       `json:"rank"`                               // 排序，表示top1, top2等
	Confidence  float64   `json:"confidence"`                         // 可信度（可选）
	CreatedAt   time.Time `json:"created_at"`                         // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`                         // 更新时间
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
	CreatedAt   time.Time `gorm:"autoCreateTime"`                   // CreatedAt 自动设置为当前时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`                   // UpdatedAt 自动更新时间
}

func InitDB() {
	var err error
	mysqlUser := config.GetString("database.mysql_user")
	mysqlPassWord := config.GetString("database.mysql_password")
	mysqlHost := config.GetString("database.mysql_host")
	mysqlPort := config.GetString("database.mysql_port")
	mysqlDBName := config.GetString("database.mysql_db")

	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", mysqlUser, mysqlPassWord, mysqlHost, mysqlPort, mysqlDBName)
	DB, err = gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{})
	if err != nil {
		logrus.Errorf("Failed to connect to database: %v", err)

		dbPath := config.GetString("storage.path")

		if err = ensureDirForFile(dbPath); err != nil {
			logrus.Fatalf("Failed to ensure database directory: %v", err)
		}

		DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			logrus.Fatalf("Failed to connect to database: %v", err)
		}
	}

	err = DB.AutoMigrate(&Task{}, &FaultInjectionSchedule{}, &ExecutionResult{}, &GranularityResult{}, &Detector{})
	if err != nil {
		logrus.Fatalf("Failed to migrate database: %v", err)
	}
}

func ensureDirForFile(filePath string) error {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if mkdirErr := os.MkdirAll(dir, os.ModePerm); mkdirErr != nil {
			return mkdirErr
		}
	}
	return nil
}
