# rcabench.openapi.AlgorithmsApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_algorithms_get**](AlgorithmsApi.md#api_v2_algorithms_get) | **GET** /api/v2/algorithms | List algorithms
[**api_v2_algorithms_search_post**](AlgorithmsApi.md#api_v2_algorithms_search_post) | **POST** /api/v2/algorithms/search | Search algorithms


# **api_v2_algorithms_get**
> DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse api_v2_algorithms_get(page=page, size=size)

List algorithms

Get a simple list of all active algorithms without complex filtering

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_algorithm_response import DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse
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
    api_instance = rcabench.openapi.AlgorithmsApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)

    try:
        # List algorithms
        api_response = api_instance.api_v2_algorithms_get(page=page, size=size)
        print("The response of AlgorithmsApi->api_v2_algorithms_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmsApi->api_v2_algorithms_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]

### Return type

[**DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse**](DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Algorithms retrieved successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_algorithms_search_post**
> DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse api_v2_algorithms_search_post(request)

Search algorithms

Search algorithms with complex filtering, sorting and pagination. Algorithms are containers with type 'algorithm'

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_algorithm_search_request import DtoAlgorithmSearchRequest
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_algorithm_response import DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse
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
    api_instance = rcabench.openapi.AlgorithmsApi(api_client)
    request = rcabench.openapi.DtoAlgorithmSearchRequest() # DtoAlgorithmSearchRequest | Algorithm search request

    try:
        # Search algorithms
        api_response = api_instance.api_v2_algorithms_search_post(request)
        print("The response of AlgorithmsApi->api_v2_algorithms_search_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmsApi->api_v2_algorithms_search_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoAlgorithmSearchRequest**](DtoAlgorithmSearchRequest.md)| Algorithm search request | 

### Return type

[**DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse**](DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Algorithms retrieved successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

