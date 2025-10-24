# rcabench.openapi.TaskApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_tasks_get**](TaskApi.md#api_v1_tasks_get) | **GET** /api/v1/tasks | Get task list
[**api_v1_tasks_queue_get**](TaskApi.md#api_v1_tasks_queue_get) | **GET** /api/v1/tasks/queue | Get queued tasks
[**api_v1_tasks_task_id_get**](TaskApi.md#api_v1_tasks_task_id_get) | **GET** /api/v1/tasks/{task_id} | Get task detail


# **api_v1_tasks_get**
> DtoGenericResponseDtoListTasksResp api_v1_tasks_get(task_id=task_id, trace_id=trace_id, group_id=group_id, task_type=task_type, status=status, immediate=immediate, sort_field=sort_field, sort_order=sort_order, limit=limit, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

Get task list

Paginate and get task list by multiple conditions. Supports exact query by task ID, trace ID, group ID, or filter by type, status, etc.

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_list_tasks_resp import DtoGenericResponseDtoListTasksResp
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.TaskApi(api_client)
    task_id = 'task_id_example' # str | Task ID - exact match (mutually exclusive with trace_id, group_id) (optional)
    trace_id = 'trace_id_example' # str | Trace ID - find all tasks in the same trace (mutually exclusive with task_id, group_id) (optional)
    group_id = 'group_id_example' # str | Group ID - find all tasks in the same group (mutually exclusive with task_id, trace_id) (optional)
    task_type = 'task_type_example' # str | Task type filter (optional)
    status = 'status_example' # str | Task status filter (optional)
    immediate = True # bool | Immediate execution - true: immediate, false: delayed (optional)
    sort_field = 'created_at' # str | Sort field, default created_at (optional) (default to 'created_at')
    sort_order = desc # str | Sort order, default desc (optional) (default to desc)
    limit = 56 # int | Result limit, controls number of records returned (optional)
    lookback = 'lookback_example' # str | Time range query, supports relative time (1h/24h/7d) or custom, default unset (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | Custom start time, RFC3339 format, required if lookback=custom (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | Custom end time, RFC3339 format, required if lookback=custom (optional)

    try:
        # Get task list
        api_response = api_instance.api_v1_tasks_get(task_id=task_id, trace_id=trace_id, group_id=group_id, task_type=task_type, status=status, immediate=immediate, sort_field=sort_field, sort_order=sort_order, limit=limit, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of TaskApi->api_v1_tasks_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TaskApi->api_v1_tasks_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **task_id** | **str**| Task ID - exact match (mutually exclusive with trace_id, group_id) | [optional] 
 **trace_id** | **str**| Trace ID - find all tasks in the same trace (mutually exclusive with task_id, group_id) | [optional] 
 **group_id** | **str**| Group ID - find all tasks in the same group (mutually exclusive with task_id, trace_id) | [optional] 
 **task_type** | **str**| Task type filter | [optional] 
 **status** | **str**| Task status filter | [optional] 
 **immediate** | **bool**| Immediate execution - true: immediate, false: delayed | [optional] 
 **sort_field** | **str**| Sort field, default created_at | [optional] [default to &#39;created_at&#39;]
 **sort_order** | **str**| Sort order, default desc | [optional] [default to desc]
 **limit** | **int**| Result limit, controls number of records returned | [optional] 
 **lookback** | **str**| Time range query, supports relative time (1h/24h/7d) or custom, default unset | [optional] 
 **custom_start_time** | **datetime**| Custom start time, RFC3339 format, required if lookback&#x3D;custom | [optional] 
 **custom_end_time** | **datetime**| Custom end time, RFC3339 format, required if lookback&#x3D;custom | [optional] 

### Return type

[**DtoGenericResponseDtoListTasksResp**](DtoGenericResponseDtoListTasksResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Successfully returned fault injection record list |  -  |
**400** | Request parameter error, e.g. invalid format or validation failed |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_tasks_queue_get**
> DtoGenericResponseDtoPaginationRespDtoUnifiedTask api_v1_tasks_queue_get(page_num=page_num, page_size=page_size)

Get queued tasks

Paginate and get the list of tasks waiting in the queue

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_pagination_resp_dto_unified_task import DtoGenericResponseDtoPaginationRespDtoUnifiedTask
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.TaskApi(api_client)
    page_num = 1 # int | Page number (optional) (default to 1)
    page_size = 10 # int | Page size (optional) (default to 10)

    try:
        # Get queued tasks
        api_response = api_instance.api_v1_tasks_queue_get(page_num=page_num, page_size=page_size)
        print("The response of TaskApi->api_v1_tasks_queue_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TaskApi->api_v1_tasks_queue_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page_num** | **int**| Page number | [optional] [default to 1]
 **page_size** | **int**| Page size | [optional] [default to 10]

### Return type

[**DtoGenericResponseDtoPaginationRespDtoUnifiedTask**](DtoGenericResponseDtoPaginationRespDtoUnifiedTask.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | OK |  -  |
**400** | Bad Request |  -  |
**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_tasks_task_id_get**
> DtoGenericResponseDtoTaskDetailResp api_v1_tasks_task_id_get(task_id)

Get task detail

Get detailed information of a task by task ID, including basic info and execution logs

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_task_detail_resp import DtoGenericResponseDtoTaskDetailResp
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.TaskApi(api_client)
    task_id = 'task_id_example' # str | Task ID

    try:
        # Get task detail
        api_response = api_instance.api_v1_tasks_task_id_get(task_id)
        print("The response of TaskApi->api_v1_tasks_task_id_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TaskApi->api_v1_tasks_task_id_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **task_id** | **str**| Task ID | 

### Return type

[**DtoGenericResponseDtoTaskDetailResp**](DtoGenericResponseDtoTaskDetailResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | OK |  -  |
**400** | Invalid task ID |  -  |
**404** | Task not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

