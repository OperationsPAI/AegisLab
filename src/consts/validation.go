package consts

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
