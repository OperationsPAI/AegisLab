# rcabench.openapi.TaskApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_tasks_get**](TaskApi.md#api_v1_tasks_get) | **GET** /api/v1/tasks | 获取任务列表
[**api_v1_tasks_queue_get**](TaskApi.md#api_v1_tasks_queue_get) | **GET** /api/v1/tasks/queue | 获取队列中的任务
[**api_v1_tasks_task_id_get**](TaskApi.md#api_v1_tasks_task_id_get) | **GET** /api/v1/tasks/{task_id} | 获取任务详情


# **api_v1_tasks_get**
> DtoGenericResponseDtoListTasksResp api_v1_tasks_get(task_id=task_id, trace_id=trace_id, group_id=group_id, task_type=task_type, status=status, immediate=immediate, sort_field=sort_field, sort_order=sort_order, limit=limit, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

获取任务列表

根据多种条件分页获取任务列表。支持按任务ID、跟踪ID、组ID进行精确查询，或按任务类型、状态等进行过滤查询

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_list_tasks_resp import DtoGenericResponseDtoListTasksResp
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8080/api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8080/api/v1"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.TaskApi(api_client)
    task_id = 'task_id_example' # str | 任务ID - 精确匹配特定任务 (与trace_id、group_id互斥) (optional)
    trace_id = 'trace_id_example' # str | 跟踪ID - 查找属于同一跟踪的所有任务 (与task_id、group_id互斥) (optional)
    group_id = 'group_id_example' # str | 组ID - 查找属于同一组的所有任务 (与task_id、trace_id互斥) (optional)
    task_type = 'task_type_example' # str | 任务类型过滤 (optional)
    status = 'status_example' # str | 任务状态过滤 (optional)
    immediate = True # bool | 是否立即执行 - true:立即执行任务, false:延时执行任务 (optional)
    sort_field = 'created_at' # str | 排序字段，默认created_at (optional) (default to 'created_at')
    sort_order = desc # str | 排序方式，默认desc (optional) (default to desc)
    limit = 56 # int | 结果数量限制，用于控制返回记录数量 (optional)
    lookback = 'lookback_example' # str | 时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置 (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | 自定义开始时间，RFC3339格式，当lookback=custom时必需 (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | 自定义结束时间，RFC3339格式，当lookback=custom时必需 (optional)

    try:
        # 获取任务列表
        api_response = api_instance.api_v1_tasks_get(task_id=task_id, trace_id=trace_id, group_id=group_id, task_type=task_type, status=status, immediate=immediate, sort_field=sort_field, sort_order=sort_order, limit=limit, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of TaskApi->api_v1_tasks_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TaskApi->api_v1_tasks_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **task_id** | **str**| 任务ID - 精确匹配特定任务 (与trace_id、group_id互斥) | [optional] 
 **trace_id** | **str**| 跟踪ID - 查找属于同一跟踪的所有任务 (与task_id、group_id互斥) | [optional] 
 **group_id** | **str**| 组ID - 查找属于同一组的所有任务 (与task_id、trace_id互斥) | [optional] 
 **task_type** | **str**| 任务类型过滤 | [optional] 
 **status** | **str**| 任务状态过滤 | [optional] 
 **immediate** | **bool**| 是否立即执行 - true:立即执行任务, false:延时执行任务 | [optional] 
 **sort_field** | **str**| 排序字段，默认created_at | [optional] [default to &#39;created_at&#39;]
 **sort_order** | **str**| 排序方式，默认desc | [optional] [default to desc]
 **limit** | **int**| 结果数量限制，用于控制返回记录数量 | [optional] 
 **lookback** | **str**| 时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置 | [optional] 
 **custom_start_time** | **datetime**| 自定义开始时间，RFC3339格式，当lookback&#x3D;custom时必需 | [optional] 
 **custom_end_time** | **datetime**| 自定义结束时间，RFC3339格式，当lookback&#x3D;custom时必需 | [optional] 

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
**200** | 成功返回故障注入记录列表 |  -  |
**400** | 请求参数错误，如参数格式不正确、验证失败等 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_tasks_queue_get**
> DtoGenericResponseDtoPaginationRespDtoUnifiedTask api_v1_tasks_queue_get(page_num=page_num, page_size=page_size)

获取队列中的任务

分页获取队列中等待执行的任务列表

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_pagination_resp_dto_unified_task import DtoGenericResponseDtoPaginationRespDtoUnifiedTask
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8080/api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8080/api/v1"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.TaskApi(api_client)
    page_num = 1 # int | 页码 (optional) (default to 1)
    page_size = 10 # int | 每页大小 (optional) (default to 10)

    try:
        # 获取队列中的任务
        api_response = api_instance.api_v1_tasks_queue_get(page_num=page_num, page_size=page_size)
        print("The response of TaskApi->api_v1_tasks_queue_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TaskApi->api_v1_tasks_queue_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page_num** | **int**| 页码 | [optional] [default to 1]
 **page_size** | **int**| 每页大小 | [optional] [default to 10]

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

获取任务详情

根据任务ID获取任务详细信息,包括任务基本信息和执行日志

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_task_detail_resp import DtoGenericResponseDtoTaskDetailResp
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8080/api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8080/api/v1"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.TaskApi(api_client)
    task_id = 'task_id_example' # str | 任务ID

    try:
        # 获取任务详情
        api_response = api_instance.api_v1_tasks_task_id_get(task_id)
        print("The response of TaskApi->api_v1_tasks_task_id_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TaskApi->api_v1_tasks_task_id_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **task_id** | **str**| 任务ID | 

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
**400** | 无效的任务ID |  -  |
**404** | 任务不存在 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

