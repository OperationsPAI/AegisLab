package database

import (
	"aegis/consts"
	"aegis/utils"
	"fmt"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"gorm.io/gorm"
)

// =====================================================================
// Core Entities
// =====================================================================

type Container struct {
	ID     int                  `gorm:"primaryKey;autoIncrement"`
	Name   string               `gorm:"index;not null;size:128"`
	Type   consts.ContainerType `gorm:"index;not null;size:64"`
	README string               `gorm:"type:mediumtext"`

	IsPublic  bool              `gorm:"not null;default:false;index"`
	Status    consts.StatusType `gorm:"not null;default:1;index"`
	CreatedAt time.Time         `gorm:"autoCreateTime;index"`
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`

	ActiveName string `gorm:"type:varchar(150) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN name ELSE NULL END) STORED;uniqueIndex:idx_active_container_name"`

	// Many-to-many relationship with labels
	Labels []Label `gorm:"many2many:container_labels"`
}

type ContainerVersion struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"not null;index;size:32;default:'v1.0.0'"`
	NameMajor int    `gorm:"index:idx_container_name_order"`
	NameMinor int    `gorm:"index:idx_container_name_order"`
	NamePatch int    `gorm:"index:idx_container_name_order"`

	GithubLink  string `gorm:"size:512"`
	Registry    string `gorm:"not null;default:'docker.io';index;size:64"`
	Namespace   string `gorm:"index;size:128"`
	Repository  string `gorm:"not null;index;size:128"`
	Tag         string `gorm:"not null;size:128"`
	Command     string `gorm:"type:text"`
	EnvVars     string `gorm:"type:text"`
	Usage       int    `gorm:"column:usage_count;default:0;index"`
	ContainerID int    `gorm:"not null;index"`
	UserID      int    `gorm:"not null;index"`

	Status    consts.StatusType `gorm:"not null;default:1;index"`
	CreatedAt time.Time         `gorm:"autoCreateTime;index"`
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`

	ActiveVersionKey string `gorm:"type:varchar(40) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN CONCAT(container_id, ':', name) ELSE NULL END) STORED;uniqueIndex:idx_active_version_unique"`

	ImageRef string `gorm:"-"`

	// Foreign key association
	Container *Container `gorm:"foreignKey:ContainerID"`
	User      *User      `gorm:"foreignKey:UserID"`

	// One-to-one relationship with HelmConfig
	HelmConfig *HelmConfig `gorm:"foreignKey:ContainerVersionID;references:ID"`
}

func (cv *ContainerVersion) BeforeCreate(tx *gorm.DB) error {
	if cv.Name != "" {
		major, minor, patch, err := utils.ParseSemanticVersion(cv.Name)
		if err != nil {
			return fmt.Errorf("invalid semantic version: %w", err)
		}

		cv.NameMajor = major
		cv.NameMinor = minor
		cv.NamePatch = patch
	}

	if cv.ImageRef != "" {
		registry, namespace, repository, tag, err := utils.ParseFullImageRefernce(cv.ImageRef)
		if err != nil {
			return fmt.Errorf("invalid image reference: %w", err)
		}

		cv.Registry = registry
		cv.Namespace = namespace
		cv.Repository = repository
		cv.Tag = tag
	}
	return nil
}

// AfterFind GORM hook - set the Image field after retrieving from DB
func (c *ContainerVersion) AfterFind(tx *gorm.DB) error {
	if c.Namespace == "" {
		c.ImageRef = fmt.Sprintf("%s/%s:%s", c.Registry, c.Repository, c.Tag)
	} else {
		c.ImageRef = fmt.Sprintf("%s/%s/%s:%s", c.Registry, c.Namespace, c.Repository, c.Tag)
	}
	return nil
}

type HelmConfig struct {
	ID int `gorm:"primaryKey;autoIncrement"` // Unique identifier

	// Helm chart information
	RepoURL   string `gorm:"not null;size:512"` // Repository URL
	RepoName  string `gorm:"size:128"`          // Repository name
	ChartName string `gorm:"not null;size:128"` // Helm chart name

	// Deployment configuration
	NsPrefix     string `gorm:"not null;size:64"` // Namespace prefix for deployments
	PortTemplate string `gorm:"size:32"`          // Port template for dynamic port assignment, e.g., "31%03d"
	Values       string `gorm:"type:longtext"`    // Helm values in JSON format

	ContainerVersionID int `gorm:"uniqueIndex"` // Associated ContainerVersion ID (one-to-one relationship)

	FullChart string `gorm:"-"` // Full chart reference (not stored in DB, used for display)

	// Foreign key association
	ContainerVersion *ContainerVersion `gorm:"foreignKey:ContainerVersionID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate GORM hook - validate NsPrefix before creating a new record
func (h *HelmConfig) BeforeCreate(tx *gorm.DB) error {
	if !utils.CheckNsPrefixExists(h.NsPrefix) {
		return fmt.Errorf("invalid namespace prefix: %s", h.NsPrefix)
	}

	return nil
}

func (h *HelmConfig) AfterFind(tx *gorm.DB) error {
	h.FullChart = fmt.Sprintf("%s/%s", h.RepoName, h.ChartName)
	return nil
}

// Dataset table, is designed to store multiple versions of a dataset(a series of datapack). Only admin can create a dataset, so there is no user id foreign key.
type Dataset struct {
	ID          int    `gorm:"primaryKey;autoIncrement"` // Unique identifier
	Name        string `gorm:"index;not null;size:128"`  // Dataset name with size limit
	Type        string `gorm:"index;not null;size:64"`   // Dataset type (e.g., "microservice", "database", "network")
	Description string `gorm:"type:mediumtext"`          // Dataset description

	IsPublic  bool              `gorm:"not null;default:false;index"` // Whether public
	Status    consts.StatusType `gorm:"not null;default:1;index"`     // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time         `gorm:"autoCreateTime;index"`         // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`               // Update time

	ActiveName string `gorm:"type:varchar(150) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN name ELSE NULL END) STORED;uniqueIndex:idx_active_dataset_name"`

	// Many-to-many relationships - use explicit intermediate tables for better control
	Labels []Label `gorm:"many2many:dataset_labels"`
}

type DatasetVersion struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"not null;index;size:32;default:'v1.0.0'"`
	NameMajor int    `gorm:"index:idx_container_name_order"`
	NameMinor int    `gorm:"index:idx_container_name_order"`
	NamePatch int    `gorm:"index:idx_container_name_order"`

	DownloadURL string `gorm:"size:512"`                        // Download link with size limit
	Checksum    string `gorm:"type:varchar(64)"`                // File checksum
	FileCount   int    `gorm:"default:0;check:file_count >= 0"` // File count with validation
	Format      string `gorm:"default:'json';size:32"`          // Data format (json, csv, parquet, etc.)
	DatasetID   int    `gorm:"not null;index"`                  // Associated Dataset ID
	UserID      int    `gorm:"not null;index"`                  // Creator User ID

	ActiveVersionKey string `gorm:"type:varchar(40) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN CONCAT(dataset_id, ':', name) ELSE NULL END) STORED;uniqueIndex:idx_active_version_unique"`

	Status    consts.StatusType `gorm:"not null;default:1;index"` // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time         `gorm:"autoCreateTime;index"`     // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`           // Update time

	// Foreign key association
	Dataset    *Dataset         `gorm:"foreignKey:DatasetID"`
	User       *User            `gorm:"foreignKey:UserID"`
	Injections []FaultInjection `gorm:"many2many:dataset_version_injections"`
}

