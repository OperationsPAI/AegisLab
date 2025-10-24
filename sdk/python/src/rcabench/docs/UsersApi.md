# rcabench.openapi.UsersApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_users_get**](UsersApi.md#api_v2_users_get) | **GET** /api/v2/users | List users
[**api_v2_users_id_delete**](UsersApi.md#api_v2_users_id_delete) | **DELETE** /api/v2/users/{id} | Delete user
[**api_v2_users_id_get**](UsersApi.md#api_v2_users_id_get) | **GET** /api/v2/users/{id} | Get user by ID
[**api_v2_users_id_projects_post**](UsersApi.md#api_v2_users_id_projects_post) | **POST** /api/v2/users/{id}/projects | Assign user to project
[**api_v2_users_id_projects_project_id_delete**](UsersApi.md#api_v2_users_id_projects_project_id_delete) | **DELETE** /api/v2/users/{id}/projects/{project_id} | Remove user from project
[**api_v2_users_id_put**](UsersApi.md#api_v2_users_id_put) | **PUT** /api/v2/users/{id} | Update user
[**api_v2_users_post**](UsersApi.md#api_v2_users_post) | **POST** /api/v2/users | Create a new user
[**api_v2_users_search_post**](UsersApi.md#api_v2_users_search_post) | **POST** /api/v2/users/search | Search users


# **api_v2_users_get**
> DtoGenericResponseDtoUserListResponse api_v2_users_get(page=page, size=size, status=status, is_active=is_active, username=username, email=email, full_name=full_name)

List users

Get paginated list of users with optional filtering

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_user_list_response import DtoGenericResponseDtoUserListResponse
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
    api_instance = rcabench.openapi.UsersApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)
    status = 56 # int | Filter by status (optional)
    is_active = True # bool | Filter by active status (optional)
    username = 'username_example' # str | Filter by username (optional)
    email = 'email_example' # str | Filter by email (optional)
    full_name = 'full_name_example' # str | Filter by full name (optional)

    try:
        # List users
        api_response = api_instance.api_v2_users_get(page=page, size=size, status=status, is_active=is_active, username=username, email=email, full_name=full_name)
        print("The response of UsersApi->api_v2_users_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling UsersApi->api_v2_users_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]
 **status** | **int**| Filter by status | [optional] 
 **is_active** | **bool**| Filter by active status | [optional] 
 **username** | **str**| Filter by username | [optional] 
 **email** | **str**| Filter by email | [optional] 
 **full_name** | **str**| Filter by full name | [optional] 

### Return type

[**DtoGenericResponseDtoUserListResponse**](DtoGenericResponseDtoUserListResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Users retrieved successfully |  -  |
**400** | Invalid request parameters |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_users_id_delete**
> DtoGenericResponseAny api_v2_users_id_delete(id)

Delete user

Delete a user (soft delete by setting status to -1)

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
    api_instance = rcabench.openapi.UsersApi(api_client)
    id = 56 # int | User ID

    try:
        # Delete user
        api_response = api_instance.api_v2_users_id_delete(id)
        print("The response of UsersApi->api_v2_users_id_delete:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling UsersApi->api_v2_users_id_delete: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| User ID | 

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
**200** | User deleted successfully |  -  |
**400** | Invalid user ID |  -  |
**404** | User not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_users_id_get**
> DtoGenericResponseDtoUserResponse api_v2_users_id_get(id)

Get user by ID

Get detailed information about a specific user

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_user_response import DtoGenericResponseDtoUserResponse
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
    api_instance = rcabench.openapi.UsersApi(api_client)
    id = 56 # int | User ID

    try:
        # Get user by ID
        api_response = api_instance.api_v2_users_id_get(id)
        print("The response of UsersApi->api_v2_users_id_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling UsersApi->api_v2_users_id_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| User ID | 

### Return type

[**DtoGenericResponseDtoUserResponse**](DtoGenericResponseDtoUserResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | User retrieved successfully |  -  |
**400** | Invalid user ID |  -  |
**404** | User not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_users_id_projects_post**
> DtoGenericResponseAny api_v2_users_id_projects_post(id, request)

Assign user to project

Assign a user to a project with a specific role

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_assign_user_to_project_request import DtoAssignUserToProjectRequest
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
    api_instance = rcabench.openapi.UsersApi(api_client)
    id = 56 # int | User ID
    request = rcabench.openapi.DtoAssignUserToProjectRequest() # DtoAssignUserToProjectRequest | Project assignment request

    try:
        # Assign user to project
        api_response = api_instance.api_v2_users_id_projects_post(id, request)
        print("The response of UsersApi->api_v2_users_id_projects_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling UsersApi->api_v2_users_id_projects_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| User ID | 
 **request** | [**DtoAssignUserToProjectRequest**](DtoAssignUserToProjectRequest.md)| Project assignment request | 

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
**200** | User assigned to project successfully |  -  |
**400** | Invalid request |  -  |
**404** | User not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_users_id_projects_project_id_delete**
> DtoGenericResponseAny api_v2_users_id_projects_project_id_delete(id, project_id)

Remove user from project

Remove a user from a project

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
    api_instance = rcabench.openapi.UsersApi(api_client)
    id = 56 # int | User ID
    project_id = 56 # int | Project ID

    try:
        # Remove user from project
        api_response = api_instance.api_v2_users_id_projects_project_id_delete(id, project_id)
        print("The response of UsersApi->api_v2_users_id_projects_project_id_delete:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling UsersApi->api_v2_users_id_projects_project_id_delete: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| User ID | 
 **project_id** | **int**| Project ID | 

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
**200** | User removed from project successfully |  -  |
**400** | Invalid ID |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_users_id_put**
> DtoGenericResponseDtoUserResponse api_v2_users_id_put(id, request)

Update user

Update user information (partial update supported)

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_user_response import DtoGenericResponseDtoUserResponse
from rcabench.openapi.models.dto_update_user_request import DtoUpdateUserRequest
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
    api_instance = rcabench.openapi.UsersApi(api_client)
    id = 56 # int | User ID
    request = rcabench.openapi.DtoUpdateUserRequest() # DtoUpdateUserRequest | User update request

    try:
        # Update user
        api_response = api_instance.api_v2_users_id_put(id, request)
        print("The response of UsersApi->api_v2_users_id_put:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling UsersApi->api_v2_users_id_put: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| User ID | 
 **request** | [**DtoUpdateUserRequest**](DtoUpdateUserRequest.md)| User update request | 

### Return type

[**DtoGenericResponseDtoUserResponse**](DtoGenericResponseDtoUserResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | User updated successfully |  -  |
**400** | Invalid request |  -  |
**404** | User not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_users_post**
> DtoGenericResponseDtoUserResponse api_v2_users_post(request)

Create a new user

Create a new user account with specified details

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_create_user_request import DtoCreateUserRequest
from rcabench.openapi.models.dto_generic_response_dto_user_response import DtoGenericResponseDtoUserResponse
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
    api_instance = rcabench.openapi.UsersApi(api_client)
    request = rcabench.openapi.DtoCreateUserRequest() # DtoCreateUserRequest | User creation request

    try:
        # Create a new user
        api_response = api_instance.api_v2_users_post(request)
        print("The response of UsersApi->api_v2_users_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling UsersApi->api_v2_users_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoCreateUserRequest**](DtoCreateUserRequest.md)| User creation request | 

### Return type

[**DtoGenericResponseDtoUserResponse**](DtoGenericResponseDtoUserResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | User created successfully |  -  |
**400** | Invalid request |  -  |
**409** | User already exists |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_users_search_post**
> DtoGenericResponseDtoSearchResponseDtoUserResponse api_v2_users_search_post(request)

Search users

Search users with complex filtering and sorting

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_user_response import DtoGenericResponseDtoSearchResponseDtoUserResponse
from rcabench.openapi.models.dto_user_search_request import DtoUserSearchRequest
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
    api_instance = rcabench.openapi.UsersApi(api_client)
    request = rcabench.openapi.DtoUserSearchRequest() # DtoUserSearchRequest | User search request

    try:
        # Search users
        api_response = api_instance.api_v2_users_search_post(request)
        print("The response of UsersApi->api_v2_users_search_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling UsersApi->api_v2_users_search_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoUserSearchRequest**](DtoUserSearchRequest.md)| User search request | 

### Return type

[**DtoGenericResponseDtoSearchResponseDtoUserResponse**](DtoGenericResponseDtoSearchResponseDtoUserResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Users retrieved successfully |  -  |
**400** | Invalid request |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

