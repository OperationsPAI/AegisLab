package database

import (
	"dagger/rcabench/config"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	DatasetInitial = 0
	DatasetSuccess = 1
	DatasetFailed  = 2
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
	ID              string    `gorm:"primaryKey" json:"id"`         // 唯一标识
	FaultType       int       `json:"fault_type"`                   // 故障类型
	Config          string    `json:"config"`                       // 配置 JSON 格式
	Duration        int       `json:"duration"`                     // 故障持续时间
	StartTime       time.Time `json:"start_time"`                   // 预计故障开始时间
	EndTime         time.Time `json:"end_time"`                     // 预计故障结束时间
	Status          int       `json:"status"`                       // 0: 初始状态，没有检查 1: 检查了，注入结束且成功 2: 检查了，注入结束且失败; 如果状态是 1，则可以用于数据集查询
	Description     string    `json:"description"`                  // 描述（可选字段）
	InjectionName   string    `gorm:"unique" json:"injection_name"` // 在k8s资源里注入的名字
	ProposedEndTime time.Time `json:"proposed_end_time"`            //预计结束时间
	CreatedAt       time.Time `json:"created_at"`                   // 创建时间
	UpdatedAt       time.Time `json:"updated_at"`                   // 更新时间
}

func InitDB() {
	var err error
	dbPath := config.GetString("storage.path")

	if err = ensureDirForFile(dbPath); err != nil {
		log.Fatalf("Failed to ensure database directory: %v", err)
	}

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(&Task{}, &FaultInjectionSchedule{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
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
