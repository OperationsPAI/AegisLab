package database

import (
	"fmt"
	"time"

	"aegis/consts"
	"aegis/utils"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"gorm.io/gorm"
)

// =====================================================================
// Dynamic Configuration Entities
// =====================================================================

// DynamicConfig stores configuration items that can be dynamically modified at runtime
type DynamicConfig struct {
	ID          int                    `gorm:"primaryKey;autoIncrement"`                                          // Unique identifier
	Key         string                 `gorm:"column:config_key;not null;index:idx_config_key_category;size:255"` // Configuration key (e.g., "debugging.enabled", "rate_limiting.max_concurrent_builds")
	Value       string                 `gorm:"column:config_value;not null;type:text"`                            // Current configuration value (stored as string, type conversion happens in application layer)
	ValueType   consts.ConfigValueType `gorm:"not null;default:0"`                                                // Value type: string, bool, int, float64, []string
	Category    string                 `gorm:"not null;size:64;index:idx_config_key_category;default:'app'"`      // Configuration category: app, system, monitor, rate_limiting, database, k8s, etc.
	Description string                 `gorm:"type:text"`                                                         // Human-readable description of what this configuration does
	IsSecret    bool                   `gorm:"not null;default:false"`                                            // Whether this is sensitive data (passwords, tokens, etc.)
	UpdatedBy   *int                   // User ID who last updated this config

	DefaultValue string   `gorm:"not null;type:text"` // Default value for this configuration
	MinValue     *float64 `gorm:"type:decimal(20,4)"` // Minimum value (for numeric types)
	MaxValue     *float64 `gorm:"type:decimal(20,4)"` // Maximum value (for numeric types)
	Pattern      string   `gorm:"size:512"`           // Regex pattern for string validation
	Options      string   `gorm:"type:text"`          // JSON array of allowed values (for enum-like configs)

	CreatedAt time.Time `gorm:"autoCreateTime"` // Creation time
	UpdatedAt time.Time `gorm:"autoUpdateTime"` // Last update time

	// Foreign key association
	UpdatedByUser *User `gorm:"foreignKey:UpdatedBy"`

	// One-to-many relationship with ConfigHistory
	History []ConfigHistory `gorm:"foreignKey:ConfigID"`

	// Many-to-many relationship with labels
	Labels []Label `gorm:"many2many:config_labels"`
}

// ConfigHistory records all changes made to configuration items
type ConfigHistory struct {
	ID               int                             `gorm:"primaryKey;autoIncrement"` // Unique identifier
	ChangeType       consts.ConfigHistoryChangeType  `gorm:"not null;default:0"`       // Type of change: update, create, delete, rollback
	ChangeField      consts.ConfigHistoryChangeField `gorm:"not null;default:0"`       // Specific field that was changed (if applicable)
	OldValue         string                          `gorm:"not null;type:text"`       // Previous value
	NewValue         string                          `gorm:"not null;type:text"`       // New value after change
	Reason           string                          `gorm:"type:text"`                // Reason for this change (provided by operator)
	ConfigID         int                             `gorm:"not null;index"`           // Foreign key to DynamicConfig
	RolledBackFromID *int                            // If this is a rollback, references the history entry being rolled back from

	OperatorID *int   // User ID who made this change
	IPAddress  string `gorm:"size:64"`  // IP address from which the change was made
	UserAgent  string `gorm:"size:512"` // User agent of the client making the change

	CreatedAt time.Time `gorm:"autoCreateTime;not null"` // When this change was made

	// Foreign key associations
	Config         *DynamicConfig `gorm:"foreignKey:ConfigID"`
	Operator       *User          `gorm:"foreignKey:OperatorID"`
	RolledBackFrom *ConfigHistory `gorm:"foreignKey:RolledBackFromID"`
}

// =====================================================================
// Core Entities
// =====================================================================

