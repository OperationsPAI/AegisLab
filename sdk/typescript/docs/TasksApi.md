# TasksApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getTaskById**](#gettaskbyid) | **GET** /api/v2/tasks/{task_id} | Get task by ID|
|[**listTasks**](#listtasks) | **GET** /api/v2/tasks | List tasks|

# **getTaskById**
> GenericResponseTaskDetailResp getTaskById()

Get detailed information about a specific task

### Example

```typescript
import {
    TasksApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new TasksApi(configuration);

let taskId: string; //Task ID (default to undefined)

const { status, data } = await apiInstance.getTaskById(
    taskId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **taskId** | [**string**] | Task ID | defaults to undefined|


### Return type

**GenericResponseTaskDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Task retrieved successfully |  -  |
|**400** | Invalid task ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Task not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listTasks**
> GenericResponseListTaskResp listTasks()

Get a simple list of tasks with basic filtering via query parameters

### Example

```typescript
import {
    TasksApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new TasksApi(configuration);

let page: number; //Page number (optional) (default to 1)
let size: number; //Page size (optional) (default to 20)
let taskType: 0 | 1 | 2 | 3 | 4 | 5 | 6; //Filter by task type (optional) (default to undefined)
let immediate: boolean; //Filter by immediate execution (optional) (default to undefined)
let traceId: string; //Filter by trace ID (uuid format) (optional) (default to undefined)
let groupId: string; //Filter by group ID (uuid format) (optional) (default to undefined)
let projectId: number; //Filter by project ID (optional) (default to undefined)
let state: TaskState; //Filter by state (optional) (default to undefined)
let status: StatusType; //Filter by status (optional) (default to undefined)

const { status, data } = await apiInstance.listTasks(
    page,
    size,
    taskType,
    immediate,
    traceId,
    groupId,
    projectId,
    state,
    status
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **size** | [**number**] | Page size | (optional) defaults to 20|
| **taskType** | [**0 | 1 | 2 | 3 | 4 | 5 | 6**]**Array<0 &#124; 1 &#124; 2 &#124; 3 &#124; 4 &#124; 5 &#124; 6>** | Filter by task type | (optional) defaults to undefined|
| **immediate** | [**boolean**] | Filter by immediate execution | (optional) defaults to undefined|
| **traceId** | [**string**] | Filter by trace ID (uuid format) | (optional) defaults to undefined|
| **groupId** | [**string**] | Filter by group ID (uuid format) | (optional) defaults to undefined|
| **projectId** | [**number**] | Filter by project ID | (optional) defaults to undefined|
| **state** | **TaskState** | Filter by state | (optional) defaults to undefined|
| **status** | **StatusType** | Filter by status | (optional) defaults to undefined|


### Return type

**GenericResponseListTaskResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Tasks retrieved successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

