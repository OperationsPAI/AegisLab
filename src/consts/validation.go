package consts

var ValidActions = map[ActionName]struct{}{
	// Basic CRUD Operations
	ActionCreate: {},
	ActionRead:   {},
	ActionUpdate: {},
	ActionDelete: {},

	// Execution and State Management
	ActionExecute:  {},
	ActionStop:     {},
	ActionRestart:  {},
	ActionActivate: {},
	ActionSuspend:  {},

	// File and Data Operations
	ActionUpload:   {},
	ActionDownload: {},
	ActionImport:   {},
	ActionExport:   {},

	// Permission and Membership Management
	ActionAssign: {},
	ActionGrant:  {},
	ActionRevoke: {},

	// Configuration and Administration
	ActionConfigure: {},
	ActionManage:    {},

	// Collaboration and Sharing
	ActionShare: {},
	ActionClone: {},

	// Monitoring and Analysis
	ActionMonitor: {},
	ActionAnalyze: {},
	ActionAudit:   {},
}

var ValidResourceNames = map[ResourceName]struct{}{
	ResourceSystem:           {},
	ResourceAudit:            {},
	ResourceConfiguration:    {},
	ResourceContainer:        {},
	ResourceContainerVersion: {},
	ResourceDataset:          {},
	ResourceDatasetVersion:   {},
	ResourceProject:          {},
	ResourceTeam:             {},
	ResourceLabel:            {},
	ResourceUser:             {},
	ResourceRole:             {},
	ResourcePermission:       {},
	ResourceTask:             {},
	ResourceTrace:            {},
	ResourceInjection:        {},
	ResourceExecution:        {},
}

var ValidResourceScopes = map[ResourceScope]struct{}{
	ScopeOwn:     {},
	ScopeProject: {},
	ScopeTeam:    {},
	ScopeAll:     {},
}

var ValidAuditLogStates = map[AuditLogState]struct{}{
	AuditLogStateFailed:  {},
	AuditLogStateSuccess: {},
}

var ValidContainerTypes = map[ContainerType]struct{}{
	ContainerTypeAlgorithm: {},
	ContainerTypeBenchmark: {},
	ContainerTypePedestal:  {},
}

var ValidConfigHistoryChanteTypes = map[ConfigHistoryChangeType]struct{}{
	ChangeTypeUpdate:   {},
	ChangeTypeRollback: {},
}

var ValidConfigScopes = map[ConfigScope]struct{}{
	ConfigScopeProducer: {},
	ConfigScopeConsumer: {},
}

var ValidDatapackStates = map[DatapackState]struct{}{
	DatapackInitial:       {},
	DatapackInjectFailed:  {},
	DatapackInjectSuccess: {},
	DatapackBuildFailed:   {},
	DatapackBuildSuccess:  {},
}

var ValidDynamicConfigTypes = map[ConfigValueType]struct{}{
	ConfigValueTypeBool:        {},
	ConfigValueTypeInt:         {},
	ConfigValueTypeFloat:       {},
	ConfigValueTypeString:      {},
	ConfigValueTypeStringArray: {},
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
	ConfigCategory:    {},
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
	ParameterTypeFixed:   {},
	ParameterTypeDynamic: {},
}

var ValidParameterCategories = map[ParameterCategory]struct{}{
	ParameterCategoryEnvVars:    {},
	ParameterCategoryHelmValues: {},
}

var ValidResourceTypes = map[ResourceType]struct{}{
	ResourceTypeSystem: {},
	ResourceTypeTable:  {},
}

var ValidResourceCategories = map[ResourceCategory]struct{}{
	ResourceCategoryChaos:    {},
	ResourceCategoryAsset:    {},
	ResourceCategoryPlatform: {},
	ResourceCategorySystem:   {},
}

var ValidStatuses = map[StatusType]struct{}{
	CommonDeleted:  {},
	CommonDisabled: {},
	CommonEnabled:  {},
}

var ValidVolumeMountNames = map[VolumeMountName]struct{}{
	VolumeMountKubeConfig:        {},
	VolumeMountDataset:           {},
	VolumeMountExperimentStorage: {},
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
