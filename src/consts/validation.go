package consts

var ValidActions = map[ActionName]struct{}{
	ActionRead:    {},
	ActionWrite:   {},
	ActionDelete:  {},
	ActionManage:  {},
	ActionExecute: {},
}

var ValidCommonStatus = map[int]struct{}{
	CommonDeleted:  {},
	CommonDisabled: {},
	CommonEnabled:  {},
}

func GetValidStatusKeys() []int {
	keys := make([]int, 0, len(ValidCommonStatus))
	for k := range ValidCommonStatus {
		keys = append(keys, k)
	}
	return keys
}

var ValidContainerTypes = map[ContainerType]struct{}{
	ContainerTypeAlgorithm: {},
	ContainerTypeBenchmark: {},
	ContainerTypePedestal:  {},
}

var ValidTaskEvents = map[TaskType][]EventType{
	TaskTypeBuildDataset: {
		EventDatapackBuildSucceed,
	},
	TaskTypeCollectResult: {
		EventDatapackResultCollection,
		EventDatapackNoAnomaly,
		EventDatapackNoDetectorData,
	},
	TaskTypeFaultInjection: {
		EventFaultInjectionStarted,
		EventFaultInjectionCompleted,
		EventFaultInjectionFailed,
	},
	TaskTypeRunAlgorithm: {
		EventAlgoRunSucceed,
	},
	TaskTypeRestartService: {
		EventNoNamespaceAvailable,
		EventRestartServiceStarted,
		EventRestartServiceCompleted,
		EventRestartServiceFailed,
	},
}
