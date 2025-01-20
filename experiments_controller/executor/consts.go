package executor

type TaskStatus string

// 定义任务类型
type TaskType string

const (
	DatasetInitial = 0
	DatasetSuccess = 1
	DatasetFailed  = 2
	DatesetDeleted = 3
)

const (
	TaskStatusCanceled TaskStatus = "Canceled"
	TaskStatusError    TaskStatus = "Error"
	TaskStatusRunning  TaskStatus = "Running"
)

const (
	TaskTypeRunAlgorithm   TaskType = "RunAlgorithm"
	TaskTypeFaultInjection TaskType = "FaultInjection"
	TaskTypeBuildImages    TaskType = "BuildImages"
	TaskTypeBuildDataset   TaskType = "BuildDataset"
	TaskTypeCollectResult  TaskType = "CollectResult"
)

// Redis 消息字段
const (
	RdbMsgTaskID       = "taskID"
	RdbMsgTaskType     = "taskType"
	RdbMsgPayload      = "payload"
	RdbMsgParentTaskID = "parentTaskID"
)

// 不同任务类型的 Payload 键
const (
	CollectDataset     = "dataset"
	CollectExecutionID = "execution_id"

	EvalBench   = "benchmark"
	EvalAlgo    = "algorithm"
	EvalDataset = "dataset"

	InjectDuration  = "duration"
	InjectFaultType = "faultType"
	InjectNamespace = "injectNamespace"
	InjectPod       = "injectPod"
	InjectSpec      = "spec"
)

// Redis 流和消费者组配置
const (
	StreamName   = "task_stream"
	GroupName    = "task_consumer_group"
	ConsumerName = "task_consumer"
)

const (
	LabelDataset     = "dataset"
	LabelExecutionID = "execution_id"
	LabelJobType     = "job_type"
	LabelTaskID      = "task_id"
)
