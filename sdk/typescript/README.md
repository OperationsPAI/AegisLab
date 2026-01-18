## rcabench-client@0.0.1

This generator creates TypeScript/JavaScript client that utilizes [axios](https://github.com/axios/axios). The generated Node module can be used in the following environments:

Environment
* Node.js
* Webpack
* Browserify

Language level
* ES5 - you must have a Promises/A+ library installed
* ES6

Module system
* CommonJS
* ES6 module system

It can be used in both TypeScript and JavaScript. In TypeScript, the definition will be automatically resolved via `package.json`. ([Reference](https://www.typescriptlang.org/docs/handbook/declaration-files/consumption.html))

### Building

To build and compile the typescript sources to javascript use:
```
npm install
npm run build
```

### Publishing

First build the package then run `npm publish`

### Consuming

navigate to the folder of your consuming project and run one of the following commands.

_published:_

```
npm install rcabench-client@0.0.1 --save
```

_unPublished (not recommended):_

```
npm install PATH_TO_GENERATED_PACKAGE --save
```

### Documentation for API Endpoints

All URIs are relative to *http://http://localhost:8082*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*AuthenticationApi* | [**login**](docs/AuthenticationApi.md#login) | **POST** /api/v2/auth/login | User login
*AuthenticationApi* | [**registerUser**](docs/AuthenticationApi.md#registeruser) | **POST** /api/v2/auth/register | User registration
*ContainersApi* | [**buildContainerImage**](docs/ContainersApi.md#buildcontainerimage) | **POST** /api/v2/containers/build | Submit container building
*ContainersApi* | [**createContainer**](docs/ContainersApi.md#createcontainer) | **POST** /api/v2/containers | Create container
*ContainersApi* | [**createContainerVersion**](docs/ContainersApi.md#createcontainerversion) | **POST** /api/v2/containers/{container_id}/versions | Create container version
*ContainersApi* | [**getContainerById**](docs/ContainersApi.md#getcontainerbyid) | **GET** /api/v2/containers/{container_id} | Get container by ID
*ContainersApi* | [**getContainerVersionById**](docs/ContainersApi.md#getcontainerversionbyid) | **GET** /api/v2/containers/{container_id}/versions/{version_id} | Get container version by ID
*ContainersApi* | [**listContainerVersions**](docs/ContainersApi.md#listcontainerversions) | **GET** /api/v2/containers/{container_id}/versions | List container versions
*ContainersApi* | [**listContainers**](docs/ContainersApi.md#listcontainers) | **GET** /api/v2/containers | List containers
*DatasetsApi* | [**createDataset**](docs/DatasetsApi.md#createdataset) | **POST** /api/v2/datasets | Create dataset
*DatasetsApi* | [**createDatasetVersion**](docs/DatasetsApi.md#createdatasetversion) | **POST** /api/v2/datasets/{dataset_id}/versions | Create dataset version
*DatasetsApi* | [**downloadDatasetVersion**](docs/DatasetsApi.md#downloaddatasetversion) | **GET** /api/v2/datasets/{dataset_id}/versions/{version_id}/download | Download dataset version
*DatasetsApi* | [**getDatasetById**](docs/DatasetsApi.md#getdatasetbyid) | **GET** /api/v2/datasets/{dataset_id} | Get dataset by ID
*DatasetsApi* | [**getDatasetVersionById**](docs/DatasetsApi.md#getdatasetversionbyid) | **GET** /api/v2/datasets/{dataset_id}/versions/{version_id} | Get dataset version by ID
*DatasetsApi* | [**listDatasetVersions**](docs/DatasetsApi.md#listdatasetversions) | **GET** /api/v2/datasets/{dataset_id}/versions | List dataset versions
*DatasetsApi* | [**listDatasets**](docs/DatasetsApi.md#listdatasets) | **GET** /api/v2/datasets | List datasets
*DatasetsApi* | [**manageDatasetVersionInjections**](docs/DatasetsApi.md#managedatasetversioninjections) | **PATCH** /api/v2/datasets/{dataset_id}/version/{version_id}/injections | Manage dataset injections
*DatasetsApi* | [**searchDatasets**](docs/DatasetsApi.md#searchdatasets) | **POST** /api/v2/datasets/search | Search datasets
*DocumentationApi* | [**apiDocsModelsGet**](docs/DocumentationApi.md#apidocsmodelsget) | **GET** /api/_docs/models | API Model Definitions
*EvaluationsApi* | [**evaluateAlgorithmOnDatapacks**](docs/EvaluationsApi.md#evaluatealgorithmondatapacks) | **POST** /api/v2/evaluations/datapacks | List Datapack Evaluation Results
*EvaluationsApi* | [**evaluateAlgorithmOnDatasets**](docs/EvaluationsApi.md#evaluatealgorithmondatasets) | **POST** /api/v2/evaluations/datasets | List Dataset Evaluation Results
*ExecutionsApi* | [**getExecutionById**](docs/ExecutionsApi.md#getexecutionbyid) | **GET** /api/v2/executions/{id} | Get execution by ID
*ExecutionsApi* | [**listExecutionLabels**](docs/ExecutionsApi.md#listexecutionlabels) | **GET** /api/v2/executions/labels | List execution labels
*ExecutionsApi* | [**listExecutions**](docs/ExecutionsApi.md#listexecutions) | **GET** /api/v2/executions | List executions
*ExecutionsApi* | [**runAlgorithm**](docs/ExecutionsApi.md#runalgorithm) | **POST** /api/v2/executions/execute | Submit batch algorithm execution
*ExecutionsApi* | [**uploadDetectionResults**](docs/ExecutionsApi.md#uploaddetectionresults) | **POST** /api/v2/executions/{execution_id}/detector_results | Upload detector results
*ExecutionsApi* | [**uploadLocalizationResults**](docs/ExecutionsApi.md#uploadlocalizationresults) | **POST** /api/v2/executions/{execution_id}/granularity_results | Upload granularity results
*InjectionsApi* | [**batchManageInjectionLabels**](docs/InjectionsApi.md#batchmanageinjectionlabels) | **PATCH** /api/v2/injections/labels/batch | Batch manage injection labels
*InjectionsApi* | [**buildDatapack**](docs/InjectionsApi.md#builddatapack) | **POST** /api/v2/injections/build | Submit batch datapack buildings
*InjectionsApi* | [**getInjectionById**](docs/InjectionsApi.md#getinjectionbyid) | **GET** /api/v2/injections/{id} | Get injection by ID
*InjectionsApi* | [**getInjectionMetadata**](docs/InjectionsApi.md#getinjectionmetadata) | **GET** /api/v2/injections/metadata | Get Injection Metadata
*InjectionsApi* | [**injectFault**](docs/InjectionsApi.md#injectfault) | **POST** /api/v2/injections/inject | Submit batch fault injections
*InjectionsApi* | [**listFailedInjections**](docs/InjectionsApi.md#listfailedinjections) | **GET** /api/v2/injections/analysis/no-issues | Query Fault Injection Records Without Issues
*InjectionsApi* | [**listInjections**](docs/InjectionsApi.md#listinjections) | **GET** /api/v2/injections | List injections
*InjectionsApi* | [**listSuccessfulInjections**](docs/InjectionsApi.md#listsuccessfulinjections) | **GET** /api/v2/injections/analysis/with-issues | Query Fault Injection Records With Issues
*InjectionsApi* | [**manageInjectionLabels**](docs/InjectionsApi.md#manageinjectionlabels) | **PATCH** /api/v2/injections/{id}/labels | Manage injection custom labels
*InjectionsApi* | [**searchInjections**](docs/InjectionsApi.md#searchinjections) | **POST** /api/v2/injections/search | Search injections
*LabelsApi* | [**batchDeleteLabels**](docs/LabelsApi.md#batchdeletelabels) | **POST** /api/v2/labels/batch-delete | Batch delete labels
*LabelsApi* | [**createLabel**](docs/LabelsApi.md#createlabel) | **POST** /api/v2/labels | Create label
*LabelsApi* | [**deleteLabel**](docs/LabelsApi.md#deletelabel) | **DELETE** /api/v2/labels/{label_id} | Delete label
*LabelsApi* | [**getLabelById**](docs/LabelsApi.md#getlabelbyid) | **GET** /api/v2/labels/{label_id} | Get label by ID
*LabelsApi* | [**listLabels**](docs/LabelsApi.md#listlabels) | **GET** /api/v2/labels | List labels
*LabelsApi* | [**updateLabel**](docs/LabelsApi.md#updatelabel) | **PATCH** /api/v2/labels/{label_id} | Update label
*PermissionsApi* | [**createPermission**](docs/PermissionsApi.md#createpermission) | **POST** /api/v2/permissions | Create a new permission
*PermissionsApi* | [**deletePermission**](docs/PermissionsApi.md#deletepermission) | **DELETE** /api/v2/permissions/{id} | Delete permission
*PermissionsApi* | [**getPermissionById**](docs/PermissionsApi.md#getpermissionbyid) | **GET** /api/v2/permissions/{id} | Get permission by ID
*PermissionsApi* | [**listPermissions**](docs/PermissionsApi.md#listpermissions) | **GET** /api/v2/permissions | List permissions
*PermissionsApi* | [**listRolesWithPermission**](docs/PermissionsApi.md#listroleswithpermission) | **GET** /api/v2/permissions/{permission_id}/roles | List roles from permission
*PermissionsApi* | [**updatePermission**](docs/PermissionsApi.md#updatepermission) | **PUT** /api/v2/permissions/{id} | Update permission
*ProjectsApi* | [**createProject**](docs/ProjectsApi.md#createproject) | **POST** /api/v2/projects | Create a new project
*ProjectsApi* | [**getProjectById**](docs/ProjectsApi.md#getprojectbyid) | **GET** /api/v2/projects/{project_id} | Get project by ID
*ProjectsApi* | [**listProjects**](docs/ProjectsApi.md#listprojects) | **GET** /api/v2/projects | List projects
*ResourcesApi* | [**getResourceById**](docs/ResourcesApi.md#getresourcebyid) | **GET** /api/v2/resources/{id} | Get resource by ID
*ResourcesApi* | [**listResourcePermissions**](docs/ResourcesApi.md#listresourcepermissions) | **GET** /api/v2/resources/{id}/permissions | List permissions from resource
*ResourcesApi* | [**listResources**](docs/ResourcesApi.md#listresources) | **GET** /api/v2/resources | List resources
*RolesApi* | [**createRole**](docs/RolesApi.md#createrole) | **POST** /api/v2/roles | Create a new role
*RolesApi* | [**deleteRole**](docs/RolesApi.md#deleterole) | **DELETE** /api/v2/roles/{id} | Delete role
*RolesApi* | [**getRoleById**](docs/RolesApi.md#getrolebyid) | **GET** /api/v2/roles/{id} | Get role by ID
*RolesApi* | [**grantPermissionsToRole**](docs/RolesApi.md#grantpermissionstorole) | **POST** /api/v2/roles/{role_id}/permissions/assign | Assign permissions to role
*RolesApi* | [**listRoles**](docs/RolesApi.md#listroles) | **GET** /api/v2/roles | List roles
*RolesApi* | [**revokePermissionsFromRole**](docs/RolesApi.md#revokepermissionsfromrole) | **POST** /api/v2/roles/{role_id}/permissions/remove | Remove permissions from role
*RolesApi* | [**updateRole**](docs/RolesApi.md#updaterole) | **PATCH** /api/v2/roles/{id} | Update role
*SystemApi* | [**getSystemHealth**](docs/SystemApi.md#getsystemhealth) | **GET** /system/health | System health check
*SystemApi* | [**getSystemMetrics**](docs/SystemApi.md#getsystemmetrics) | **GET** /api/v2/system/metrics | Get current system metrics
*SystemApi* | [**getSystemMetricsHistory**](docs/SystemApi.md#getsystemmetricshistory) | **GET** /api/v2/system/metrics/history | Get historical system metrics
*TasksApi* | [**getTaskById**](docs/TasksApi.md#gettaskbyid) | **GET** /api/v2/tasks/{task_id} | Get task by ID
*TasksApi* | [**listTasks**](docs/TasksApi.md#listtasks) | **GET** /api/v2/tasks | List tasks
*TracesApi* | [**getGroupStats**](docs/TracesApi.md#getgroupstats) | **GET** /api/v2/traces/group/stats | Get statistics for a group of traces
*TracesApi* | [**getTraceEvents**](docs/TracesApi.md#gettraceevents) | **GET** /api/v2/traces/{trace_id}/stream | Stream trace events in real-time
*UsersApi* | [**createUser**](docs/UsersApi.md#createuser) | **POST** /api/v2/users | Create a new user
*UsersApi* | [**deleteUser**](docs/UsersApi.md#deleteuser) | **DELETE** /api/v2/users/{id} | Delete user
*UsersApi* | [**getUserById**](docs/UsersApi.md#getuserbyid) | **GET** /api/v2/users/{id}/detail | Get user by ID
*UsersApi* | [**listUsers**](docs/UsersApi.md#listusers) | **GET** /api/v2/users | List users
*UsersApi* | [**updateUser**](docs/UsersApi.md#updateuser) | **PATCH** /api/v2/users/{id} | Update user


### Documentation For Models

 - [ActionName](docs/ActionName.md)
 - [AssignRolePermissionReq](docs/AssignRolePermissionReq.md)
 - [BatchDeleteLabelReq](docs/BatchDeleteLabelReq.md)
 - [BatchEvaluateDatapackReq](docs/BatchEvaluateDatapackReq.md)
 - [BatchEvaluateDatapackResp](docs/BatchEvaluateDatapackResp.md)
 - [BatchEvaluateDatasetReq](docs/BatchEvaluateDatasetReq.md)
 - [BatchEvaluateDatasetResp](docs/BatchEvaluateDatasetResp.md)
 - [BatchManageInjectionLabelReq](docs/BatchManageInjectionLabelReq.md)
 - [BatchManageInjectionLabelResp](docs/BatchManageInjectionLabelResp.md)
 - [BuildOptions](docs/BuildOptions.md)
 - [BuildingSpec](docs/BuildingSpec.md)
 - [ChaosChaosResourceMapping](docs/ChaosChaosResourceMapping.md)
 - [ChaosChaosType](docs/ChaosChaosType.md)
 - [ChaosGroundtruth](docs/ChaosGroundtruth.md)
 - [ChaosNode](docs/ChaosNode.md)
 - [ChaosPair](docs/ChaosPair.md)
 - [ChaosSystemResource](docs/ChaosSystemResource.md)
 - [ChaosSystemType](docs/ChaosSystemType.md)
 - [ContainerDetailResp](docs/ContainerDetailResp.md)
 - [ContainerRef](docs/ContainerRef.md)
 - [ContainerResp](docs/ContainerResp.md)
 - [ContainerSpec](docs/ContainerSpec.md)
 - [ContainerType](docs/ContainerType.md)
 - [ContainerVersionDetailResp](docs/ContainerVersionDetailResp.md)
 - [ContainerVersionItem](docs/ContainerVersionItem.md)
 - [ContainerVersionResp](docs/ContainerVersionResp.md)
 - [CreateContainerReq](docs/CreateContainerReq.md)
 - [CreateContainerVersionReq](docs/CreateContainerVersionReq.md)
 - [CreateDatasetReq](docs/CreateDatasetReq.md)
 - [CreateDatasetVersionReq](docs/CreateDatasetVersionReq.md)
 - [CreateHelmConfigReq](docs/CreateHelmConfigReq.md)
 - [CreateLabelReq](docs/CreateLabelReq.md)
 - [CreateParameterConfigReq](docs/CreateParameterConfigReq.md)
 - [CreatePermissionReq](docs/CreatePermissionReq.md)
 - [CreateProjectReq](docs/CreateProjectReq.md)
 - [CreateRoleReq](docs/CreateRoleReq.md)
 - [CreateUserReq](docs/CreateUserReq.md)
 - [DatapackInfo](docs/DatapackInfo.md)
 - [DatapackResult](docs/DatapackResult.md)
 - [DatapackState](docs/DatapackState.md)
 - [DatasetDetailResp](docs/DatasetDetailResp.md)
 - [DatasetRef](docs/DatasetRef.md)
 - [DatasetResp](docs/DatasetResp.md)
 - [DatasetVersionDetailResp](docs/DatasetVersionDetailResp.md)
 - [DatasetVersionResp](docs/DatasetVersionResp.md)
 - [DateRange](docs/DateRange.md)
 - [DetectorResultItem](docs/DetectorResultItem.md)
 - [EvaluateDatapackItem](docs/EvaluateDatapackItem.md)
 - [EvaluateDatapackRef](docs/EvaluateDatapackRef.md)
 - [EvaluateDatapackSpec](docs/EvaluateDatapackSpec.md)
 - [EvaluateDatasetItem](docs/EvaluateDatasetItem.md)
 - [EvaluateDatasetSpec](docs/EvaluateDatasetSpec.md)
 - [EventType](docs/EventType.md)
 - [ExecutionDetailResp](docs/ExecutionDetailResp.md)
 - [ExecutionInfo](docs/ExecutionInfo.md)
 - [ExecutionRef](docs/ExecutionRef.md)
 - [ExecutionResp](docs/ExecutionResp.md)
 - [ExecutionResult](docs/ExecutionResult.md)
 - [ExecutionSpec](docs/ExecutionSpec.md)
 - [ExecutionState](docs/ExecutionState.md)
 - [GenericResponseAny](docs/GenericResponseAny.md)
 - [GenericResponseArrayInjectionNoIssuesResp](docs/GenericResponseArrayInjectionNoIssuesResp.md)
 - [GenericResponseArrayInjectionWithIssuesResp](docs/GenericResponseArrayInjectionWithIssuesResp.md)
 - [GenericResponseArrayLabelItem](docs/GenericResponseArrayLabelItem.md)
 - [GenericResponseArrayPermissionResp](docs/GenericResponseArrayPermissionResp.md)
 - [GenericResponseArrayRoleResp](docs/GenericResponseArrayRoleResp.md)
 - [GenericResponseBatchEvaluateDatapackResp](docs/GenericResponseBatchEvaluateDatapackResp.md)
 - [GenericResponseBatchEvaluateDatasetResp](docs/GenericResponseBatchEvaluateDatasetResp.md)
 - [GenericResponseBatchManageInjectionLabelResp](docs/GenericResponseBatchManageInjectionLabelResp.md)
 - [GenericResponseContainerDetailResp](docs/GenericResponseContainerDetailResp.md)
 - [GenericResponseContainerResp](docs/GenericResponseContainerResp.md)
 - [GenericResponseContainerVersionDetailResp](docs/GenericResponseContainerVersionDetailResp.md)
 - [GenericResponseContainerVersionResp](docs/GenericResponseContainerVersionResp.md)
 - [GenericResponseDatasetDetailResp](docs/GenericResponseDatasetDetailResp.md)
 - [GenericResponseDatasetResp](docs/GenericResponseDatasetResp.md)
 - [GenericResponseDatasetVersionDetailResp](docs/GenericResponseDatasetVersionDetailResp.md)
 - [GenericResponseDatasetVersionResp](docs/GenericResponseDatasetVersionResp.md)
 - [GenericResponseExecutionDetailResp](docs/GenericResponseExecutionDetailResp.md)
 - [GenericResponseGroupStats](docs/GenericResponseGroupStats.md)
 - [GenericResponseHealthCheckResp](docs/GenericResponseHealthCheckResp.md)
 - [GenericResponseInjectionDetailResp](docs/GenericResponseInjectionDetailResp.md)
 - [GenericResponseInjectionMetadataResp](docs/GenericResponseInjectionMetadataResp.md)
 - [GenericResponseInjectionResp](docs/GenericResponseInjectionResp.md)
 - [GenericResponseLabelDetailResp](docs/GenericResponseLabelDetailResp.md)
 - [GenericResponseLabelResp](docs/GenericResponseLabelResp.md)
 - [GenericResponseListContainerResp](docs/GenericResponseListContainerResp.md)
 - [GenericResponseListContainerVersionResp](docs/GenericResponseListContainerVersionResp.md)
 - [GenericResponseListDatasetDetailResp](docs/GenericResponseListDatasetDetailResp.md)
 - [GenericResponseListDatasetResp](docs/GenericResponseListDatasetResp.md)
 - [GenericResponseListDatasetVersionResp](docs/GenericResponseListDatasetVersionResp.md)
 - [GenericResponseListExecutionResp](docs/GenericResponseListExecutionResp.md)
 - [GenericResponseListInjectionResp](docs/GenericResponseListInjectionResp.md)
 - [GenericResponseListLabelResp](docs/GenericResponseListLabelResp.md)
 - [GenericResponseListProjectResp](docs/GenericResponseListProjectResp.md)
 - [GenericResponseListResourceResp](docs/GenericResponseListResourceResp.md)
 - [GenericResponseListRoleResp](docs/GenericResponseListRoleResp.md)
 - [GenericResponseListTaskResp](docs/GenericResponseListTaskResp.md)
 - [GenericResponseListUserResp](docs/GenericResponseListUserResp.md)
 - [GenericResponseLoginResp](docs/GenericResponseLoginResp.md)
 - [GenericResponsePermissionDetailResp](docs/GenericResponsePermissionDetailResp.md)
 - [GenericResponsePermissionResp](docs/GenericResponsePermissionResp.md)
 - [GenericResponseProjectDetailResp](docs/GenericResponseProjectDetailResp.md)
 - [GenericResponseProjectResp](docs/GenericResponseProjectResp.md)
 - [GenericResponseResourceResp](docs/GenericResponseResourceResp.md)
 - [GenericResponseRoleDetailResp](docs/GenericResponseRoleDetailResp.md)
 - [GenericResponseRoleResp](docs/GenericResponseRoleResp.md)
 - [GenericResponseSearchRespInjectionDetailResp](docs/GenericResponseSearchRespInjectionDetailResp.md)
 - [GenericResponseSubmitContainerBuildResp](docs/GenericResponseSubmitContainerBuildResp.md)
 - [GenericResponseSubmitDatapackBuildingResp](docs/GenericResponseSubmitDatapackBuildingResp.md)
 - [GenericResponseSubmitExecutionResp](docs/GenericResponseSubmitExecutionResp.md)
 - [GenericResponseSubmitInjectionResp](docs/GenericResponseSubmitInjectionResp.md)
 - [GenericResponseSystemMetricsHistoryResp](docs/GenericResponseSystemMetricsHistoryResp.md)
 - [GenericResponseSystemMetricsResp](docs/GenericResponseSystemMetricsResp.md)
 - [GenericResponseTaskDetailResp](docs/GenericResponseTaskDetailResp.md)
 - [GenericResponseUploadExecutionResultResp](docs/GenericResponseUploadExecutionResultResp.md)
 - [GenericResponseUserDetailResp](docs/GenericResponseUserDetailResp.md)
 - [GenericResponseUserInfo](docs/GenericResponseUserInfo.md)
 - [GenericResponseUserResp](docs/GenericResponseUserResp.md)
 - [GranularityResultItem](docs/GranularityResultItem.md)
 - [GroupStats](docs/GroupStats.md)
 - [HealthCheckResp](docs/HealthCheckResp.md)
 - [HelmConfigDetailResp](docs/HelmConfigDetailResp.md)
 - [HelmConfigItem](docs/HelmConfigItem.md)
 - [InfoPayloadTemplate](docs/InfoPayloadTemplate.md)
 - [InjectionDetailResp](docs/InjectionDetailResp.md)
 - [InjectionItem](docs/InjectionItem.md)
 - [InjectionLabelOperation](docs/InjectionLabelOperation.md)
 - [InjectionMetadataResp](docs/InjectionMetadataResp.md)
 - [InjectionNoIssuesResp](docs/InjectionNoIssuesResp.md)
 - [InjectionResp](docs/InjectionResp.md)
 - [InjectionWithIssuesResp](docs/InjectionWithIssuesResp.md)
 - [JobMessage](docs/JobMessage.md)
 - [LabelCategory](docs/LabelCategory.md)
 - [LabelDetailResp](docs/LabelDetailResp.md)
 - [LabelItem](docs/LabelItem.md)
 - [LabelResp](docs/LabelResp.md)
 - [ListContainerResp](docs/ListContainerResp.md)
 - [ListContainerVersionResp](docs/ListContainerVersionResp.md)
 - [ListDatasetDetailResp](docs/ListDatasetDetailResp.md)
 - [ListDatasetResp](docs/ListDatasetResp.md)
 - [ListDatasetVersionResp](docs/ListDatasetVersionResp.md)
 - [ListExecutionResp](docs/ListExecutionResp.md)
 - [ListInjectionResp](docs/ListInjectionResp.md)
 - [ListLabelResp](docs/ListLabelResp.md)
 - [ListProjectResp](docs/ListProjectResp.md)
 - [ListResourceResp](docs/ListResourceResp.md)
 - [ListRoleResp](docs/ListRoleResp.md)
 - [ListTaskResp](docs/ListTaskResp.md)
 - [ListUserResp](docs/ListUserResp.md)
 - [LoginReq](docs/LoginReq.md)
 - [LoginResp](docs/LoginResp.md)
 - [ManageDatasetVersionInjectionReq](docs/ManageDatasetVersionInjectionReq.md)
 - [ManageInjectionLabelReq](docs/ManageInjectionLabelReq.md)
 - [MetricValue](docs/MetricValue.md)
 - [PageSize](docs/PageSize.md)
 - [PaginationInfo](docs/PaginationInfo.md)
 - [ParameterCategory](docs/ParameterCategory.md)
 - [ParameterItem](docs/ParameterItem.md)
 - [ParameterSpec](docs/ParameterSpec.md)
 - [ParameterType](docs/ParameterType.md)
 - [PermissionDetailResp](docs/PermissionDetailResp.md)
 - [PermissionResp](docs/PermissionResp.md)
 - [ProjectDetailResp](docs/ProjectDetailResp.md)
 - [ProjectResp](docs/ProjectResp.md)
 - [RegisterReq](docs/RegisterReq.md)
 - [RemoveRolePermissionReq](docs/RemoveRolePermissionReq.md)
 - [ResourceCategory](docs/ResourceCategory.md)
 - [ResourceName](docs/ResourceName.md)
 - [ResourceResp](docs/ResourceResp.md)
 - [ResourceType](docs/ResourceType.md)
 - [RoleDetailResp](docs/RoleDetailResp.md)
 - [RoleResp](docs/RoleResp.md)
 - [SSEEventName](docs/SSEEventName.md)
 - [SearchDatasetReq](docs/SearchDatasetReq.md)
 - [SearchInjectionReq](docs/SearchInjectionReq.md)
 - [SearchRespInjectionDetailResp](docs/SearchRespInjectionDetailResp.md)
 - [SortDirection](docs/SortDirection.md)
 - [SortOption](docs/SortOption.md)
 - [StatusType](docs/StatusType.md)
 - [StreamEvent](docs/StreamEvent.md)
 - [SubmitBuildContainerReq](docs/SubmitBuildContainerReq.md)
 - [SubmitBuildingItem](docs/SubmitBuildingItem.md)
 - [SubmitContainerBuildResp](docs/SubmitContainerBuildResp.md)
 - [SubmitDatapackBuildingReq](docs/SubmitDatapackBuildingReq.md)
 - [SubmitDatapackBuildingResp](docs/SubmitDatapackBuildingResp.md)
 - [SubmitExecutionItem](docs/SubmitExecutionItem.md)
 - [SubmitExecutionReq](docs/SubmitExecutionReq.md)
 - [SubmitExecutionResp](docs/SubmitExecutionResp.md)
 - [SubmitInjectionItem](docs/SubmitInjectionItem.md)
 - [SubmitInjectionReq](docs/SubmitInjectionReq.md)
 - [SubmitInjectionResp](docs/SubmitInjectionResp.md)
 - [SystemMetricsHistoryResp](docs/SystemMetricsHistoryResp.md)
 - [SystemMetricsResp](docs/SystemMetricsResp.md)
 - [TaskDetailResp](docs/TaskDetailResp.md)
 - [TaskResp](docs/TaskResp.md)
 - [TaskState](docs/TaskState.md)
 - [TaskType](docs/TaskType.md)
 - [TraceStatsItem](docs/TraceStatsItem.md)
 - [UpdateLabelReq](docs/UpdateLabelReq.md)
 - [UpdatePermissionReq](docs/UpdatePermissionReq.md)
 - [UpdateRoleReq](docs/UpdateRoleReq.md)
 - [UpdateUserReq](docs/UpdateUserReq.md)
 - [UploadDetectorResultReq](docs/UploadDetectorResultReq.md)
 - [UploadExecutionResultResp](docs/UploadExecutionResultResp.md)
 - [UploadGranularityResultReq](docs/UploadGranularityResultReq.md)
 - [UserContainerInfo](docs/UserContainerInfo.md)
 - [UserDatasetInfo](docs/UserDatasetInfo.md)
 - [UserDetailResp](docs/UserDetailResp.md)
 - [UserInfo](docs/UserInfo.md)
 - [UserProjectInfo](docs/UserProjectInfo.md)
 - [UserResp](docs/UserResp.md)
 - [ValueDataType](docs/ValueDataType.md)


<a id="documentation-for-authorization"></a>
## Documentation For Authorization


Authentication schemes defined for the API:
<a id="BearerAuth"></a>
### BearerAuth

- **Type**: API key
- **API key parameter name**: Authorization
- **Location**: HTTP header

