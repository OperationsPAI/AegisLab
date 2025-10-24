# rcabench.openapi.AnalyzerApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_analyzers_injections_get**](AnalyzerApi.md#api_v1_analyzers_injections_get) | **GET** /api/v1/analyzers/injections | Analyze fault injection data


# **api_v1_analyzers_injections_get**
> DtoGenericResponseDtoAnalyzeInjectionsResp api_v1_analyzers_injections_get(project_name=project_name, env=env, batch=batch, tag=tag, benchmark=benchmark, status=status, fault_type=fault_type, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

Analyze fault injection data

Analyze fault injection data using various filtering conditions, returning statistical information including efficiency, diversity, distance between seeds, etc.

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_analyze_injections_resp import DtoGenericResponseDtoAnalyzeInjectionsResp
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
    api_instance = rcabench.openapi.AnalyzerApi(api_client)
    project_name = 'project_name_example' # str | Project name filter (optional)
    env = prod # str | Environment label filter (optional) (default to prod)
    batch = 'batch_example' # str | Batch label filter (optional)
    tag = train # str | Classification label filter (optional) (default to train)
    benchmark = clickhouse # str | Benchmark type filter (optional) (default to clickhouse)
    status = 0 # int | Status filter, refer to field mapping interface (/mapping) for specific values (optional) (default to 0)
    fault_type = 0 # int | Fault type filter, refer to field mapping interface (/mapping) for specific values (optional) (default to 0)
    lookback = 'lookback_example' # str | Time range query, supports custom relative time (1h/24h/7d) or custom, default is not set (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | Custom start time, RFC3339 format, required when lookback=custom (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | Custom end time, RFC3339 format, required when lookback=custom (optional)

    try:
        # Analyze fault injection data
        api_response = api_instance.api_v1_analyzers_injections_get(project_name=project_name, env=env, batch=batch, tag=tag, benchmark=benchmark, status=status, fault_type=fault_type, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of AnalyzerApi->api_v1_analyzers_injections_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AnalyzerApi->api_v1_analyzers_injections_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **project_name** | **str**| Project name filter | [optional] 
 **env** | **str**| Environment label filter | [optional] [default to prod]
 **batch** | **str**| Batch label filter | [optional] 
 **tag** | **str**| Classification label filter | [optional] [default to train]
 **benchmark** | **str**| Benchmark type filter | [optional] [default to clickhouse]
 **status** | **int**| Status filter, refer to field mapping interface (/mapping) for specific values | [optional] [default to 0]
 **fault_type** | **int**| Fault type filter, refer to field mapping interface (/mapping) for specific values | [optional] [default to 0]
 **lookback** | **str**| Time range query, supports custom relative time (1h/24h/7d) or custom, default is not set | [optional] 
 **custom_start_time** | **datetime**| Custom start time, RFC3339 format, required when lookback&#x3D;custom | [optional] 
 **custom_end_time** | **datetime**| Custom end time, RFC3339 format, required when lookback&#x3D;custom | [optional] 

### Return type

[**DtoGenericResponseDtoAnalyzeInjectionsResp**](DtoGenericResponseDtoAnalyzeInjectionsResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Returns fault injection analysis statistics |  -  |
**400** | Request parameter error, such as incorrect parameter format, validation failure, etc. |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

