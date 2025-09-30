package consts

import "time"

// ResourceName is the type for resource names, used for permission checks
type ResourceName string

// System resource name constants
const (
	ResourceProject        ResourceName = "project"         // project resource
	ResourceDataset        ResourceName = "dataset"         // dataset resource
	ResourceFaultInjection ResourceName = "fault_injection" // fault injection resource
	ResourceContainer      ResourceName = "container"       // container resource
	ResourceTask           ResourceName = "task"            // task resource
	ResourceUser           ResourceName = "user"            // user resource
	ResourceRole           ResourceName = "role"            // role resource
	ResourcePermission     ResourceName = "permission"      // permission resource
	ResourceLabel          ResourceName = "label"           // label resource
	ResourceSystem         ResourceName = "system"          // system resource
	ResourceAudit          ResourceName = "audit"           // audit resource
)

// String returns the string representation of the resource name
func (r ResourceName) String() string {
	return string(r)
}

// ActionName is the type for permission action names, used for permission checks
type ActionName string

// System permission action constants
const (
	ActionRead    ActionName = "read"    // read action
	ActionWrite   ActionName = "write"   // write action
	ActionDelete  ActionName = "delete"  // delete action
	ActionExecute ActionName = "execute" // execute action
	ActionManage  ActionName = "manage"  // manage action

	// Execution result label keys
	ExecutionLabelSource = "source"

	// Execution result label values
	ExecutionSourceManual = "manual" // User manually uploaded
	ExecutionSourceSystem = "system" // RCABench internally managed

	ExecutionManualDescription = "Manual execution result created via API"
	ExecutionSystemDescription = "System-managed execution result created by RCABench"
)

// Database Label
const (
	LabelExecution = "execution"
	LabelInjection = "injection"
	LabelSystem    = "system"
)

// Injection label keys
const (
	LabelKeyTag   = "tag"   // User-defined tag for injection
	LabelKeyEnv   = "env"   // Environment label key
	LabelKeyBatch = "batch" // Batch label key
)

// Custom label description templates
const (
	CustomLabelDescriptionTemplate = "Custom label '%s' created for injection"
)

// String returns the string representation of the action name
func (a ActionName) String() string {
	return string(a)
}

/*
Permission check usage example:

// Basic permission check (type-safe approach recommended)
checker := repository.NewPermissionChecker(userID, nil)

// Use type-safe constants for permission checks
canRead, err := checker.CanReadResource(consts.ResourceContainer)
canWrite, err := checker.CanWriteResource(consts.ResourceTask)
canDelete, err := checker.CanDeleteResource(consts.ResourceProject)

// General permission check
hasPermission, err := checker.HasPermissionTyped(consts.ActionRead, consts.ResourceContainer)

// Compatible string approach (not recommended, but still supported)
canRead, err := checker.CanRead("container")
*/

// Define task types
type TaskType string

type ContainerType string

const (
	ContainerTypeAlgorithm ContainerType = "algorithm"
	ContainerTypeBenchmark ContainerType = "benchmark"
	ContainerTypeNamespace ContainerType = "namespace"
)

type BuildSourceType string

const (
	BuildSourceTypeFile   BuildSourceType = "file"
	BuildSourceTypeGitHub BuildSourceType = "github"
	BuildSourceTypeHarbor BuildSourceType = "harbor"
)

const (
	DefaultBenchmark = "clickhouse"
	DefaultTimeUnit  = time.Minute
)

const (
	DatapackDeleted       = -1
	DatapackInitial       = 0
	DatapackInjectFailed  = 1
	DatapackInjectSuccess = 2
	DatapackBuildFailed   = 3
	DatapackBuildSuccess  = 4
)

// project status: 0:disabled 1:enabled -1:deleted
const (
	ProjectDisabled = 0
	ProjectEnabled  = 1
	ProjectDeleted  = -1
)

// label status: 0:disabled 1:enabled -1:deleted
const (
	LabelDisabled = 0
	LabelEnabled  = 1
	LabelDeleted  = -1
)

// dataset status: 0:disabled 1:enabled -1:deleted
const (
	DatasetDisabled = 0
	DatasetEnabled  = 1
	DatasetDeleted  = -1
)

// container status: 0:disabled 1:enabled -1:deleted
const (
	ContainerDisabled = 0
	ContainerEnabled  = 1
	ContainerDeleted  = -1
)

const (
	ExecutionDeleted = -1
	ExecutionInitial = 0
	ExecutionFailed  = 1
	ExecutionSuccess = 2
)