func (dv *DatasetVersion) BeforeCreate(tx *gorm.DB) error {
	if dv.Name != "" {
		major, minor, patch, err := utils.ParseSemanticVersion(dv.Name)
		if err != nil {
			return fmt.Errorf("invalid semantic version: %w", err)
		}

		dv.NameMajor = major
		dv.NameMinor = minor
		dv.NamePatch = patch
	}
	return nil
}

type Project struct {
	ID          int    `gorm:"primaryKey"`
	Name        string `gorm:"unique,index;not null;size:128"` // Project name with size limit
	Description string `gorm:"type:text"`                      // Project description

	IsPublic  bool              `gorm:"not null;default:false;index:idx_project_visibility"` // Whether publicly visible
	Status    consts.StatusType `gorm:"not null;default:1;index"`                            // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time         `gorm:"autoCreateTime;index"`                                // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`                                      // Update time

	Containers []Container `gorm:"many2many:project_containers"`
	Datasets   []Dataset   `gorm:"many2many:project_datasets"`
	Labels     []Label     `gorm:"many2many:project_labels"`
}

// Label table - Unified label management
type Label struct {
	ID          int                  `gorm:"primaryKey;autoIncrement"`                           // Unique identifier
	Key         string               `gorm:"column:label_key;not null;type:varchar(20);index"`   // Label key
	Value       string               `gorm:"column:label_value;not null;type:varchar(64);index"` // Label value
	Category    consts.LabelCategory `gorm:"index"`                                              // Label category (dataset, fault_injection, algorithm, container, etc.)
	Description string               `gorm:"type:text"`                                          // Label description
	Color       string               `gorm:"type:varchar(7);default:'#1890ff'"`                  // Label color (hex format)
	Usage       int                  `gorm:"not null;column:usage_count;default:0;index"`        // Usage count

	IsSystem  bool              `gorm:"not null;default:false;index"` // Whether system label
	Status    consts.StatusType `gorm:"not null;default:1;index"`     // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time         `gorm:"autoCreateTime"`               // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`               // Update time

	ActiveKeyValue string `gorm:"type:varchar(100) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN CONCAT(label_key, ':', label_value) ELSE NULL END) STORED;uniqueIndex:idx_key_value_unique"`
}

