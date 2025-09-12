# rcabench.openapi.PermissionsApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_permissions_get**](PermissionsApi.md#api_v2_permissions_get) | **GET** /api/v2/permissions | List permissions
[**api_v2_permissions_id_delete**](PermissionsApi.md#api_v2_permissions_id_delete) | **DELETE** /api/v2/permissions/{id} | Delete permission
[**api_v2_permissions_id_get**](PermissionsApi.md#api_v2_permissions_id_get) | **GET** /api/v2/permissions/{id} | Get permission by ID
[**api_v2_permissions_id_put**](PermissionsApi.md#api_v2_permissions_id_put) | **PUT** /api/v2/permissions/{id} | Update permission
[**api_v2_permissions_id_roles_get**](PermissionsApi.md#api_v2_permissions_id_roles_get) | **GET** /api/v2/permissions/{id}/roles | Get permission roles
[**api_v2_permissions_post**](PermissionsApi.md#api_v2_permissions_post) | **POST** /api/v2/permissions | Create a new permission
[**api_v2_permissions_resource_resource_id_get**](PermissionsApi.md#api_v2_permissions_resource_resource_id_get) | **GET** /api/v2/permissions/resource/{resource_id} | Get permissions by resource
[**api_v2_permissions_search_post**](PermissionsApi.md#api_v2_permissions_search_post) | **POST** /api/v2/permissions/search | Search permissions


# **api_v2_permissions_get**
> DtoGenericResponseDtoPermissionListResponse api_v2_permissions_get(page=page, size=size, action=action, resource_id=resource_id, status=status, is_system=is_system, name=name)

List permissions

Get paginated list of permissions with optional filtering

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_permission_list_response import DtoGenericResponseDtoPermissionListResponse
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
    api_instance = rcabench.openapi.PermissionsApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)
    action = 'action_example' # str | Filter by action (optional)
    resource_id = 56 # int | Filter by resource ID (optional)
    status = 56 # int | Filter by status (optional)
    is_system = True # bool | Filter by system permission (optional)
    name = 'name_example' # str | Filter by permission name (optional)

    try:
        # List permissions
        api_response = api_instance.api_v2_permissions_get(page=page, size=size, action=action, resource_id=resource_id, status=status, is_system=is_system, name=name)
        print("The response of PermissionsApi->api_v2_permissions_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling PermissionsApi->api_v2_permissions_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]
 **action** | **str**| Filter by action | [optional] 
 **resource_id** | **int**| Filter by resource ID | [optional] 
 **status** | **int**| Filter by status | [optional] 
 **is_system** | **bool**| Filter by system permission | [optional] 
 **name** | **str**| Filter by permission name | [optional] 

### Return type

[**DtoGenericResponseDtoPermissionListResponse**](DtoGenericResponseDtoPermissionListResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Permissions retrieved successfully |  -  |
**400** | Invalid request parameters |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_permissions_id_delete**
> DtoGenericResponseAny api_v2_permissions_id_delete(id)

Delete permission

Delete a permission (soft delete by setting status to -1)

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
    api_instance = rcabench.openapi.PermissionsApi(api_client)
    id = 56 # int | Permission ID

    try:
        # Delete permission
        api_response = api_instance.api_v2_permissions_id_delete(id)
        print("The response of PermissionsApi->api_v2_permissions_id_delete:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling PermissionsApi->api_v2_permissions_id_delete: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Permission ID | 

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
**200** | Permission deleted successfully |  -  |
**400** | Invalid permission ID |  -  |
**403** | Cannot delete system permission |  -  |
**404** | Permission not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_permissions_id_get**
> DtoGenericResponseDtoPermissionResponse api_v2_permissions_id_get(id)

Get permission by ID

Get detailed information about a specific permission

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_permission_response import DtoGenericResponseDtoPermissionResponse
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
    api_instance = rcabench.openapi.PermissionsApi(api_client)
    id = 56 # int | Permission ID

    try:
        # Get permission by ID
        api_response = api_instance.api_v2_permissions_id_get(id)
        print("The response of PermissionsApi->api_v2_permissions_id_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling PermissionsApi->api_v2_permissions_id_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Permission ID | 

### Return type

[**DtoGenericResponseDtoPermissionResponse**](DtoGenericResponseDtoPermissionResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Permission retrieved successfully |  -  |
**400** | Invalid permission ID |  -  |
**404** | Permission not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_permissions_id_put**
> DtoGenericResponseDtoPermissionResponse api_v2_permissions_id_put(id, request)

Update permission

Update permission information (partial update supported)

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_permission_response import DtoGenericResponseDtoPermissionResponse
from rcabench.openapi.models.dto_update_permission_request import DtoUpdatePermissionRequest
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
    api_instance = rcabench.openapi.PermissionsApi(api_client)
    id = 56 # int | Permission ID
    request = rcabench.openapi.DtoUpdatePermissionRequest() # DtoUpdatePermissionRequest | Permission update request

    try:
        # Update permission
        api_response = api_instance.api_v2_permissions_id_put(id, request)
        print("The response of PermissionsApi->api_v2_permissions_id_put:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling PermissionsApi->api_v2_permissions_id_put: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Permission ID | 
 **request** | [**DtoUpdatePermissionRequest**](DtoUpdatePermissionRequest.md)| Permission update request | 

### Return type

[**DtoGenericResponseDtoPermissionResponse**](DtoGenericResponseDtoPermissionResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Permission updated successfully |  -  |
**400** | Invalid request |  -  |
**404** | Permission not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_permissions_id_roles_get**
> DtoGenericResponseArrayDtoRoleResponse api_v2_permissions_id_roles_get(id)

Get permission roles

Get list of roles that have been assigned a specific permission

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_array_dto_role_response import DtoGenericResponseArrayDtoRoleResponse
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
    api_instance = rcabench.openapi.PermissionsApi(api_client)
    id = 56 # int | Permission ID

    try:
        # Get permission roles
        api_response = api_instance.api_v2_permissions_id_roles_get(id)
        print("The response of PermissionsApi->api_v2_permissions_id_roles_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling PermissionsApi->api_v2_permissions_id_roles_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Permission ID | 

### Return type

[**DtoGenericResponseArrayDtoRoleResponse**](DtoGenericResponseArrayDtoRoleResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Roles retrieved successfully |  -  |
**400** | Invalid permission ID |  -  |
**404** | Permission not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_permissions_post**
> DtoGenericResponseDtoPermissionResponse api_v2_permissions_post(request)

Create a new permission

Create a new permission with specified resource and action

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_create_permission_request import DtoCreatePermissionRequest
from rcabench.openapi.models.dto_generic_response_dto_permission_response import DtoGenericResponseDtoPermissionResponse
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
    api_instance = rcabench.openapi.PermissionsApi(api_client)
    request = rcabench.openapi.DtoCreatePermissionRequest() # DtoCreatePermissionRequest | Permission creation request

    try:
        # Create a new permission
        api_response = api_instance.api_v2_permissions_post(request)
        print("The response of PermissionsApi->api_v2_permissions_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling PermissionsApi->api_v2_permissions_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoCreatePermissionRequest**](DtoCreatePermissionRequest.md)| Permission creation request | 

### Return type

[**DtoGenericResponseDtoPermissionResponse**](DtoGenericResponseDtoPermissionResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | Permission created successfully |  -  |
**400** | Invalid request |  -  |
**409** | Permission already exists |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_permissions_resource_resource_id_get**
> DtoGenericResponseArrayDtoPermissionResponse api_v2_permissions_resource_resource_id_get(resource_id)

Get permissions by resource

Get list of permissions associated with a specific resource

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_array_dto_permission_response import DtoGenericResponseArrayDtoPermissionResponse
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
    api_instance = rcabench.openapi.PermissionsApi(api_client)
    resource_id = 56 # int | Resource ID

    try:
        # Get permissions by resource
        api_response = api_instance.api_v2_permissions_resource_resource_id_get(resource_id)
        print("The response of PermissionsApi->api_v2_permissions_resource_resource_id_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling PermissionsApi->api_v2_permissions_resource_resource_id_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **resource_id** | **int**| Resource ID | 

### Return type

[**DtoGenericResponseArrayDtoPermissionResponse**](DtoGenericResponseArrayDtoPermissionResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Permissions retrieved successfully |  -  |
**400** | Invalid resource ID |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_permissions_search_post**
> DtoGenericResponseDtoSearchResponseDtoPermissionResponse api_v2_permissions_search_post(request)

Search permissions

Search permissions with complex filtering and sorting

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_permission_response import DtoGenericResponseDtoSearchResponseDtoPermissionResponse
from rcabench.openapi.models.dto_permission_search_request import DtoPermissionSearchRequest
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
    api_instance = rcabench.openapi.PermissionsApi(api_client)
    request = rcabench.openapi.DtoPermissionSearchRequest() # DtoPermissionSearchRequest | Permission search request

    try:
        # Search permissions
        api_response = api_instance.api_v2_permissions_search_post(request)
        print("The response of PermissionsApi->api_v2_permissions_search_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling PermissionsApi->api_v2_permissions_search_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoPermissionSearchRequest**](DtoPermissionSearchRequest.md)| Permission search request | 

### Return type

[**DtoGenericResponseDtoSearchResponseDtoPermissionResponse**](DtoGenericResponseDtoSearchResponseDtoPermissionResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Permissions retrieved successfully |  -  |
**400** | Invalid request |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

