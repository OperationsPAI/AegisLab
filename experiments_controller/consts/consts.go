package consts

import "time"

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
	HarborURL      = "http://%s/api/v2.0/projects/%s/repositories/%s/artifacts?page_size=100"
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

	EventNoNamespaceAvailable    EventType = "no.namespace.available"
	EventRestartServiceStarted   EventType = "restart.service.started"
	EventRestartServiceCompleted EventType = "restart.service.completed"
	EventRestartServiceFailed    EventType = "restart.service.failed"

	EventFaultInjectionStarted   EventType = "fault.injection.started"
	EventFaultInjectionCompleted EventType = "fault.injection.completed"
	EventFaultInjectionFailed    EventType = "fault.injection.failed"

	EventAcquireLock EventType = "acquire.lock"
	EventReleaseLock EventType = "release.lock"
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

// K8s CRD Label 字段
const (
	CRDTaskID      = "task_id"
	CRDTraceID     = "trace_id"
	CRDGroupID     = "group_id"
	CRDBenchmark   = "benchmark"
	CRDPreDuration = "pre_duration"
)

// K8s Job Label 字段
const (
	LabelTaskID      = "task_id"
	LabelTraceID     = "trace_id"
	LabelGroupID     = "group_id"
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
