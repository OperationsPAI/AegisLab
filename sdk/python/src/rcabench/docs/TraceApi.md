# rcabench.openapi.TraceApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_analyzers_traces_get**](TraceApi.md#api_v1_analyzers_traces_get) | **GET** /api/v1/analyzers/traces | Analyze trace data


# **api_v1_analyzers_traces_get**
> DtoGenericResponseDtoTraceStats api_v1_analyzers_traces_get(first_task_type=first_task_type, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

Analyze trace data

Analyze trace data using various filtering conditions, returning statistical information including traces ending with fault injection

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_trace_stats import DtoGenericResponseDtoTraceStats
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
    first_task_type = 'first_task_type_example' # str | First task type filter (optional)
    lookback = 'lookback_example' # str | Time range query, supports custom relative time (1h/24h/7d) or custom, default is not set (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | Custom start time, RFC3339 format, required when lookback=custom (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | Custom end time, RFC3339 format, required when lookback=custom (optional)

    try:
        # Analyze trace data
        api_response = api_instance.api_v1_analyzers_traces_get(first_task_type=first_task_type, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of TraceApi->api_v1_analyzers_traces_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TraceApi->api_v1_analyzers_traces_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **first_task_type** | **str**| First task type filter | [optional] 
 **lookback** | **str**| Time range query, supports custom relative time (1h/24h/7d) or custom, default is not set | [optional] 
 **custom_start_time** | **datetime**| Custom start time, RFC3339 format, required when lookback&#x3D;custom | [optional] 
 **custom_end_time** | **datetime**| Custom end time, RFC3339 format, required when lookback&#x3D;custom | [optional] 

### Return type

[**DtoGenericResponseDtoTraceStats**](DtoGenericResponseDtoTraceStats.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Returns trace analysis statistics |  -  |
**400** | Request parameter error, such as incorrect parameter format, validation failure, etc. |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

