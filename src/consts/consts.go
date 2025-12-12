package consts

import (
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
)

const InitialFilename = "data.json"
const DetectorKey = "algo.detector"

const (
	DefaultContainerVersion = "v1.0.0"
	DefaultContainerTag     = "latest"
	DefaultInvalidID        = 0
	DefaultLabelUsage       = 1
	DefaultTimeUnit         = time.Minute
)

// ResourceName is the type for resource names, used for permission checks
type ResourceName string

// System resource name constants
const (
	ResourceSystem           ResourceName = "system"            // system resource
	ResourceAudit            ResourceName = "audit"             // audit resource
	ResourceContainer        ResourceName = "container"         // container resource
	ResourceContainerVersion ResourceName = "container_version" // container version resource
	ResourceDataset          ResourceName = "dataset"           // dataset resource
	ResourceDatasetVersion   ResourceName = "dataset_version"   // dataset version resource
	ResourceProject          ResourceName = "project"           // project resource
	ResourceLabel            ResourceName = "label"             // label resource
	ResourceUser             ResourceName = "user"              // user resource
	ResourceRole             ResourceName = "role"              // role resource
	ResourcePermission       ResourceName = "permission"        // permission resource
	ResourceTask             ResourceName = "task"              // task resource
	ResourceTrace            ResourceName = "trace"             // trace resource
	ResourceInjection        ResourceName = "injection"         // fault injection resource
	ResourceExecution        ResourceName = "execution"         // execution resource
)

type ResourceType int

const (
	ResourceTypeSystem ResourceType = iota
	ResourceTypeTable
)

type ResourceCategory int

