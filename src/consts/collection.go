package consts

var AuditLogStateMap = map[AuditLogState]string{
	AuditLogStateFailed:  "failed",
	AuditLogStateSuccess: "success",
}

var containerTypeMap = map[ContainerType]string{
	ContainerTypeAlgorithm: "algorithm",
	ContainerTypeBenchmark: "benchmark",
	ContainerTypePedestal:  "pedestal",
}

var DatapackStateMap = map[DatapackState]string{
	DatapackInitial:         "initial",
	DatapackInjectFailed:    "inject_failed",
	DatapackInjectSuccess:   "inject_success",
	DatapackBuildFailed:     "build_failed",
	DatapackBuildSuccess:    "build_success",
	DatapackDetectorFailed:  "detector_failed",
	DatapackDetectorSuccess: "detector_success",
}

var ExecuteStateMap = map[ExecutionState]string{
	ExecutionInitial: "initial",
	ExecutionFailed:  "failed",
	ExecutionSuccess: "success",
}

var GrantTypeMap = map[GrantType]string{
	GrantTypeGrant: "grant",
	GrantTypeDeny:  "deny",
}

var LabelCategoryMap = map[LabelCategory]string{
	SystemCategory:    "system",
	ContainerCategory: "container",
	DatasetCategory:   "dataset",
	ProjectCategory:   "project",
	InjectionCategory: "injection",
	ExecutionCategory: "execution",
}

var ParameterTypeMap = map[ParameterType]string{
	ParameterTypeFixed:   "fixed",
	ParameterTypeDynamic: "dynamic",
}

var ParameterCategoryMap = map[ParameterCategory]string{
	ParameterCategoryEnvVars:    "env_vars",
	ParameterCategoryHelmValues: "helm_values",
}

var ValueDataTypeMap = map[ValueDataType]string{
	ValueDataTypeString: "string",
	ValueDataTypeInt:    "int",
	ValueDataTypeBool:   "bool",
	ValueDataTypeFloat:  "float",
	ValueDataTypeArray:  "array",
	ValueDataTypeObject: "object",
}

var ResourceDisplayNameMap = map[ResourceName]string{
	ResourceSystem:           "System",
	ResourceAudit:            "Audit",
	ResourceContainer:        "Container",
	ResourceContainerVersion: "Container Version",
	ResourceDataset:          "Dataset",
	ResourceDatasetVersion:   "Dataset Version",
	ResourceProject:          "Project",
	ResourceLabel:            "Label",
	ResourceUser:             "User",
	ResourceRole:             "Role",
	ResourcePermission:       "Permission",
	ResourceTask:             "Task",
	ResourceTrace:            "Trace",
	ResourceInjection:        "Fault Injection",
	ResourceExecution:        "Execution",
}

var ResouceTypeMap = map[ResourceType]string{
	ResourceTypeSystem: "system",
	ResourceTypeTable:  "table",
}

var ResourceCategoryMap = map[ResourceCategory]string{
	ResourceCore:  "core",
	ResourceAdmin: "admin",
}

var StatusTypeMap = map[StatusType]string{
	CommonDeleted:  "deleted",
	CommonDisabled: "disabled",
	CommonEnabled:  "enabled",
}

var TaskStateMap = map[TaskState]string{
	TaskCancelled:   "Cancelled",
	TaskError:       "Error",
	TaskPending:     "Pending",
	TaskRescheduled: "Rescheduled",
	TaskRunning:     "Running",
	TaskCompleted:   "Completed",
}

var TaskTypeMap = map[TaskType]string{
	TaskTypeBuildContainer:  "BuildContainer",
	TaskTypeRestartPedestal: "RestartPedestal",
	TaskTypeFaultInjection:  "FaultInjection",
	TaskTypeRunAlgorithm:    "RunAlgorithm",
	TaskTypeBuildDatapack:   "BuildDatapack",
	TaskTypeCollectResult:   "CollectResult",
	TaskTypeCronJob:         "CronJob",
}

var TraceTypeMap = map[TraceType]string{
	TraceTypeFullPipeline:   "FullPipeline",
	TraceTypeFaultInjection: "FaultInjection",
	TraceTypeDatapackBuild:  "DatapackBuild",
	TraceTypeAlgorithmRun:   "AlgorithmRun",
}

var TraceTypeHeightMap = map[TraceType]int{
	TraceTypeFullPipeline:   7,
	TraceTypeFaultInjection: 5,
	TraceTypeDatapackBuild:  3,
	TraceTypeAlgorithmRun:   2,
}

var TraceStateMap = map[TraceState]string{
	TracePending:   "Pending",
	TraceRunning:   "Running",
	TraceCompleted: "Completed",
	TraceFailed:    "Failed",
}

// SystemRoleDisplayNames maps system role names to their display names
var SystemRoleDisplayNames = map[RoleName]string{
	RoleSuperAdmin:         "Super Admin",
	RoleAdmin:              "Admin",
	RoleContainerAdmin:     "Container Admin",
	RoleContainerDeveloper: "Container Developer",
	RoleContainerViewer:    "Container Viewer",
	RoleDatasetAdmin:       "Dataset Admin",
	RoleDatasetDeveloper:   "Dataset Developer",
	RoleDatasetViewer:      "Dataset Viewer",
	RoleProjectAdmin:       "Project Admin",
	RoleProjectDeveloper:   "Project Developer",
	RoleProjectViewer:      "Project Viewer",
}

