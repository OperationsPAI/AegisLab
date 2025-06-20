# rcabench.openapi.TraceApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_traces_analyze_get**](TraceApi.md#api_v1_traces_analyze_get) | **GET** /api/v1/traces/analyze | 分析链路数据
[**api_v1_traces_completed_get**](TraceApi.md#api_v1_traces_completed_get) | **GET** /api/v1/traces/completed | 获取完成状态的链路


# **api_v1_traces_analyze_get**
> DtoGenericResponseAny api_v1_traces_analyze_get(first_task_type=first_task_type, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

分析链路数据

使用多种筛选条件分析链路数据

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny
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
    api_instance = rcabench.openapi.TraceApi(api_client)
    first_task_type = 'first_task_type_example' # str | 子任务类型筛选 (optional)
    lookback = 'lookback_example' # str | 相对时间查询，如 1h, 24h, 7d或者是custom (optional)
    custom_start_time = 'custom_start_time_example' # str | 当lookback=custom时必需，自定义开始时间(RFC3339格式) (optional)
    custom_end_time = 'custom_end_time_example' # str | 当lookback=custom时必需，自定义结束时间(RFC3339格式) (optional)

    try:
        # 分析链路数据
        api_response = api_instance.api_v1_traces_analyze_get(first_task_type=first_task_type, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of TraceApi->api_v1_traces_analyze_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TraceApi->api_v1_traces_analyze_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **first_task_type** | **str**| 子任务类型筛选 | [optional] 
 **lookback** | **str**| 相对时间查询，如 1h, 24h, 7d或者是custom | [optional] 
 **custom_start_time** | **str**| 当lookback&#x3D;custom时必需，自定义开始时间(RFC3339格式) | [optional] 
 **custom_end_time** | **str**| 当lookback&#x3D;custom时必需，自定义结束时间(RFC3339格式) | [optional] 

### Return type

[**DtoGenericResponseAny**](DtoGenericResponseAny.md)

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

# **api_v1_traces_completed_get**
> DtoGenericResponseDtoGetCompletedMapResp api_v1_traces_completed_get(lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

获取完成状态的链路

根据指定的时间范围获取完成状态的链路

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_get_completed_map_resp import DtoGenericResponseDtoGetCompletedMapResp
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
    api_instance = rcabench.openapi.TraceApi(api_client)
    lookback = 'lookback_example' # str | 相对时间查询，如 1h, 24h, 7d或者是custom (optional)
    custom_start_time = 'custom_start_time_example' # str | 当lookback=custom时必需，自定义开始时间(RFC3339格式) (optional)
    custom_end_time = 'custom_end_time_example' # str | 当lookback=custom时必需，自定义结束时间(RFC3339格式) (optional)

    try:
        # 获取完成状态的链路
        api_response = api_instance.api_v1_traces_completed_get(lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of TraceApi->api_v1_traces_completed_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TraceApi->api_v1_traces_completed_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **lookback** | **str**| 相对时间查询，如 1h, 24h, 7d或者是custom | [optional] 
 **custom_start_time** | **str**| 当lookback&#x3D;custom时必需，自定义开始时间(RFC3339格式) | [optional] 
 **custom_end_time** | **str**| 当lookback&#x3D;custom时必需，自定义结束时间(RFC3339格式) | [optional] 

### Return type

[**DtoGenericResponseDtoGetCompletedMapResp**](DtoGenericResponseDtoGetCompletedMapResp.md)

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