const (
	TaskStatusCanceled    string = "Cancelled"
	TaskStatusCompleted   string = "Completed"
	TaskStatusError       string = "Error"
	TaskStatusPending     string = "Pending"
	TaskStatusRunning     string = "Running"
	TaskStautsRescheduled string = "Rescheduled"
)

const (
	TaskTypeRestartService TaskType = "RestartService"
	TaskTypeRunAlgorithm   TaskType = "RunAlgorithm"
	TaskTypeFaultInjection TaskType = "FaultInjection"
	TaskTypeBuildImage     TaskType = "BuildImage"
	TaskTypeBuildDataset   TaskType = "BuildDataset"
	TaskTypeCollectResult  TaskType = "CollectResult"
)

const (
	TaskMsgCompleted string = "Task %s completed"
	TaskMsgFailed    string = "Task %s failed"
)

// Payload keys for different task types
const (
	BuildBenchmark   = "benchmark"
	BuildDataset     = "name"
	BuildPreDuration = "pre_duration"
	BuildStartTime   = "start_time"
	BuildEndTime     = "end_time"
	BuildEnvVars     = "env_vars"

	BuildContainerType = "container_type"
	BuildName          = "name"
	BuildImage         = "image"
	BuildTag           = "tag"
	BuildCommand       = "command"
	BuildImageEnvVars  = "env_vars"
	BuildSourcePath    = "source_path"
	BuildBuildOptions  = "build_options"

	BuildOptionContextDir     = "context_dir"
	BuildOptionDockerfilePath = "dockerfile_path"
	BuildOptionTarget         = "target"
	BuildOptionBuildArgs      = "build_args"
	BuildOptionForceRebuild   = "force_rebuild"

	CollectAlgorithm   = "algorithm"
	CollectDataset     = "dataset"
	CollectExecutionID = "execution_id"
	CollectTimestamp   = "timestamp"

	EvaluateLabel = "app_name"
	EvaluateLevel = "level"

	AnnotationAlgorithm = "algorithm"
	ExecuteAlgorithm    = "algorithm"
	ExecuteDataset      = "dataset"
	ExecuteEnvVars      = "env_vars"

	InjectAlgorithms  = "algorithms"
	InjectBenchmark   = "benchmark"
	InjectFaultType   = "fault_type"
	InjectNamespace   = "namespace"
	InjectPreDuration = "pre_duration"
	InjectDisplayData = "display_data"
	InjectConf        = "conf"
	InjectNode        = "node"
	InjectLabels      = "labels"

	RestartIntarval      = "interval"
	RestartFaultDuration = "fault_duration"
	RestartInjectPayload = "inject_payload"
)

// Environment variable names
const (
	BuildEnvVarNamespace = "NAMESPACE"
)

const (
	HarborTimeout  = 30
	HarborTimeUnit = time.Second
)

const (
	InjectionAlgorithmsKey = "injection:algorithms"
)

// Redis stream channels and fields
const (
	StreamLogKey = "trace:%s:log"

	RdbEventTaskID   = "task_id"
	RdbEventTaskType = "task_type"
	RdbEventStatus   = "status"
	RdbEventFileName = "file_name"
	RdbEventLine     = "line"
	RdbEventName     = "name"
	RdbEventPayload  = "payload"
	RdbEventFn       = "function_name"
)

const (
	RestartServiceRateLimitKey = "rate_limit:restart_service"
	RestartServiceTokenBucket  = "token_bucket:restart_service"
	MaxConcurrentRestarts      = 2
	TokenWaitTimeout           = 10
	DelayRetryMinutes          = 5

	// Build container rate limiting
	BuildContainerTokenBucket = "token_bucket:build_container"
	MaxConcurrentBuilds       = 3

	// Algorithm execution rate limiting
	AlgoExecutionTokenBucket   = "token_bucket:algo_execution"
	MaxConcurrentAlgoExecution = 5
)

type EventType string

const (
	// when adding the consts, remember to update the consts in python sdk, const.py
	EventAlgoRunSucceed       EventType = "algorithm.run.succeed"
	EventAlgoRunFailed        EventType = "algorithm.run.failed"
	EventAlgoResultCollection EventType = "algorithm.result.collection"
	EventAlgoNoResultData     EventType = "algorithm.no_result_data"

	EventDatapackBuildSucceed     EventType = "datapack.build.succeed"
	EventDatapackBuildFailed      EventType = "datapack.build.failed"
	EventDatapackResultCollection EventType = "datapack.result.collection"
	EventDatapackNoAnomaly        EventType = "datapack.no_anomaly"
	EventDatapackNoDetectorData   EventType = "datapack.no_detector_data"

	EventImageBuildSucceed EventType = "image.build.succeed"

	EventTaskStatusUpdate EventType = "task.status.update"
	EventTaskRetryStatus  EventType = "task.retry.status"
	EventTaskStarted      EventType = "task.started"

	EventNoTokenAvailable        EventType = "no.token.available"
	EventNoNamespaceAvailable    EventType = "no.namespace.available"
	EventRestartServiceStarted   EventType = "restart.service.started"
	EventRestartServiceCompleted EventType = "restart.service.completed"
	EventRestartServiceFailed    EventType = "restart.service.failed"

	EventFaultInjectionStarted   EventType = "fault.injection.started"
	EventFaultInjectionCompleted EventType = "fault.injection.completed"
	EventFaultInjectionFailed    EventType = "fault.injection.failed"

	EventAcquireLock EventType = "acquire.lock"
	EventReleaseLock EventType = "release.lock"

	EventJobLogsRecorded EventType = "job.logs.recorded"
)

