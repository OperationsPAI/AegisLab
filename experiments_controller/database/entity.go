package database

import (
	"time"
)

type Project struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"unique,index;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	IsPublic    bool      `gorm:"default:false;index:idx_project_visibility" json:"is_public"` // 是否公开可见
	Status      int       `gorm:"default:1;index" json:"status"`                               // 0:禁用 1:启用 -1:删除
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// Task 模型
type Task struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Type        string    `gorm:"index:idx_task_type_status" json:"type"` // 添加复合索引
	Immediate   bool      `json:"immediate"`
	ExecuteTime int64     `gorm:"index" json:"execute_time"` // 添加执行时间索引
	CronExpr    string    `json:"cron_expr,omitempty"`
	Payload     string    `json:"payload"`
	Status      string    `gorm:"index:idx_task_type_status;index:idx_task_project_status" json:"status"` // 添加多个复合索引
	TraceID     string    `gorm:"index" json:"trace_id"`                                                  // 添加追踪ID索引
	GroupID     string    `gorm:"index" json:"group_id"`                                                  // 添加组ID索引
	ProjectID   int       `gorm:"not null;index:idx_task_project_status" json:"project_id"`               // 任务必须属于某个项目
	CreatedAt   time.Time `gorm:"autoCreateTime;index" json:"created_at"`                                 // 添加时间索引
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// 外键关联
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

// FaultInjectionSchedule 模型
type FaultInjectionSchedule struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`                                      // 唯一标识
	TaskID        string    `gorm:"index:idx_fault_task_status;index:idx_fault_task_type" json:"task_id"`    // 从属什么 taskid，添加复合索引
	FaultType     int       `gorm:"index:idx_fault_task_type;index:idx_fault_type_status" json:"fault_type"` // 故障类型，添加复合索引
	DisplayConfig string    `json:"display_config"`                                                          // 面向用户的展示配置
	EngineConfig  string    `json:"engine_config"`                                                           // 面向系统的运行配置
	PreDuration   int       `json:"pre_duration"`                                                            // 正常数据时间
	StartTime     time.Time `gorm:"default:null;index" json:"start_time"`                                    // 预计故障开始时间，添加时间索引
	EndTime       time.Time `gorm:"default:null;index" json:"end_time"`                                      // 预计故障结束时间，添加时间索引
	Status        int       `gorm:"index:idx_fault_task_status;index:idx_fault_type_status" json:"status"`   // 状态，添加复合索引
	Description   string    `json:"description"`                                                             // 描述（可选字段）
	Benchmark     string    `gorm:"index" json:"benchmark"`                                                  // 基准数据库，添加索引
	InjectionName string    `gorm:"unique,index" json:"injection_name"`                                      // 在k8s资源里注入的名字
	CreatedAt     time.Time `gorm:"autoCreateTime;index" json:"created_at"`                                  // 创建时间，添加时间索引
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                        // 更新时间

	// 外键关联
	Task *Task `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}