// User table
type User struct {
	ID          int        `gorm:"primaryKey;autoIncrement"`       // Unique identifier
	Username    string     `gorm:"unique;not null;index;size:64"`  // Username (unique) with size limit
	Email       string     `gorm:"unique;not null;index;size:128"` // Email (unique) with size limit
	Password    string     `gorm:"not null;size:255"`              // Password (not returned to frontend) with size limit
	FullName    string     `gorm:"not null;size:128"`              // Full name with size limit
	Avatar      string     `gorm:"size:512"`                       // Avatar URL with size limit
	Phone       string     `gorm:"index;size:32"`                  // Phone number
	LastLoginAt *time.Time // Last login time

	IsActive  bool              `gorm:"not null;default:true;index"` // Whether active
	Status    consts.StatusType `gorm:"not null;default:1;index"`    // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time         `gorm:"autoCreateTime;index"`        // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`              // Update time

	ActiveUsername string `gorm:"type:varchar(64) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN username ELSE NULL END) STORED;uniqueIndex:idx_active_username"`
}

// BeforeCreate GORM hook - hash the password before creating a new user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u.Password = hashedPassword
	return nil
}

// Role table
type Role struct {
	ID          int    `gorm:"primaryKey;autoIncrement"` // Unique identifier
	Name        string `gorm:"not null;index;size:32"`   // Role name (unique)
	DisplayName string `gorm:"not null"`                 // Display name
	Description string `gorm:"type:text"`                // Role description

	IsSystem  bool              `gorm:"not null;default:false;index"` // Whether system role
	Status    consts.StatusType `gorm:"not null;default:1;index"`     // 0:disabled 1:enabled -1:deleted
	CreatedAt time.Time         `gorm:"autoCreateTime"`               // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`               // Update time

	ActiveName string `gorm:"type:varchar(32) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN name ELSE NULL END) STORED;uniqueIndex:idx_active_role_name"`
}

