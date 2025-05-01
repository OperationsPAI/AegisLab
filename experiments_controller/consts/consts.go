package consts

import "time"

// 定义任务类型
type TaskType string

const (
	DefaultTimeUnit = time.Minute
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
	TaskStatusCanceled    string = "Canceled"
	TaskStatusCompleted   string = "Completed"
	TaskStatusError       string = "Error"
	TaskStatusPending     string = "Pending"
	TaskStatusRunning     string = "Running"
	TaskStautsRescheduled string = "Rescheduled"
	TaskStatusScheduled   string = "Scheduled"
)

const (
	TaskTypeRestartService TaskType = "RestartService"
	TaskTypeRunAlgorithm   TaskType = "RunAlgorithm"
	TaskTypeFaultInjection TaskType = "FaultInjection"
	TaskTypeBuildImages    TaskType = "BuildImages"
	TaskTypeBuildDataset   TaskType = "BuildDataset"
	TaskTypeCollectResult  TaskType = "CollectResult"
)

const (
	TaskMsgCompleted string = "Task %s completed"
	TaskMsgFailed    string = "Task %s failed"
)

// 不同任务类型的 Payload 键
const (
	BuildBenchmark     = "benchmark"
	BuildDataset       = "name"
	BuildPreDuration   = "pre_duration"
	BuildStartTime     = "start_time"
	BuildEndTime       = "end_time"
	BuildEnvVars       = "env_vars"
	BuildAlgorithm     = "algorithm"
	BuildAlgorithmPath = "algorithm_path"

	CollectAlgorithm   = "algorithm"
	CollectDataset     = "dataset"
	CollectExecutionID = "execution_id"

	EvaluateLabel = "app_name"
	EvaluateLevel = "level"

	ExecuteImage   = "image"
	ExecuteTag     = "tag"
	ExecuteDataset = "dataset"
	ExecuteEnvVars = "env_vars"

	InjectBenchmark   = "benchmark"
	InjectFaultType   = "fault_type"
	InjectNamespace   = "namespace"
	InjectPreDuration = "pre_duration"
	InjectDisplayData = "display_data"
	InjectConf        = "conf"
	InjectNode        = "node"

	RestartIntarval      = "interval"
	RestartFaultDuration = "fault_duration"
	RestartInjectPayload = "inject_payload"
)

// 环境变量名称
const (
	BuildEnvVarNamespace = "NAMESPACE"
	BuildEnvVarService   = "SERVICE"

	ExecuteEnvVarAlgorithm = "ALGORITHM"
	ExecuteEnvVarService   = "SERVICE"
)

const (
	HarborURL      = "http://%s/api/v2.0/projects/%s/repositories/%s/artifacts?page_size=100"
	HarborTimeout  = 30
	HarborTimeUnit = time.Second
)

// Redis stream 频道和字段
const (
	StreamLogKey = "trace:%s:log"

	RdbEventTaskID   = "task_id"
	RdbEventTaskType = "task_type"
	RdbEventStatus   = "status"
	RdbEventFileName = "file_name"
	RdbEventLine     = "event_line"
	RdbEventName     = "event_name"
	RdbEventPayload  = "event_payload"

	RdbPayloadErr            = "error"
	RdbPayloadDataset        = "dataset"
	RdbPayloadExecutionID    = "execution_id"
	RdbPayloadDetectorResult = "detector_result"
)

type EventType string

const (
	EventAlgoResultCollection    EventType = "algorithm.collect_result"
	EventDatasetResultCollection EventType = "algorithm.dataset_result"
	EventTaskStatusUpdate        EventType = "task.status.update"
	EventTaskStarted             EventType = "task.started"
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
	LabelAlgorithm   = "algorithm"
	LabelDataset     = "dataset"
	LabelExecutionID = "execution_id"
	LabelService     = "service"
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
