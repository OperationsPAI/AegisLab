# rcabench.openapi.TracesApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**stream_trace_events**](TracesApi.md#stream_trace_events) | **GET** /api/v2/traces/{trace_id}/stream | Stream trace events in real-time


# **stream_trace_events**
> stream_trace_events(trace_id, last_id=last_id)

Stream trace events in real-time

Establishes a Server-Sent Events (SSE) connection to stream trace logs and task execution events in real-time. Returns historical events first, then switches to live monitoring.

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
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
    api_instance = rcabench.openapi.TracesApi(api_client)
    trace_id = 'trace_id_example' # str | Trace ID
    last_id = '"0"' # str | Last event ID received (optional) (default to '"0"')

    try:
        # Stream trace events in real-time
        api_instance.stream_trace_events(trace_id, last_id=last_id)
    except Exception as e:
        print("Exception when calling TracesApi->stream_trace_events: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trace_id** | **str**| Trace ID | 
 **last_id** | **str**| Last event ID received | [optional] [default to &#39;&quot;0&quot;&#39;]

### Return type

void (empty response body)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/event-stream

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**400** | Invalid trace ID or invalid request format/parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

