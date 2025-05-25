# rcabench.openapi.InjectionApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_injections_conf_get**](InjectionApi.md#api_v1_injections_conf_get) | **GET** /api/v1/injections/conf | 获取故障注入配置
[**api_v1_injections_configs_get**](InjectionApi.md#api_v1_injections_configs_get) | **GET** /api/v1/injections/configs | 获取故障注入配置列表
[**api_v1_injections_get**](InjectionApi.md#api_v1_injections_get) | **GET** /api/v1/injections | 分页查询注入记录列表
[**api_v1_injections_ns_status_get**](InjectionApi.md#api_v1_injections_ns_status_get) | **GET** /api/v1/injections/ns/status | 获取命名空间锁状态
[**api_v1_injections_post**](InjectionApi.md#api_v1_injections_post) | **POST** /api/v1/injections | 注入故障
[**api_v1_injections_query_get**](InjectionApi.md#api_v1_injections_query_get) | **GET** /api/v1/injections/query | 查询故障注入记录
[**api_v1_injections_task_id_cancel_put**](InjectionApi.md#api_v1_injections_task_id_cancel_put) | **PUT** /api/v1/injections/{task_id}/cancel | 取消故障注入任务


# **api_v1_injections_conf_get**
> DtoGenericResponseHandlerNode api_v1_injections_conf_get(namespace, mode)

获取故障注入配置

获取指定命名空间的故障注入配置信息

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_handler_node import DtoGenericResponseHandlerNode
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    namespace = 'namespace_example' # str | 命名空间
    mode = 'mode_example' # str | 显示模式(display/engine)

    try:
        # 获取故障注入配置
        api_response = api_instance.api_v1_injections_conf_get(namespace, mode)
        print("The response of InjectionApi->api_v1_injections_conf_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_conf_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **namespace** | **str**| 命名空间 | 
 **mode** | **str**| 显示模式(display/engine) | 

### Return type

[**DtoGenericResponseHandlerNode**](DtoGenericResponseHandlerNode.md)

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

# **api_v1_injections_configs_get**
> DtoGenericResponseAny api_v1_injections_configs_get(trace_ids)

获取故障注入配置列表

根据多个 TraceID 获取对应的故障注入配置信息

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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    trace_ids = ['trace_ids_example'] # List[str] | Trace ID 列表

    try:
        # 获取故障注入配置列表
        api_response = api_instance.api_v1_injections_configs_get(trace_ids)
        print("The response of InjectionApi->api_v1_injections_configs_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_configs_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trace_ids** | [**List[str]**](str.md)| Trace ID 列表 | 

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

# **api_v1_injections_get**
> DtoGenericResponseDtoPaginationRespDtoInjectionItem api_v1_injections_get(page_num=page_num, page_size=page_size)

分页查询注入记录列表

获取注入记录列表（支持分页参数）

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_pagination_resp_dto_injection_item import DtoGenericResponseDtoPaginationRespDtoInjectionItem
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    page_num = 1 # int | 页码 (optional) (default to 1)
    page_size = 10 # int | 每页大小 (optional) (default to 10)

    try:
        # 分页查询注入记录列表
        api_response = api_instance.api_v1_injections_get(page_num=page_num, page_size=page_size)
        print("The response of InjectionApi->api_v1_injections_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page_num** | **int**| 页码 | [optional] [default to 1]
 **page_size** | **int**| 每页大小 | [optional] [default to 10]

### Return type

[**DtoGenericResponseDtoPaginationRespDtoInjectionItem**](DtoGenericResponseDtoPaginationRespDtoInjectionItem.md)

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

# **api_v1_injections_ns_status_get**
> DtoGenericResponseAny api_v1_injections_ns_status_get()

获取命名空间锁状态

获取命名空间锁状态信息

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
    api_instance = rcabench.openapi.InjectionApi(api_client)

    try:
        # 获取命名空间锁状态
        api_response = api_instance.api_v1_injections_ns_status_get()
        print("The response of InjectionApi->api_v1_injections_ns_status_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_ns_status_get: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

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
**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_post**
> DtoGenericResponseDtoSubmitResp api_v1_injections_post(body)

注入故障

注入故障

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_submit_resp import DtoGenericResponseDtoSubmitResp
from rcabench.openapi.models.dto_injection_submit_req import DtoInjectionSubmitReq
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    body = rcabench.openapi.DtoInjectionSubmitReq() # DtoInjectionSubmitReq | 请求体

    try:
        # 注入故障
        api_response = api_instance.api_v1_injections_post(body)
        print("The response of InjectionApi->api_v1_injections_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**DtoInjectionSubmitReq**](DtoInjectionSubmitReq.md)| 请求体 | 

### Return type

[**DtoGenericResponseDtoSubmitResp**](DtoGenericResponseDtoSubmitResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**202** | Accepted |  -  |
**400** | Bad Request |  -  |
**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_query_get**
> DtoGenericResponseDtoInjectionItem api_v1_injections_query_get(name=name, task_id=task_id)

查询故障注入记录

根据名称或任务ID查询故障注入记录详情

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_injection_item import DtoGenericResponseDtoInjectionItem
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    name = 'name_example' # str | 注入名称 (optional)
    task_id = 'task_id_example' # str | 任务ID (optional)

    try:
        # 查询故障注入记录
        api_response = api_instance.api_v1_injections_query_get(name=name, task_id=task_id)
        print("The response of InjectionApi->api_v1_injections_query_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_query_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **str**| 注入名称 | [optional] 
 **task_id** | **str**| 任务ID | [optional] 

### Return type

[**DtoGenericResponseDtoInjectionItem**](DtoGenericResponseDtoInjectionItem.md)

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

# **api_v1_injections_task_id_cancel_put**
> DtoGenericResponseDtoInjectCancelResp api_v1_injections_task_id_cancel_put(task_id)

取消故障注入任务

取消指定的故障注入任务

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_inject_cancel_resp import DtoGenericResponseDtoInjectCancelResp
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    task_id = 'task_id_example' # str | 任务ID

    try:
        # 取消故障注入任务
        api_response = api_instance.api_v1_injections_task_id_cancel_put(task_id)
        print("The response of InjectionApi->api_v1_injections_task_id_cancel_put:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_task_id_cancel_put: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **task_id** | **str**| 任务ID | 

### Return type

[**DtoGenericResponseDtoInjectCancelResp**](DtoGenericResponseDtoInjectCancelResp.md)

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

