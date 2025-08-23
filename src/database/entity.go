package database

import (
	"time"
)

type Project struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"unique,index;not null;size:128" json:"name"`                  // Project name with size limit
	Description string    `gorm:"type:text" json:"description"`                                // Project description
	IsPublic    bool      `gorm:"default:false;index:idx_project_visibility" json:"is_public"` // Whether publicly visible
	Status      int       `gorm:"default:1;index" json:"status"`                               // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt   time.Time `gorm:"autoCreateTime;index" json:"created_at"`                      // Creation time
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`                            // Update time
}

// Task model
type Task struct {
	ID          string    `gorm:"primaryKey;size:64" json:"id"`                                                   // Task ID with size limit
	Type        string    `gorm:"index:idx_task_type_status;size:64" json:"type"`                                 // Task type with size limit
	Immediate   bool      `json:"immediate"`                                                                      // Whether to execute immediately
	ExecuteTime int64     `gorm:"index" json:"execute_time"`                                                      // Execution time timestamp
	CronExpr    string    `gorm:"size:128" json:"cron_expr,omitempty"`                                            // Cron expression with size limit
	Payload     string    `gorm:"type:text" json:"payload"`                                                       // Task payload
	Status      string    `gorm:"index:idx_task_type_status;index:idx_task_project_status;size:32" json:"status"` // Status: Pending, Running, Completed, Error, Cancelled, Rescheduled
	TraceID     string    `gorm:"index;size:64" json:"trace_id"`                                                  // Trace ID with size limit
	GroupID     string    `gorm:"index;size:64" json:"group_id"`                                                  // Group ID with size limit
	ProjectID   *int      `gorm:"index:idx_task_project_status" json:"project_id,omitempty"`                      // Task can belong to a project (optional)
	CreatedAt   time.Time `gorm:"autoCreateTime;index" json:"created_at"`                                         // Creation time with index
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                               // Update time

	// Foreign key association
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

// FaultInjectionSchedule model
type FaultInjectionSchedule struct {
	ID            int        `gorm:"primaryKey;autoIncrement" json:"id"`                                                            // Unique identifier
	TaskID        string     `gorm:"index:idx_fault_task_status;index:idx_fault_task_type;size:64" json:"task_id"`                  // Associated task ID, add composite index
	FaultType     int        `gorm:"index:idx_fault_task_type;index:idx_fault_type_status" json:"fault_type"`                       // Fault type, add composite index
	DisplayConfig string     `json:"display_config"`                                                                                // User-facing display configuration
	EngineConfig  string     `json:"engine_config"`                                                                                 // System-facing runtime configuration
	PreDuration   int        `json:"pre_duration"`                                                                                  // Normal data duration
	StartTime     *time.Time `gorm:"index;check:start_time IS NULL OR end_time IS NULL OR start_time < end_time" json:"start_time"` // Expected fault start time, nullable with validation
	EndTime       *time.Time `gorm:"index" json:"end_time"`                                                                         // Expected fault end time, nullable
	Status        int        `gorm:"default:1;index:idx_fault_task_status;index:idx_fault_type_status" json:"status"`               // Status: -1:deleted 0:disabled 1:enabled
	Description   string     `gorm:"type:text" json:"description"`                                                                  // Description (optional field)
	Benchmark     string     `gorm:"index;size:128" json:"benchmark"`                                                               // Benchmark database, add index and size limit
	InjectionName string     `gorm:"unique,index;size:128;not null" json:"injection_name"`                                          // Name injected in k8s resources with size limit
	CreatedAt     time.Time  `gorm:"autoCreateTime;index" json:"created_at"`                                                        // Creation time, add time index
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`                                                              // Update time

	// Foreign key association
	Task *Task `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}

type Container struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`                                             // Unique identifier
	Type      string    `gorm:"index;not null;index:idx_container_unique,unique;size:64" json:"type"`           // Image type
	Name      string    `gorm:"index;not null;index:idx_container_unique,unique;size:128" json:"name"`          // Name with size limit
	Image     string    `gorm:"not null;index:idx_container_unique,unique;size:256" json:"image"`               // Image name with size limit
	Tag       string    `gorm:"not null;default:'latest';index:idx_container_unique,unique;size:64" json:"tag"` // Image tag with size limit
	Command   string    `gorm:"type:text" json:"command"`                                                       // Startup command
	EnvVars   string    `gorm:"default:''" json:"env_vars"`                                                     // List of environment variable names
	UserID    int       `gorm:"not null;index:idx_container_user" json:"user_id"`                               // Container must belong to a user
	IsPublic  bool      `gorm:"default:false;index:idx_container_visibility" json:"is_public"`                  // Whether publicly visible
	Status    int       `gorm:"default:1" json:"status"`                                                        // Status: -1:deleted 0:disabled 1:active
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`                                         // Creation time
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                               // Update time

	// Foreign key association
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type ExecutionResult struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`                                         // Unique identifier
	TaskID      *string   `gorm:"index:idx_exec_task_status;index:idx_exec_task_algo;size:64" json:"task_id"` // Associated task ID, add composite index, nullable
	AlgorithmID int       `gorm:"index:idx_exec_task_algo;index:idx_exec_algo_dataset" json:"algorithm_id"`   // Algorithm used, add composite index
	DatapackID  int       `gorm:"index:idx_exec_algo_dataset" json:"datapack_id"`                             // Data package identifier, add composite index
	Duration    int       `gorm:"default:0;index:idx_exec_duration" json:"duration"`                          // Execution duration
	Status      int       `gorm:"default:0;index:idx_exec_task_status" json:"status"`                         // Status: -1:deleted 0:initial 1:failed 2:success
	CreatedAt   time.Time `gorm:"autoCreateTime;index:idx_exec_created_at" json:"created_at"`                 // Creation time, add time index for partitioning
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                           // Update time

	// Foreign key association
	Task      *Task                   `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	Algorithm *Container              `gorm:"foreignKey:AlgorithmID" json:"algorithm,omitempty"`
	Datapack  *FaultInjectionSchedule `gorm:"foreignKey:DatapackID" json:"datapack,omitempty"`

	// Many-to-many relationship with labels
	Labels []Label `gorm:"many2many:execution_result_labels;" json:"labels,omitempty"`
}

