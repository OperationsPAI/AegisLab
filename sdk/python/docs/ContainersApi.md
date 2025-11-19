# openapi.ContainersApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**build_container_image**](ContainersApi.md#build_container_image) | **POST** /api/v2/containers/build | Submit container building
[**create_container**](ContainersApi.md#create_container) | **POST** /api/v2/containers | Create container
[**create_container_version**](ContainersApi.md#create_container_version) | **POST** /api/v2/containers/{container_id}/versions | Create container version
[**get_container_by_id**](ContainersApi.md#get_container_by_id) | **GET** /api/v2/containers/{container_id} | Get container by ID
[**get_container_version_by_id**](ContainersApi.md#get_container_version_by_id) | **GET** /api/v2/containers/{container_id}/versions/{version_id} | Get container version by ID
[**list_container_versions**](ContainersApi.md#list_container_versions) | **GET** /api/v2/containers/{container_id}/versions | List container versions
[**list_containers**](ContainersApi.md#list_containers) | **GET** /api/v2/containers | List containers
[**search_containers**](ContainersApi.md#search_containers) | **POST** /api/v2/containers/search | Search containers


# **build_container_image**
> GenericResponseSubmitContainerBuildResp build_container_image(request)

Submit container building

Submit a container build task to build a container image from provided source files.

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.generic_response_submit_container_build_resp import GenericResponseSubmitContainerBuildResp
from openapi.models.submit_build_container_req import SubmitBuildContainerReq
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
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
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.ContainersApi(api_client)
    request = openapi.SubmitBuildContainerReq() # SubmitBuildContainerReq | Container build request

    try:
        # Submit container building
        api_response = api_instance.build_container_image(request)
        print("The response of ContainersApi->build_container_image:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->build_container_image: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**SubmitBuildContainerReq**](SubmitBuildContainerReq.md)| Container build request | 

### Return type

[**GenericResponseSubmitContainerBuildResp**](GenericResponseSubmitContainerBuildResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Container build task submitted successfully |  -  |
**400** | Invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Required files not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **create_container**
> GenericResponseContainerResp create_container(request)

Create container

Create a new container without build configuration.

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.create_container_req import CreateContainerReq
from openapi.models.generic_response_container_resp import GenericResponseContainerResp
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
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
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.ContainersApi(api_client)
    request = openapi.CreateContainerReq() # CreateContainerReq | Container creation request

    try:
        # Create container
        api_response = api_instance.create_container(request)
        print("The response of ContainersApi->create_container:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->create_container: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**CreateContainerReq**](CreateContainerReq.md)| Container creation request | 

### Return type

[**GenericResponseContainerResp**](GenericResponseContainerResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | Container created successfully |  -  |
**400** | Invalid request |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**409** | Conflict error |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **create_container_version**
> GenericResponseContainerVersionResp create_container_version(container_id, request)

Create container version

Create a new container version for an existing container.

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.create_container_version_req import CreateContainerVersionReq
from openapi.models.generic_response_container_version_resp import GenericResponseContainerVersionResp
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
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
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.ContainersApi(api_client)
    container_id = 56 # int | Container ID
    request = openapi.CreateContainerVersionReq() # CreateContainerVersionReq | Container version creation request

    try:
        # Create container version
        api_response = api_instance.create_container_version(container_id, request)
        print("The response of ContainersApi->create_container_version:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->create_container_version: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **container_id** | **int**| Container ID | 
 **request** | [**CreateContainerVersionReq**](CreateContainerVersionReq.md)| Container version creation request | 

### Return type

[**GenericResponseContainerVersionResp**](GenericResponseContainerVersionResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | Container version created successfully |  -  |
**400** | Invalid container ID or invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**409** | Conflict error |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_container_by_id**
> GenericResponseContainerDetailResp get_container_by_id(container_id)

Get container by ID

Get detailed information about a specific container

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.generic_response_container_detail_resp import GenericResponseContainerDetailResp
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
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
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.ContainersApi(api_client)
    container_id = 56 # int | Container ID

    try:
        # Get container by ID
        api_response = api_instance.get_container_by_id(container_id)
        print("The response of ContainersApi->get_container_by_id:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->get_container_by_id: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **container_id** | **int**| Container ID | 

### Return type

[**GenericResponseContainerDetailResp**](GenericResponseContainerDetailResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Container retrieved successfully |  -  |
**400** | Invalid container ID |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Container not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_container_version_by_id**
> GenericResponseContainerVersionDetailResp get_container_version_by_id(container_id, version_id)

Get container version by ID

Get detailed information about a specific container version

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.generic_response_container_version_detail_resp import GenericResponseContainerVersionDetailResp
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
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
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.ContainersApi(api_client)
    container_id = 56 # int | Container ID
    version_id = 56 # int | Container Version ID

    try:
        # Get container version by ID
        api_response = api_instance.get_container_version_by_id(container_id, version_id)
        print("The response of ContainersApi->get_container_version_by_id:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->get_container_version_by_id: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **container_id** | **int**| Container ID | 
 **version_id** | **int**| Container Version ID | 

### Return type

[**GenericResponseContainerVersionDetailResp**](GenericResponseContainerVersionDetailResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Container version retrieved successfully |  -  |
**400** | Invalid container ID/container version ID |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Container or version not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_container_versions**
> GenericResponseListContainerVersionResp list_container_versions(container_id, page=page, size=size, status=status)

List container versions

Get paginated list of container versions for a specific container

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.generic_response_list_container_version_resp import GenericResponseListContainerVersionResp
from openapi.models.status_type import StatusType
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
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
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.ContainersApi(api_client)
    container_id = 56 # int | Container ID
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)
    status = openapi.StatusType() # StatusType | Container version status filter (optional)

    try:
        # List container versions
        api_response = api_instance.list_container_versions(container_id, page=page, size=size, status=status)
        print("The response of ContainersApi->list_container_versions:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->list_container_versions: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **container_id** | **int**| Container ID | 
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]
 **status** | [**StatusType**](.md)| Container version status filter | [optional] 

### Return type

[**GenericResponseListContainerVersionResp**](GenericResponseListContainerVersionResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Container versions retrieved successfully |  -  |
**400** | Invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_containers**
> GenericResponseListContainerResp list_containers(page=page, size=size, type=type, is_public=is_public, status=status)

List containers

Get paginated list of containers with pagination and filtering

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.container_type import ContainerType
from openapi.models.generic_response_list_container_resp import GenericResponseListContainerResp
from openapi.models.page_size import PageSize
from openapi.models.status_type import StatusType
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
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
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.ContainersApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = openapi.PageSize() # PageSize | Page size (optional)
    type = openapi.ContainerType() # ContainerType | Container type filter (optional)
    is_public = True # bool | Container public visibility filter (optional)
    status = openapi.StatusType() # StatusType | Container status filter (optional)

    try:
        # List containers
        api_response = api_instance.list_containers(page=page, size=size, type=type, is_public=is_public, status=status)
        print("The response of ContainersApi->list_containers:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->list_containers: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | [**PageSize**](.md)| Page size | [optional] 
 **type** | [**ContainerType**](.md)| Container type filter | [optional] 
 **is_public** | **bool**| Container public visibility filter | [optional] 
 **status** | [**StatusType**](.md)| Container status filter | [optional] 

### Return type

[**GenericResponseListContainerResp**](GenericResponseListContainerResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Containers retrieved successfully |  -  |
**400** | Invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **search_containers**
> GenericResponseSearchRespContainerResp search_containers(request)

Search containers

Search containers with complex filtering, sorting and pagination. Supports all container types (algorithm, benchmark, etc.)

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.generic_response_search_resp_container_resp import GenericResponseSearchRespContainerResp
from openapi.models.search_container_req import SearchContainerReq
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
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
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.ContainersApi(api_client)
    request = openapi.SearchContainerReq() # SearchContainerReq | Container search request

    try:
        # Search containers
        api_response = api_instance.search_containers(request)
        print("The response of ContainersApi->search_containers:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->search_containers: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**SearchContainerReq**](SearchContainerReq.md)| Container search request | 

### Return type

[**GenericResponseSearchRespContainerResp**](GenericResponseSearchRespContainerResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

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