// Permission table
type Permission struct {
	ID          int    `gorm:"primaryKey;autoIncrement"` // Unique identifier
	Name        string `gorm:"not null;index;size:128"`  // Permission name (unique)
	DisplayName string `gorm:"not null"`                 // Display name
	Description string `gorm:"type:text"`                // Permission description
	Action      string `gorm:"not null;index"`           // Action (read, write, delete, execute, etc.)
	ResourceID  int    `gorm:"not null;index"`           // Associated resource ID

	IsSystem  bool              `gorm:"not null;default:false;index"` // Whether system permission
	Status    consts.StatusType `gorm:"not null;default:1;index"`     // 0:disabled 1:enabled -1:deleted
	CreatedAt time.Time         `gorm:"autoCreateTime"`               // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`               // Update time

	ActiveName string `gorm:"type:varchar(128) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN name ELSE NULL END) STORED;uniqueIndex:idx_active_permission_name"`

	// Foreign key association
	Resource *Resource `gorm:"foreignKey:ResourceID"`
}

// Resource table
type Resource struct {
	ID          int                     `gorm:"primaryKey;autoIncrement"`      // Unique identifier
	Name        consts.ResourceName     `gorm:"not null;uniqueIndex;size:64" ` // Resource name (unique)
	DisplayName string                  `gorm:"not null"`                      // Display name
	Description string                  `gorm:"type:text"`                     // Resource description
	Type        consts.ResourceType     `gorm:"not null;index"`                // Resource type (table, api, function, etc.)
	Category    consts.ResourceCategory `gorm:"not null;index"`                // Resource category
	ParentID    *int                    `gorm:"index"`                         // Parent resource ID (supports hierarchy)

	CreatedAt time.Time `gorm:"autoCreateTime"` // Creation time

	// Foreign key association
	Parent *Resource `gorm:"foreignKey:ParentID"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID                 int    `gorm:"primaryKey;autoIncrement" json:"id"`
	IPAddress          string `gorm:"not null;default:'127.0.0.1';index" json:"ip_address"` // IP address of the client
	UserAgent          string `gorm:"not null;type:text" json:"user_agent"`                 // User agent of the client
	Duration           int    `json:"duration"`                                             // Duration in milliseconds
	Action             string `gorm:"not null;index" json:"action"`                         // Action performed (CREATE, UPDATE, DELETE, etc.)
	Details            string `gorm:"type:text" json:"details"`                             // Additional details in JSON format
	ErrorMsg           string `gorm:"type:text" json:"error_msg,omitempty"`                 // Error message if state is FAILED
	UserID             int    `gorm:"not null;index" json:"user_id"`                        // User who performed the action (nullable for system actions)
	ResourceID         int    `gorm:"not null;index" json:"resource_id"`                    // ID of the affected resource
	ResourceInstanceID *int   `gorm:"index" json:"resource_instance_id,omitempty"`          // Actual business object ID (e.g., dataset_id=5, container_id=10)
	ResourceInstance   string `gorm:"type:varchar(128)" json:"resource_instance"`           // Composite identifier, e.g., "datasets:5", "containers:10"

	State     consts.AuditLogState `gorm:"not null;default:0;index" json:"state"`  // SUCCESS, FAILED, WARNING
	Status    consts.StatusType    `gorm:"not null;default:1;index" json:"status"` // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time            `gorm:"autoCreateTime;index" json:"created_at"` // When the action was performed

	// Foreign key association
	User     *User     `gorm:"foreignKey:UserID"`
	Resource *Resource `gorm:"foreignKey:ResourceID"`
}

// =====================================================================
// Business Entities
// =====================================================================

// Task model
type Task struct {
	ID          string          `gorm:"primaryKey;size:64"`         // Task ID with size limit
	Type        consts.TaskType `gorm:"index:idx_task_type_status"` // Task type with size limit
	Immediate   bool            // Whether to execute immediately
	ExecuteTime int64           `gorm:"index"`                                                      // Execution time timestamp
	CronExpr    string          `gorm:"size:128"`                                                   // Cron expression with size limit
	Payload     string          `gorm:"type:text"`                                                  // Task payload
	TraceID     string          `gorm:"index;size:64"`                                              // Trace ID with size limit
	GroupID     string          `gorm:"index;size:64"`                                              // Group ID with size limit
	ProjectID   int             `gorm:"index:idx_task_project_state;index:idx_task_project_status"` // Task can belong to a project (optional)

	State     consts.TaskState  `gorm:"not null;default:0;index:idx_task_type_state;index:idx_task_project_state"`   // Event type for the task Running
	Status    consts.StatusType `gorm:"not null;default:1;index:idx_task_type_status;index:idx_task_project_status"` // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time         `gorm:"autoCreateTime;index"`                                                        // Creation time with index
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`                                                              // Update time

	// Foreign key association
	Project *Project `gorm:"foreignKey:ProjectID"`

	// Many-to-many relationship with labels - 添加这一行
	Labels []Label `gorm:"many2many:task_labels"`

	// One-to-one back reference with cascade delete
	FaultInjection *FaultInjection `gorm:"foreignKey:TaskID;references:ID;constraint:OnDelete:CASCADE"`
	Execution      *Execution      `gorm:"foreignKey:TaskID;references:ID;constraint:OnDelete:CASCADE"`
}