type Container struct {
	ID     int                  `gorm:"primaryKey;autoIncrement"`
	Name   string               `gorm:"not null;size:128"`
	Type   consts.ContainerType `gorm:"not null;size:64"`
	README string               `gorm:"type:mediumtext"`

	IsPublic  bool              `gorm:"not null;default:false"`
	Status    consts.StatusType `gorm:"not null;default:1;index:idx_container_status_type"`
	CreatedAt time.Time         `gorm:"autoCreateTime"`
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`

	ActiveName string `gorm:"type:varchar(150) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN name ELSE NULL END) VIRTUAL;uniqueIndex:idx_active_container_name"`

	// Many-to-many relationship with labels
	Versions []ContainerVersion `gorm:"foreignKey:ContainerID"`
	Labels   []Label            `gorm:"many2many:container_labels"`
}

func (c *Container) BeforeCreate(tx *gorm.DB) error {
	if c.Type == consts.ContainerTypePedestal {
		system := chaos.SystemType(c.Name)
		if !system.IsValid() {
			return fmt.Errorf("invalid pedestal name: %s", c.Name)
		}
	}
	return nil
}

type ContainerVersion struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"not null;size:32;default:'1.0.0'"`
	NameMajor int    `gorm:"index:idx_container_version_name_order"`
	NameMinor int    `gorm:"index:idx_container_version_name_order"`
	NamePatch int    `gorm:"index:idx_container_version_name_order"`

	GithubLink  string `gorm:"size:512"`
	Registry    string `gorm:"not null;default:'docker.io';size:64"`
	Namespace   string `gorm:"size:128"`
	Repository  string `gorm:"not null;size:128"`
	Tag         string `gorm:"not null;size:128"`
	Command     string `gorm:"type:text"`
	Usage       int    `gorm:"column:usage_count;check:usage_count >= 0;default:0"`
	ContainerID int    `gorm:"not null;index:idx_cv_container_status"`
	UserID      int    `gorm:"not null;index"`

	Status    consts.StatusType `gorm:"not null;default:1;index:idx_cv_container_status"`
	CreatedAt time.Time         `gorm:"autoCreateTime"`
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`

	ActiveVersionKey string `gorm:"type:varchar(40) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN CONCAT(container_id, ':', name) ELSE NULL END) VIRTUAL;uniqueIndex:idx_active_version_unique"`

	ImageRef string `gorm:"-"`

	// Foreign key association
	Container *Container `gorm:"foreignKey:ContainerID"`
	User      *User      `gorm:"foreignKey:UserID"`

	// One-to-one relationship with HelmConfig
	HelmConfig *HelmConfig `gorm:"foreignKey:ContainerVersionID;references:ID"`

	// Many-to-many relationship with ParameterConfig (for environment variables)
	EnvVars []ParameterConfig `gorm:"many2many:container_version_env_vars"`
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
	ID                 int    `gorm:"primaryKey;autoIncrement"` // Unique identifier
	RepoURL            string `gorm:"not null;size:512"`        // Repository URL
	RepoName           string `gorm:"size:128"`                 // Repository name
	ChartName          string `gorm:"not null;size:128"`        // Helm chart name
	ContainerVersionID int    `gorm:"not null;uniqueIndex"`     // Associated ContainerVersion ID (one-to-one relationship)
	ValueFile          string // Values file path

	FullChart string `gorm:"-"` // Full chart reference (not stored in DB, used for display)

	// Foreign key association
	ContainerVersion *ContainerVersion `gorm:"foreignKey:ContainerVersionID;constraint:OnDelete:CASCADE"`

	// Many-to-many relationship with ParameterConfig (for Helm values)
	DynamicValues []ParameterConfig `gorm:"many2many:helm_config_values"`
}

func (h *HelmConfig) AfterFind(tx *gorm.DB) error {
	h.FullChart = fmt.Sprintf("%s/%s", h.RepoName, h.ChartName)
	return nil
}

type ParameterConfig struct {
	ID             int                      `gorm:"primaryKey;autoIncrement"`
	Key            string                   `gorm:"column:config_key;not null;size:128;uniqueIndex:idx_unique_config"`
	Type           consts.ParameterType     `gorm:"not null;default:0;uniqueIndex:idx_unique_config"`
	Category       consts.ParameterCategory `gorm:"not null;uniqueIndex:idx_unique_config"`
	ValueType      consts.ValueDataType     `gorm:"not null;default:0"`
	Description    string                   `gorm:"type:text"`
	DefaultValue   *string                  `gorm:"type:text"`
	TemplateString *string                  `gorm:"type:text"`
	Required       bool                     `gorm:"not null;default:false"`
	Overridable    bool                     `gorm:"not null;default:true"`
}

func (p *ParameterConfig) BeforeCreate(tx *gorm.DB) error {
	switch p.Type {
	case consts.ParameterTypeFixed:
		if p.Required && p.DefaultValue == nil {
			return fmt.Errorf("default value is required for fixed parameters")
		}
	case consts.ParameterTypeDynamic:
		if p.TemplateString == nil || *p.TemplateString == "" {
			return fmt.Errorf("template string is required for dynamic parameters")
		}
	}

	switch p.Category {
	case consts.ParameterCategoryEnvVars:
		if err := utils.IsValidEnvVar(p.Key); err != nil {
			return fmt.Errorf("invalid environment variable key: %w", err)
		}
	case consts.ParameterCategoryHelmValues:
		if err := utils.IsValidHelmValueKey(p.Key); err != nil {
			return fmt.Errorf("invalid helm value key: %w", err)
		}
	}
	return nil
}

// ContainerVersionEnvVar Many-to-many relationship table between ContainerVersion and ParameterConfig (for environment variables)
type ContainerVersionEnvVar struct {
	ContainerVersionID int       `gorm:"primaryKey"`     // ContainerVersion ID
	ParameterConfigID  int       `gorm:"primaryKey"`     // ParameterConfig ID
	CreatedAt          time.Time `gorm:"autoCreateTime"` // Removed index - many-to-many tables rarely queried by creation time

	// Foreign key association
	ContainerVersion *ContainerVersion `gorm:"foreignKey:ContainerVersionID"`
	ParameterConfig  *ParameterConfig  `gorm:"foreignKey:ParameterConfigID"`
}

// HelmConfigValue Many-to-many relationship table between HelmConfig and ParameterConfig (for Helm values)
type HelmConfigValue struct {
	HelmConfigID      int       `gorm:"primaryKey"`     // HelmConfig ID
	ParameterConfigID int       `gorm:"primaryKey"`     // ParameterConfig ID
	CreatedAt         time.Time `gorm:"autoCreateTime"` // Removed index - many-to-many tables rarely queried by creation time

	// Foreign key association
	HelmConfig      *HelmConfig      `gorm:"foreignKey:HelmConfigID"`
	ParameterConfig *ParameterConfig `gorm:"foreignKey:ParameterConfigID"`
}

type Dataset struct {
	ID          int    `gorm:"primaryKey;autoIncrement"` // Unique identifier
	Name        string `gorm:"not null;size:128"`        // Dataset name with size limit
	Type        string `gorm:"not null;size:64"`         // Dataset type (e.g., "microservice", "database", "network")
	Description string `gorm:"type:mediumtext"`          // Dataset description

	IsPublic  bool              `gorm:"not null;default:false"`   // Whether public
	Status    consts.StatusType `gorm:"not null;default:1;index"` // Status: -1:deleted 0:disaBled 1:enabled
	CreatedAt time.Time         `gorm:"autoCreateTime"`           // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`           // Update time

	ActiveName string `gorm:"type:varchar(150) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN name ELSE NULL END) VIRTUAL;uniqueIndex:idx_active_dataset_name"`

	// Many-to-many relationships - use explicit intermediate tables for better control
	Versions []DatasetVersion `gorm:"foreignKey:DatasetID"`
	Labels   []Label          `gorm:"many2many:dataset_labels"`
}

