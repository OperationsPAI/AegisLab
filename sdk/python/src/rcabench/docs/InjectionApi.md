# rcabench.openapi.InjectionApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_injections_analysis_no_issues_get**](InjectionApi.md#api_v1_injections_analysis_no_issues_get) | **GET** /api/v1/injections/analysis/no-issues | 查询没有问题的故障注入记录
[**api_v1_injections_analysis_statistics_get**](InjectionApi.md#api_v1_injections_analysis_statistics_get) | **GET** /api/v1/injections/analysis/statistics | 获取故障注入统计信息
[**api_v1_injections_analysis_with_issues_get**](InjectionApi.md#api_v1_injections_analysis_with_issues_get) | **GET** /api/v1/injections/analysis/with-issues | 查询有问题的故障注入记录
[**api_v1_injections_conf_get**](InjectionApi.md#api_v1_injections_conf_get) | **GET** /api/v1/injections/conf | 获取故障注入配置
[**api_v1_injections_configs_get**](InjectionApi.md#api_v1_injections_configs_get) | **GET** /api/v1/injections/configs | 获取已注入故障配置列表
[**api_v1_injections_get**](InjectionApi.md#api_v1_injections_get) | **GET** /api/v1/injections | 获取故障注入记录列表
[**api_v1_injections_mapping_get**](InjectionApi.md#api_v1_injections_mapping_get) | **GET** /api/v1/injections/mapping | 获取字段映射关系
[**api_v1_injections_ns_resources_get**](InjectionApi.md#api_v1_injections_ns_resources_get) | **GET** /api/v1/injections/ns-resources | 获取命名空间资源映射
[**api_v1_injections_post**](InjectionApi.md#api_v1_injections_post) | **POST** /api/v1/injections | 提交故障注入任务
[**api_v1_injections_query_get**](InjectionApi.md#api_v1_injections_query_get) | **GET** /api/v1/injections/query | 查询单个故障注入记录
[**api_v1_injections_task_id_cancel_put**](InjectionApi.md#api_v1_injections_task_id_cancel_put) | **PUT** /api/v1/injections/{task_id}/cancel | 取消故障注入任务


# **api_v1_injections_analysis_no_issues_get**
> DtoGenericResponseArrayDtoFaultInjectionNoIssuesResp api_v1_injections_analysis_no_issues_get(env=env, batch=batch, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

查询没有问题的故障注入记录

根据时间范围查询所有没有问题的故障注入记录列表，返回包含配置信息的详细记录

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_array_dto_fault_injection_no_issues_resp import DtoGenericResponseArrayDtoFaultInjectionNoIssuesResp
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
    env = 'env_example' # str | 环境标签过滤 (optional)
    batch = 'batch_example' # str | 批次标签过滤 (optional)
    lookback = 'lookback_example' # str | 时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置 (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | 自定义开始时间，RFC3339格式，当lookback=custom时必需 (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | 自定义结束时间，RFC3339格式，当lookback=custom时必需 (optional)

    try:
        # 查询没有问题的故障注入记录
        api_response = api_instance.api_v1_injections_analysis_no_issues_get(env=env, batch=batch, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of InjectionApi->api_v1_injections_analysis_no_issues_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_analysis_no_issues_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **env** | **str**| 环境标签过滤 | [optional] 
 **batch** | **str**| 批次标签过滤 | [optional] 
 **lookback** | **str**| 时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置 | [optional] 
 **custom_start_time** | **datetime**| 自定义开始时间，RFC3339格式，当lookback&#x3D;custom时必需 | [optional] 
 **custom_end_time** | **datetime**| 自定义结束时间，RFC3339格式，当lookback&#x3D;custom时必需 | [optional] 

### Return type

[**DtoGenericResponseArrayDtoFaultInjectionNoIssuesResp**](DtoGenericResponseArrayDtoFaultInjectionNoIssuesResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | 成功返回没有问题的故障注入记录列表 |  -  |
**400** | 请求参数错误，如时间格式不正确或参数验证失败等 |  -  |
**500** | 服务器内部错 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_analysis_statistics_get**
> DtoGenericResponseDtoFaultInjectionStatisticsResp api_v1_injections_analysis_statistics_get(lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

获取故障注入统计信息

获取故障注入记录的统计信息，包括有问题、没有问题和总记录数量

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_fault_injection_statistics_resp import DtoGenericResponseDtoFaultInjectionStatisticsResp
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
    lookback = 'lookback_example' # str | 时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置 (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | 自定义开始时间，RFC3339格式，当lookback=custom时必需 (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | 自定义结束时间，RFC3339格式，当lookback=custom时必需 (optional)

    try:
        # 获取故障注入统计信息
        api_response = api_instance.api_v1_injections_analysis_statistics_get(lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of InjectionApi->api_v1_injections_analysis_statistics_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_analysis_statistics_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **lookback** | **str**| 时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置 | [optional] 
 **custom_start_time** | **datetime**| 自定义开始时间，RFC3339格式，当lookback&#x3D;custom时必需 | [optional] 
 **custom_end_time** | **datetime**| 自定义结束时间，RFC3339格式，当lookback&#x3D;custom时必需 | [optional] 

### Return type

[**DtoGenericResponseDtoFaultInjectionStatisticsResp**](DtoGenericResponseDtoFaultInjectionStatisticsResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | 成功返回故障注入统计信息 |  -  |
**400** | 请求参数错误，如时间格式不正确或参数验证失败等 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_analysis_with_issues_get**
> DtoGenericResponseArrayDtoFaultInjectionWithIssuesResp api_v1_injections_analysis_with_issues_get(env=env, batch=batch, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

查询有问题的故障注入记录

根据时间范围查询所有有问题的故障注入记录列表

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_array_dto_fault_injection_with_issues_resp import DtoGenericResponseArrayDtoFaultInjectionWithIssuesResp
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
    env = 'env_example' # str | 环境标签过滤 (optional)
    batch = 'batch_example' # str | 批次标签过滤 (optional)
    lookback = 'lookback_example' # str | 时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置 (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | 自定义开始时间，RFC3339格式，当lookback=custom时必需 (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | 自定义结束时间，RFC3339格式，当lookback=custom时必需 (optional)

    try:
        # 查询有问题的故障注入记录
        api_response = api_instance.api_v1_injections_analysis_with_issues_get(env=env, batch=batch, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of InjectionApi->api_v1_injections_analysis_with_issues_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_analysis_with_issues_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **env** | **str**| 环境标签过滤 | [optional] 
 **batch** | **str**| 批次标签过滤 | [optional] 
 **lookback** | **str**| 时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置 | [optional] 
 **custom_start_time** | **datetime**| 自定义开始时间，RFC3339格式，当lookback&#x3D;custom时必需 | [optional] 
 **custom_end_time** | **datetime**| 自定义结束时间，RFC3339格式，当lookback&#x3D;custom时必需 | [optional] 

### Return type

[**DtoGenericResponseArrayDtoFaultInjectionWithIssuesResp**](DtoGenericResponseArrayDtoFaultInjectionWithIssuesResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | OK |  -  |
**400** | 请求参数错误，如时间格式不正确或参数验证失败等 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_conf_get**
> DtoGenericResponseHandlerNode api_v1_injections_conf_get(namespace, mode=mode)

获取故障注入配置

获取指定命名空间的故障注入配置信息，支持不同显示模式的配置树结构

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
    namespace = 'namespace_example' # str | 命名空间，指定要获取配置的命名空间
    mode = engine # str | 显示模式 (optional) (default to engine)

    try:
        # 获取故障注入配置
        api_response = api_instance.api_v1_injections_conf_get(namespace, mode=mode)
        print("The response of InjectionApi->api_v1_injections_conf_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_conf_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **namespace** | **str**| 命名空间，指定要获取配置的命名空间 | 
 **mode** | **str**| 显示模式 | [optional] [default to engine]

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
**200** | 成功返回配置树结构 |  -  |
**400** | 请求参数错误，如命名空间或模式参数缺失 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_configs_get**
> DtoGenericResponseAny api_v1_injections_configs_get(trace_ids=trace_ids)

获取已注入故障配置列表

根据多个TraceID获取对应的故障注入配置信息，用于查看已提交的故障注入任务的配置详情

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
    trace_ids = ['trace_ids_example'] # List[str] | TraceID列表，支持多个值，用于查询对应的配置信息 (optional)

    try:
        # 获取已注入故障配置列表
        api_response = api_instance.api_v1_injections_configs_get(trace_ids=trace_ids)
        print("The response of InjectionApi->api_v1_injections_configs_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_configs_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trace_ids** | [**List[str]**](str.md)| TraceID列表，支持多个值，用于查询对应的配置信息 | [optional] 

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
**200** | 成功返回配置列表 |  -  |
**400** | 请求参数错误，如TraceID参数缺失或格式不正确 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_get**
> DtoGenericResponseArrayDatabaseFaultInjectionSchedule api_v1_injections_get(env=env, batch=batch, benchmark=benchmark, status=status, fault_type=fault_type, sort=sort, limit=limit, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

获取故障注入记录列表

支持排序、过滤的故障注入记录查询接口。返回数据库原始记录列表，不进行数据转换。

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_array_database_fault_injection_schedule import DtoGenericResponseArrayDatabaseFaultInjectionSchedule
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
    env = 'env_example' # str | 环境标签过滤 (optional)
    batch = 'batch_example' # str | 批次标签过滤 (optional)
    benchmark = 'benchmark_example' # str | 基准测试类型过滤 (optional)
    status = 0 # int | 状态过滤，具体值参考字段映射接口(/mapping) (optional) (default to 0)
    fault_type = 0 # int | 故障类型过滤，具体值参考字段映射接口(/mapping) (optional) (default to 0)
    sort = desc # str | 排序方式，默认desc。按created_at字段排序 (optional) (default to desc)
    limit = 56 # int | 结果数量限制，用于控制返回记录数量 (optional)
    lookback = 'lookback_example' # str | 时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置 (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | 自定义开始时间，RFC3339格式，当lookback=custom时必需 (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | 自定义结束时间，RFC3339格式，当lookback=custom时必需 (optional)

    try:
        # 获取故障注入记录列表
        api_response = api_instance.api_v1_injections_get(env=env, batch=batch, benchmark=benchmark, status=status, fault_type=fault_type, sort=sort, limit=limit, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of InjectionApi->api_v1_injections_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **env** | **str**| 环境标签过滤 | [optional] 
 **batch** | **str**| 批次标签过滤 | [optional] 
 **benchmark** | **str**| 基准测试类型过滤 | [optional] 
 **status** | **int**| 状态过滤，具体值参考字段映射接口(/mapping) | [optional] [default to 0]
 **fault_type** | **int**| 故障类型过滤，具体值参考字段映射接口(/mapping) | [optional] [default to 0]
 **sort** | **str**| 排序方式，默认desc。按created_at字段排序 | [optional] [default to desc]
 **limit** | **int**| 结果数量限制，用于控制返回记录数量 | [optional] 
 **lookback** | **str**| 时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置 | [optional] 
 **custom_start_time** | **datetime**| 自定义开始时间，RFC3339格式，当lookback&#x3D;custom时必需 | [optional] 
 **custom_end_time** | **datetime**| 自定义结束时间，RFC3339格式，当lookback&#x3D;custom时必需 | [optional] 

### Return type

[**DtoGenericResponseArrayDatabaseFaultInjectionSchedule**](DtoGenericResponseArrayDatabaseFaultInjectionSchedule.md)

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

# **api_v1_injections_mapping_get**
> DtoGenericResponseDtoInjectionFieldMappingResp api_v1_injections_mapping_get()

获取字段映射关系

获取状态和故障类型的字符串与数字映射关系，用于前端显示和API参数验证

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_injection_field_mapping_resp import DtoGenericResponseDtoInjectionFieldMappingResp
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
        # 获取字段映射关系
        api_response = api_instance.api_v1_injections_mapping_get()
        print("The response of InjectionApi->api_v1_injections_mapping_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_mapping_get: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**DtoGenericResponseDtoInjectionFieldMappingResp**](DtoGenericResponseDtoInjectionFieldMappingResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | 成功返回字段映射关系 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_ns_resources_get**
> DtoGenericResponseHandlerResources api_v1_injections_ns_resources_get(namespace=namespace)

获取命名空间资源映射

获取所有命名空间及其对应的资源信息映射，或查询指定命名空间的资源信息。返回命名空间到资源的映射表，用于故障注入配置和资源管理

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_handler_resources import DtoGenericResponseHandlerResources
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
    namespace = 'namespace_example' # str | 命名空间名称，不指定时返回所有命名空间的资源映射 (optional)

    try:
        # 获取命名空间资源映射
        api_response = api_instance.api_v1_injections_ns_resources_get(namespace=namespace)
        print("The response of InjectionApi->api_v1_injections_ns_resources_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_ns_resources_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **namespace** | **str**| 命名空间名称，不指定时返回所有命名空间的资源映射 | [optional] 

### Return type

[**DtoGenericResponseHandlerResources**](DtoGenericResponseHandlerResources.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | 指定命名空间时返回该命名空间的资源信息 |  -  |
**404** | 指定的命名空间不存在 |  -  |
**500** | 服务器内部错误，无法获取资源映射 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_post**
> DtoGenericResponseDtoSubmitInjectionResp api_v1_injections_post(body)

提交故障注入任务

提交故障注入任务，支持批量提交多个故障配置，系统会自动去重并返回提交结果

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_submit_injection_resp import DtoGenericResponseDtoSubmitInjectionResp
from rcabench.openapi.models.dto_submit_injection_req import DtoSubmitInjectionReq
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
    body = rcabench.openapi.DtoSubmitInjectionReq() # DtoSubmitInjectionReq | 故障注入请求体

    try:
        # 提交故障注入任务
        api_response = api_instance.api_v1_injections_post(body)
        print("The response of InjectionApi->api_v1_injections_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**DtoSubmitInjectionReq**](DtoSubmitInjectionReq.md)| 故障注入请求体 | 

### Return type

[**DtoGenericResponseDtoSubmitInjectionResp**](DtoGenericResponseDtoSubmitInjectionResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**202** | 成功提交故障注入任务 |  -  |
**400** | 请求参数错误，如JSON格式不正确、参数验证失败或算法无效等 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_query_get**
> DtoGenericResponseDtoQueryInjectionResp api_v1_injections_query_get(name=name, task_id=task_id)

查询单个故障注入记录

根据名称或任务ID查询故障注入记录详情，两个参数至少提供一个

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_query_injection_resp import DtoGenericResponseDtoQueryInjectionResp
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
    name = 'name_example' # str | 故障注入名称 (optional)
    task_id = 'task_id_example' # str | 任务ID (optional)

    try:
        # 查询单个故障注入记录
        api_response = api_instance.api_v1_injections_query_get(name=name, task_id=task_id)
        print("The response of InjectionApi->api_v1_injections_query_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_query_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **str**| 故障注入名称 | [optional] 
 **task_id** | **str**| 任务ID | [optional] 

### Return type

[**DtoGenericResponseDtoQueryInjectionResp**](DtoGenericResponseDtoQueryInjectionResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | 成功返回故障注入记录详情 |  -  |
**400** | 请求参数错误，如参数缺失、格式不正确或验证失败等 |  -  |
**500** | 服务器内部错误 |  -  |

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

