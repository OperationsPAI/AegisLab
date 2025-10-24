# rcabench.openapi.RolesApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_roles_get**](RolesApi.md#api_v2_roles_get) | **GET** /api/v2/roles | List roles
[**api_v2_roles_id_delete**](RolesApi.md#api_v2_roles_id_delete) | **DELETE** /api/v2/roles/{id} | Delete role
[**api_v2_roles_id_get**](RolesApi.md#api_v2_roles_id_get) | **GET** /api/v2/roles/{id} | Get role by ID
[**api_v2_roles_id_permissions_delete**](RolesApi.md#api_v2_roles_id_permissions_delete) | **DELETE** /api/v2/roles/{id}/permissions | Remove permissions from role
[**api_v2_roles_id_permissions_post**](RolesApi.md#api_v2_roles_id_permissions_post) | **POST** /api/v2/roles/{id}/permissions | Assign permissions to role
[**api_v2_roles_id_put**](RolesApi.md#api_v2_roles_id_put) | **PUT** /api/v2/roles/{id} | Update role
[**api_v2_roles_id_users_get**](RolesApi.md#api_v2_roles_id_users_get) | **GET** /api/v2/roles/{id}/users | Get role users
[**api_v2_roles_post**](RolesApi.md#api_v2_roles_post) | **POST** /api/v2/roles | Create a new role
[**api_v2_roles_search_post**](RolesApi.md#api_v2_roles_search_post) | **POST** /api/v2/roles/search | Search roles


# **api_v2_roles_get**
> DtoGenericResponseDtoRoleListResponse api_v2_roles_get(page=page, size=size, type=type, status=status, is_system=is_system, name=name)

List roles

Get paginated list of roles with optional filtering

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_role_list_response import DtoGenericResponseDtoRoleListResponse
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
    api_instance = rcabench.openapi.RolesApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)
    type = 'type_example' # str | Filter by role type (optional)
    status = 56 # int | Filter by status (optional)
    is_system = True # bool | Filter by system role (optional)
    name = 'name_example' # str | Filter by role name (optional)

    try:
        # List roles
        api_response = api_instance.api_v2_roles_get(page=page, size=size, type=type, status=status, is_system=is_system, name=name)
        print("The response of RolesApi->api_v2_roles_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RolesApi->api_v2_roles_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]
 **type** | **str**| Filter by role type | [optional] 
 **status** | **int**| Filter by status | [optional] 
 **is_system** | **bool**| Filter by system role | [optional] 
 **name** | **str**| Filter by role name | [optional] 

### Return type

[**DtoGenericResponseDtoRoleListResponse**](DtoGenericResponseDtoRoleListResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Roles retrieved successfully |  -  |
**400** | Invalid request parameters |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_roles_id_delete**
> DtoGenericResponseAny api_v2_roles_id_delete(id)

Delete role

Delete a role (soft delete by setting status to -1)

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny
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
    api_instance = rcabench.openapi.RolesApi(api_client)
    id = 56 # int | Role ID

    try:
        # Delete role
        api_response = api_instance.api_v2_roles_id_delete(id)
        print("The response of RolesApi->api_v2_roles_id_delete:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RolesApi->api_v2_roles_id_delete: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Role ID | 

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
**200** | Role deleted successfully |  -  |
**400** | Invalid role ID |  -  |
**403** | Cannot delete system role |  -  |
**404** | Role not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_roles_id_get**
> DtoGenericResponseDtoRoleResponse api_v2_roles_id_get(id)

Get role by ID

Get detailed information about a specific role

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_role_response import DtoGenericResponseDtoRoleResponse
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
    api_instance = rcabench.openapi.RolesApi(api_client)
    id = 56 # int | Role ID

    try:
        # Get role by ID
        api_response = api_instance.api_v2_roles_id_get(id)
        print("The response of RolesApi->api_v2_roles_id_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RolesApi->api_v2_roles_id_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Role ID | 

### Return type

[**DtoGenericResponseDtoRoleResponse**](DtoGenericResponseDtoRoleResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Role retrieved successfully |  -  |
**400** | Invalid role ID |  -  |
**404** | Role not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_roles_id_permissions_delete**
> DtoGenericResponseAny api_v2_roles_id_permissions_delete(id, request)

Remove permissions from role

Remove multiple permissions from a role

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny
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
    api_instance = rcabench.openapi.RolesApi(api_client)
    id = 56 # int | Role ID
    request = None # object | Permission removal request

    try:
        # Remove permissions from role
        api_response = api_instance.api_v2_roles_id_permissions_delete(id, request)
        print("The response of RolesApi->api_v2_roles_id_permissions_delete:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RolesApi->api_v2_roles_id_permissions_delete: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Role ID | 
 **request** | **object**| Permission removal request | 

### Return type

[**DtoGenericResponseAny**](DtoGenericResponseAny.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Permissions removed successfully |  -  |
**400** | Invalid request |  -  |
**404** | Role not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_roles_id_permissions_post**
> DtoGenericResponseAny api_v2_roles_id_permissions_post(id, request)

Assign permissions to role

Assign multiple permissions to a role

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny
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
    api_instance = rcabench.openapi.RolesApi(api_client)
    id = 56 # int | Role ID
    request = None # object | Permission assignment request

    try:
        # Assign permissions to role
        api_response = api_instance.api_v2_roles_id_permissions_post(id, request)
        print("The response of RolesApi->api_v2_roles_id_permissions_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RolesApi->api_v2_roles_id_permissions_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Role ID | 
 **request** | **object**| Permission assignment request | 

### Return type

[**DtoGenericResponseAny**](DtoGenericResponseAny.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Permissions assigned successfully |  -  |
**400** | Invalid request |  -  |
**404** | Role not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_roles_id_put**
> DtoGenericResponseDtoRoleResponse api_v2_roles_id_put(id, request)

Update role

Update role information (partial update supported)

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_role_response import DtoGenericResponseDtoRoleResponse
from rcabench.openapi.models.dto_update_role_request import DtoUpdateRoleRequest
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
    api_instance = rcabench.openapi.RolesApi(api_client)
    id = 56 # int | Role ID
    request = rcabench.openapi.DtoUpdateRoleRequest() # DtoUpdateRoleRequest | Role update request

    try:
        # Update role
        api_response = api_instance.api_v2_roles_id_put(id, request)
        print("The response of RolesApi->api_v2_roles_id_put:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RolesApi->api_v2_roles_id_put: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Role ID | 
 **request** | [**DtoUpdateRoleRequest**](DtoUpdateRoleRequest.md)| Role update request | 

### Return type

[**DtoGenericResponseDtoRoleResponse**](DtoGenericResponseDtoRoleResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Role updated successfully |  -  |
**400** | Invalid request |  -  |
**404** | Role not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_roles_id_users_get**
> DtoGenericResponseArrayDtoUserResponse api_v2_roles_id_users_get(id)

Get role users

Get list of users assigned to a specific role

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_array_dto_user_response import DtoGenericResponseArrayDtoUserResponse
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
    api_instance = rcabench.openapi.RolesApi(api_client)
    id = 56 # int | Role ID

    try:
        # Get role users
        api_response = api_instance.api_v2_roles_id_users_get(id)
        print("The response of RolesApi->api_v2_roles_id_users_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RolesApi->api_v2_roles_id_users_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Role ID | 

### Return type

[**DtoGenericResponseArrayDtoUserResponse**](DtoGenericResponseArrayDtoUserResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Users retrieved successfully |  -  |
**400** | Invalid role ID |  -  |
**404** | Role not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_roles_post**
> DtoGenericResponseDtoRoleResponse api_v2_roles_post(request)

Create a new role

Create a new role with specified permissions

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_create_role_request import DtoCreateRoleRequest
from rcabench.openapi.models.dto_generic_response_dto_role_response import DtoGenericResponseDtoRoleResponse
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
    api_instance = rcabench.openapi.RolesApi(api_client)
    request = rcabench.openapi.DtoCreateRoleRequest() # DtoCreateRoleRequest | Role creation request

    try:
        # Create a new role
        api_response = api_instance.api_v2_roles_post(request)
        print("The response of RolesApi->api_v2_roles_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RolesApi->api_v2_roles_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoCreateRoleRequest**](DtoCreateRoleRequest.md)| Role creation request | 

### Return type

[**DtoGenericResponseDtoRoleResponse**](DtoGenericResponseDtoRoleResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | Role created successfully |  -  |
**400** | Invalid request |  -  |
**409** | Role already exists |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_roles_search_post**
> DtoGenericResponseDtoSearchResponseDtoRoleResponse api_v2_roles_search_post(request)

Search roles

Search roles with complex filtering and sorting

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_role_response import DtoGenericResponseDtoSearchResponseDtoRoleResponse
from rcabench.openapi.models.dto_role_search_request import DtoRoleSearchRequest
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
    api_instance = rcabench.openapi.RolesApi(api_client)
    request = rcabench.openapi.DtoRoleSearchRequest() # DtoRoleSearchRequest | Role search request

    try:
        # Search roles
        api_response = api_instance.api_v2_roles_search_post(request)
        print("The response of RolesApi->api_v2_roles_search_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RolesApi->api_v2_roles_search_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoRoleSearchRequest**](DtoRoleSearchRequest.md)| Role search request | 

### Return type

[**DtoGenericResponseDtoSearchResponseDtoRoleResponse**](DtoGenericResponseDtoSearchResponseDtoRoleResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Roles retrieved successfully |  -  |
**400** | Invalid request |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