type GranularityResult struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"` // Unique identifier
	ExecutionID int       `gorm:"index,unique" json:"execution_id"`   // Associated ExecutionResult ID
	Level       string    `json:"level"`                              // Granularity type (e.g., "service", "pod", "span", "metric")
	Result      string    `json:"result"`                             // Localization result, comma-separated
	Rank        int       `json:"rank"`                               // Ranking, representing top1, top2, etc.
	Confidence  float64   `json:"confidence"`                         // Confidence level (optional)
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`   // Creation time
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`   // Update time

	// Foreign key association
	Execution *ExecutionResult `gorm:"foreignKey:ExecutionID" json:"execution,omitempty"`
}

type Detector struct {
	ID                  int       `gorm:"primaryKey" json:"id"`                    // Unique identifier
	ExecutionID         int       `gorm:"index,unique" json:"execution_id"`        // ExecutionID is primary key
	SpanName            string    `gorm:"type:varchar(255)" json:"span_name"`      // SpanName database field type
	Issues              string    `gorm:"type:text" json:"issues"`                 // Issues field type is text
	AbnormalAvgDuration *float64  `gorm:"type:float" json:"abnormal_avg_duration"` // Average duration during abnormal period
	NormalAvgDuration   *float64  `gorm:"type:float" json:"normal_avg_duration"`   // Average duration during normal period
	AbnormalSuccRate    *float64  `gorm:"type:float" json:"abnormal_succ_rate"`    // Success rate during abnormal period
	NormalSuccRate      *float64  `gorm:"type:float" json:"normal_succ_rate"`      // Success rate during normal period
	AbnormalP90         *float64  `gorm:"type:float" json:"abnormal_p90"`          // P90 during abnormal period
	NormalP90           *float64  `gorm:"type:float" json:"normal_p90"`            // P90 during normal period
	AbnormalP95         *float64  `gorm:"type:float" json:"abnormal_p95"`          // P95 during abnormal period
	NormalP95           *float64  `gorm:"type:float" json:"normal_p95"`            // P95 during normal period
	AbnormalP99         *float64  `gorm:"type:float" json:"abnormal_p99"`          // P99 during abnormal period
	NormalP99           *float64  `gorm:"type:float" json:"normal_p99"`            // P99 during normal period
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`        // CreatedAt automatically set to current time
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updated_at"`        // UpdatedAt automatically updates time

	// Foreign key association
	Execution *ExecutionResult `gorm:"foreignKey:ExecutionID" json:"execution,omitempty"`
}

