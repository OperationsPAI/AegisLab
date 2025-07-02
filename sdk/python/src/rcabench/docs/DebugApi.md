# rcabench.openapi.DebugApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_debug_ns_status_get**](DebugApi.md#api_v1_debug_ns_status_get) | **GET** /api/v1/debug/ns/status | 获取命名空间锁状态


# **api_v1_debug_ns_status_get**
> DtoGenericResponseAny api_v1_debug_ns_status_get()

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
    api_instance = rcabench.openapi.DebugApi(api_client)

    try:
        # 获取命名空间锁状态
        api_response = api_instance.api_v1_debug_ns_status_get()
        print("The response of DebugApi->api_v1_debug_ns_status_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DebugApi->api_v1_debug_ns_status_get: %s\n" % e)
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

