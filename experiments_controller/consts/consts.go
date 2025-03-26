package consts

// 定义任务类型
type TaskType string

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

	ExecuteAlgo    = "algorithm"
	ExecuteDataset = "dataset"
	ExecuteService = "service"
	ExecuteTag     = "tag"

	InjectFaultType     = "fault_type"
	InjectNamespace     = "inject_namespace"
	InjectPod           = "inject_pod"
	InjectSpec          = "spec"
	InjectExectuionTime = "execution_time"
	InjectPreDuration   = "pre_duration"
	InjectFaultDuration = "fault_duration"
	InjectBenchmark     = "benchmark"
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
	RdbMsgError       = "error"
	RdbMsgExecutionID = "execution_id"
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
	LabelStartTime   = "start_time"
	LabelEndTime     = "end_time"
)

// sse 事件类型
const (
	EventEnd    = "end"
	EventError  = "error"
	EventUpdate = "update"
)