// Dataset table, is designed to store multiple versions of a dataset(a series of datapack). Only admin can create a dataset, so there is no user id foreign key.
type Dataset struct {
	ID          int    `gorm:"primaryKey;autoIncrement" json:"id"`                                                                  // Unique identifier
	Name        string `gorm:"not null;index:idx_dataset_name_version_status,unique;size:128" json:"name"`                          // Dataset name with size limit
	Version     string `gorm:"not null;default:'v1.0';index:idx_dataset_name_version_status,unique;size:32" json:"dataset_version"` // Dataset version with size limit
	Description string `gorm:"type:text" json:"description"`                                                                        // Dataset description
	Type        string `gorm:"index;size:64" json:"type"`                                                                           // Dataset type (e.g., "microservice", "database", "network")
	FileCount   int    `gorm:"default:0;check:file_count >= 0" json:"file_count"`                                                   // File count with validation
	DataSource  string `gorm:"type:text" json:"data_source"`                                                                        // Data source description
	Format      string `gorm:"default:'json';size:32" json:"format"`                                                                // Data format (json, csv, parquet, etc.)

	Status      int       `gorm:"default:1;index:idx_dataset_name_version_status,unique" json:"status"` // Status: -1:deleted 0:disabled 1:enabled
	IsPublic    bool      `gorm:"default:false;index:idx_dataset_visibility" json:"is_public"`          // Whether public
	DownloadURL string    `gorm:"size:512" json:"download_url,omitempty"`                               // Download link with size limit
	Checksum    string    `gorm:"type:varchar(64)" json:"checksum,omitempty"`                           // File checksum
	CreatedAt   time.Time `gorm:"autoCreateTime;index" json:"created_at"`                               // Creation time
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                     // Update time

	// Many-to-many relationships - use explicit intermediate tables for better control
	Labels          []Label                  `gorm:"many2many:dataset_labels;" json:"labels,omitempty"`
	FaultInjections []FaultInjectionSchedule `gorm:"many2many:dataset_fault_injections;foreignKey:ID;joinForeignKey:DatasetID;References:ID;joinReferences:FaultInjectionID" json:"fault_injections,omitempty"`
}

// Label table - Unified label management
type Label struct {
	ID          int    `gorm:"primaryKey;autoIncrement" json:"id"`                                               // Unique identifier
	Key         string `gorm:"column:label_key;not null;index:idx_label_key_value_status,unique" json:"key"`     // Label key
	Value       string `gorm:"column:label_value;not null;index:idx_label_key_value_status,unique" json:"value"` // Label value
	Category    string `gorm:"index" json:"category"`                                                            // Label category (dataset, fault_injection, algorithm, container, etc.)
	Description string `gorm:"type:text" json:"description"`                                                     // Label description
	Color       string `gorm:"type:varchar(7);default:'#1890ff'" json:"color"`                                   // Label color (hex format)
	IsSystem    bool   `gorm:"default:false;index" json:"is_system"`                                             // Whether system label
	Usage       int    `gorm:"column:usage_count;default:0;index" json:"usage"`                                  // Usage count

	Status    int       `gorm:"default:1;index:idx_label_key_value_status,unique" json:"status"` // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                                // Creation time
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                // Update time
}

// DatasetFaultInjection Many-to-many relationship table between Dataset and FaultInjectionSchedule
type DatasetFaultInjection struct {
	ID               int       `gorm:"primaryKey;autoIncrement" json:"id"`                                       // Unique identifier
	DatasetID        int       `gorm:"not null;index:idx_dataset_fault_unique,unique" json:"dataset_id"`         // Dataset ID
	FaultInjectionID int       `gorm:"not null;index:idx_dataset_fault_unique,unique" json:"fault_injection_id"` // Fault injection ID
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`                                         // Creation time
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                         // Update time

	// Foreign key association - keep explicit associations for manual queries
	Dataset                *Dataset                `gorm:"foreignKey:DatasetID" json:"dataset,omitempty"`
	FaultInjectionSchedule *FaultInjectionSchedule `gorm:"foreignKey:FaultInjectionID" json:"fault_injection,omitempty"`
}

// DatasetLabel Many-to-many relationship table between Dataset and Label
type DatasetLabel struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`                               // Unique identifier
	DatasetID int       `gorm:"not null;index:idx_dataset_label_unique,unique" json:"dataset_id"` // Dataset ID
	LabelID   int       `gorm:"not null;index:idx_dataset_label_unique,unique" json:"label_id"`   // Label ID
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                                 // Creation time
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                 // Update time

	// Foreign key association - keep explicit associations for manual queries
	Dataset *Dataset `gorm:"foreignKey:DatasetID" json:"dataset,omitempty"`
	Label   *Label   `gorm:"foreignKey:LabelID" json:"label,omitempty"`
}

