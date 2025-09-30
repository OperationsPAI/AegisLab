# rcabench.openapi.ContainersApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_containers_get**](ContainersApi.md#api_v2_containers_get) | **GET** /api/v2/containers | List containers
[**api_v2_containers_id_delete**](ContainersApi.md#api_v2_containers_id_delete) | **DELETE** /api/v2/containers/{id} | Delete container
[**api_v2_containers_id_get**](ContainersApi.md#api_v2_containers_id_get) | **GET** /api/v2/containers/{id} | Get container by ID
[**api_v2_containers_id_put**](ContainersApi.md#api_v2_containers_id_put) | **PUT** /api/v2/containers/{id} | Update container
[**api_v2_containers_name_name_latest_get**](ContainersApi.md#api_v2_containers_name_name_latest_get) | **GET** /api/v2/containers/name/{name}/latest | Get latest container by name
[**api_v2_containers_post**](ContainersApi.md#api_v2_containers_post) | **POST** /api/v2/containers | Create or update container
[**api_v2_containers_search_post**](ContainersApi.md#api_v2_containers_search_post) | **POST** /api/v2/containers/search | Search containers


# **api_v2_containers_get**
> DtoGenericResponseDtoSearchResponseDtoContainerResponse api_v2_containers_get(page=page, size=size, type=type, status=status)

List containers

Get a simple list of containers with basic filtering

### Example

* Api Key Authentication (BearerAuth):

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

# **api_v2_containers_id_delete**
> DtoGenericResponseAny api_v2_containers_id_delete(id)

Delete container

Delete a container (soft delete by setting status to false)

### Example

* Api Key Authentication (BearerAuth):

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
    api_instance = rcabench.openapi.ContainersApi(api_client)
    id = 56 # int | Container ID

    try:
        # Delete container
        api_response = api_instance.api_v2_containers_id_delete(id)
        print("The response of ContainersApi->api_v2_containers_id_delete:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->api_v2_containers_id_delete: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Container ID | 

### Return type

[**DtoGenericResponseAny**](DtoGenericResponseAny.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Container deleted successfully |  -  |
**400** | Invalid container ID |  -  |
**403** | Permission denied |  -  |
**404** | Container not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_containers_id_get**
> DtoGenericResponseDtoContainerResponse api_v2_containers_id_get(id)

Get container by ID

Get detailed information about a specific container

### Example

* Api Key Authentication (BearerAuth):

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

[BearerAuth](../README.md#BearerAuth)

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

# **api_v2_containers_id_put**
> DtoGenericResponseDtoContainerResponse api_v2_containers_id_put(id, request)

Update container

Update an existing container's information

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_container_response import DtoGenericResponseDtoContainerResponse
from rcabench.openapi.models.dto_update_container_request import DtoUpdateContainerRequest
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8080/api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8080/api/v1"
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
    api_instance = rcabench.openapi.ContainersApi(api_client)
    id = 56 # int | Container ID
    request = rcabench.openapi.DtoUpdateContainerRequest() # DtoUpdateContainerRequest | Container update request

    try:
        # Update container
        api_response = api_instance.api_v2_containers_id_put(id, request)
        print("The response of ContainersApi->api_v2_containers_id_put:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->api_v2_containers_id_put: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Container ID | 
 **request** | [**DtoUpdateContainerRequest**](DtoUpdateContainerRequest.md)| Container update request | 

### Return type

[**DtoGenericResponseDtoContainerResponse**](DtoGenericResponseDtoContainerResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Container updated successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**404** | Container not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_containers_name_name_latest_get**
> DtoGenericResponseDtoContainerResponse api_v2_containers_name_name_latest_get(name)

Get latest container by name

Get the most recently updated container record for a given algorithm name

### Example

* Api Key Authentication (BearerAuth):

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
    api_instance = rcabench.openapi.ContainersApi(api_client)
    name = 'name_example' # str | Container/Algorithm name

    try:
        # Get latest container by name
        api_response = api_instance.api_v2_containers_name_name_latest_get(name)
        print("The response of ContainersApi->api_v2_containers_name_name_latest_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->api_v2_containers_name_name_latest_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **str**| Container/Algorithm name | 

### Return type

[**DtoGenericResponseDtoContainerResponse**](DtoGenericResponseDtoContainerResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Latest container retrieved successfully |  -  |
**400** | Invalid container name |  -  |
**403** | Permission denied |  -  |
**404** | Container not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_containers_post**
> DtoGenericResponseDtoSubmitResp api_v2_containers_post(type, name, image, tag=tag, command=command, env_vars=env_vars, is_public=is_public, build_source_type=build_source_type, file=file, github_repository=github_repository, github_branch=github_branch, github_commit=github_commit, github_path=github_path, github_token=github_token, harbor_image=harbor_image, harbor_tag=harbor_tag, context_dir=context_dir, dockerfile_path=dockerfile_path)

Create or update container

Create a new container with build configuration or update existing one if it already exists. Containers are associated with the authenticated user. If a container with the same name, type, image, and tag already exists, it will be updated instead of creating a new one.

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_submit_resp import DtoGenericResponseDtoSubmitResp
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8080/api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8080/api/v1"
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
    api_instance = rcabench.openapi.ContainersApi(api_client)
    type = algorithm # str | Container type (default to algorithm)
    name = 'name_example' # str | Container name
    image = 'image_example' # str | Docker image name
    tag = 'latest' # str | Docker image tag (optional) (default to 'latest')
    command = '/bin/bash' # str | Container startup command (optional) (default to '/bin/bash')
    env_vars = ['env_vars_example'] # List[str] | Environment variables (can be specified multiple times) (optional)
    is_public = False # bool | Whether the container is public (optional) (default to False)
    build_source_type = file # str | Build source type (optional) (default to file)
    file = None # bytearray | Source code file (zip or tar.gz format, max 5MB) - required when build_source_type=file (optional)
    github_repository = 'github_repository_example' # str | GitHub repository (owner/repo) - required when build_source_type=github (optional)
    github_branch = 'main' # str | GitHub branch name (optional) (default to 'main')
    github_commit = 'github_commit_example' # str | GitHub commit hash (if specified, branch is ignored) (optional)
    github_path = '.' # str | Path within repository (optional) (default to '.')
    github_token = 'github_token_example' # str | GitHub access token for private repositories (optional)
    harbor_image = 'harbor_image_example' # str | Harbor image name - required when build_source_type=harbor (optional)
    harbor_tag = 'harbor_tag_example' # str | Harbor image tag - required when build_source_type=harbor (optional)
    context_dir = '.' # str | Docker build context directory (optional) (default to '.')
    dockerfile_path = 'Dockerfile' # str | Dockerfile path relative to source root (optional) (default to 'Dockerfile')

    try:
        # Create or update container
        api_response = api_instance.api_v2_containers_post(type, name, image, tag=tag, command=command, env_vars=env_vars, is_public=is_public, build_source_type=build_source_type, file=file, github_repository=github_repository, github_branch=github_branch, github_commit=github_commit, github_path=github_path, github_token=github_token, harbor_image=harbor_image, harbor_tag=harbor_tag, context_dir=context_dir, dockerfile_path=dockerfile_path)
        print("The response of ContainersApi->api_v2_containers_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainersApi->api_v2_containers_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **type** | **str**| Container type | [default to algorithm]
 **name** | **str**| Container name | 
 **image** | **str**| Docker image name | 
 **tag** | **str**| Docker image tag | [optional] [default to &#39;latest&#39;]
 **command** | **str**| Container startup command | [optional] [default to &#39;/bin/bash&#39;]
 **env_vars** | [**List[str]**](str.md)| Environment variables (can be specified multiple times) | [optional] 
 **is_public** | **bool**| Whether the container is public | [optional] [default to False]
 **build_source_type** | **str**| Build source type | [optional] [default to file]
 **file** | **bytearray**| Source code file (zip or tar.gz format, max 5MB) - required when build_source_type&#x3D;file | [optional] 
 **github_repository** | **str**| GitHub repository (owner/repo) - required when build_source_type&#x3D;github | [optional] 
 **github_branch** | **str**| GitHub branch name | [optional] [default to &#39;main&#39;]
 **github_commit** | **str**| GitHub commit hash (if specified, branch is ignored) | [optional] 
 **github_path** | **str**| Path within repository | [optional] [default to &#39;.&#39;]
 **github_token** | **str**| GitHub access token for private repositories | [optional] 
 **harbor_image** | **str**| Harbor image name - required when build_source_type&#x3D;harbor | [optional] 
 **harbor_tag** | **str**| Harbor image tag - required when build_source_type&#x3D;harbor | [optional] 
 **context_dir** | **str**| Docker build context directory | [optional] [default to &#39;.&#39;]
 **dockerfile_path** | **str**| Dockerfile path relative to source root | [optional] [default to &#39;Dockerfile&#39;]

### Return type

[**DtoGenericResponseDtoSubmitResp**](DtoGenericResponseDtoSubmitResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Container information updated successfully from Harbor |  -  |
**202** | Container creation/update task submitted successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_containers_search_post**
> DtoGenericResponseDtoSearchResponseDtoContainerResponse api_v2_containers_search_post(request)

Search containers

Search containers with complex filtering, sorting and pagination. Supports all container types (algorithm, benchmark, etc.)

### Example

* Api Key Authentication (BearerAuth):

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

