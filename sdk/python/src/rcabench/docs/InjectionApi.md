# rcabench.openapi.InjectionApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_injections_analysis_no_issues_get**](InjectionApi.md#api_v1_injections_analysis_no_issues_get) | **GET** /api/v1/injections/analysis/no-issues | Query Fault Injection Records Without Issues
[**api_v1_injections_analysis_stats_get**](InjectionApi.md#api_v1_injections_analysis_stats_get) | **GET** /api/v1/injections/analysis/stats | Get Fault Injection Statistics
[**api_v1_injections_analysis_with_issues_get**](InjectionApi.md#api_v1_injections_analysis_with_issues_get) | **GET** /api/v1/injections/analysis/with-issues | Query Fault Injection Records With Issues
[**api_v1_injections_conf_get**](InjectionApi.md#api_v1_injections_conf_get) | **GET** /api/v1/injections/conf | Get Fault Injection Configuration
[**api_v1_injections_configs_get**](InjectionApi.md#api_v1_injections_configs_get) | **GET** /api/v1/injections/configs | Get Injected Fault Configuration List
[**api_v1_injections_get**](InjectionApi.md#api_v1_injections_get) | **GET** /api/v1/injections | Get Fault Injection Record List
[**api_v1_injections_mapping_get**](InjectionApi.md#api_v1_injections_mapping_get) | **GET** /api/v1/injections/mapping | Get Field Mapping
[**api_v1_injections_ns_resources_get**](InjectionApi.md#api_v1_injections_ns_resources_get) | **GET** /api/v1/injections/ns-resources | Get Namespace Resource Mapping
[**api_v1_injections_post**](InjectionApi.md#api_v1_injections_post) | **POST** /api/v1/injections | Submit Fault Injection Task
[**api_v1_injections_query_get**](InjectionApi.md#api_v1_injections_query_get) | **GET** /api/v1/injections/query | Query Single Fault Injection Record
[**api_v1_injections_task_id_cancel_put**](InjectionApi.md#api_v1_injections_task_id_cancel_put) | **PUT** /api/v1/injections/{task_id}/cancel | Cancel Fault Injection Task


# **api_v1_injections_analysis_no_issues_get**
> DtoGenericResponseArrayDtoFaultInjectionNoIssuesResp api_v1_injections_analysis_no_issues_get(env=env, batch=batch, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

Query Fault Injection Records Without Issues

Query all fault injection records without issues based on time range, returning detailed records including configuration information

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_array_dto_fault_injection_no_issues_resp import DtoGenericResponseArrayDtoFaultInjectionNoIssuesResp
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    env = 'env_example' # str | Environment label filter (optional)
    batch = 'batch_example' # str | Batch label filter (optional)
    lookback = 'lookback_example' # str | Time range query, supports custom relative time (1h/24h/7d) or custom, default not set (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | Custom start time, RFC3339 format, required when lookback=custom (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | Custom end time, RFC3339 format, required when lookback=custom (optional)

    try:
        # Query Fault Injection Records Without Issues
        api_response = api_instance.api_v1_injections_analysis_no_issues_get(env=env, batch=batch, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of InjectionApi->api_v1_injections_analysis_no_issues_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_analysis_no_issues_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **env** | **str**| Environment label filter | [optional] 
 **batch** | **str**| Batch label filter | [optional] 
 **lookback** | **str**| Time range query, supports custom relative time (1h/24h/7d) or custom, default not set | [optional] 
 **custom_start_time** | **datetime**| Custom start time, RFC3339 format, required when lookback&#x3D;custom | [optional] 
 **custom_end_time** | **datetime**| Custom end time, RFC3339 format, required when lookback&#x3D;custom | [optional] 

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
**200** | Successfully returned fault injection records without issues |  -  |
**400** | Request parameter error, such as incorrect time format or parameter validation failure, etc. |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_analysis_stats_get**
> DtoGenericResponseDtoInjectionStatsResp api_v1_injections_analysis_stats_get(lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

Get Fault Injection Statistics

Get statistical information of fault injection records, including counts of records with issues, without issues, and total records

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_injection_stats_resp import DtoGenericResponseDtoInjectionStatsResp
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    lookback = 'lookback_example' # str | Time range query, supports custom relative time (1h/24h/7d) or custom, default not set (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | Custom start time, RFC3339 format, required when lookback=custom (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | Custom end time, RFC3339 format, required when lookback=custom (optional)

    try:
        # Get Fault Injection Statistics
        api_response = api_instance.api_v1_injections_analysis_stats_get(lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of InjectionApi->api_v1_injections_analysis_stats_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_analysis_stats_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **lookback** | **str**| Time range query, supports custom relative time (1h/24h/7d) or custom, default not set | [optional] 
 **custom_start_time** | **datetime**| Custom start time, RFC3339 format, required when lookback&#x3D;custom | [optional] 
 **custom_end_time** | **datetime**| Custom end time, RFC3339 format, required when lookback&#x3D;custom | [optional] 

### Return type

[**DtoGenericResponseDtoInjectionStatsResp**](DtoGenericResponseDtoInjectionStatsResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Successfully returned fault injection statistics |  -  |
**400** | Request parameter error, such as incorrect time format or parameter validation failure, etc. |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_analysis_with_issues_get**
> DtoGenericResponseArrayDtoFaultInjectionWithIssuesResp api_v1_injections_analysis_with_issues_get(env=env, batch=batch, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

Query Fault Injection Records With Issues

Query all fault injection records with issues based on time range

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_array_dto_fault_injection_with_issues_resp import DtoGenericResponseArrayDtoFaultInjectionWithIssuesResp
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    env = 'env_example' # str | Environment label filter (optional)
    batch = 'batch_example' # str | Batch label filter (optional)
    lookback = 'lookback_example' # str | Time range query, supports custom relative time (1h/24h/7d) or custom, default not set (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | Custom start time, RFC3339 format, required when lookback=custom (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | Custom end time, RFC3339 format, required when lookback=custom (optional)

    try:
        # Query Fault Injection Records With Issues
        api_response = api_instance.api_v1_injections_analysis_with_issues_get(env=env, batch=batch, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of InjectionApi->api_v1_injections_analysis_with_issues_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_analysis_with_issues_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **env** | **str**| Environment label filter | [optional] 
 **batch** | **str**| Batch label filter | [optional] 
 **lookback** | **str**| Time range query, supports custom relative time (1h/24h/7d) or custom, default not set | [optional] 
 **custom_start_time** | **datetime**| Custom start time, RFC3339 format, required when lookback&#x3D;custom | [optional] 
 **custom_end_time** | **datetime**| Custom end time, RFC3339 format, required when lookback&#x3D;custom | [optional] 

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
**400** | Request parameter error, such as incorrect time format or parameter validation failure, etc. |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_conf_get**
> DtoGenericResponseHandlerNode api_v1_injections_conf_get(namespace, mode=mode)

Get Fault Injection Configuration

Get fault injection configuration for the specified namespace, supporting different display modes for configuration tree structure

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_handler_node import DtoGenericResponseHandlerNode
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    namespace = 'namespace_example' # str | Namespace, specifies the namespace to get configuration for
    mode = engine # str | Display mode (optional) (default to engine)

    try:
        # Get Fault Injection Configuration
        api_response = api_instance.api_v1_injections_conf_get(namespace, mode=mode)
        print("The response of InjectionApi->api_v1_injections_conf_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_conf_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **namespace** | **str**| Namespace, specifies the namespace to get configuration for | 
 **mode** | **str**| Display mode | [optional] [default to engine]

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
**200** | Successfully returned configuration tree structure |  -  |
**400** | Request parameter error, such as missing namespace or mode parameter |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_configs_get**
> DtoGenericResponseAny api_v1_injections_configs_get(trace_ids=trace_ids)

Get Injected Fault Configuration List

Get fault injection configuration information based on multiple TraceIDs, used to view configuration details of submitted fault injection tasks

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    trace_ids = ['trace_ids_example'] # List[str] | TraceID list, supports multiple values, used to query corresponding configuration information (optional)

    try:
        # Get Injected Fault Configuration List
        api_response = api_instance.api_v1_injections_configs_get(trace_ids=trace_ids)
        print("The response of InjectionApi->api_v1_injections_configs_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_configs_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trace_ids** | [**List[str]**](str.md)| TraceID list, supports multiple values, used to query corresponding configuration information | [optional] 

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
**200** | Successfully returned configuration list |  -  |
**400** | Request parameter error, such as missing TraceID parameter or incorrect format |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_get**
> DtoGenericResponseDtoListInjectionsResp api_v1_injections_get(project_name=project_name, env=env, batch=batch, tag=tag, benchmark=benchmark, status=status, fault_type=fault_type, sort_field=sort_field, sort_order=sort_order, limit=limit, page_num=page_num, page_size=page_size, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

Get Fault Injection Record List

Fault injection record query interface supporting sorting and filtering. Returns the original database record list without data conversion.

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_list_injections_resp import DtoGenericResponseDtoListInjectionsResp
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    project_name = 'project_name_example' # str | Project name filter (optional)
    env = prod # str | Environment label filter (optional) (default to prod)
    batch = 'batch_example' # str | Batch label filter (optional)
    tag = train # str | Category label filter (optional) (default to train)
    benchmark = clickhouse # str | Benchmark type filter (optional) (default to clickhouse)
    status = 0 # int | Status filter, refer to field mapping interface (/mapping) for specific values (optional) (default to 0)
    fault_type = 0 # int | Fault type filter, refer to field mapping interface (/mapping) for specific values (optional) (default to 0)
    sort_field = 'created_at' # str | Sort field, default created_at (optional) (default to 'created_at')
    sort_order = desc # str | Sort order, default desc (optional) (default to desc)
    limit = 0 # int | Result quantity limit, used to control the number of returned records (optional) (default to 0)
    page_num = 0 # int | Pagination query, page number (optional) (default to 0)
    page_size = 0 # int | Pagination query, records per page (optional) (default to 0)
    lookback = 'lookback_example' # str | Time range query, supports custom relative time (1h/24h/7d) or custom, default not set (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | Custom start time, RFC3339 format, required when lookback=custom (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | Custom end time, RFC3339 format, required when lookback=custom (optional)

    try:
        # Get Fault Injection Record List
        api_response = api_instance.api_v1_injections_get(project_name=project_name, env=env, batch=batch, tag=tag, benchmark=benchmark, status=status, fault_type=fault_type, sort_field=sort_field, sort_order=sort_order, limit=limit, page_num=page_num, page_size=page_size, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of InjectionApi->api_v1_injections_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **project_name** | **str**| Project name filter | [optional] 
 **env** | **str**| Environment label filter | [optional] [default to prod]
 **batch** | **str**| Batch label filter | [optional] 
 **tag** | **str**| Category label filter | [optional] [default to train]
 **benchmark** | **str**| Benchmark type filter | [optional] [default to clickhouse]
 **status** | **int**| Status filter, refer to field mapping interface (/mapping) for specific values | [optional] [default to 0]
 **fault_type** | **int**| Fault type filter, refer to field mapping interface (/mapping) for specific values | [optional] [default to 0]
 **sort_field** | **str**| Sort field, default created_at | [optional] [default to &#39;created_at&#39;]
 **sort_order** | **str**| Sort order, default desc | [optional] [default to desc]
 **limit** | **int**| Result quantity limit, used to control the number of returned records | [optional] [default to 0]
 **page_num** | **int**| Pagination query, page number | [optional] [default to 0]
 **page_size** | **int**| Pagination query, records per page | [optional] [default to 0]
 **lookback** | **str**| Time range query, supports custom relative time (1h/24h/7d) or custom, default not set | [optional] 
 **custom_start_time** | **datetime**| Custom start time, RFC3339 format, required when lookback&#x3D;custom | [optional] 
 **custom_end_time** | **datetime**| Custom end time, RFC3339 format, required when lookback&#x3D;custom | [optional] 

### Return type

[**DtoGenericResponseDtoListInjectionsResp**](DtoGenericResponseDtoListInjectionsResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Successfully returned fault injection record list |  -  |
**400** | Request parameter error, such as incorrect parameter format, validation failure, etc. |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_mapping_get**
> DtoGenericResponseDtoInjectionFieldMappingResp api_v1_injections_mapping_get()

Get Field Mapping

Get string-to-number mapping relationships for status and fault types, used for frontend display and API parameter validation

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_injection_field_mapping_resp import DtoGenericResponseDtoInjectionFieldMappingResp
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
    api_instance = rcabench.openapi.InjectionApi(api_client)

    try:
        # Get Field Mapping
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
**200** | Successfully returned field mapping relationships |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_ns_resources_get**
> DtoGenericResponseHandlerResources api_v1_injections_ns_resources_get(namespace=namespace)

Get Namespace Resource Mapping

Get mapping of all namespaces and their corresponding resource information, or query resource information for a specific namespace. Returns a mapping table from namespace to resources, used for fault injection configuration and resource management

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_handler_resources import DtoGenericResponseHandlerResources
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    namespace = 'namespace_example' # str | Namespace name, returns resource mappings for all namespaces if not specified (optional)

    try:
        # Get Namespace Resource Mapping
        api_response = api_instance.api_v1_injections_ns_resources_get(namespace=namespace)
        print("The response of InjectionApi->api_v1_injections_ns_resources_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_ns_resources_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **namespace** | **str**| Namespace name, returns resource mappings for all namespaces if not specified | [optional] 

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
**200** | Returns resource information for the specified namespace when a namespace is provided |  -  |
**404** | The specified namespace does not exist |  -  |
**500** | Internal server error, unable to get resource mapping |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_post**
> DtoGenericResponseDtoSubmitInjectionResp api_v1_injections_post(body)

Submit Fault Injection Task

Submit fault injection task, supporting batch submission of multiple fault configurations, the system will automatically deduplicate and return submission results

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_submit_injection_resp import DtoGenericResponseDtoSubmitInjectionResp
from rcabench.openapi.models.dto_submit_injection_req import DtoSubmitInjectionReq
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    body = rcabench.openapi.DtoSubmitInjectionReq() # DtoSubmitInjectionReq | Fault injection request body

    try:
        # Submit Fault Injection Task
        api_response = api_instance.api_v1_injections_post(body)
        print("The response of InjectionApi->api_v1_injections_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**DtoSubmitInjectionReq**](DtoSubmitInjectionReq.md)| Fault injection request body | 

### Return type

[**DtoGenericResponseDtoSubmitInjectionResp**](DtoGenericResponseDtoSubmitInjectionResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**202** | Successfully submitted fault injection task |  -  |
**400** | Request parameter error, such as incorrect JSON format, parameter validation failure, or invalid algorithm, etc. |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_query_get**
> DtoGenericResponseDtoQueryInjectionResp api_v1_injections_query_get(name=name, task_id=task_id)

Query Single Fault Injection Record

Query fault injection record details by name or task ID, at least one of the two parameters must be provided

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_query_injection_resp import DtoGenericResponseDtoQueryInjectionResp
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    name = 'name_example' # str | Fault injection name (optional)
    task_id = 'task_id_example' # str | Task ID (optional)

    try:
        # Query Single Fault Injection Record
        api_response = api_instance.api_v1_injections_query_get(name=name, task_id=task_id)
        print("The response of InjectionApi->api_v1_injections_query_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_query_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **str**| Fault injection name | [optional] 
 **task_id** | **str**| Task ID | [optional] 

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
**200** | Successfully returned fault injection record details |  -  |
**400** | Request parameter error, such as missing parameters, incorrect format, or validation failure, etc. |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_injections_task_id_cancel_put**
> DtoGenericResponseDtoInjectCancelResp api_v1_injections_task_id_cancel_put(task_id)

Cancel Fault Injection Task

Cancel the specified fault injection task

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_inject_cancel_resp import DtoGenericResponseDtoInjectCancelResp
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
    api_instance = rcabench.openapi.InjectionApi(api_client)
    task_id = 'task_id_example' # str | Task ID

    try:
        # Cancel Fault Injection Task
        api_response = api_instance.api_v1_injections_task_id_cancel_put(task_id)
        print("The response of InjectionApi->api_v1_injections_task_id_cancel_put:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionApi->api_v1_injections_task_id_cancel_put: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **task_id** | **str**| Task ID | 

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