// SystemRolePermissions defines the default permissions for each system role
var SystemRolePermissions = map[RoleName][]PermissionName{
	RoleSuperAdmin: {},
	RoleAdmin: {
		PermissionReadProject, PermissionWriteProject, PermissionDeleteProject, PermissionManageProject,
		PermissionReadDataset, PermissionWriteDataset, PermissionDeleteDataset, PermissionManageDataset,
		PermissionReadFaultInjection, PermissionWriteFaultInjection, PermissionDeleteFaultInjection, PermissionExecuteFaultInjection,
		PermissionReadContainer, PermissionWriteContainer, PermissionDeleteContainer, PermissionManageContainer,
		PermissionReadTask, PermissionWriteTask, PermissionDeleteTask, PermissionExecuteTask,
		PermissionReadRole,
		PermissionReadPermission,
	},
	RoleContainerAdmin: {
		PermissionReadContainer, PermissionWriteContainer, PermissionDeleteContainer, PermissionManageContainer,
		PermissionReadContainerVersion, PermissionWriteContainerVersion, PermissionDeleteContainerVersion, PermissionManageContainerVersion,
	},
	RoleContainerDeveloper: {
		PermissionReadContainer, PermissionWriteContainer,
		PermissionReadContainerVersion, PermissionWriteContainerVersion,
	},
	RoleContainerViewer: {
		PermissionReadContainer,
		PermissionReadContainerVersion,
	},
	RoleProjectAdmin: {
		PermissionReadProject, PermissionWriteProject, PermissionDeleteProject, PermissionManageProject,
		PermissionReadDataset, PermissionWriteDataset, PermissionDeleteDataset, PermissionManageDataset,
		PermissionReadFaultInjection, PermissionWriteFaultInjection, PermissionDeleteFaultInjection, PermissionExecuteFaultInjection,
		PermissionReadTask, PermissionWriteTask, PermissionExecuteTask,
	},
	RoleProjectDeveloper: {
		PermissionReadProject,
		PermissionReadDataset, PermissionWriteDataset,
		PermissionReadFaultInjection, PermissionWriteFaultInjection, PermissionExecuteFaultInjection,
		PermissionReadTask, PermissionWriteTask, PermissionExecuteTask,
	},
	RoleProjectViewer: {
		PermissionReadProject,
		PermissionReadDataset,
		PermissionReadFaultInjection,
		PermissionReadTask,
	},
}

// ------------------- Functions to get names ------------------

func GetAuditLogStateName(state AuditLogState) string {
	if name, exists := AuditLogStateMap[state]; exists {
		return name
	}
	return "unknown"
}

func GetContainerTypeName(containerType ContainerType) string {
	if name, exists := containerTypeMap[containerType]; exists {
		return name
	}
	return "unknown"
}

func GetDatapackStateName(state DatapackState) string {
	if name, exists := DatapackStateMap[state]; exists {
		return name
	}
	return "unknown"
}

func GetExecuteStateName(state ExecutionState) string {
	if name, exists := ExecuteStateMap[state]; exists {
		return name
	}
	return "unknown"
}

func GetGrantTypeName(grantType GrantType) string {
	if name, exists := GrantTypeMap[grantType]; exists {
		return name
	}
	return "unknown"
}

func GetLabelCategoryName(category LabelCategory) string {
	if name, exists := LabelCategoryMap[category]; exists {
		return name
	}
	return "unknown"
}

func GetParameterTypeName(paramType ParameterType) string {
	if name, exists := ParameterTypeMap[paramType]; exists {
		return name
	}
	return "unknown"
}

func GetParameterCategoryName(category ParameterCategory) string {
	if name, exists := ParameterCategoryMap[category]; exists {
		return name
	}
	return "unknown"
}

func GetValueDataTypeName(valueType ValueDataType) string {
	if name, exists := ValueDataTypeMap[valueType]; exists {
		return name
	}
	return "unknown"
}

func GetResourceDisplayName(resourceName ResourceName) string {
	if displayName, exists := ResourceDisplayNameMap[resourceName]; exists {
		return displayName
	}
	return "Unknown"
}

func GetResourceTypeName(resourceType ResourceType) string {
	if name, exists := ResouceTypeMap[resourceType]; exists {
		return name
	}
	return "unknown"
}

func GetResourceCategoryName(category ResourceCategory) string {
	if name, exists := ResourceCategoryMap[category]; exists {
		return name
	}
	return "unknown"
}

func GetStatusTypeName(status StatusType) string {
	if name, exists := StatusTypeMap[status]; exists {
		return name
	}
	return "unknown"
}

func GetTaskStateName(state TaskState) string {
	if name, exists := TaskStateMap[state]; exists {
		return name
	}
	return "Unknown"
}

func GetTaskStateByName(name string) *TaskState {
	taskStateNameToStateMap := make(map[string]TaskState, len(TaskStateMap))
	for tState, name := range TaskStateMap {
		taskStateNameToStateMap[name] = tState
	}

	taskState, exists := taskStateNameToStateMap[name]
	if exists {
		return &taskState
	}

	return nil
}

func GetTaskTypeName(taskType TaskType) string {
	if name, exists := TaskTypeMap[taskType]; exists {
		return name
	}
	return "Unknown"
}

func GetTaskTypeByName(name string) *TaskType {
	taskTypeNameToTypeMap := make(map[string]TaskType, len(TaskTypeMap))
	for tType, name := range TaskTypeMap {
		taskTypeNameToTypeMap[name] = tType
	}

	taskType, exists := taskTypeNameToTypeMap[name]
	if exists {
		return &taskType
	}

	return nil
}

func GetTraceTypeName(traceType TraceType) string {
	if name, exists := TraceTypeMap[traceType]; exists {
		return name
	}
	return "Unknown"
}

func GetTraceStateName(state TraceState) string {
	if name, exists := TraceStateMap[state]; exists {
		return name
	}
	return "Unknown"
}