type Container struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`                                    // 唯一标识
	Type      string    `gorm:"index;not null;uniqueIndex:idx_container_unique" json:"type"`           // 镜像类型
	Name      string    `gorm:"index;not null;uniqueIndex:idx_container_unique" json:"name"`           // 名称
	Image     string    `gorm:"not null;uniqueIndex:idx_container_unique" json:"image"`                // 镜像名
	Tag       string    `gorm:"not null;default:'latest';uniqueIndex:idx_container_unique" json:"tag"` // 镜像标签
	Command   string    `gorm:"type:text;default:'bash /entrypoint.sh'" json:"command"`                // 启动命令
	EnvVars   string    `gorm:"default:''" json:"env_vars"`                                            // 环境变量名称列表
	UserID    int       `gorm:"not null;index:idx_container_user" json:"user_id"`                      // 容器必须属于某个用户
	IsPublic  bool      `gorm:"default:false;index:idx_container_visibility" json:"is_public"`         // 是否公开可见
	Status    bool      `gorm:"default:true" json:"status"`                                            // 0: 已删除 1: 活跃
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                                      // 创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                      // 更新时间

	// 外键关联
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type ExecutionResult struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`                                       // 唯一标识
	TaskID      string    `gorm:"index:idx_exec_task_status;index:idx_exec_task_algo" json:"task_id"`       // 从属什么 taskid，添加复合索引
	AlgorithmID int       `gorm:"index:idx_exec_task_algo;index:idx_exec_algo_dataset" json:"container_id"` // 使用的算法，添加复合索引
	DatasetID   int       `gorm:"index:idx_exec_algo_dataset" json:"dataset_id"`                            // 数据集标识，添加复合索引
	Status      int       `gorm:"default:0;index:idx_exec_task_status" json:"status"`                       // 状态，添加复合索引
	CreatedAt   time.Time `gorm:"autoCreateTime;index" json:"created_at"`                                   // 创建时间，添加时间索引
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                         // 更新时间

	// 外键关联
	Task      *Task      `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	Algorithm *Container `gorm:"foreignKey:AlgorithmID" json:"algorithm,omitempty"`
	Dataset   *Dataset   `gorm:"foreignKey:DatasetID" json:"dataset,omitempty"`
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

	// 外键关联
	Execution *ExecutionResult `gorm:"foreignKey:ExecutionID" json:"execution,omitempty"`
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

	// 外键关联
	Execution *ExecutionResult `gorm:"foreignKey:ExecutionID" json:"execution,omitempty"`
}

// Dataset 数据集表
type Dataset struct {
	ID          int    `gorm:"primaryKey;autoIncrement" json:"id"`                                           // 唯一标识
	Name        string `gorm:"not null;index:idx_dataset_name_version,unique" json:"name"`                   // 数据集名称
	Version     string `gorm:"not null;default:'v1.0';index:idx_dataset_name_version,unique" json:"version"` // 数据集版本
	Description string `gorm:"type:text" json:"description"`                                                 // 数据集描述
	Type        string `gorm:"index" json:"type"`                                                            // 数据集类型 (e.g., "microservice", "database", "network")
	FileCount   int    `gorm:"default:0" json:"file_count"`                                                  // 文件数量
	DataSource  string `gorm:"type:text" json:"data_source"`                                                 // 数据来源描述
	Format      string `gorm:"default:'json'" json:"format"`                                                 // 数据格式 (json, csv, parquet等)
	ProjectID   int    `gorm:"not null;index:idx_dataset_project" json:"project_id"`                         // 数据集必须属于某个项目

	Status      int       `gorm:"default:1;index" json:"status"`                               // 0:禁用 1:启用 -1:删除
	IsPublic    bool      `gorm:"default:false;index:idx_dataset_visibility" json:"is_public"` // 是否公开
	DownloadURL string    `json:"download_url,omitempty"`                                      // 下载链接
	Checksum    string    `gorm:"type:varchar(64)" json:"checksum,omitempty"`                  // 文件校验和
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`                            // 创建时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`                            // 更新时间

	// 外键关联
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

// Label 标签表 - 统一的标签管理
type Label struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`                     // 唯一标识
	Key         string    `gorm:"not null;index:idx_label_key_value,unique" json:"key"`   // 标签键
	Value       string    `gorm:"not null;index:idx_label_key_value,unique" json:"value"` // 标签值
	Category    string    `gorm:"index" json:"category"`                                  // 标签分类 (dataset, fault_injection, algorithm, container等)
	Description string    `gorm:"type:text" json:"description"`                           // 标签描述
	Color       string    `gorm:"type:varchar(7);default:'#1890ff'" json:"color"`         // 标签颜色 (hex格式)
	IsSystem    bool      `gorm:"default:false;index" json:"is_system"`                   // 是否为系统标签
	Usage       int       `gorm:"default:0;index" json:"usage"`                           // 使用次数
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`                       // 创建时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`                       // 更新时间
}

// DatasetFaultInjection Dataset与FaultInjectionSchedule的多对多关系表
type DatasetFaultInjection struct {
	ID               int       `gorm:"primaryKey;autoIncrement" json:"id"`                                       // 唯一标识
	DatasetID        int       `gorm:"not null;index:idx_dataset_fault_unique,unique" json:"dataset_id"`         // 数据集ID
	FaultInjectionID int       `gorm:"not null;index:idx_dataset_fault_unique,unique" json:"fault_injection_id"` // 故障注入ID
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`                                         // 创建时间
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                         // 更新时间

	// 外键关联
	Dataset                *Dataset                `gorm:"foreignKey:DatasetID" json:"dataset,omitempty"`
	FaultInjectionSchedule *FaultInjectionSchedule `gorm:"foreignKey:FaultInjectionID" json:"fault_injection,omitempty"`
}

// DatasetLabel Dataset与Label的多对多关系表
type DatasetLabel struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`                               // 唯一标识
	DatasetID int       `gorm:"not null;index:idx_dataset_label_unique,unique" json:"dataset_id"` // 数据集ID
	LabelID   int       `gorm:"not null;index:idx_dataset_label_unique,unique" json:"label_id"`   // 标签ID
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                                 // 创建时间

	// 外键关联
	Dataset *Dataset `gorm:"foreignKey:DatasetID" json:"dataset,omitempty"`
	Label   *Label   `gorm:"foreignKey:LabelID" json:"label,omitempty"`
}

