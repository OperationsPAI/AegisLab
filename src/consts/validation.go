package consts

var ValidActions = map[ActionName]struct{}{
	ActionRead:    {},
	ActionWrite:   {},
	ActionDelete:  {},
	ActionManage:  {},
	ActionExecute: {},
}

var ValidAuditLogStates = map[AuditLogState]struct{}{
	AuditLogStateFailed:  {},
	AuditLogStateSuccess: {},
}

var ValidBenchmarks = map[string]struct{}{
	"clickhouse": {},
}

var ValidContainerTypes = map[ContainerType]struct{}{
	ContainerTypeAlgorithm: {},
	ContainerTypeBenchmark: {},
	ContainerTypePedestal:  {},
}

var ValidDatapackStates = map[DatapackState]struct{}{
	DatapackInitial:       {},
	DatapackInjectFailed:  {},
	DatapackInjectSuccess: {},
	DatapackBuildFailed:   {},
	DatapackBuildSuccess:  {},
}

var ValidExecutionStates = map[ExecutionState]struct{}{
	ExecutionInitial: {},
	ExecutionFailed:  {},
	ExecutionSuccess: {},
}

var ValidGrantTypes = map[GrantType]struct{}{
	GrantTypeGrant: {},
	GrantTypeDeny:  {},
}

var ValidLabelCategories = map[LabelCategory]struct{}{
	SystemCategory:    {},
	ContainerCategory: {},
	DatasetCategory:   {},
	ProjectCategory:   {},
	InjectionCategory: {},
	ExecutionCategory: {},
}

var ValidPageSizes = map[PageSize]struct{}{
	PageSizeSmall:  {},
	PageSizeMedium: {},
	PageSizeLarge:  {},
}

var ValidParameterTypes = map[ParameterType]struct{}{
	ParamTypeFixed:   {},
	ParamTypeDynamic: {},
}

var ValidParameterCategories = map[ParameterCategory]struct{}{
	ParamCategoryEnvVars:    {},
	ParamCategoryHelmValues: {},
}

var ValidResourceTypes = map[ResourceType]struct{}{
	ResourceTypeSystem: {},
	ResourceTypeTable:  {},
}

var ValidResourceCategories = map[ResourceCategory]struct{}{
	ResourceCore:  {},
	ResourceAdmin: {},
}

var ValidStatuses = map[StatusType]struct{}{
	CommonDeleted:  {},
	CommonDisabled: {},
	CommonEnabled:  {},
}

var ValidTaskEvents = map[TaskType][]EventType{
	TaskTypeBuildDatapack: {
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
	TaskTypeRestartPedestal: {
		EventNoNamespaceAvailable,
		EventRestartPedestalStarted,
		EventRestartPedestalCompleted,
		EventRestartPedestalFailed,
	},
}

var ValidTaskStates = map[TaskState]struct{}{
	TaskCancelled:   {},
	TaskError:       {},
	TaskPending:     {},
	TaskRescheduled: {},
	TaskRunning:     {},
	TaskCompleted:   {},
}

var ValidTaskTypes = map[TaskType]struct{}{
	TaskTypeBuildContainer:  {},
	TaskTypeRestartPedestal: {},
	TaskTypeBuildDatapack:   {},
	TaskTypeFaultInjection:  {},
	TaskTypeRunAlgorithm:    {},
	TaskTypeCollectResult:   {},
}
