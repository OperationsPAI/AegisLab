package consts

import "time"

// 定义任务类型
type TaskType string

const (
	DefaultTimeUnit = time.Minute
)

const (
	DatasetInitial       = 0
	DatasetInjectSuccess = 1
	DatasetInjectFailed  = 2
	DatasetBuildSuccess  = 3
	DatasetBuildFailed   = 4
	DatasetDeleted       = 5
)

const (
	TaskStatusCanceled  string = "Canceled"
	TaskStatusCompleted string = "Completed"
	TaskStatusError     string = "Error"
	TaskStatusPending   string = "Pending"
	TaskStatusRunning   string = "Running"
)

const (
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
	InjectPreDuration = "pre_duration"
	InjectRawConf     = "raw_conf"
	InjectConf        = "conf"
)

// 环境变量名称
const (
	BuildEnvVarNamespace = "NAMESPACE"
	BuildEnvVarService   = "SERVICE"

	ExecuteEnvVarAlgorithm = "ALGORITHM"
	ExecuteEnvVarService   = "SERVICE"
)

// Redis 流和消费者组配置
const (
	StreamName   = "task_stream"
	GroupName    = "task_consumer_group"
	ConsumerName = "task_consumer"
)

// Redis 记录名称
const (
	LogFormat  = "[%s] %s"
	LogKey     = "task:%s:logs"
	MetaKey    = "task:%s:meta"
	StatusKey  = "task:%s:status"
	SubChannel = "trace:%s:channel"
)

// Redis 订阅消息字段
const (
	RdbMsgStatus            = "status"
	RdbMsgTaskID            = "task_id"
	RdbMsgTaskType          = "task_type"
	RdbMsgDataset           = "dataset"
	RdbMsgExecutionID       = "execution_id"
	RdbMsgHasDetectorResult = "has_detector_result"
	RdbMsgError             = "error"
)

// K8s Job 名称
const (
	DatasetJobName = "dataset"
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
	DownloadFilename       = "package"
	DetectorConclusionFile = "conclusion.csv"
	ExecutionResultFile    = "result.csv"
)
