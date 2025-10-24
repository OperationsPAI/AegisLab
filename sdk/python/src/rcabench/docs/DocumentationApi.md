# rcabench.openapi.DocumentationApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_docs_models_get**](DocumentationApi.md#api_docs_models_get) | **GET** /api/_docs/models | API Model Definitions


# **api_docs_models_get**
> ConstsSSEEventName api_docs_models_get()

API Model Definitions

Virtual endpoint for including all DTO type definitions in Swagger documentation. DO NOT USE in production.

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.consts_sse_event_name import ConstsSSEEventName
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
    api_instance = rcabench.openapi.DocumentationApi(api_client)

    try:
        # API Model Definitions
        api_response = api_instance.api_docs_models_get()
        print("The response of DocumentationApi->api_docs_models_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DocumentationApi->api_docs_models_get: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**ConstsSSEEventName**](ConstsSSEEventName.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | SSE event name constants |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

