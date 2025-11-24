# rcabench.openapi.AuthenticationApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**login**](AuthenticationApi.md#login) | **POST** /api/v2/auth/login | User login
[**register_user**](AuthenticationApi.md#register_user) | **POST** /api/v2/auth/register | User registration


# **login**
> GenericResponseLoginResp login(request)

User login

Authenticate user with username and password

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.generic_response_login_resp import GenericResponseLoginResp
from rcabench.openapi.models.login_req import LoginReq
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.AuthenticationApi(api_client)
    request = rcabench.openapi.LoginReq() # LoginReq | Login credentials

    try:
        # User login
        api_response = api_instance.login(request)
        print("The response of AuthenticationApi->login:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AuthenticationApi->login: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**LoginReq**](LoginReq.md)| Login credentials | 

### Return type

[**GenericResponseLoginResp**](GenericResponseLoginResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Login successful |  -  |
**400** | Invalid request format |  -  |
**401** | Invalid user name or password |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **register_user**
> GenericResponseUserInfo register_user(request)

User registration

Register a new user account

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.generic_response_user_info import GenericResponseUserInfo
from rcabench.openapi.models.register_req import RegisterReq
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.AuthenticationApi(api_client)
    request = rcabench.openapi.RegisterReq() # RegisterReq | Registration details

    try:
        # User registration
        api_response = api_instance.register_user(request)
        print("The response of AuthenticationApi->register_user:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AuthenticationApi->register_user: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**RegisterReq**](RegisterReq.md)| Registration details | 

### Return type

[**GenericResponseUserInfo**](GenericResponseUserInfo.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | Registration successful |  -  |
**400** | Invalid request format/parameters |  -  |
**409** | User already exists |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

