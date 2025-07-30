# rcabench.openapi.AlgorithmApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_algorithms_get**](AlgorithmApi.md#api_v1_algorithms_get) | **GET** /api/v1/algorithms | Get algorithm list
[**api_v1_algorithms_post**](AlgorithmApi.md#api_v1_algorithms_post) | **POST** /api/v1/algorithms | Submit algorithm execution task


# **api_v1_algorithms_get**
> DtoGenericResponseDtoListAlgorithmsResp api_v1_algorithms_get()

Get algorithm list

Get all available algorithms in the system, including image info, tags, and update time. Only returns containers with active status.

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_list_algorithms_resp import DtoGenericResponseDtoListAlgorithmsResp
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
    api_instance = rcabench.openapi.AlgorithmApi(api_client)

    try:
        # Get algorithm list
        api_response = api_instance.api_v1_algorithms_get()
        print("The response of AlgorithmApi->api_v1_algorithms_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmApi->api_v1_algorithms_get: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**DtoGenericResponseDtoListAlgorithmsResp**](DtoGenericResponseDtoListAlgorithmsResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Successfully returned algorithm list |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_algorithms_post**
> DtoGenericResponseDtoSubmitResp api_v1_algorithms_post(body)

Submit algorithm execution task

Batch submit algorithm execution tasks, supporting multiple algorithm and dataset combinations. The system assigns a unique TraceID for each execution task to track status and results.

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_submit_resp import DtoGenericResponseDtoSubmitResp
from rcabench.openapi.models.dto_submit_execution_req import DtoSubmitExecutionReq
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
    api_instance = rcabench.openapi.AlgorithmApi(api_client)
    body = rcabench.openapi.DtoSubmitExecutionReq() # DtoSubmitExecutionReq | Algorithm execution request list, including algorithm name, dataset, and environment variables

    try:
        # Submit algorithm execution task
        api_response = api_instance.api_v1_algorithms_post(body)
        print("The response of AlgorithmApi->api_v1_algorithms_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmApi->api_v1_algorithms_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**DtoSubmitExecutionReq**](DtoSubmitExecutionReq.md)| Algorithm execution request list, including algorithm name, dataset, and environment variables | 

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
**202** | Successfully submitted algorithm execution task, returns task tracking info |  -  |
**400** | Request parameter error, such as invalid JSON format, algorithm name or dataset name, unsupported environment variable name, etc. |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

