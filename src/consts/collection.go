package consts

var auditLogStateMap = map[AuditLogState]string{
	AuditLogStateFailed:  "failed",
	AuditLogStateSuccess: "success",
}

var containerTypeMap = map[ContainerType]string{
	ContainerTypeAlgorithm: "algorithm",
	ContainerTypeBenchmark: "benchmark",
	ContainerTypePedestal:  "pedestal",
}

var configHistoryChangeTypeMap = map[ConfigHistoryChangeType]string{
	ChangeTypeUpdate:   "update",
	ChangeTypeRollback: "rollback",
}

var ConfigScopeMap = map[ConfigScope]string{
	ConfigScopeProducer: "producer",
	ConfigScopeConsumer: "consumer",
	ConfigScopeGlobal:   "global",
}

var dynamicConfigTypeMap = map[ConfigValueType]string{
	ConfigValueTypeBool:        "bool",
	ConfigValueTypeInt:         "int",
	ConfigValueTypeFloat:       "float",
	ConfigValueTypeString:      "string",
	ConfigValueTypeStringArray: "string_array",
}

var datapackStateMap = map[DatapackState]string{
	DatapackInitial:         "initial",
	DatapackInjectFailed:    "inject_failed",
	DatapackInjectSuccess:   "inject_success",
	DatapackBuildFailed:     "build_failed",
	DatapackBuildSuccess:    "build_success",
	DatapackDetectorFailed:  "detector_failed",
	DatapackDetectorSuccess: "detector_success",
}

var executeStateMap = map[ExecutionState]string{
	ExecutionInitial: "initial",
	ExecutionFailed:  "failed",
	ExecutionSuccess: "success",
}

var grantTypeMap = map[GrantType]string{
	GrantTypeGrant: "grant",
	GrantTypeDeny:  "deny",
}

var labelCategoryMap = map[LabelCategory]string{
	SystemCategory:    "system",
	ConfigCategory:    "config",
	ContainerCategory: "container",
	DatasetCategory:   "dataset",
	ProjectCategory:   "project",
	InjectionCategory: "injection",
	ExecutionCategory: "execution",
}

var parameterTypeMap = map[ParameterType]string{
	ParameterTypeFixed:   "fixed",
	ParameterTypeDynamic: "dynamic",
}

var parameterCategoryMap = map[ParameterCategory]string{
	ParameterCategoryEnvVars:    "env_vars",
	ParameterCategoryHelmValues: "helm_values",
}

var valueDataTypeMap = map[ValueDataType]string{
	ValueDataTypeString: "string",
	ValueDataTypeInt:    "int",
	ValueDataTypeBool:   "bool",
	ValueDataTypeFloat:  "float",
	ValueDataTypeArray:  "array",
	ValueDataTypeObject: "object",
}