// FaultInjectionLabel Many-to-many relationship table between FaultInjectionSchedule and Label
type FaultInjectionLabel struct {
	ID               int       `gorm:"primaryKey;autoIncrement" json:"id"`                                     // Unique identifier
	FaultInjectionID int       `gorm:"not null;index:idx_fault_label_unique,unique" json:"fault_injection_id"` // Fault injection ID
	LabelID          int       `gorm:"not null;index:idx_fault_label_unique,unique" json:"label_id"`           // Label ID
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`                                       // Creation time

	// Foreign key association
	FaultInjectionSchedule *FaultInjectionSchedule `gorm:"foreignKey:FaultInjectionID" json:"fault_injection,omitempty"`
	Label                  *Label                  `gorm:"foreignKey:LabelID" json:"label,omitempty"`
}

// ContainerLabel Many-to-many relationship table between Container and Label
type ContainerLabel struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`                                   // Unique identifier
	ContainerID int       `gorm:"not null;index:idx_container_label_unique,unique" json:"container_id"` // Container ID
	LabelID     int       `gorm:"not null;index:idx_container_label_unique,unique" json:"label_id"`     // Label ID
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`                                     // Creation time

	// Foreign key association
	Container *Container `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
	Label     *Label     `gorm:"foreignKey:LabelID" json:"label,omitempty"`
}

type ProjectContainer struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID   int       `gorm:"not null;index:idx_project_container_unique,unique" json:"project_id"`
	ContainerID int       `gorm:"not null;index:idx_project_container_unique,unique" json:"container_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Project   *Project   `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Container *Container `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
}

type ProjectDataset struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID int       `gorm:"not null;index:idx_project_dataset_unique,unique" json:"project_id"`
	DatasetID int       `gorm:"not null;index:idx_project_dataset_unique,unique" json:"dataset_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Dataset *Dataset `gorm:"foreignKey:DatasetID" json:"dataset,omitempty"`
}

// ProjectLabel Many-to-many relationship table between Project and Label
type ProjectLabel struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`                               // Unique identifier
	ProjectID int       `gorm:"not null;index:idx_project_label_unique,unique" json:"project_id"` // Project ID
	LabelID   int       `gorm:"not null;index:idx_project_label_unique,unique" json:"label_id"`   // Label ID
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                                 // Creation time

	// Foreign key association
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Label   *Label   `gorm:"foreignKey:LabelID" json:"label,omitempty"`
}

// User User table
type User struct {
	ID          int        `gorm:"primaryKey;autoIncrement" json:"id"`            // Unique identifier
	Username    string     `gorm:"unique;not null;index;size:64" json:"username"` // Username (unique) with size limit
	Email       string     `gorm:"unique;not null;index;size:128" json:"email"`   // Email (unique) with size limit
	Password    string     `gorm:"not null;size:255" json:"-"`                    // Password (not returned to frontend) with size limit
	FullName    string     `gorm:"not null;size:128" json:"full_name"`            // Full name with size limit
	Avatar      string     `gorm:"size:512" json:"avatar,omitempty"`              // Avatar URL with size limit
	Phone       string     `gorm:"index;size:32" json:"phone,omitempty"`          // Phone number
	Status      int        `gorm:"default:1;index" json:"status"`                 // Status: -1:deleted 0:disabled 1:enabled
	IsActive    bool       `gorm:"default:true;index" json:"is_active"`           // Whether active
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`                       // Last login time
	CreatedAt   time.Time  `gorm:"autoCreateTime;index" json:"created_at"`        // Creation time
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`              // Update time
}

// Role Role table
type Role struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`   // Unique identifier
	Name        string    `gorm:"unique;not null;index" json:"name"`    // Role name (unique)
	DisplayName string    `gorm:"not null" json:"display_name"`         // Display name
	Description string    `gorm:"type:text" json:"description"`         // Role description
	Type        string    `gorm:"default:'custom';index" json:"type"`   // Role type (system, custom)
	IsSystem    bool      `gorm:"default:false;index" json:"is_system"` // Whether system role
	Status      int       `gorm:"default:1;index" json:"status"`        // 0:disabled 1:enabled -1:deleted
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`     // Creation time
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`     // Update time
}

