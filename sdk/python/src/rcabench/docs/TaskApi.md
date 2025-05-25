# rcabench.openapi.TaskApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_tasks_list_get**](TaskApi.md#api_v1_tasks_list_get) | **GET** /api/v1/tasks/list | 获取任务列表
[**api_v1_tasks_queue_get**](TaskApi.md#api_v1_tasks_queue_get) | **GET** /api/v1/tasks/queue | 获取队列中的任务
[**api_v1_tasks_task_id_get**](TaskApi.md#api_v1_tasks_task_id_get) | **GET** /api/v1/tasks/{task_id} | 获取任务详情


# **api_v1_tasks_list_get**
> DtoGenericResponseDtoPaginationRespDtoTaskItem api_v1_tasks_list_get(page_num=page_num, page_size=page_size, sort_field=sort_field)

获取任务列表

分页获取任务列表

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_pagination_resp_dto_task_item import DtoGenericResponseDtoPaginationRespDtoTaskItem
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
    sort_field = 'sort_field_example' # str | 排序字段 (optional)

    try:
        # 获取任务列表
        api_response = api_instance.api_v1_tasks_list_get(page_num=page_num, page_size=page_size, sort_field=sort_field)
        print("The response of TaskApi->api_v1_tasks_list_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TaskApi->api_v1_tasks_list_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page_num** | **int**| 页码 | [optional] [default to 1]
 **page_size** | **int**| 每页大小 | [optional] [default to 10]
 **sort_field** | **str**| 排序字段 | [optional] 

### Return type

[**DtoGenericResponseDtoPaginationRespDtoTaskItem**](DtoGenericResponseDtoPaginationRespDtoTaskItem.md)

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

