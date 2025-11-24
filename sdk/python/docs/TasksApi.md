# rcabench.openapi.TasksApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**get_task_by_id**](TasksApi.md#get_task_by_id) | **GET** /api/v2/tasks/{task_id} | Get task by ID
[**list_tasks**](TasksApi.md#list_tasks) | **GET** /api/v2/tasks | List tasks


# **get_task_by_id**
> GenericResponseTaskDetailResp get_task_by_id(task_id)

Get task by ID

Get detailed information about a specific task

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.generic_response_task_detail_resp import GenericResponseTaskDetailResp
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure API key authorization: BearerAuth
configuration.api_key['BearerAuth'] = os.environ["API_KEY"]

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['BearerAuth'] = 'Bearer'

# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.TasksApi(api_client)
    task_id = 'task_id_example' # str | Task ID

    try:
        # Get task by ID
        api_response = api_instance.get_task_by_id(task_id)
        print("The response of TasksApi->get_task_by_id:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TasksApi->get_task_by_id: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **task_id** | **str**| Task ID | 

### Return type

[**GenericResponseTaskDetailResp**](GenericResponseTaskDetailResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Task retrieved successfully |  -  |
**400** | Invalid task ID |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Task not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_tasks**
> GenericResponseListTaskResp list_tasks(page=page, size=size, task_type=task_type, immediate=immediate, trace_id=trace_id, group_id=group_id, project_id=project_id, state=state, status=status)

List tasks

Get a simple list of tasks with basic filtering via query parameters

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.generic_response_list_task_resp import GenericResponseListTaskResp
from rcabench.openapi.models.status_type import StatusType
from rcabench.openapi.models.task_state import TaskState
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure API key authorization: BearerAuth
configuration.api_key['BearerAuth'] = os.environ["API_KEY"]

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['BearerAuth'] = 'Bearer'

# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.TasksApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)
    task_type = 56 # int | Filter by task type (optional)
    immediate = True # bool | Filter by immediate execution (optional)
    trace_id = 'trace_id_example' # str | Filter by trace ID (uuid format) (optional)
    group_id = 'group_id_example' # str | Filter by group ID (uuid format) (optional)
    project_id = 56 # int | Filter by project ID (optional)
    state = rcabench.openapi.TaskState() # TaskState | Filter by state (optional)
    status = rcabench.openapi.StatusType() # StatusType | Filter by status (optional)

    try:
        # List tasks
        api_response = api_instance.list_tasks(page=page, size=size, task_type=task_type, immediate=immediate, trace_id=trace_id, group_id=group_id, project_id=project_id, state=state, status=status)
        print("The response of TasksApi->list_tasks:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TasksApi->list_tasks: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]
 **task_type** | **int**| Filter by task type | [optional] 
 **immediate** | **bool**| Filter by immediate execution | [optional] 
 **trace_id** | **str**| Filter by trace ID (uuid format) | [optional] 
 **group_id** | **str**| Filter by group ID (uuid format) | [optional] 
 **project_id** | **int**| Filter by project ID | [optional] 
 **state** | [**TaskState**](.md)| Filter by state | [optional] 
 **status** | [**StatusType**](.md)| Filter by status | [optional] 

### Return type

[**GenericResponseListTaskResp**](GenericResponseListTaskResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Tasks retrieved successfully |  -  |
**400** | Invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

