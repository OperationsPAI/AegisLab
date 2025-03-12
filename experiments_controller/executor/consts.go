package executor

// 定义任务类型
type TaskType string

const (
	DatasetInitial = 0
	DatasetSuccess = 1
	DatasetFailed  = 2
	DatesetDeleted = 3
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
	BuildBenchmark = "benchmark"
	BuildDataset   = "dataset"
	BuildNamespace = "namespace"
	BuildStartTime = "start_time"
	BuildEndTime   = "end_time"

	CollectAlgorithm   = "algorithm"
	CollectDataset     = "dataset"
	CollectExecutionID = "execution_id"

	EvalBench   = "benchmark"
	EvalAlgo    = "algorithm"
	EvalDataset = "dataset"

	InjectDuration       = "duration"
	InjectFaultType      = "faultType"
	InjectNamespace      = "injectNamespace"
	InjectPod            = "injectPod"
	InjectSpec           = "spec"
	InjectDatasetPayload = "dataset_payload"
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
	StatusKey  = "task:%s:status"
	SubChannel = "trace:%s:channel"
)

// Redis 订阅消息字段
const (
	RdbMsgStatus      = "status"
	RdbMsgTaskType    = "task_type"
	RdbMsgDataset     = "dataset"
	RdbMsgError       = "error"
	RdbMsgExecutionID = "execution_id"
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
