# rcabench.openapi.SystemApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**system_health_get**](SystemApi.md#system_health_get) | **GET** /system/health | System health check


# **system_health_get**
> GenericResponseHealthCheckResp system_health_get()

System health check

Get system health status and service information

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.generic_response_health_check_resp import GenericResponseHealthCheckResp
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
    api_instance = rcabench.openapi.SystemApi(api_client)

    try:
        # System health check
        api_response = api_instance.system_health_get()
        print("The response of SystemApi->system_health_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling SystemApi->system_health_get: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**GenericResponseHealthCheckResp**](GenericResponseHealthCheckResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Health check successful |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