type DatasetVersion struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"not null;size:32;default:'1.0.0'"`
	NameMajor int    `gorm:"index:idx_dataset_version_name_order"`
	NameMinor int    `gorm:"index:idx_dataset_version_name_order"`
	NamePatch int    `gorm:"index:idx_dataset_version_name_order"`

	Checksum  string `gorm:"type:varchar(64)"`                         // File checksum
	FileCount int    `gorm:"not null;default:0;check:file_count >= 0"` // File count with validation
	DatasetID int    `gorm:"not null;index:idx_dv_dataset_status"`     // Associated Dataset ID
	UserID    int    `gorm:"not null;index"`                           // Creator User ID

	Status    consts.StatusType `gorm:"not null;default:1;index:idx_dv_dataset_status"` // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time         `gorm:"autoCreateTime"`                                 // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`                                 // Update time

	ActiveVersionKey string `gorm:"type:varchar(40) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN CONCAT(dataset_id, ':', name) ELSE NULL END) VIRTUAL;uniqueIndex:idx_active_version_unique"`

	// Foreign key association
	Dataset *Dataset `gorm:"foreignKey:DatasetID"`
	User    *User    `gorm:"foreignKey:UserID"`

	Datapacks []FaultInjection `gorm:"many2many:dataset_version_injections"`
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

	IsPublic  bool              `gorm:"not null;default:false"`   // Whether publicly visible
	Status    consts.StatusType `gorm:"not null;default:1;index"` // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time         `gorm:"autoCreateTime"`           // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`           // Update time

	Containers []Container `gorm:"many2many:project_containers"`
	Datasets   []Dataset   `gorm:"many2many:project_datasets"`
	Labels     []Label     `gorm:"many2many:project_labels"`
}

