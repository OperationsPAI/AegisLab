# rcabench.openapi.AuthenticationApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_auth_change_password_post**](AuthenticationApi.md#api_v2_auth_change_password_post) | **POST** /api/v2/auth/change-password | Change user password
[**api_v2_auth_login_post**](AuthenticationApi.md#api_v2_auth_login_post) | **POST** /api/v2/auth/login | User login
[**api_v2_auth_logout_post**](AuthenticationApi.md#api_v2_auth_logout_post) | **POST** /api/v2/auth/logout | User logout
[**api_v2_auth_profile_get**](AuthenticationApi.md#api_v2_auth_profile_get) | **GET** /api/v2/auth/profile | Get current user profile
[**api_v2_auth_refresh_post**](AuthenticationApi.md#api_v2_auth_refresh_post) | **POST** /api/v2/auth/refresh | Refresh JWT token
[**api_v2_auth_register_post**](AuthenticationApi.md#api_v2_auth_register_post) | **POST** /api/v2/auth/register | User registration


# **api_v2_auth_change_password_post**
> DtoGenericResponseAny api_v2_auth_change_password_post(request)

Change user password

Change password for authenticated user

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_change_password_request import DtoChangePasswordRequest
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
    api_instance = rcabench.openapi.AuthenticationApi(api_client)
    request = rcabench.openapi.DtoChangePasswordRequest() # DtoChangePasswordRequest | Password change request

    try:
        # Change user password
        api_response = api_instance.api_v2_auth_change_password_post(request)
        print("The response of AuthenticationApi->api_v2_auth_change_password_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AuthenticationApi->api_v2_auth_change_password_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoChangePasswordRequest**](DtoChangePasswordRequest.md)| Password change request | 

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
**200** | Password changed successfully |  -  |
**400** | Invalid request |  -  |
**401** | Unauthorized or invalid old password |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_auth_login_post**
> DtoGenericResponseDtoLoginResponse api_v2_auth_login_post(request)

User login

Authenticate user with username and password

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_login_response import DtoGenericResponseDtoLoginResponse
from rcabench.openapi.models.dto_login_request import DtoLoginRequest
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
    api_instance = rcabench.openapi.AuthenticationApi(api_client)
    request = rcabench.openapi.DtoLoginRequest() # DtoLoginRequest | Login credentials

    try:
        # User login
        api_response = api_instance.api_v2_auth_login_post(request)
        print("The response of AuthenticationApi->api_v2_auth_login_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AuthenticationApi->api_v2_auth_login_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoLoginRequest**](DtoLoginRequest.md)| Login credentials | 

### Return type

[**DtoGenericResponseDtoLoginResponse**](DtoGenericResponseDtoLoginResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Login successful |  -  |
**400** | Invalid request |  -  |
**401** | Invalid credentials |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_auth_logout_post**
> DtoGenericResponseAny api_v2_auth_logout_post(request)

User logout

Logout user and invalidate token

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny
from rcabench.openapi.models.dto_logout_request import DtoLogoutRequest
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
    api_instance = rcabench.openapi.AuthenticationApi(api_client)
    request = rcabench.openapi.DtoLogoutRequest() # DtoLogoutRequest | Logout request

    try:
        # User logout
        api_response = api_instance.api_v2_auth_logout_post(request)
        print("The response of AuthenticationApi->api_v2_auth_logout_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AuthenticationApi->api_v2_auth_logout_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoLogoutRequest**](DtoLogoutRequest.md)| Logout request | 

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
**200** | Logout successful |  -  |
**400** | Invalid request |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_auth_profile_get**
> DtoGenericResponseDtoUserResponse api_v2_auth_profile_get()

Get current user profile

Get profile information for authenticated user

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_user_response import DtoGenericResponseDtoUserResponse
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
    api_instance = rcabench.openapi.AuthenticationApi(api_client)

    try:
        # Get current user profile
        api_response = api_instance.api_v2_auth_profile_get()
        print("The response of AuthenticationApi->api_v2_auth_profile_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AuthenticationApi->api_v2_auth_profile_get: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

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
**200** | Profile retrieved successfully |  -  |
**401** | Unauthorized |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_auth_refresh_post**
> DtoGenericResponseDtoTokenRefreshResponse api_v2_auth_refresh_post(request)

Refresh JWT token

Refresh an existing JWT token

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_token_refresh_response import DtoGenericResponseDtoTokenRefreshResponse
from rcabench.openapi.models.dto_token_refresh_request import DtoTokenRefreshRequest
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
    api_instance = rcabench.openapi.AuthenticationApi(api_client)
    request = rcabench.openapi.DtoTokenRefreshRequest() # DtoTokenRefreshRequest | Token refresh request

    try:
        # Refresh JWT token
        api_response = api_instance.api_v2_auth_refresh_post(request)
        print("The response of AuthenticationApi->api_v2_auth_refresh_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AuthenticationApi->api_v2_auth_refresh_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoTokenRefreshRequest**](DtoTokenRefreshRequest.md)| Token refresh request | 

### Return type

[**DtoGenericResponseDtoTokenRefreshResponse**](DtoGenericResponseDtoTokenRefreshResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Token refreshed successfully |  -  |
**400** | Invalid request |  -  |
**401** | Invalid token |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_auth_register_post**
> DtoGenericResponseDtoUserInfo api_v2_auth_register_post(request)

User registration

Register a new user account

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_user_info import DtoGenericResponseDtoUserInfo
from rcabench.openapi.models.dto_register_request import DtoRegisterRequest
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
    api_instance = rcabench.openapi.AuthenticationApi(api_client)
    request = rcabench.openapi.DtoRegisterRequest() # DtoRegisterRequest | Registration details

    try:
        # User registration
        api_response = api_instance.api_v2_auth_register_post(request)
        print("The response of AuthenticationApi->api_v2_auth_register_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AuthenticationApi->api_v2_auth_register_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoRegisterRequest**](DtoRegisterRequest.md)| Registration details | 

### Return type

[**DtoGenericResponseDtoUserInfo**](DtoGenericResponseDtoUserInfo.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | Registration successful |  -  |
**400** | Invalid request or validation error |  -  |
**409** | User already exists |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