// FaultInjectionLabel FaultInjectionSchedule与Label的多对多关系表
type FaultInjectionLabel struct {
	ID               int       `gorm:"primaryKey;autoIncrement" json:"id"`                                     // 唯一标识
	FaultInjectionID int       `gorm:"not null;index:idx_fault_label_unique,unique" json:"fault_injection_id"` // 故障注入ID
	LabelID          int       `gorm:"not null;index:idx_fault_label_unique,unique" json:"label_id"`           // 标签ID
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`                                       // 创建时间

	// 外键关联
	FaultInjectionSchedule *FaultInjectionSchedule `gorm:"foreignKey:FaultInjectionID" json:"fault_injection,omitempty"`
	Label                  *Label                  `gorm:"foreignKey:LabelID" json:"label,omitempty"`
}

// ContainerLabel Container与Label的多对多关系表
type ContainerLabel struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`                                   // 唯一标识
	ContainerID int       `gorm:"not null;index:idx_container_label_unique,unique" json:"container_id"` // 容器ID
	LabelID     int       `gorm:"not null;index:idx_container_label_unique,unique" json:"label_id"`     // 标签ID
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`                                     // 创建时间

	// 外键关联
	Container *Container `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
	Label     *Label     `gorm:"foreignKey:LabelID" json:"label,omitempty"`
}

// ProjectLabel Project与Label的多对多关系表
type ProjectLabel struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`                               // 唯一标识
	ProjectID int       `gorm:"not null;index:idx_project_label_unique,unique" json:"project_id"` // 项目ID
	LabelID   int       `gorm:"not null;index:idx_project_label_unique,unique" json:"label_id"`   // 标签ID
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                                 // 创建时间

	// 外键关联
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Label   *Label   `gorm:"foreignKey:LabelID" json:"label,omitempty"`
}

// User 用户表
type User struct {
	ID          int        `gorm:"primaryKey;autoIncrement" json:"id"`    // 唯一标识
	Username    string     `gorm:"unique;not null;index" json:"username"` // 用户名（唯一）
	Email       string     `gorm:"unique;not null;index" json:"email"`    // 邮箱（唯一）
	Password    string     `gorm:"not null" json:"-"`                     // 密码（不返回给前端）
	FullName    string     `gorm:"not null" json:"full_name"`             // 全名
	Avatar      string     `json:"avatar,omitempty"`                      // 头像URL
	Phone       string     `gorm:"index" json:"phone,omitempty"`          // 电话号码
	Status      int        `gorm:"default:1;index" json:"status"`         // 0:禁用 1:启用 -1:删除
	IsActive    bool       `gorm:"default:true;index" json:"is_active"`   // 是否激活
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`               // 最后登录时间
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`      // 创建时间
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`      // 更新时间
}

// Role 角色表
type Role struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`   // 唯一标识
	Name        string    `gorm:"unique;not null;index" json:"name"`    // 角色名称（唯一）
	DisplayName string    `gorm:"not null" json:"display_name"`         // 显示名称
	Description string    `gorm:"type:text" json:"description"`         // 角色描述
	Type        string    `gorm:"default:'custom';index" json:"type"`   // 角色类型 (system, custom)
	IsSystem    bool      `gorm:"default:false;index" json:"is_system"` // 是否为系统角色
	Status      int       `gorm:"default:1;index" json:"status"`        // 0:禁用 1:启用 -1:删除
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`     // 创建时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`     // 更新时间
}