// FaultInjectionSchedule model
type FaultInjection struct {
	ID            int             `gorm:"primaryKey;autoIncrement"`                                                    // Unique identifier
	Name          string          `gorm:"size:128;not null;uniqueIndex"`                                               // Schedule name, add unique index
	FaultType     chaos.ChaosType `gorm:"not null;index;index:idx_fault_type_state"`                                   // Fault type, add composite index
	Description   string          `gorm:"type:text"`                                                                   // Description
	DisplayConfig *string         `gorm:"type:longtext"`                                                               // User-facing display configuration
	EngineConfig  string          `gorm:"type:longtext;not null"`                                                      // System-facing runtime configuration
	PreDuration   int             `gorm:"not null"`                                                                    // Normal data duration
	StartTime     *time.Time      `gorm:"index;check:start_time IS NULL OR end_time IS NULL OR start_time < end_time"` // Expected fault start time, nullable with validation
	EndTime       *time.Time      `gorm:"index"`                                                                       // Expected fault end time, nullable
	BenchmarkID   int             `gorm:"not null;index"`                                                              // Associated benchmark ID, add index
	PedestalID    int             `gorm:"not null;index"`                                                              // Associated pedestal ID, add index
	TaskID        string          `gorm:"not null;uniqueIndex;size:64"`                                                // Associated task ID, add composite index

	State     consts.DatapackState `gorm:"not null;default:0;index;index:idx_fault_type_state"` // Datapack state
	Status    consts.StatusType    `gorm:"not null;default:1;index"`                            // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time            `gorm:"autoCreateTime;index"`                                // Creation time, add time index
	UpdatedAt time.Time            `gorm:"autoUpdateTime;index"`                                // Update time

	ActiveName string `gorm:"type:varchar(150) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN name ELSE NULL END) STORED;uniqueIndex:idx_active_injection_name"`

	// Foreign key association with cascade
	Benchmark *ContainerVersion `gorm:"foreignKey:BenchmarkID;constraint:OnDelete:RESTRICT"`
	Pedestal  *ContainerVersion `gorm:"foreignKey:PedestalID;constraint:OnDelete:RESTRICT"`
	Task      *Task             `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`
}

