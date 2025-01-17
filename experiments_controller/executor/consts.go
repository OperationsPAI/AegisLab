package executor

// 定义任务类型
type TaskType string

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
	EvalPayloadAlgo    = "algorithm"
	EvalPayloadDataset = "dataset"
	EvalPayloadBench   = "benchmark"

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
