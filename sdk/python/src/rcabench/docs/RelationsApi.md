# rcabench.openapi.RelationsApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_relations_batch_post**](RelationsApi.md#api_v2_relations_batch_post) | **POST** /api/v2/relations/batch | Batch relationship operations
[**api_v2_relations_get**](RelationsApi.md#api_v2_relations_get) | **GET** /api/v2/relations | List relationships
[**api_v2_relations_role_permissions_delete**](RelationsApi.md#api_v2_relations_role_permissions_delete) | **DELETE** /api/v2/relations/role-permissions | Remove permissions from role
[**api_v2_relations_role_permissions_post**](RelationsApi.md#api_v2_relations_role_permissions_post) | **POST** /api/v2/relations/role-permissions | Assign permissions to role
[**api_v2_relations_statistics_get**](RelationsApi.md#api_v2_relations_statistics_get) | **GET** /api/v2/relations/statistics | Get relationship statistics
[**api_v2_relations_user_permissions_delete**](RelationsApi.md#api_v2_relations_user_permissions_delete) | **DELETE** /api/v2/relations/user-permissions | Remove permission from user
[**api_v2_relations_user_permissions_post**](RelationsApi.md#api_v2_relations_user_permissions_post) | **POST** /api/v2/relations/user-permissions | Assign permission to user
[**api_v2_relations_user_roles_delete**](RelationsApi.md#api_v2_relations_user_roles_delete) | **DELETE** /api/v2/relations/user-roles | Remove role from user
[**api_v2_relations_user_roles_post**](RelationsApi.md#api_v2_relations_user_roles_post) | **POST** /api/v2/relations/user-roles | Assign role to user


# **api_v2_relations_batch_post**
> DtoGenericResponseAny api_v2_relations_batch_post(request)

Batch relationship operations

Perform multiple relationship operations in a single request

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_batch_relation_request import DtoBatchRelationRequest
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny
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
    api_instance = rcabench.openapi.RelationsApi(api_client)
    request = rcabench.openapi.DtoBatchRelationRequest() # DtoBatchRelationRequest | Batch relation operations request

    try:
        # Batch relationship operations
        api_response = api_instance.api_v2_relations_batch_post(request)
        print("The response of RelationsApi->api_v2_relations_batch_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RelationsApi->api_v2_relations_batch_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoBatchRelationRequest**](DtoBatchRelationRequest.md)| Batch relation operations request | 

### Return type

[**DtoGenericResponseAny**](DtoGenericResponseAny.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Batch operations completed successfully |  -  |
**400** | Invalid request |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_relations_get**
> DtoGenericResponseDtoRelationListResponse api_v2_relations_get(page=page, size=size, type=type, source_type=source_type, target_type=target_type, source_id=source_id, target_id=target_id)

List relationships

Get paginated list of relationships with optional filtering

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_relation_list_response import DtoGenericResponseDtoRelationListResponse
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
    api_instance = rcabench.openapi.RelationsApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)
    type = 'type_example' # str | Relationship type (optional)
    source_type = 'source_type_example' # str | Source entity type (optional)
    target_type = 'target_type_example' # str | Target entity type (optional)
    source_id = 56 # int | Source entity ID (optional)
    target_id = 56 # int | Target entity ID (optional)

    try:
        # List relationships
        api_response = api_instance.api_v2_relations_get(page=page, size=size, type=type, source_type=source_type, target_type=target_type, source_id=source_id, target_id=target_id)
        print("The response of RelationsApi->api_v2_relations_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RelationsApi->api_v2_relations_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]
 **type** | **str**| Relationship type | [optional] 
 **source_type** | **str**| Source entity type | [optional] 
 **target_type** | **str**| Target entity type | [optional] 
 **source_id** | **int**| Source entity ID | [optional] 
 **target_id** | **int**| Target entity ID | [optional] 

### Return type

[**DtoGenericResponseDtoRelationListResponse**](DtoGenericResponseDtoRelationListResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Relations retrieved successfully |  -  |
**400** | Invalid request parameters |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_relations_role_permissions_delete**
> DtoGenericResponseAny api_v2_relations_role_permissions_delete(request)

Remove permissions from role

Remove multiple permissions from a role

### Example


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


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.RelationsApi(api_client)
    request = None # object | Role permission removal request

    try:
        # Remove permissions from role
        api_response = api_instance.api_v2_relations_role_permissions_delete(request)
        print("The response of RelationsApi->api_v2_relations_role_permissions_delete:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RelationsApi->api_v2_relations_role_permissions_delete: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | **object**| Role permission removal request | 

### Return type

[**DtoGenericResponseAny**](DtoGenericResponseAny.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Permissions removed successfully |  -  |
**400** | Invalid request |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_relations_role_permissions_post**
> DtoGenericResponseAny api_v2_relations_role_permissions_post(request)

Assign permissions to role

Assign multiple permissions to a role

### Example


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


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.RelationsApi(api_client)
    request = None # object | Role permission assignment request

    try:
        # Assign permissions to role
        api_response = api_instance.api_v2_relations_role_permissions_post(request)
        print("The response of RelationsApi->api_v2_relations_role_permissions_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RelationsApi->api_v2_relations_role_permissions_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | **object**| Role permission assignment request | 

### Return type

[**DtoGenericResponseAny**](DtoGenericResponseAny.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Permissions assigned successfully |  -  |
**400** | Invalid request |  -  |
**404** | Role or permission not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_relations_statistics_get**
> DtoGenericResponseDtoRelationStatisticsResponse api_v2_relations_statistics_get()

Get relationship statistics

Get statistics about all relationship types in the system

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_relation_statistics_response import DtoGenericResponseDtoRelationStatisticsResponse
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
    api_instance = rcabench.openapi.RelationsApi(api_client)

    try:
        # Get relationship statistics
        api_response = api_instance.api_v2_relations_statistics_get()
        print("The response of RelationsApi->api_v2_relations_statistics_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RelationsApi->api_v2_relations_statistics_get: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**DtoGenericResponseDtoRelationStatisticsResponse**](DtoGenericResponseDtoRelationStatisticsResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Statistics retrieved successfully |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_relations_user_permissions_delete**
> DtoGenericResponseAny api_v2_relations_user_permissions_delete(request)

Remove permission from user

Remove a permission directly from a user

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny
from rcabench.openapi.models.dto_remove_user_permission_request import DtoRemoveUserPermissionRequest
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
    api_instance = rcabench.openapi.RelationsApi(api_client)
    request = rcabench.openapi.DtoRemoveUserPermissionRequest() # DtoRemoveUserPermissionRequest | User permission removal request

    try:
        # Remove permission from user
        api_response = api_instance.api_v2_relations_user_permissions_delete(request)
        print("The response of RelationsApi->api_v2_relations_user_permissions_delete:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RelationsApi->api_v2_relations_user_permissions_delete: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoRemoveUserPermissionRequest**](DtoRemoveUserPermissionRequest.md)| User permission removal request | 

### Return type

[**DtoGenericResponseAny**](DtoGenericResponseAny.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Permission removed successfully |  -  |
**400** | Invalid request |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_relations_user_permissions_post**
> DtoGenericResponseAny api_v2_relations_user_permissions_post(request)

Assign permission to user

Assign a permission directly to a user (with optional project scope)

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_assign_user_permission_request import DtoAssignUserPermissionRequest
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny
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
    api_instance = rcabench.openapi.RelationsApi(api_client)
    request = rcabench.openapi.DtoAssignUserPermissionRequest() # DtoAssignUserPermissionRequest | User permission assignment request

    try:
        # Assign permission to user
        api_response = api_instance.api_v2_relations_user_permissions_post(request)
        print("The response of RelationsApi->api_v2_relations_user_permissions_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RelationsApi->api_v2_relations_user_permissions_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoAssignUserPermissionRequest**](DtoAssignUserPermissionRequest.md)| User permission assignment request | 

### Return type

[**DtoGenericResponseAny**](DtoGenericResponseAny.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Permission assigned successfully |  -  |
**400** | Invalid request |  -  |
**404** | User or permission not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_relations_user_roles_delete**
> DtoGenericResponseAny api_v2_relations_user_roles_delete(request)

Remove role from user

Remove a role from a user (global role removal)

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny
from rcabench.openapi.models.dto_remove_user_role_request import DtoRemoveUserRoleRequest
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
    api_instance = rcabench.openapi.RelationsApi(api_client)
    request = rcabench.openapi.DtoRemoveUserRoleRequest() # DtoRemoveUserRoleRequest | User role removal request

    try:
        # Remove role from user
        api_response = api_instance.api_v2_relations_user_roles_delete(request)
        print("The response of RelationsApi->api_v2_relations_user_roles_delete:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RelationsApi->api_v2_relations_user_roles_delete: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoRemoveUserRoleRequest**](DtoRemoveUserRoleRequest.md)| User role removal request | 

### Return type

[**DtoGenericResponseAny**](DtoGenericResponseAny.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Role removed successfully |  -  |
**400** | Invalid request |  -  |
**404** | User or role not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_relations_user_roles_post**
> DtoGenericResponseAny api_v2_relations_user_roles_post(request)

Assign role to user

Assign a role to a user (global role assignment)

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_assign_user_role_request import DtoAssignUserRoleRequest
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny
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
    api_instance = rcabench.openapi.RelationsApi(api_client)
    request = rcabench.openapi.DtoAssignUserRoleRequest() # DtoAssignUserRoleRequest | User role assignment request

    try:
        # Assign role to user
        api_response = api_instance.api_v2_relations_user_roles_post(request)
        print("The response of RelationsApi->api_v2_relations_user_roles_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling RelationsApi->api_v2_relations_user_roles_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoAssignUserRoleRequest**](DtoAssignUserRoleRequest.md)| User role assignment request | 

### Return type

[**DtoGenericResponseAny**](DtoGenericResponseAny.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Role assigned successfully |  -  |
**400** | Invalid request |  -  |
**404** | User or role not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