// Label table - Unified label management
type Label struct {
	ID          int                  `gorm:"primaryKey;autoIncrement"`                                                // Unique identifier
	Key         string               `gorm:"column:label_key;not null;type:varchar(64);index:idx_label_key_category"` // Label key
	Value       string               `gorm:"column:label_value;not null;type:varchar(64)"`                            // Label value
	Category    consts.LabelCategory `gorm:"index:idx_label_key_category"`                                            // Label category (dataset, fault_injection, algorithm, container, etc.)
	Description string               `gorm:"type:text"`                                                               // Label description
	Color       string               `gorm:"type:varchar(7);default:'#1890ff'"`                                       // Label color (hex format)
	Usage       int                  `gorm:"not null;column:usage_count;default:0;check:usage_count>=0"`              // Usage count

	IsSystem  bool              `gorm:"not null;default:false"`   // Whether system label
	Status    consts.StatusType `gorm:"not null;default:1;index"` // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time         `gorm:"autoCreateTime"`           // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`           // Update time

	ActiveKeyValue string `gorm:"type:varchar(100) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN CONCAT(label_key, ':', label_value) ELSE NULL END) VIRTUAL;uniqueIndex:idx_key_value_unique"`
}

// User table
type User struct {
	ID          int        `gorm:"primaryKey;autoIncrement"` // Unique identifier
	Username    string     `gorm:"unique;not null;size:64"`  // Username (unique) with size limit
	Email       string     `gorm:"unique;not null;size:128"` // Email (unique) with size limit
	Password    string     `gorm:"not null;size:255"`        // Password (not returned to frontend) with size limit
	FullName    string     `gorm:"not null;size:128"`        // Full name with size limit
	Avatar      string     `gorm:"size:512"`                 // Avatar URL with size limit
	Phone       string     `gorm:"size:32"`                  // Phone number
	LastLoginAt *time.Time // Last login time

	IsActive  bool              `gorm:"not null;default:true"`    // Whether active
	Status    consts.StatusType `gorm:"not null;default:1;index"` // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time         `gorm:"autoCreateTime"`           // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`           // Update time

	ActiveUsername string `gorm:"type:varchar(64) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN username ELSE NULL END) VIRTUAL;uniqueIndex:idx_active_username"`
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
	Name        string `gorm:"not null;size:32"`         // Role name (unique)
	DisplayName string `gorm:"not null"`                 // Display name
	Description string `gorm:"type:text"`                // Role description

	IsSystem  bool              `gorm:"not null;default:false"`   // Whether system role
	Status    consts.StatusType `gorm:"not null;default:1;index"` // 0:disabled 1:enabled -1:deleted
	CreatedAt time.Time         `gorm:"autoCreateTime"`           // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`           // Update time

	ActiveName string `gorm:"type:varchar(32) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN name ELSE NULL END) VIRTUAL;uniqueIndex:idx_active_role_name"`
}