type Execution struct {
	ID          int     `gorm:"primaryKey;autoIncrement"`             // Unique identifier
	Duration    float64 `gorm:"not null;default:0;index"`             // Execution duration
	TaskID      string  `gorm:"not null;uniqueIndex;size:64"`         // Associated task ID, add composite index
	AlgorithmID int     `gorm:"not null;index:idx_exec_algo_dataset"` // Algorithm ID, add composite index
	DatapackID  int     `gorm:"not null;index:idx_exec_algo_dataset"` // Datapack identifier, add composite index
	DatasetID   int     `gorm:"not null;index"`                       // Dataset identifier

	State     consts.ExecuteState `gorm:"not null;default:0;index"` // Execution state
	Status    consts.StatusType   `gorm:"not null;default:1;index"` // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time           `gorm:"autoCreateTime"`           // CreatedAt automatically set to current time
	UpdatedAt time.Time           `gorm:"autoUpdateTime"`           // UpdatedAt automatically updates time

	// Foreign key association with cascade
	Task      *Task             `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`
	Algorithm *ContainerVersion `gorm:"foreignKey:AlgorithmID;constraint:OnDelete:RESTRICT"`
	Datapack  *FaultInjection   `gorm:"foreignKey:DatapackID;constraint:OnDelete:RESTRICT"`
	Dataset   *DatasetVersion   `gorm:"foreignKey:DatasetID;constraint:OnDelete:RESTRICT"`

	DetectorResults    []DetectorResult    `gorm:"foreignKey:ExecutionID"`
	GranularityResults []GranularityResult `gorm:"foreignKey:ExecutionID"`
}

type DetectorResult struct {
	ID                  int      `gorm:"primaryKey"`        // Unique identifier
	SpanName            string   `gorm:"type:varchar(255)"` // SpanName database field type
	Issues              string   `gorm:"type:text"`         // Issues field type is text
	AbnormalAvgDuration *float64 `gorm:"type:float"`        // Average duration during abnormal period
	NormalAvgDuration   *float64 `gorm:"type:float"`        // Average duration during normal period
	AbnormalSuccRate    *float64 `gorm:"type:float"`        // Success rate during abnormal period
	NormalSuccRate      *float64 `gorm:"type:float"`        // Success rate during normal period
	AbnormalP90         *float64 `gorm:"type:float"`        // P90 during abnormal period
	NormalP90           *float64 `gorm:"type:float"`        // P90 during normal period
	AbnormalP95         *float64 `gorm:"type:float"`        // P95 during abnormal period
	NormalP95           *float64 `gorm:"type:float"`        // P95 during normal period
	AbnormalP99         *float64 `gorm:"type:float"`        // P99 during abnormal period
	NormalP99           *float64 `gorm:"type:float"`        // P99 during normal period
	ExecutionID         int      `gorm:"uniqueIndex"`       // Associated Execution ID

	// Foreign key association
	Execution *Execution `gorm:"foreignKey:ExecutionID;constraint:OnDelete:CASCADE"`
}

type GranularityResult struct {
	ID          int     `gorm:"primaryKey;autoIncrement"`        // Unique identifier
	Level       string  `gorm:"not null;type:varchar(50);index"` // Granularity type (e.g., "service", "pod", "span", "metric")
	Result      string  // Localization result, comma-separated
	Rank        int     // Ranking, representing top1, top2, etc.
	Confidence  float64 // Confidence level (optional)
	ExecutionID int     `gorm:"index"` // Associated Execution ID

	// Foreign key association
	Execution *Execution `gorm:"foreignKey:ExecutionID;constraint:OnDelete:CASCADE"`
}

// =====================================================================
// Many-to-many Relationship Tables
// =====================================================================

// ContainerLabel Many-to-many relationship table between Container and Label
type ContainerLabel struct {
	ContainerID int       `gorm:"primaryKey"`     // Container ID
	LabelID     int       `gorm:"primaryKey"`     // Label ID
	CreatedAt   time.Time `gorm:"autoCreateTime"` // Creation time

	// Foreign key association
	Container *Container `gorm:"foreignKey:ContainerID"`
	Label     *Label     `gorm:"foreignKey:LabelID"`
}

// DatasetLabel Many-to-many relationship table between Dataset and Label
type DatasetLabel struct {
	DatasetID int       `gorm:"primaryKey"`     // Dataset ID
	LabelID   int       `gorm:"primaryKey"`     // Label ID
	CreatedAt time.Time `gorm:"autoCreateTime"` // Creation time

	// Foreign key association
	Dataset *Dataset `gorm:"foreignKey:DatasetID"`
	Label   *Label   `gorm:"foreignKey:LabelID"`
}