// K8s Job name
const (
	DatasetJobName = "dataset"
)

const (
	TaskCarrier  = "task_carrier"
	TraceCarrier = "trace_carrier"
	GroupCarrier = "group_carrier"
)

// K8s label fields
const (
	LabelTaskID    = "task_id"
	LabelTraceID   = "trace_id"
	LabelGroupID   = "group_id"
	LabelProjectID = "project_id"

	// CRD label fields
	LabelBenchmark   = "benchmark"
	LabelPreDuration = "pre_duration"

	// Job label fields
	LabelTaskType    = "task_type"
	LabelDataset     = "dataset"
	LabelExecutionID = "execution_id"
	LabelTimestamp   = "timestamp"
)

const (
	JobSucceed = "succeed"
	JobFailed  = "failed"
)

// SSE event types
const (
	EventEnd    = "end"
	EventUpdate = "update"
)

const (
	SpanStatusDescription = "task %s %s"
)

const (
	DownloadFilename       = "package"
	DetectorConclusionFile = "conclusion.csv"
	ExecutionResultFile    = "result.csv"
)

const (
	DurationNodeKey       = "0"
	NamespaceDefaultValue = 1
	NamespaceNodeKey      = "1"
)

// span attribute keys
const (
	// TaskIDKey is the key for the task ID attribute.
	TaskIDKey = "task.task_id"
	// TaskTypeKey is the key for the task type attribute.
	TaskTypeKey = "task.task_type"
	// TaskStatusKey is the key for the task status attribute.
	TaskStatusKey = "task.task_status"
)

// PermissionName is the type for permission constants
type PermissionName string

// Permission constants for system permissions
const (
	PermissionReadProject   PermissionName = "read_project"   // Read project
	PermissionWriteProject  PermissionName = "write_project"  // Write project
	PermissionDeleteProject PermissionName = "delete_project" // Delete project
	PermissionManageProject PermissionName = "manage_project" // Manage project

	PermissionReadDataset   PermissionName = "read_dataset"   // Read dataset
	PermissionWriteDataset  PermissionName = "write_dataset"  // Write dataset
	PermissionDeleteDataset PermissionName = "delete_dataset" // Delete dataset
	PermissionManageDataset PermissionName = "manage_dataset" // Manage dataset

	PermissionReadFaultInjection    PermissionName = "read_fault_injection"    // Read fault injection
	PermissionWriteFaultInjection   PermissionName = "write_fault_injection"   // Write fault injection
	PermissionDeleteFaultInjection  PermissionName = "delete_fault_injection"  // Delete fault injection
	PermissionExecuteFaultInjection PermissionName = "execute_fault_injection" // Execute fault injection

	PermissionReadContainer   PermissionName = "read_container"   // Read container
	PermissionWriteContainer  PermissionName = "write_container"  // Write container
	PermissionDeleteContainer PermissionName = "delete_container" // Delete container
	PermissionManageContainer PermissionName = "manage_container" // Manage container

	PermissionReadTask    PermissionName = "read_task"    // Read task
	PermissionWriteTask   PermissionName = "write_task"   // Write task
	PermissionDeleteTask  PermissionName = "delete_task"  // Delete task
	PermissionExecuteTask PermissionName = "execute_task" // Execute task

	PermissionReadRole       PermissionName = "read_role"       // Read role
	PermissionReadPermission PermissionName = "read_permission" // Read permission
)

// RoleName is the type for role constants
type RoleName string

// Role constants for system roles
const (
	RoleSuperAdmin   RoleName = "super_admin"   // Super Admin
	RoleAdmin        RoleName = "admin"         // Admin
	RoleProjectAdmin RoleName = "project_admin" // Project Admin
	RoleDeveloper    RoleName = "developer"     // Developer
	RoleViewer       RoleName = "viewer"        // Viewer
)