// Permission table
type Permission struct {
	ID          int    `gorm:"primaryKey;autoIncrement"`                // Unique identifier
	Name        string `gorm:"not null;size:128"`                       // Permission name (unique)
	DisplayName string `gorm:"not null"`                                // Display name
	Description string `gorm:"type:text"`                               // Permission description
	Action      string `gorm:"not null;index:idx_perm_action_resource"` // Action (read, write, delete, execute, etc.)
	ResourceID  int    `gorm:"not null;index:idx_perm_action_resource"` // Associated resource ID

	IsSystem  bool              `gorm:"not null;default:false"`   // Whether system permission
	Status    consts.StatusType `gorm:"not null;default:1;index"` // 0:disabled 1:enabled -1:deleted
	CreatedAt time.Time         `gorm:"autoCreateTime"`           // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`           // Update time

	ActiveName string `gorm:"type:varchar(128) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN name ELSE NULL END) VIRTUAL;uniqueIndex:idx_active_permission_name"`

	// Foreign key association
	Resource *Resource `gorm:"foreignKey:ResourceID"`
}

// Resource table
type Resource struct {
	ID          int                     `gorm:"primaryKey;autoIncrement"`               // Unique identifier
	Name        consts.ResourceName     `gorm:"not null;uniqueIndex;size:64" `          // Resource name (unique)
	DisplayName string                  `gorm:"not null"`                               // Display name
	Description string                  `gorm:"type:text"`                              // Resource description
	Type        consts.ResourceType     `gorm:"not null"`                               // Resource type (table, api, function, etc.)
	Category    consts.ResourceCategory `gorm:"not null;index:idx_res_category_parent"` // Resource category
	ParentID    *int                    `gorm:"index:idx_res_category_parent"`          // Parent resource ID (supports hierarchy)

	CreatedAt time.Time `gorm:"autoCreateTime"` // Creation time

	// Foreign key association
	Parent *Resource `gorm:"foreignKey:ParentID"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID                 int    `gorm:"primaryKey;autoIncrement" json:"id"`
	IPAddress          string `gorm:"not null;default:'127.0.0.1'" json:"ip_address"`      // IP address of the client
	UserAgent          string `gorm:"not null;type:text" json:"user_agent"`                // User agent of the client
	Duration           int    `json:"duration"`                                            // Duration in milliseconds
	Action             string `gorm:"not null;index:idx_audit_action_state" json:"action"` // Action performed (CREATE, UPDATE, DELETE, etc.)
	Details            string `gorm:"type:text" json:"details"`                            // Additional details in JSON format
	ErrorMsg           string `gorm:"type:text" json:"error_msg,omitempty"`                // Error message if state is FAILED
	UserID             int    `gorm:"not null;index:idx_audit_user_time" json:"user_id"`   // User who performed the action (nullable for system actions)
	ResourceID         int    `gorm:"not null;index" json:"resource_id"`                   // ID of the affected resource
	ResourceInstanceID *int   `json:"resource_instance_id,omitempty"`                      // Actual business object ID (e.g., dataset_id=5, container_id=10)
	ResourceInstance   string `gorm:"type:varchar(128)" json:"resource_instance"`          // Composite identifier, e.g., "datasets:5", "containers:10"

	State     consts.AuditLogState `gorm:"not null;default:0;index:idx_audit_action_state" json:"state"` // SUCCESS, FAILED, WARNING
	Status    consts.StatusType    `gorm:"not null;default:1;index" json:"status"`                       // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time            `gorm:"autoCreateTime;index:idx_audit_user_time" json:"created_at"`   // When the action was performed

	// Foreign key association
	User     *User     `gorm:"foreignKey:UserID"`
	Resource *Resource `gorm:"foreignKey:ResourceID"`
}

// =====================================================================
// Business Entities
// =====================================================================

// Trace model - Represents execution flow of related tasks
type Trace struct {
	ID        string           `gorm:"primaryKey;size:64"`                  // Trace ID (unique identifier for a workflow)
	Type      consts.TraceType `gorm:"not null;index:idx_trace_type_state"` // Trace type (datapack_build, algorithm_run, full_pipeline)
	LastEvent consts.EventType `gorm:"size:128"`                            // Last event type received (for quick status check)
	StartTime time.Time        `gorm:"not null"`                            // Trace start time
	EndTime   *time.Time       // Trace end time (null if not completed)
	GroupID   string           `gorm:"index;size:64"`                 // Group ID for batch operations
	ProjectID int              `gorm:"index:idx_trace_project_state"` // Associated project (optional)

	LeafNum int `gorm:"not null;default:1"` // Number of leaf nodes in the trace DAG

	State     consts.TraceState `gorm:"not null;default:0;index:idx_trace_type_state;index:idx_trace_project_state"` // Trace state (pending, running, completed, failed)
	Status    consts.StatusType `gorm:"not null;default:1;index"`                                                    // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time         `gorm:"autoCreateTime"`                                                              // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`                                                              // Update time

	// Foreign key association
	Project *Project `gorm:"foreignKey:ProjectID"`

	// One-to-many relationship with tasks
	Tasks []Task `gorm:"foreignKey:TraceID;references:ID"`
}