const (
	ResourceCore ResourceCategory = iota
	ResourceAdmin
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

type LabelCategory int

const (
	SystemCategory LabelCategory = iota
	ContainerCategory
	DatasetCategory
	ProjectCategory
	InjectionCategory
	ExecutionCategory
)

const (
	LabelKeyTag = "tag"

	CustomLabelDescriptionTemplate = "Custom label '%s' created %s"
)

type AuditLogState int

const (
	AuditLogStateFailed AuditLogState = iota
	AuditLogStateSuccess
)

type BuildSourceType string

const (
	BuildSourceTypeFile   BuildSourceType = "file"
	BuildSourceTypeGitHub BuildSourceType = "github"
	BuildSourceTypeHarbor BuildSourceType = "harbor"
)

type ContainerType int

const (
	ContainerTypeAlgorithm ContainerType = iota
	ContainerTypeBenchmark
	ContainerTypePedestal
)

type ParameterType int

const (
	ParameterTypeFixed ParameterType = iota
	ParameterTypeDynamic
)

type ParameterCategory int

const (
	ParameterCategoryEnvVars ParameterCategory = iota
	ParameterCategoryHelmValues
)

type DatapackState int

const (
	DatapackInitial DatapackState = iota
	DatapackInjectFailed
	DatapackInjectSuccess
	DatapackBuildFailed
	DatapackBuildSuccess
	DatapackDetectorFailed
	DatapackDetectorSuccess
)

type ExecutionState int

const (
	ExecutionInitial ExecutionState = iota
	ExecutionFailed
	ExecutionSuccess
)

type FaultType chaos.ChaosType

type GrantType int

const (
	GrantTypeGrant GrantType = iota
	GrantTypeDeny
)

const (
	DetectorNoAnomaly = "no_anomaly"
)

type PageSize int

const (
	PageSizeSmall  PageSize = 10
	PageSizeMedium PageSize = 20
	PageSizeLarge  PageSize = 50
)

type TaskState int

const (
	TaskCancelled   TaskState = -2
	TaskError       TaskState = -1
	TaskPending     TaskState = 0
	TaskRescheduled TaskState = 1
	TaskRunning     TaskState = 2
	TaskCompleted   TaskState = 3
)

type TaskType int

const (
	TaskTypeBuildContainer TaskType = iota
	TaskTypeRestartPedestal
	TaskTypeFaultInjection
	TaskTypeRunAlgorithm
	TaskTypeBuildDatapack
	TaskTypeCollectResult
	TaskTypeCronJob
)

type StatusType int

// common status: 0:disabled 1:enabled -1:deleted
const (
	CommonDeleted  StatusType = -1
	CommonDisabled StatusType = 0
	CommonEnabled  StatusType = 1
)

const (
	TaskMsgCompleted string = "Task %s completed"
	TaskMsgFailed    string = "Task %s failed"
)

// Payload keys for different task types
const (
	BuildImageRef     = "image_ref"
	BuildSourcePath   = "source_path"
	BuildBuildOptions = "build_options"

	BuildOptionContextDir     = "context_dir"
	BuildOptionDockerfilePath = "dockerfile_path"
	BuildOptionTarget         = "target"
	BuildOptionBuildArgs      = "build_args"
	BuildOptionForceRebuild   = "force_rebuild"

	RestartPedestal      = "pedestal_version"
	RestartHelmConfig    = "helm_config"
	RestartIntarval      = "interval"
	RestartFaultDuration = "fault_duration"
	RestartInjectPayload = "inject_payload"

	InjectBenchmark   = "benchmark_version"
	InjectPreDuration = "pre_duration"
	InjectNode        = "node"
	InjectNamespace   = "namespace"
	InjectPedestalID  = "pedestal_id"
	InjectLabels      = "labels"

	BuildBenchmark        = "benchmark"
	BuildDatapack         = "datapack"
	BuildDatasetVersionID = "dataset_version_id"
	BuildLabels           = "labels"

	ExecuteAlgorithm        = "algorithm"
	ExecuteDatapack         = "datapack"
	ExecuteDatasetVersionID = "dataset_version_id"
	ExecuteEnvVars          = "env_vars"
	ExecuteLabels           = "labels"

	CollectAlgorithm   = "algorithm"
	CollectDatapack    = "datapack"
	CollectExecutionID = "execution_id"

	EvaluateLabel = "app_name"
	EvaluateLevel = "level"
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
	TokenWaitTimeout = 10

	RestartPedestalTokenBucket   = "token_bucket:restart_service"
	MaxTokensKeyRestartPedestal  = "rate_limiting.max_concurrent_restarts_pedestal"
	MaxConcurrentRestartPedestal = 2
	RestartPedestalServiceName   = "restart_pedestal"

	BuildContainerTokenBucket   = "token_bucket:build_container"
	MaxTokensKeyBuildContainer  = "rate_limiting.max_concurrent_build_container"
	MaxConcurrentBuildContainer = 3
	BuildContainerServiceName   = "build_container"

	// Algorithm execution rate limiting
	AlgoExecutionTokenBucket   = "token_bucket:algo_execution"
	MaxTokensKeyAlgoExecution  = "rate_limiting.max_concurrent_algo_execution"
	MaxConcurrentAlgoExecution = 5
	AlgoExecutionServiceName   = "algo_execution"
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

	EventTaskStateUpdate EventType = "task.state.update"
	EventTaskRetryStatus EventType = "task.retry.status"
	EventTaskStarted     EventType = "task.started"

	EventNoTokenAvailable         EventType = "no.token.available"
	EventNoNamespaceAvailable     EventType = "no.namespace.available"
	EventRestartPedestalStarted   EventType = "restart.pedestal.started"
	EventRestartPedestalCompleted EventType = "restart.pedestal.completed"
	EventRestartPedestalFailed    EventType = "restart.pedestal.failed"

	EventFaultInjectionStarted   EventType = "fault.injection.started"
	EventFaultInjectionCompleted EventType = "fault.injection.completed"
	EventFaultInjectionFailed    EventType = "fault.injection.failed"

	EventAcquireLock EventType = "acquire.lock"
	EventReleaseLock EventType = "release.lock"

	EventJobSucceed EventType = "k8s.job.succeed"
	EventJobFailed  EventType = "k8s.job.failed"
)

const (
	TaskCarrier  = "task_carrier"
	TraceCarrier = "trace_carrier"
	GroupCarrier = "group_carrier"
)

// K8s fields
const (
	// Annotation fields
	CRDAnnotationBenchmark = "benchmark"
	JobAnnotationAlgorithm = "algorithm"
	JobAnnotationDatapack  = "datapack"

	K8sLabelAppID = "rcabench_app_id"

	// CRD label fields
	CRDLabelInjectionID = "injection_id"

	// Job label common fields
	JobLabelName      = "job-name"
	JobLabelTaskID    = "task_id"
	JobLabelTraceID   = "trace_id"
	JobLabelGroupID   = "group_id"
	JobLabelProjectID = "project_id"
	JobLabelUserID    = "user_id"
	JobLabelTaskType  = "task_type"

	// Job label custom fields
	JobLabelDatapack    = "datapack"
	JobLabelDatasetID   = "dataset_id"
	JobLabelExecutionID = "execution_id"
	JobLabelTimestamp   = "timestamp"
)

type VolumeMountName string

const (
	VolumeMountKubeConfig        VolumeMountName = "kube_config"
	VolumeMountDataset           VolumeMountName = "dataset"
	VolumeMountExperimentStorage VolumeMountName = "experiment_storage"
)

type SSEEventName string

// SSE event types
const (
	EventEnd    SSEEventName = "end"
	EventUpdate SSEEventName = "update"
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
	// TaskStateKey is the key for the task status attribute.
	TaskStateKey = "task.task_state"
)

// RoleName is the type for role constants
type RoleName string

// Role constants for system roles
const (
	RoleSuperAdmin         RoleName = "super_admin"         // Super Admin
	RoleAdmin              RoleName = "admin"               // Admin
	RoleContainerAdmin     RoleName = "container_admin"     // Container Admin
	RoleContainerDeveloper RoleName = "container_developer" // Container Developer
	RoleContainerViewer    RoleName = "container_viewer"    // Container Viewer
	RoleDatasetAdmin       RoleName = "dataset_admin"       // Dataset Admin
	RoleDatasetDeveloper   RoleName = "dataset_developer"   // Dataset Developer
	RoleDatasetViewer      RoleName = "dataset_viewer"      // Dataset Viewer
	RoleProjectAdmin       RoleName = "project_admin"       // Project Admin
	RoleProjectDeveloper   RoleName = "project_developer"   // Project Developer
	RoleProjectViewer      RoleName = "project_viewer"      // Project Viewer
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

	PermissionReadContainerVersion   PermissionName = "read_container_version"   // Read container version
	PermissionWriteContainerVersion  PermissionName = "write_container_version"  // Write container version
	PermissionDeleteContainerVersion PermissionName = "delete_container_version" // Delete container version
	PermissionManageContainerVersion PermissionName = "manage_container_version" // Manage container version

	PermissionReadTask    PermissionName = "read_task"    // Read task
	PermissionWriteTask   PermissionName = "write_task"   // Write task
	PermissionDeleteTask  PermissionName = "delete_task"  // Delete task
	PermissionExecuteTask PermissionName = "execute_task" // Execute task

	PermissionReadRole       PermissionName = "read_role"       // Read role
	PermissionReadPermission PermissionName = "read_permission" // Read permission
)

const (
	URLPathID           = "id"
	URLPathUserID       = "user_id"
	URLPathRoleID       = "role_id"
	URLPathPermissionID = "permission_id"
	URLPathContainerID  = "container_id"
	URLPathVersionID    = "version_id"
	URLPathDatasetID    = "dataset_id"
	URLPathProjectID    = "project_id"
	URLPathTaskID       = "task_id"
	URLPathDatapackID   = "datapack_id"
	URLPathExecutionID  = "execution_id"
	URLPathAlgorithmID  = "algorithm_id"
	URLPathInjectionID  = "injection_id"
	URLPathTraceID      = "trace_id"
	URLPathLabelID      = "label_id"
	URLPathResourceID   = "resource_id"
	URLPathName         = "name"
)

var AppID string
var InitialTime *time.Time