// Permission 权限表
type Permission struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`   // 唯一标识
	Name        string    `gorm:"unique;not null;index" json:"name"`    // 权限名称（唯一）
	DisplayName string    `gorm:"not null" json:"display_name"`         // 显示名称
	Description string    `gorm:"type:text" json:"description"`         // 权限描述
	Action      string    `gorm:"not null;index" json:"action"`         // 动作 (read, write, delete, execute等)
	ResourceID  int       `gorm:"index" json:"resource_id"`             // 关联的资源ID
	IsSystem    bool      `gorm:"default:false;index" json:"is_system"` // 是否为系统权限
	Status      int       `gorm:"default:1;index" json:"status"`        // 0:禁用 1:启用 -1:删除
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`     // 创建时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`     // 更新时间

	// 外键关联
	Resource *Resource `gorm:"foreignKey:ResourceID" json:"resource,omitempty"`
}

// Resource 资源表
type Resource struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`   // 唯一标识
	Name        string    `gorm:"unique;not null;index" json:"name"`    // 资源名称（唯一）
	DisplayName string    `gorm:"not null" json:"display_name"`         // 显示名称
	Description string    `gorm:"type:text" json:"description"`         // 资源描述
	Type        string    `gorm:"not null;index" json:"type"`           // 资源类型 (table, api, function等)
	Category    string    `gorm:"index" json:"category"`                // 资源分类
	ParentID    *int      `gorm:"index" json:"parent_id,omitempty"`     // 父资源ID（支持层级结构）
	IsSystem    bool      `gorm:"default:false;index" json:"is_system"` // 是否为系统资源
	Status      int       `gorm:"default:1;index" json:"status"`        // 0:禁用 1:启用 -1:删除
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`     // 创建时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`     // 更新时间

	// 外键关联
	Parent *Resource `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
}

// UserProject 用户与项目的多对多关系表（包含项目级权限）
type UserProject struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`                              // 唯一标识
	UserID    int       `gorm:"not null;index:idx_user_project_unique,unique" json:"user_id"`    // 用户ID
	ProjectID int       `gorm:"not null;index:idx_user_project_unique,unique" json:"project_id"` // 项目ID
	RoleID    int       `gorm:"index" json:"role_id"`                                            // 在该项目中的角色ID
	JoinedAt  time.Time `gorm:"autoCreateTime" json:"joined_at"`                                 // 加入时间
	Status    int       `gorm:"default:1;index" json:"status"`                                   // 0:禁用 1:启用 -1:退出
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                                // 创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                // 更新时间

	// 外键关联
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Role    *Role    `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// UserRole 用户与全局角色的多对多关系表
type UserRole struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`                        // 唯一标识
	UserID    int       `gorm:"not null;index:idx_user_role_unique,unique" json:"user_id"` // 用户ID
	RoleID    int       `gorm:"not null;index:idx_user_role_unique,unique" json:"role_id"` // 角色ID
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                          // 创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`                          // 更新时间

	// 外键关联
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role *Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// RolePermission 角色与权限的多对多关系表
type RolePermission struct {
	ID           int       `gorm:"primaryKey;autoIncrement" json:"id"`                                    // 唯一标识
	RoleID       int       `gorm:"not null;index:idx_role_permission_unique,unique" json:"role_id"`       // 角色ID
	PermissionID int       `gorm:"not null;index:idx_role_permission_unique,unique" json:"permission_id"` // 权限ID
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`                                      // 创建时间
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                      // 更新时间

	// 外键关联
	Role       *Role       `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Permission *Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
}

// UserPermission 用户直接权限表（补充角色权限的不足，支持特殊权限分配）
type UserPermission struct {
	ID           int        `gorm:"primaryKey;autoIncrement" json:"id"`                                    // 唯一标识
	UserID       int        `gorm:"not null;index:idx_user_permission_unique,unique" json:"user_id"`       // 用户ID
	PermissionID int        `gorm:"not null;index:idx_user_permission_unique,unique" json:"permission_id"` // 权限ID
	ProjectID    *int       `gorm:"index:idx_user_permission_unique,unique" json:"project_id,omitempty"`   // 项目ID（项目级权限，为空表示全局权限）
	GrantType    string     `gorm:"default:'grant';index" json:"grant_type"`                               // 授权类型 (grant, deny)
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`                                                  // 过期时间
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`                                      // 创建时间
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`                                      // 更新时间

	// 外键关联
	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Permission *Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
	Project    *Project    `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}
