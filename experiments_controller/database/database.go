package database

import (
	"log"
	"time"

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
	ID          string    `gorm:"primaryKey" json:"id"` // 唯一标识
	FaultType   int       `json:"fault_type"`           // 故障类型
	Config      string    `json:"config"`               // 配置 JSON 格式
	LastTime    time.Time `json:"last_time"`            // 故障持续时间
	StartTime   time.Time `json:"start_time"`           // 故障开始时间
	EndTime     time.Time `json:"end_time"`             // 故障结束时间
	Description string    `json:"description"`          // 描述（可选字段）
	CreatedAt   time.Time `json:"created_at"`           // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`           // 更新时间
}

// 初始化数据库
func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("tasks.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移
	err = DB.AutoMigrate(&Task{}, &FaultInjectionSchedule{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
}