// Permission Permission table
type Permission struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`   // Unique identifier
	Name        string    `gorm:"unique;not null;index" json:"name"`    // Permission name (unique)
	DisplayName string    `gorm:"not null" json:"display_name"`         // Display name
	Description string    `gorm:"type:text" json:"description"`         // Permission description
	Action      string    `gorm:"not null;index" json:"action"`         // Action (read, write, delete, execute, etc.)
	ResourceID  int       `gorm:"index" json:"resource_id"`             // Associated resource ID
	IsSystem    bool      `gorm:"default:false;index" json:"is_system"` // Whether system permission
	Status      int       `gorm:"default:1;index" json:"status"`        // 0:disabled 1:enabled -1:deleted
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`     // Creation time
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`     // Update time

	// Foreign key association
	Resource *Resource `gorm:"foreignKey:ResourceID" json:"resource,omitempty"`
}

// Resource Resource table
type Resource struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`   // Unique identifier
	Name        string    `gorm:"unique;not null;index" json:"name"`    // Resource name (unique)
	DisplayName string    `gorm:"not null" json:"display_name"`         // Display name
	Description string    `gorm:"type:text" json:"description"`         // Resource description
	Type        string    `gorm:"not null;index" json:"type"`           // Resource type (table, api, function, etc.)
	Category    string    `gorm:"index" json:"category"`                // Resource category
	ParentID    *int      `gorm:"index" json:"parent_id,omitempty"`     // Parent resource ID (supports hierarchy)
	IsSystem    bool      `gorm:"default:false;index" json:"is_system"` // Whether system resource
	Status      int       `gorm:"default:1;index" json:"status"`        // 0:disabled 1:enabled -1:deleted
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`     // Creation time
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`     // Update time

	// Foreign key association
	Parent *Resource `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
}

// UserProject Many-to-many relationship table between User and Project (includes project-level permissions)
type UserProject struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`                              // Unique identifier
	UserID    int       `gorm:"not null;index:idx_user_project_unique,unique" json:"user_id"`    // User ID
	ProjectID int       `gorm:"not null;index:idx_user_project_unique,unique" json:"project_id"` // Project ID
	RoleID    int       `gorm:"index" json:"role_id"`                                            // Role ID in this project
	JoinedAt  time.Time `gorm:"autoCreateTime" json:"joined_at"`                                 // Join time
	Status    int       `gorm:"default:1;index" json:"status"`                                   // 0:disabled 1:enabled -1:quit
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                                // Creation time
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                // Update time

	// Foreign key association
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Role    *Role    `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// UserRole Many-to-many relationship table between User and global roles
type UserRole struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`                        // Unique identifier
	UserID    int       `gorm:"not null;index:idx_user_role_unique,unique" json:"user_id"` // User ID
	RoleID    int       `gorm:"not null;index:idx_user_role_unique,unique" json:"role_id"` // Role ID
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                          // Creation time
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`                          // Update time

	// Foreign key association
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role *Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// RolePermission Many-to-many relationship table between Role and Permission
type RolePermission struct {
	ID           int       `gorm:"primaryKey;autoIncrement" json:"id"`                                    // Unique identifier
	RoleID       int       `gorm:"not null;index:idx_role_permission_unique,unique" json:"role_id"`       // Role ID
	PermissionID int       `gorm:"not null;index:idx_role_permission_unique,unique" json:"permission_id"` // Permission ID
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`                                      // Creation time
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`                                      // Update time

	// Foreign key association
	Role       *Role       `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Permission *Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
}

// UserPermission User direct permission table (supplements role permissions, supports special permission assignment)
type UserPermission struct {
	ID           int        `gorm:"primaryKey;autoIncrement" json:"id"`                                    // Unique identifier
	UserID       int        `gorm:"not null;index:idx_user_permission_unique,unique" json:"user_id"`       // User ID
	PermissionID int        `gorm:"not null;index:idx_user_permission_unique,unique" json:"permission_id"` // Permission ID
	ProjectID    *int       `gorm:"index:idx_user_permission_unique,unique" json:"project_id,omitempty"`   // Project ID (project-level permission, empty means global permission)
	GrantType    string     `gorm:"default:'grant';index;size:16" json:"grant_type"`                       // Grant type: grant, deny
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`                                                  // Expiration time
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`                                      // Creation time
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`                                      // Update time

	// Foreign key association
	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Permission *Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
	Project    *Project    `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

// ExecutionResultLabel Many-to-many relationship table between ExecutionResult and Label
type ExecutionResultLabel struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`                                               // Unique identifier
	ExecutionID int       `gorm:"index:idx_exec_label_exec;index:idx_exec_label_unique,unique" json:"execution_id"` // Execution result ID
	LabelID     int       `gorm:"index:idx_exec_label_label;index:idx_exec_label_unique,unique" json:"label_id"`    // Label ID
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`                                                 // Creation time

	// Foreign key associations
	ExecutionResult *ExecutionResult `gorm:"foreignKey:ExecutionID" json:"execution_result,omitempty"` // Associated execution result
	Label           *Label           `gorm:"foreignKey:LabelID" json:"label,omitempty"`                // Associated label
}
