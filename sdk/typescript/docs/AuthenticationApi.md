# AuthenticationApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**login**](#login) | **POST** /api/v2/auth/login | User login|
|[**registerUser**](#registeruser) | **POST** /api/v2/auth/register | User registration|

# **login**
> GenericResponseLoginResp login(request)

Authenticate user with username and password

### Example

```typescript
import {
    AuthenticationApi,
    Configuration,
    LoginReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new AuthenticationApi(configuration);

let request: LoginReq; //Login credentials

const { status, data } = await apiInstance.login(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **LoginReq**| Login credentials | |


### Return type

**GenericResponseLoginResp**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Login successful |  -  |
|**400** | Invalid request format |  -  |
|**401** | Invalid user name or password |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **registerUser**
> GenericResponseUserInfo registerUser(request)

Register a new user account

### Example

```typescript
import {
    AuthenticationApi,
    Configuration,
    RegisterReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new AuthenticationApi(configuration);

let request: RegisterReq; //Registration details

const { status, data } = await apiInstance.registerUser(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **RegisterReq**| Registration details | |


### Return type

**GenericResponseUserInfo**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Registration successful |  -  |
|**400** | Invalid request format/parameters |  -  |
|**409** | User already exists |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