// ProjectLabel Many-to-many relationship table between Project and Label
type ProjectLabel struct {
	ProjectID int       `gorm:"primaryKey"`     // Project ID
	LabelID   int       `gorm:"primaryKey"`     // Label ID
	CreatedAt time.Time `gorm:"autoCreateTime"` // Creation time

	// Foreign key association
	Project *Project `gorm:"foreignKey:ProjectID"`
	Label   *Label   `gorm:"foreignKey:LabelID"`
}

// DatasetVersionInjection Many-to-many relationship table between DatasetVersion and FaultInjection
type DatasetVersionInjection struct {
	DatasetVersionID int       `gorm:"primaryKey"`
	InjectionID      int       `gorm:"primaryKey"`
	CreatedAt        time.Time `gorm:"autoCreateTime"` // Creation time

	// Foreign key associations
	DatasetVersion *DatasetVersion `gorm:"foreignKey:DatasetVersionID"`
	Injection      *FaultInjection `gorm:"foreignKey:InjectionID"`
}

// TaskLabel Many-to-many relationship table between Task and Label
type TaskLabel struct {
	TaskID    string    `gorm:"primaryKey;size:64"` // Task ID
	LabelID   int       `gorm:"primaryKey"`         // Label ID
	CreatedAt time.Time `gorm:"autoCreateTime"`     // Creation time

	// Foreign key association
	Task  *Task  `gorm:"foreignKey:TaskID"`
	Label *Label `gorm:"foreignKey:LabelID"`
}

// UserContainer Many-to-many relationship table between User and Container (includes container-level permissions)
type UserContainer struct {
	ID          int `gorm:"primaryKey;autoIncrement"` // Unique identifier
	UserID      int `gorm:"not null;index"`           // User ID
	ContainerID int `gorm:"not null;index"`           // Container ID
	RoleID      int `gorm:"index"`                    // Role ID for this container

	Status    consts.StatusType `gorm:"not null;default:1;index"` // 0:disabled 1:enabled -1:quit
	CreatedAt time.Time         `gorm:"autoCreateTime;index"`     // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`           // Update time

	ActiveUserContainer string `gorm:"type:varchar(32) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN CONCAT(user_id, ':', container_id, ':', role_id) ELSE NULL END) STORED;uniqueIndex:idx_user_container_unique"`

	// Foreign key association
	User      *User      `gorm:"foreignKey:UserID"`
	Container *Container `gorm:"foreignKey:ContainerID"`
	Role      *Role      `gorm:"foreignKey:RoleID"`
}

type UserDataset struct {
	ID        int `gorm:"primaryKey;autoIncrement"` // Unique identifier
	UserID    int `gorm:"not null;index"`           // User ID
	DatasetID int `gorm:"not null;index"`           // DatasetID
	RoleID    int `gorm:"index"`                    // Role ID for this dataset

	Status    consts.StatusType `gorm:"not null;default:1;index"` // 0:disabled 1:enabled -1:quit
	CreatedAt time.Time         `gorm:"autoCreateTime;index"`     // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`           // Update time

	ActiveUserDataset string `gorm:"type:varchar(32) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN CONCAT(user_id, ':', dataset_id, ':', role_id) ELSE NULL END) STORED;uniqueIndex:idx_user_dataset_unique"`

	// Foreign key association
	User    *User    `gorm:"foreignKey:UserID"`
	Dataset *Dataset `gorm:"foreignKey:DatasetID"`
	Role    *Role    `gorm:"foreignKey:RoleID"`
}