// Task model
type Task struct {
	ID          string          `gorm:"primaryKey;size:64"`         // Task ID with size limit
	Type        consts.TaskType `gorm:"index:idx_task_type_status"` // Task type with size limit
	Immediate   bool            // Whether to execute immediately
	ExecuteTime int64           `gorm:"index"`                                                      // Execution time timestamp
	CronExpr    string          `gorm:"size:128"`                                                   // Cron expression with size limit
	Payload     string          `gorm:"type:text"`                                                  // Task payload
	TraceID     string          `gorm:"index;size:64"`                                              // Trace ID with size limit
	ProjectID   int             `gorm:"index:idx_task_project_state;index:idx_task_project_status"` // Task can belong to a project (optional)

	ParentTaskID *string `gorm:"index;size:64"`      // Parent task ID for sub-tasks
	Level        int     `gorm:"not null;default:0"` // Task level in the trace
	Sequence     int     `gorm:"not null;default:0"` // Task sequence in the trace

	State     consts.TaskState  `gorm:"not null;default:0;index:idx_task_type_state;index:idx_task_project_state"`   // Event type for the task Running
	Status    consts.StatusType `gorm:"not null;default:1;index:idx_task_type_status;index:idx_task_project_status"` // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time         `gorm:"autoCreateTime"`                                                              // Creation time with index
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`

	// Foreign key association
	Project    *Project `gorm:"foreignKey:ProjectID"`
	Trace      *Trace   `gorm:"foreignKey:TraceID;references:ID;constraint:OnDelete:CASCADE"`
	ParentTask *Task    `gorm:"foreignKey:ParentTaskID;references:ID;constraint:OnDelete:CASCADE"`

	// One-to-one back reference with cascade delete
	FaultInjection *FaultInjection `gorm:"foreignKey:TaskID;references:ID;constraint:OnDelete:CASCADE"`
	Execution      *Execution      `gorm:"foreignKey:TaskID;references:ID;constraint:OnDelete:CASCADE"`

	// One-to-many relationship with sub-tasks
	SubTasks []Task `gorm:"foreignKey:ParentTaskID;references:ID"`
}

// FaultInjectionSchedule model
type FaultInjection struct {
	ID            int              `gorm:"primaryKey;autoIncrement"`                                              // Unique identifier
	Name          string           `gorm:"size:128;not null"`                                                     // Schedule name, add unique index
	FaultType     chaos.ChaosType  `gorm:"not null;index:idx_fault_type_state"`                                   // Fault type, add composite index
	Category      chaos.SystemType `gorm:"not null"`                                                              // System category
	Description   string           `gorm:"type:text"`                                                             // Description
	DisplayConfig *string          `gorm:"type:longtext"`                                                         // User-facing display configuration
	EngineConfig  string           `gorm:"type:longtext;not null"`                                                // System-facing runtime configuration
	Groundtruths  []Groundtruth    `gorm:"type:json;serializer:json"`                                             // Expected impact groundtruth (service, pod, container, metric, function, span)
	PreDuration   int              `gorm:"not null"`                                                              // Normal data duration
	StartTime     *time.Time       `gorm:"check:start_time IS NULL OR end_time IS NULL OR start_time < end_time"` // Expected fault start time, nullable with validation
	EndTime       *time.Time       // Expected fault end time, nullable
	BenchmarkID   int              `gorm:"not null;index:idx_fault_bench_ped"` // Associated benchmark ID, add index
	PedestalID    int              `gorm:"not null;index:idx_fault_bench_ped"` // Associated pedestal ID, add index
	TaskID        *string          `gorm:"index;size:64"`                      // Associated task ID, add composite index

	State     consts.DatapackState `gorm:"not null;default:0;index:idx_fault_type_state"` // Datapack state
	Status    consts.StatusType    `gorm:"not null;default:1;index"`                      // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time            `gorm:"autoCreateTime"`                                // Creation time, add time index
	UpdatedAt time.Time            `gorm:"autoUpdateTime"`                                // Update time (removed index - rarely queried)

	// Foreign key association with cascade
	Benchmark *ContainerVersion `gorm:"foreignKey:BenchmarkID;constraint:OnDelete:RESTRICT"`
	Pedestal  *ContainerVersion `gorm:"foreignKey:PedestalID;constraint:OnDelete:RESTRICT"`
	Task      *Task             `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`

	// Many-to-many relationship with labels
	Labels []Label `gorm:"many2many:fault_injection_labels"`
}

type Execution struct {
	ID                 int     `gorm:"primaryKey;autoIncrement"`              // Unique identifier
	Duration           float64 `gorm:"not null;default:0"`                    // Execution duration
	TaskID             *string `gorm:"index;size:64"`                         // Associated task ID, add composite index
	AlgorithmVersionID int     `gorm:"not null;index:idx_exec_algo_datapack"` // Algorithm ID, add composite index
	DatapackID         int     `gorm:"not null;index:idx_exec_algo_datapack"` // Datapack identifier, add composite index
	DatasetVersionID   *int    // Dataset identifier (optional, for dataset-based executions)

	State     consts.ExecutionState `gorm:"not null;default:0;index"` // Execution state
	Status    consts.StatusType     `gorm:"not null;default:1;index"` // Status: -1:deleted 0:disabled 1:enabled
	CreatedAt time.Time             `gorm:"autoCreateTime"`           // CreatedAt automatically set to current time
	UpdatedAt time.Time             `gorm:"autoUpdateTime"`           // UpdatedAt automatically updates time

	// Foreign key association with cascade
	Task             *Task             `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`
	AlgorithmVersion *ContainerVersion `gorm:"foreignKey:AlgorithmVersionID;constraint:OnDelete:RESTRICT"`
	Datapack         *FaultInjection   `gorm:"foreignKey:DatapackID;constraint:OnDelete:RESTRICT"`
	DatasetVersion   *DatasetVersion   `gorm:"foreignKey:DatasetVersionID;constraint:OnDelete:SET NULL"`

	DetectorResults    []DetectorResult    `gorm:"foreignKey:ExecutionID"`
	GranularityResults []GranularityResult `gorm:"foreignKey:ExecutionID"`

	// Many-to-many relationship with labels
	Labels []Label `gorm:"many2many:execution_injection_labels"`
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
	ExecutionID         int      // Associated Execution ID

	// Foreign key association
	Execution *Execution `gorm:"foreignKey:ExecutionID;constraint:OnDelete:CASCADE"`
}

type GranularityResult struct {
	ID          int     `gorm:"primaryKey;autoIncrement"`  // Unique identifier
	Level       string  `gorm:"not null;type:varchar(50)"` // Granularity type (e.g., "service", "pod", "span", "metric")
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

// FaultInjectionLabel Many-to-many relationship table between FaultInjection and Label
type FaultInjectionLabel struct {
	FaultInjectionID int `gorm:"primaryKey"` // Fault injection ID
	LabelID          int `gorm:"primaryKey"` // Label ID

	// Foreign key association
	FaultInjection *FaultInjection `gorm:"foreignKey:FaultInjectionID;constraint:OnDelete:CASCADE"`
	Label          *Label          `gorm:"foreignKey:LabelID"`
}

// ExecutionInjectionLabel Many-to-many relationship table between Execution and Label
type ExecutionInjectionLabel struct {
	ExecutionID int       `gorm:"primaryKey"`     // Execution ID
	LabelID     int       `gorm:"primaryKey"`     // Label ID
	CreatedAt   time.Time `gorm:"autoCreateTime"` // Creation time

	// Foreign key association
	Execution *Execution `gorm:"foreignKey:ExecutionID;constraint:OnDelete:CASCADE"`
	Label     *Label     `gorm:"foreignKey:LabelID"`
}

// ConfigLabel Many-to-many relationship table between DynamicConfig and Label
type ConfigLabel struct {
	ConfigID  int       `gorm:"primaryKey"`     // Config ID
	LabelID   int       `gorm:"primaryKey"`     // Label ID
	CreatedAt time.Time `gorm:"autoCreateTime"` // Creation time

	// Foreign key association
	Config *DynamicConfig `gorm:"foreignKey:ConfigID"`
	Label  *Label         `gorm:"foreignKey:LabelID"`
}

// UserContainer Many-to-many relationship table between User and Container (includes container-level permissions)
type UserContainer struct {
	ID          int `gorm:"primaryKey;autoIncrement"` // Unique identifier
	UserID      int `gorm:"not null;index"`           // User ID
	ContainerID int `gorm:"not null;index"`           // Container ID
	RoleID      int // Role ID for this container

	Status    consts.StatusType `gorm:"not null;default:1;index"` // 0:disabled 1:enabled -1:quit
	CreatedAt time.Time         `gorm:"autoCreateTime"`           // Removed index - join table rarely queried by creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`           // Update time

	ActiveUserContainer string `gorm:"type:varchar(32) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN CONCAT(user_id, ':', container_id, ':', role_id) ELSE NULL END) VIRTUAL;uniqueIndex:idx_user_container_unique"`

	// Foreign key association
	User      *User      `gorm:"foreignKey:UserID"`
	Container *Container `gorm:"foreignKey:ContainerID"`
	Role      *Role      `gorm:"foreignKey:RoleID"`
}

