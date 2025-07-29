package consts

import "time"

// ResourceName 资源名称类型，用于权限检查
type ResourceName string

// 系统资源名称常量
const (
	ResourceProject        ResourceName = "project"         // 项目资源
	ResourceDataset        ResourceName = "dataset"         // 数据集资源
	ResourceFaultInjection ResourceName = "fault_injection" // 故障注入资源
	ResourceContainer      ResourceName = "container"       // 容器资源
	ResourceTask           ResourceName = "task"            // 任务资源
	ResourceUser           ResourceName = "user"            // 用户资源
	ResourceRole           ResourceName = "role"            // 角色资源
	ResourcePermission     ResourceName = "permission"      // 权限资源
)

// String 返回资源名称的字符串表示
func (r ResourceName) String() string {
	return string(r)
}

// ActionName 权限动作类型，用于权限检查
type ActionName string

// 系统权限动作常量
const (
	ActionRead    ActionName = "read"    // 读取权限
	ActionWrite   ActionName = "write"   // 写入权限
	ActionDelete  ActionName = "delete"  // 删除权限
	ActionExecute ActionName = "execute" // 执行权限
	ActionManage  ActionName = "manage"  // 管理权限
)

// String 返回动作名称的字符串表示
func (a ActionName) String() string {
	return string(a)
}

/*
权限检查使用示例：

// 基本权限检查（推荐使用类型安全的方式）
checker := repository.NewPermissionChecker(userID, nil)

// 使用类型安全的常量进行权限检查
canRead, err := checker.CanReadResource(consts.ResourceContainer)
canWrite, err := checker.CanWriteResource(consts.ResourceTask)
canDelete, err := checker.CanDeleteResource(consts.ResourceProject)

// 通用权限检查
hasPermission, err := checker.HasPermissionTyped(consts.ActionRead, consts.ResourceContainer)

// 兼容的字符串方式（不推荐，但仍然支持）
canRead, err := checker.CanRead("container")
*/

// 定义任务类型
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
	DatasetDeleted       = -1
	DatasetInitial       = 0
	DatasetInjectFailed  = 1
	DatasetInjectSuccess = 2
	DatasetBuildFailed   = 3
	DatasetBuildSuccess  = 4
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

// 不同任务类型的 Payload 键
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

// 环境变量名称
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

// Redis stream 频道和字段
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
	EventAlgoRunSucceed EventType = "algorithm.run.succeed"
	EventAlgoRunFailed  EventType = "algorithm.run.failed"

	EventAlgoResultCollection    EventType = "algorithm.collect_result"
	EventDatasetResultCollection EventType = "dataset.result.collection"
	EventDatasetNoAnomaly        EventType = "dataset.no_anomaly"
	EventDatasetNoConclusionFile EventType = "dataset.no_conclusion_file"
	EventDatasetBuildSucceed     EventType = "dataset.build.succeed"
	EventDatasetBuildFailed      EventType = "dataset.build.failed"

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

// K8s Job 名称
const (
	DatasetJobName = "dataset"
)

const (
	TaskCarrier  = "task_carrier"
	TraceCarrier = "trace_carrier"
	GroupCarrier = "group_carrier"
)

// K8s Label 字段
const (
	LabelTaskID    = "task_id"
	LabelTraceID   = "trace_id"
	LabelGroupID   = "group_id"
	LabelProjectID = "project_id"

	// CRD Label 字段
	LabelBenchmark   = "benchmark"
	LabelPreDuration = "pre_duration"

	// Job Label 字段
	LabelTaskType    = "task_type"
	LabelDataset     = "dataset"
	LabelExecutionID = "execution_id"
	LabelTimestamp   = "timestamp"
)

// sse 事件类型
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

// PermissionName type for permission constants
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

// RoleName type for role constants
type RoleName string

// Role constants for system roles
const (
	RoleSuperAdmin   RoleName = "super_admin"   // Super Admin
	RoleAdmin        RoleName = "admin"         // Admin
	RoleProjectAdmin RoleName = "project_admin" // Project Admin
	RoleDeveloper    RoleName = "developer"     // Developer
	RoleViewer       RoleName = "viewer"        // Viewer
)