// UserProject Many-to-many relationship table between User and Project (includes project-level permissions)
type UserProject struct {
	ID        int `gorm:"primaryKey;autoIncrement"` // Unique identifier
	UserID    int `gorm:"not null;index"`           // User ID
	ProjectID int `gorm:"not null;index"`           // Project ID
	RoleID    int `gorm:"index"`                    // Role ID in this project

	Status    consts.StatusType `gorm:"not null;default:1;index"` // 0:disabled 1:enabled -1:quit
	CreatedAt time.Time         `gorm:"autoCreateTime;index"`     // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`           // Update time

	ActiveUserProject string `gorm:"type:varchar(32) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN CONCAT(user_id, ':', project_id, ':', role_id) ELSE NULL END) STORED;uniqueIndex:idx_user_project_unique"`

	// Foreign key association
	User    *User    `gorm:"foreignKey:UserID"`
	Project *Project `gorm:"foreignKey:ProjectID"`
	Role    *Role    `gorm:"foreignKey:RoleID"`
}

// UserRole Many-to-many relationship table between User and global roles
type UserRole struct {
	ID     int `gorm:"primaryKey;autoIncrement"`                   // Unique identifier
	UserID int `gorm:"not null;index:idx_user_role_unique,unique"` // User ID
	RoleID int `gorm:"not null;index:idx_user_role_unique,unique"` // Role ID

	CreatedAt time.Time `gorm:"autoCreateTime"` // Creation time
	UpdatedAt time.Time `gorm:"autoUpdateTime"` // Update time

	// Foreign key association
	User *User `gorm:"foreignKey:UserID"`
	Role *Role `gorm:"foreignKey:RoleID"`
}

// RolePermission Many-to-many relationship table between Role and Permission
type RolePermission struct {
	ID           int `gorm:"primaryKey;autoIncrement"`                        // Unique identifier
	RoleID       int `gorm:"not null;uniqueIndex:idx_role_permission_unique"` // Role ID
	PermissionID int `gorm:"not null;uniqueIndex:idx_role_permission_unique"` // Permission ID

	CreatedAt time.Time `gorm:"autoCreateTime"` // Creation time
	UpdatedAt time.Time `gorm:"autoUpdateTime"` // Update time

	// Foreign key association
	Role       *Role       `gorm:"foreignKey:RoleID"`
	Permission *Permission `gorm:"foreignKey:PermissionID"`
}

// UserPermission User direct permission table (supplements role permissions, supports special permission assignment)
type UserPermission struct {
	ID           int              `gorm:"primaryKey;autoIncrement"`                                                                                         // Unique identifier
	UserID       int              `gorm:"not null;uniqueIndex:idx_up_container_unique;uniqueIndex:idx_up_dataset_unique;uniqueIndex:idx_up_project_unique"` // User ID
	PermissionID int              `gorm:"not null;uniqueIndex:idx_up_container_unique;uniqueIndex:idx_up_dataset_unique;uniqueIndex:idx_up_project_unique"` // Permission ID
	GrantType    consts.GrantType `gorm:"default:'grant';index;size:16"`                                                                                    // Grant type: grant, deny
	ExpiresAt    *time.Time       // Expiration time
	ContainerID  *int             `gorm:"uniqueIndex:idx_up_container_unique"` // Container ID (container-level permission, empty means global or project-level permission)
	DatasetID    *int             `gorm:"uniqueIndex:idx_up_dataset_unique"`   // Dataset ID (dataset-level permission, empty means global or project-level permission)
	ProjectID    *int             `gorm:"uniqueIndex:idx_up_project_unique"`   // Project ID (project-level permission, empty means global permission)

	CreatedAt time.Time `gorm:"autoCreateTime"` // Creation time
	UpdatedAt time.Time `gorm:"autoUpdateTime"` // Update time

	// Foreign key association
	User       *User       `gorm:"foreignKey:UserID"`
	Permission *Permission `gorm:"foreignKey:PermissionID"`
	Container  *Container  `gorm:"foreignKey:ContainerID"`
	Dataset    *Dataset    `gorm:"foreignKey:DatasetID"`
	Project    *Project    `gorm:"foreignKey:ProjectID"`
}
