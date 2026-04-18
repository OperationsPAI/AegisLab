package consts

import (
	"encoding/json"
	"strconv"
	"time"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

const InitialFilename = "data.yaml"
const DetectorKey = "algo.detector"

const (
	DefaultContainerVersion = "v1.0.0"
	DefaultContainerTag     = "latest"
	DefaultInvalidID        = 0
	DefaultLabelUsage       = 1
	DefaultTimeUnit         = time.Minute
)

// monitoring related constants
const (
	// Set of all monitored namespaces
	NamespacesKey = "monitor:namespaces"
	// Hash pattern for namespace status (will be monitor:ns:{namespace})
	NamespaceKeyPattern = "monitor:ns:%s"
)

// configuration update notification constants
const (
	ConfigEtcdProducerPrefix    = "/rcabench/config/producer/"
	ConfigEtcdConsumerPrefix    = "/rcabench/config/consumer/"
	ConfigEtcdGlobalPrefix      = "/rcabench/config/global/"
	ConfigUpdateResponseChannel = "config:updates:response"
)

const Hybrid chaos.ChaosType = -1

// Evaluation type constants
const (
	EvalTypeDatapack = "datapack"
	EvalTypeDataset  = "dataset"
)

type DatapackSource string

const (
	DatapackSourceInjection DatapackSource = "injection"
	DatapackSourceManual    DatapackSource = "manual"
)

const (
	GroundtruthSourceAuto     = "auto"
	GroundtruthSourceManual   = "manual"
	GroundtruthSourceImported = "imported"
)

const (
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
	ConfigCategory
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

type ConfigHistoryChangeType int

const (
	ChangeTypeUpdate ConfigHistoryChangeType = iota
	ChangeTypeRollback
)

type ConfigHistoryChangeField int

const (
	ChangeFieldValue ConfigHistoryChangeField = iota
	ChangeFieldDescription
	ChangeFieldDefaultValue
	ChangeFieldMinValue
	ChangeFieldMaxValue
	ChangeFieldPattern
	ChangeFieldOptions
)

type ContainerType int

const (
	ContainerTypeAlgorithm ContainerType = iota
	ContainerTypeBenchmark
	ContainerTypePedestal
)

type ConfigScope int

const (
	ConfigScopeProducer ConfigScope = iota
	ConfigScopeConsumer
	ConfigScopeGlobal
)

type ConfigValueType int

const (
	ConfigValueTypeString ConfigValueType = iota
	ConfigValueTypeBool
	ConfigValueTypeInt
	ConfigValueTypeFloat
	ConfigValueTypeStringArray
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

type ValueDataType int

const (
	ValueDataTypeString ValueDataType = iota
	ValueDataTypeInt
	ValueDataTypeBool
	ValueDataTypeFloat
	ValueDataTypeArray
	ValueDataTypeObject
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

func (ds DatapackState) MarshalJSON() ([]byte, error) {
	return json.Marshal(GetDatapackStateName(ds))
}

type ExecutionState int

const (
	ExecutionInitial ExecutionState = iota
	ExecutionFailed
	ExecutionSuccess
)

func (es ExecutionState) MarshalJSON() ([]byte, error) {
	return json.Marshal(GetExecutionStateName(es))
}

func (es *ExecutionState) UnmarshalJSON(data []byte) error {
	var stateName string
	if err := json.Unmarshal(data, &stateName); err != nil {
		return err
	}
	*es = *GetExecutionStateByName(stateName)
	return nil
}

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
	PageSizeTiny   PageSize = 5
	PageSizeSmall  PageSize = 10
	PageSizeMedium PageSize = 20
	PageSizeLarge  PageSize = 50
	PageSizeXLarge PageSize = 100
)

type TaskExtra string

const (
	TaskExtraInjectionAlgorithms TaskExtra = "injection_algorithms"
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

type TraceType int

const (
	TraceTypeFullPipeline   TraceType = iota // Complete fault injection + algorithm execution pipeline
	TraceTypeFaultInjection                  // Fault injection workflow
	TraceTypeDatapackBuild                   // Datapack building workflow
	TraceTypeAlgorithmRun                    // Algorithm execution workflow
)

type TraceState int

const (
	TracePending   TraceState = iota // Trace is pending/not started
	TraceRunning                     // Trace is currently running
	TraceCompleted                   // Trace completed successfully
	TraceFailed                      // Trace failed with errors
)

func (t TraceState) MarshalBinary() ([]byte, error) {
	return []byte(strconv.Itoa(int(t))), nil
}

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
	InjectNodes       = "nodes"
	InjectNamespace   = "namespace"
	InjectPedestal    = "pedestal"
	InjectPedestalID  = "pedestal_id"
	InjectLabels      = "labels"
	InjectSystem      = "system"

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
	StreamTraceLogKey     = "trace:%s:log"
	StreamGroupLogKey     = "group:%s:log"
	NotificationStreamKey = "notifications:global"

	RdbEventTaskID   = "task_id"
	RdbEventTaskType = "task_type"
	RdbEventStatus   = "status"
	RdbEventFileName = "file_name"
	RdbEventLine     = "line"
	RdbEventName     = "name"
	RdbEventPayload  = "payload"
	RdbEventFn       = "function_name"

	RdbEventTraceID        = "trace_id"
	RdbEventTraceState     = "state"
	RdbEventTraceLastEvent = "last_event"
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
	EventRestartPedestalStarted   EventType = "restart.pedestal.started"
	EventRestartPedestalCompleted EventType = "restart.pedestal.completed"
	EventRestartPedestalFailed    EventType = "restart.pedestal.failed"

	EventFaultInjectionStarted   EventType = "fault.injection.started"
	EventFaultInjectionCompleted EventType = "fault.injection.completed"
	EventFaultInjectionFailed    EventType = "fault.injection.failed"

	EventAlgoRunStarted       EventType = "algorithm.run.started"
	EventAlgoRunSucceed       EventType = "algorithm.run.succeed"
	EventAlgoRunFailed        EventType = "algorithm.run.failed"
	EventAlgoResultCollection EventType = "algorithm.result.collection"
	EventAlgoNoResultData     EventType = "algorithm.no_result_data"

	EventDatapackBuildStarted     EventType = "datapack.build.started"
	EventDatapackBuildSucceed     EventType = "datapack.build.succeed"
	EventDatapackBuildFailed      EventType = "datapack.build.failed"
	EventDatapackResultCollection EventType = "datapack.result.collection"
	EventDatapackNoAnomaly        EventType = "datapack.no_anomaly"
	EventDatapackNoDetectorData   EventType = "datapack.no_detector_data"

	EventImageBuildStarted EventType = "image.build.started"
	EventImageBuildSucceed EventType = "image.build.succeed"
	EventImageBuildFailed  EventType = "image.build.failed"

	EventTaskStarted     EventType = "task.started"
	EventTaskStateUpdate EventType = "task.state.update"
	EventTaskRetryStatus EventType = "task.retry.status"
	EventTaskScheduled   EventType = "task.scheduled"

	EventNoNamespaceAvailable EventType = "no.namespace.available"
	EventNoTokenAvailable     EventType = "no.token.available"

	EventAcquireLock EventType = "acquire.lock"
	EventReleaseLock EventType = "release.lock"

	EventJobSucceed EventType = "k8s.job.succeed"
	EventJobFailed  EventType = "k8s.job.failed"
)

func (e EventType) String() string {
	return string(e)
}

func (e EventType) MarshalBinary() ([]byte, error) {
	return []byte(e), nil
}

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
	CRDLabelBatchID  = "batch_id"
	CRDLabelIsHybrid = "is_hybrid"

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

type LogLevel string

const (
	LogLevelError LogLevel = "error"
	LogLevelWarn  LogLevel = "warn"
	LogLevelInfo  LogLevel = "info"
	LogLevelDebug LogLevel = "debug"
)

type WSLogType string

// WebSocket log message types
const (
	WSLogTypeHistory  WSLogType = "history"
	WSLogTypeRealtime WSLogType = "realtime"
	WSLogTypeEnd      WSLogType = "end"
	WSLogTypeError    WSLogType = "error"
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
	DurationNodeKey = "0"
	SystemNodeKey   = "1"
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

const (
	URLPathID           = "id"
	URLPathUserID       = "user_id"
	URLPathRoleID       = "role_id"
	URLPathPermissionID = "permission_id"
	URLPathConfigID     = "config_id"
	URLPathContainerID  = "container_id"
	URLPathVersionID    = "version_id"
	URLPathDatasetID    = "dataset_id"
	URLPathProjectID    = "project_id"
	URLPathTeamID       = "team_id"
	URLPathTaskID       = "task_id"
	URLPathDatapackID   = "datapack_id"
	URLPathExecutionID  = "execution_id"
	URLPathAlgorithmID  = "algorithm_id"
	URLPathInjectionID  = "injection_id"
	URLPathTraceID      = "trace_id"
	URLPathGroupID      = "group_id"
	URLPathLabelID      = "label_id"
	URLPathResourceID   = "resource_id"
	URLPathName         = "name"
)

var AppID string
var InitialTime *time.Time
