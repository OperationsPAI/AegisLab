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

// 初始化数据库
func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("tasks.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移
	err = DB.AutoMigrate(&Task{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
}
