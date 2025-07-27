package database

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type LabelsMap map[string]string

func (l LabelsMap) Value() (driver.Value, error) {
	if l == nil {
		return "{}", nil
	}
	return json.Marshal(l)
}

func (l *LabelsMap) Scan(value any) error {
	if value == nil {
		*l = make(LabelsMap)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into LabelsMap", value)
	}

	return json.Unmarshal(bytes, l)
}

type Project struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"unique,index;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Status      int       `gorm:"default:1;index" json:"status"` // 0:禁用 1:启用 -1:删除
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

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
	ProjectID   int       `gorm:"index" json:"project_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// FaultInjectionSchedule 模型
type FaultInjectionSchedule struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`                    // 唯一标识
	TaskID        string    `gorm:"index" json:"task_id"`                                  // 从属什么 taskid
	FaultType     int       `gorm:"index" json:"fault_type"`                               // 故障类型
	DisplayConfig string    `json:"display_config"`                                        // 面向用户的展示配置
	EngineConfig  string    `json:"engine_config"`                                         // 面向系统的运行配置
	PreDuration   int       `json:"pre_duration"`                                          // 正常数据时间
	StartTime     time.Time `gorm:"default:null" json:"start_time"`                        // 预计故障开始时间
	EndTime       time.Time `gorm:"default:null" json:"end_time"`                          // 预计故障结束时间
	Status        int       `json:"status"`                                                // -1: 已删除 0: 初始状态 1: 注入结束且失败 2: 注入结束且成功 3: 收集数据失败 4:收集数据成功
	Description   string    `json:"description"`                                           // 描述（可选字段）
	Benchmark     string    `json:"benchmark"`                                             // 基准数据库
	InjectionName string    `gorm:"unique,index" json:"injection_name"`                    // 在k8s资源里注入的名字
	Labels        LabelsMap `gorm:"type:jsonb;default:'{}';index:,type:gin" json:"labels"` // 用户自定义标签，JSONB格式存储 key-value pairs
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`                      // 创建时间
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`                      // 更新时间
}

type Container struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`                                    // 唯一标识
	Type      string    `gorm:"index;not null;uniqueIndex:idx_container_unique" json:"type"`           // 镜像类型
	Name      string    `gorm:"index;not null;uniqueIndex:idx_container_unique" json:"name"`           // 名称
	Image     string    `gorm:"not null;uniqueIndex:idx_container_unique" json:"image"`                // 镜像名
	Tag       string    `gorm:"not null;default:'latest';uniqueIndex:idx_container_unique" json:"tag"` // 镜像标签
	Command   string    `gorm:"type:text;default:'bash /entrypoint.sh'" json:"command"`                // 启动命令
	EnvVars   string    `gorm:"default:''" json:"env_vars"`                                            // 环境变量名称列表
	Status    bool      `gorm:"default:true" json:"status"`                                            // 0: 已删除 1: 活跃
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                                      // 创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                      // 更新时间
}

type ExecutionResult struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"` // 唯一标识
	TaskID      string    `gorm:"index" json:"task_id"`               // 从属什么 taskid
	AlgorithmID int       `gorm:"index" json:"container_id"`          // 使用的算法
	DatasetID   int       `gorm:"index" json:"dataset_id"`            // 数据集标识
	Status      int       `gorm:"default:0" json:"status"`            // -1: 已删除 0: 初始状态 1: 执行算法成功 2: 执行算法失败
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`   // 更新时间
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
	ID                  int       `gorm:"primaryKey" json:"id"`                    // 唯一标识
	ExecutionID         int       `gorm:"index,unique" json:"execution_id"`        // ExecutionID 是主键
	SpanName            string    `gorm:"type:varchar(255)" json:"span_name"`      // SpanName 数据库字段类型
	Issues              string    `gorm:"type:text" json:"issues"`                 // Issues 字段类型为文本
	AbnormalAvgDuration *float64  `gorm:"type:float" json:"abnormal_avg_duration"` // 异常时段的平均耗时
	NormalAvgDuration   *float64  `gorm:"type:float" json:"normal_avg_duration"`   // 正常时段的平均耗时
	AbnormalSuccRate    *float64  `gorm:"type:float" json:"abnormal_succ_rate"`    // 异常时段的成功率
	NormalSuccRate      *float64  `gorm:"type:float" json:"normal_succ_rate"`      // 正常时段的成功率
	AbnormalP90         *float64  `gorm:"type:float" json:"abnormal_p90"`          // 异常时段的P90
	NormalP90           *float64  `gorm:"type:float" json:"normal_p90"`            // 正常时段的P90
	AbnormalP95         *float64  `gorm:"type:float" json:"abnormal_p95"`          // 异常时段的P95
	NormalP95           *float64  `gorm:"type:float" json:"normal_p95"`            // 正常时段的P95
	AbnormalP99         *float64  `gorm:"type:float" json:"abnormal_p99"`          // 异常时段的P99
	NormalP99           *float64  `gorm:"type:float" json:"normal_p99"`            // 正常时段的P99
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`        // CreatedAt 自动设置为当前时间
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updated_at"`        // UpdatedAt 自动更新时间
}
