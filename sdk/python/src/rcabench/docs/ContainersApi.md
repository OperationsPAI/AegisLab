# rcabench.openapi.ContainersApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_containers_get**](ContainersApi.md#api_v2_containers_get) | **GET** /api/v2/containers | List containers
[**api_v2_containers_id_get**](ContainersApi.md#api_v2_containers_id_get) | **GET** /api/v2/containers/{id} | Get container by ID
[**api_v2_containers_search_post**](ContainersApi.md#api_v2_containers_search_post) | **POST** /api/v2/containers/search | Search containers


# **api_v2_containers_get**
> DtoGenericResponseDtoSearchResponseDtoContainerResponse api_v2_containers_get(page=page, size=size, type=type, status=status)

List containers

Get a simple list of containers with basic filtering

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_container_response import DtoGenericResponseDtoSearchResponseDtoContainerResponse
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
    api_instance = rcabench.openapi.ContainersApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)
    type = 'type_example' # str | Container type filter (optional)
    status = True # bool | Container status filter (optional)

    try:
        # List containers
        api_response = api_instance.api_v2_containers_get(page=page, size=size, type=type, status=status)
        print("The response of ContainersApi->api_v2_containers_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->api_v2_containers_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]
 **type** | **str**| Container type filter | [optional] 
 **status** | **bool**| Container status filter | [optional] 

### Return type

[**DtoGenericResponseDtoSearchResponseDtoContainerResponse**](DtoGenericResponseDtoSearchResponseDtoContainerResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Containers retrieved successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_containers_id_get**
> DtoGenericResponseDtoContainerResponse api_v2_containers_id_get(id)

Get container by ID

Get detailed information about a specific container

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_container_response import DtoGenericResponseDtoContainerResponse
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
    api_instance = rcabench.openapi.ContainersApi(api_client)
    id = 56 # int | Container ID

    try:
        # Get container by ID
        api_response = api_instance.api_v2_containers_id_get(id)
        print("The response of ContainersApi->api_v2_containers_id_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->api_v2_containers_id_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Container ID | 

### Return type

[**DtoGenericResponseDtoContainerResponse**](DtoGenericResponseDtoContainerResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Container retrieved successfully |  -  |
**400** | Invalid container ID |  -  |
**403** | Permission denied |  -  |
**404** | Container not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_containers_search_post**
> DtoGenericResponseDtoSearchResponseDtoContainerResponse api_v2_containers_search_post(request)

Search containers

Search containers with complex filtering, sorting and pagination. Supports all container types (algorithm, benchmark, etc.)

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_container_search_request import DtoContainerSearchRequest
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_container_response import DtoGenericResponseDtoSearchResponseDtoContainerResponse
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
    api_instance = rcabench.openapi.ContainersApi(api_client)
    request = rcabench.openapi.DtoContainerSearchRequest() # DtoContainerSearchRequest | Container search request

    try:
        # Search containers
        api_response = api_instance.api_v2_containers_search_post(request)
        print("The response of ContainersApi->api_v2_containers_search_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->api_v2_containers_search_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoContainerSearchRequest**](DtoContainerSearchRequest.md)| Container search request | 

### Return type

[**DtoGenericResponseDtoSearchResponseDtoContainerResponse**](DtoGenericResponseDtoSearchResponseDtoContainerResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Containers retrieved successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

