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
	BuildDataset       = "dataset"
	BuildNamespace     = "namespace"
	BuildPreDuration   = "pre_duration"
	BuildService       = "service"
	BuildStartTime     = "start_time"
	BuildEndTime       = "end_time"
	BuildAlgorithm     = "algorithm"
	BuildAlgorithmPath = "algorithm_path"

	CollectAlgorithm   = "algorithm"
	CollectDataset     = "dataset"
	CollectExecutionID = "execution_id"

	EvaluateLabel = "app_name"
	EvaluateLevel = "level"

	ExecuteAlgo    = "algorithm"
	ExecuteDataset = "dataset"
	ExecuteService = "service"
	ExecuteTag     = "tag"

	InjectBenchmark   = "benchmark"
	InjectFaultType   = "fault_type"
	InjectPreDuration = "pre_duration"
	InjectRawConf     = "raw_conf"
	InjectConf        = "conf"
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

// Redis Meta 属性名称
const (
	MetaBenchmark   = "benchmark"
	MetaPreDuration = "pre_duration"
	MetaTraceID     = "trace_id"
	MetaGroupID     = "group_id"
)

// Redis 订阅消息字段
const (
	RdbMsgStatus      = "status"
	RdbMsgTaskID      = "task_id"
	RdbMsgTaskType    = "task_type"
	RdbMsgDataset     = "dataset"
	RdbMsgExecutionID = "execution_id"
	RdbMsgError       = "error"
)

// K8s Job 名称
const (
	DatasetJobName = "dataset"
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
)

// sse 事件类型
const (
	EventEnd    = "end"
	EventUpdate = "update"
)
