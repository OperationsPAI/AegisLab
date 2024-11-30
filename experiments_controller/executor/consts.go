package executor

// 定义任务类型
type TaskType string

const (
	TaskTypeRunAlgorithm   TaskType = "RunAlgorithm"
	TaskTypeFaultInjection TaskType = "FaultInjection"
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

	InjectFaultType = "faultType"
	InjectNamespace = "injectNamespace"
	InjectPod       = "injectPod"
	InjectDuration  = "duration"
	InjectSpec      = "spec"
)

// Redis 流和消费者组配置
const (
	StreamName   = "task_stream"
	GroupName    = "task_consumer_group"
	ConsumerName = "task_consumer"
)