type UserDataset struct {
	ID        int `gorm:"primaryKey;autoIncrement"` // Unique identifier
	UserID    int `gorm:"not null;index"`           // User ID
	DatasetID int `gorm:"not null;index"`           // DatasetID
	RoleID    int // Role ID for this dataset

	Status    consts.StatusType `gorm:"not null;default:1;index"` // 0:disabled 1:enabled -1:quit
	CreatedAt time.Time         `gorm:"autoCreateTime"`           // Removed index - join table rarely queried by creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`           // Update time

	ActiveUserDataset string `gorm:"type:varchar(32) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN CONCAT(user_id, ':', dataset_id, ':', role_id) ELSE NULL END) VIRTUAL;uniqueIndex:idx_user_dataset_unique"`

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
	RoleID    int // Role ID in this project

	Status    consts.StatusType `gorm:"not null;default:1;index"` // 0:disabled 1:enabled -1:quit
	CreatedAt time.Time         `gorm:"autoCreateTime"`           // Creation time
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`           // Update time

	ActiveUserProject string `gorm:"type:varchar(32) GENERATED ALWAYS AS (CASE WHEN status >= 0 THEN CONCAT(user_id, ':', project_id, ':', role_id) ELSE NULL END) VIRTUAL;uniqueIndex:idx_user_project_unique"`

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
	GrantType    consts.GrantType `gorm:"default:0;size:16"`                                                                                                // Grant type: grant, deny
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