var resourceDisplayNameMap = map[ResourceName]string{
	ResourceSystem:           "System",
	ResourceAudit:            "Audit",
	ResourceConfiguration:    "Configuration",
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

var resouceTypeMap = map[ResourceType]string{
	ResourceTypeSystem: "system",
	ResourceTypeTable:  "table",
}

var resourceCategoryMap = map[ResourceCategory]string{
	ResourceCategoryChaos:    "chaos",
	ResourceCategoryAsset:    "asset",
	ResourceCategoryPlatform: "platform",
	ResourceCategorySystem:   "system",
}

var statusTypeMap = map[StatusType]string{
	CommonDeleted:  "deleted",
	CommonDisabled: "disabled",
	CommonEnabled:  "enabled",
}

var taskStateMap = map[TaskState]string{
	TaskCancelled:   "Cancelled",
	TaskError:       "Error",
	TaskPending:     "Pending",
	TaskRescheduled: "Rescheduled",
	TaskRunning:     "Running",
	TaskCompleted:   "Completed",
}

var taskTypeMap = map[TaskType]string{
	TaskTypeBuildContainer:  "BuildContainer",
	TaskTypeRestartPedestal: "RestartPedestal",
	TaskTypeFaultInjection:  "FaultInjection",
	TaskTypeRunAlgorithm:    "RunAlgorithm",
	TaskTypeBuildDatapack:   "BuildDatapack",
	TaskTypeCollectResult:   "CollectResult",
	TaskTypeCronJob:         "CronJob",
}

var traceTypeMap = map[TraceType]string{
	TraceTypeFullPipeline:   "FullPipeline",
	TraceTypeFaultInjection: "FaultInjection",
	TraceTypeDatapackBuild:  "DatapackBuild",
	TraceTypeAlgorithmRun:   "AlgorithmRun",
}

var traceStateMap = map[TraceState]string{
	TracePending:   "Pending",
	TraceRunning:   "Running",
	TraceCompleted: "Completed",
	TraceFailed:    "Failed",
}

// SystemRoleDisplayNames maps system role names to their display names
var SystemRoleDisplayNames = map[RoleName]string{
	// System Roles
	RoleSuperAdmin: "Super Admin",
	RoleAdmin:      "Admin",

	// Regular User Role
	RoleUser:                 "User",
	RoleContainerAdmin:       "Container Admin",
	RoleContainerDeveloper:   "Container Developer",
	RoleContainerViewer:      "Container Viewer",
	RoleDatasetAdmin:         "Dataset Admin",
	RoleDatasetDeveloper:     "Dataset Developer",
	RoleDatasetViewer:        "Dataset Viewer",
	RoleProjectAdmin:         "Project Admin",
	RoleProjectAlgoDeveloper: "Project Algorithm Developer",
	RoleProjectDataDeveloper: "Project Data Developer",
	RoleProjectViewer:        "Project Viewer",
	RoleTeamAdmin:            "Team Admin",
	RoleTeamMember:           "Team Member",
	RoleTeamViewer:           "Team Viewer",
}

// ------------------- Functions to get names ------------------

func GetAuditLogStateName(state AuditLogState) string {
	if name, exists := auditLogStateMap[state]; exists {
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

func GetConfigHistoryChangeTypeName(changeType ConfigHistoryChangeType) string {
	if name, exists := configHistoryChangeTypeMap[changeType]; exists {
		return name
	}
	return "unknown"
}

func GetConfigScopeName(scope ConfigScope) string {
	if name, exists := ConfigScopeMap[scope]; exists {
		return name
	}
	return "unknown"
}

func GetDatapackStateName(state DatapackState) string {
	if name, exists := datapackStateMap[state]; exists {
		return name
	}
	return "unknown"
}

func GetDatapackStateByName(name string) *DatapackState {
	datapackStateNameToStateMap := make(map[string]DatapackState, len(datapackStateMap))
	for dState, name := range datapackStateMap {
		datapackStateNameToStateMap[name] = dState
	}

	datapackState, exists := datapackStateNameToStateMap[name]
	if exists {
		return &datapackState
	}

	return nil
}

func GetDynamicConfigTypeName(configType ConfigValueType) string {
	if name, exists := dynamicConfigTypeMap[configType]; exists {
		return name
	}
	return "unknown"
}

func GetExecutionStateName(state ExecutionState) string {
	if name, exists := executeStateMap[state]; exists {
		return name
	}
	return "unknown"
}

func GetExecutionStateByName(name string) *ExecutionState {
	executionStateNameToStateMap := make(map[string]ExecutionState, len(executeStateMap))
	for eState, name := range executeStateMap {
		executionStateNameToStateMap[name] = eState
	}

	executionState, exists := executionStateNameToStateMap[name]
	if exists {
		return &executionState
	}

	return nil
}

func GetGrantTypeName(grantType GrantType) string {
	if name, exists := grantTypeMap[grantType]; exists {
		return name
	}
	return "unknown"
}

func GetLabelCategoryName(category LabelCategory) string {
	if name, exists := labelCategoryMap[category]; exists {
		return name
	}
	return "unknown"
}

func GetParameterTypeName(paramType ParameterType) string {
	if name, exists := parameterTypeMap[paramType]; exists {
		return name
	}
	return "unknown"
}

func GetParameterCategoryName(category ParameterCategory) string {
	if name, exists := parameterCategoryMap[category]; exists {
		return name
	}
	return "unknown"
}

func GetValueDataTypeName(valueType ValueDataType) string {
	if name, exists := valueDataTypeMap[valueType]; exists {
		return name
	}
	return "unknown"
}

func GetResourceDisplayName(resourceName ResourceName) string {
	if displayName, exists := resourceDisplayNameMap[resourceName]; exists {
		return displayName
	}
	return "Unknown"
}

func GetResourceTypeName(resourceType ResourceType) string {
	if name, exists := resouceTypeMap[resourceType]; exists {
		return name
	}
	return "unknown"
}

func GetResourceCategoryName(category ResourceCategory) string {
	if name, exists := resourceCategoryMap[category]; exists {
		return name
	}
	return "unknown"
}

func GetStatusTypeName(status StatusType) string {
	if name, exists := statusTypeMap[status]; exists {
		return name
	}
	return "unknown"
}

func GetTaskStateName(state TaskState) string {
	if name, exists := taskStateMap[state]; exists {
		return name
	}
	return "Unknown"
}

func GetTaskStateByName(name string) *TaskState {
	taskStateNameToStateMap := make(map[string]TaskState, len(taskStateMap))
	for tState, name := range taskStateMap {
		taskStateNameToStateMap[name] = tState
	}

	taskState, exists := taskStateNameToStateMap[name]
	if exists {
		return &taskState
	}

	return nil
}

func GetTaskTypeName(taskType TaskType) string {
	if name, exists := taskTypeMap[taskType]; exists {
		return name
	}
	return "Unknown"
}

func GetTaskTypeByName(name string) *TaskType {
	taskTypeNameToTypeMap := make(map[string]TaskType, len(taskTypeMap))
	for tType, name := range taskTypeMap {
		taskTypeNameToTypeMap[name] = tType
	}

	taskType, exists := taskTypeNameToTypeMap[name]
	if exists {
		return &taskType
	}

	return nil
}

func GetTraceTypeName(traceType TraceType) string {
	if name, exists := traceTypeMap[traceType]; exists {
		return name
	}
	return "Unknown"
}

func GetTraceStateName(state TraceState) string {
	if name, exists := traceStateMap[state]; exists {
		return name
	}
	return "Unknown"
}
