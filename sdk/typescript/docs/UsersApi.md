# UsersApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createUser**](#createuser) | **POST** /api/v2/users | Create a new user|
|[**deleteUser**](#deleteuser) | **DELETE** /api/v2/users/{id} | Delete user|
|[**getUserById**](#getuserbyid) | **GET** /api/v2/users/{id}/detail | Get user by ID|
|[**listUsers**](#listusers) | **GET** /api/v2/users | List users|
|[**updateUser**](#updateuser) | **PATCH** /api/v2/users/{id} | Update user|

# **createUser**
> GenericResponseUserResp createUser(request)

Create a new user account with specified details

### Example

```typescript
import {
    UsersApi,
    Configuration,
    CreateUserReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let request: CreateUserReq; //User creation request

const { status, data } = await apiInstance.createUser(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **CreateUserReq**| User creation request | |


### Return type

**GenericResponseUserResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | User created successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**409** | User already exists |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteUser**
> GenericResponseAny deleteUser()

Delete a user

### Example

```typescript
import {
    UsersApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let id: number; //User ID (default to undefined)

const { status, data } = await apiInstance.deleteUser(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**number**] | User ID | defaults to undefined|


### Return type

**GenericResponseAny**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**204** | User deleted successfully |  -  |
|**400** | Invalid user ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | User not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getUserById**
> GenericResponseUserDetailResp getUserById()

Get detailed information about a specific user

### Example

```typescript
import {
    UsersApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let id: number; //User ID (default to undefined)

const { status, data } = await apiInstance.getUserById(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**number**] | User ID | defaults to undefined|


### Return type

**GenericResponseUserDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | User retrieved successfully |  -  |
|**400** | Invalid user ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | User not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listUsers**
> GenericResponseListUserResp listUsers()

Get paginated list of users with filtering

### Example

```typescript
import {
    UsersApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let page: number; //Page number (optional) (default to 1)
let size: number; //Page size (optional) (default to 20)
let username: string; //Filter by username (optional) (default to undefined)
let email: string; //Filter by email (optional) (default to undefined)
let isActive: boolean; //Filter by active status (optional) (default to undefined)
let status: StatusType; //Filter by status (optional) (default to undefined)

const { status, data } = await apiInstance.listUsers(
    page,
    size,
    username,
    email,
    isActive,
    status
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **size** | [**number**] | Page size | (optional) defaults to 20|
| **username** | [**string**] | Filter by username | (optional) defaults to undefined|
| **email** | [**string**] | Filter by email | (optional) defaults to undefined|
| **isActive** | [**boolean**] | Filter by active status | (optional) defaults to undefined|
| **status** | **StatusType** | Filter by status | (optional) defaults to undefined|


### Return type

**GenericResponseListUserResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Users retrieved successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateUser**
> GenericResponseUserResp updateUser(request)

Update an existing user\'s information

### Example

```typescript
import {
    UsersApi,
    Configuration,
    UpdateUserReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new UsersApi(configuration);

let id: number; //User ID (default to undefined)
let request: UpdateUserReq; //User update request

const { status, data } = await apiInstance.updateUser(
    id,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **UpdateUserReq**| User update request | |
| **id** | [**number**] | User ID | defaults to undefined|


### Return type

**GenericResponseUserResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**202** | User updated successfully |  -  |
|**400** | Invalid user ID/request |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | User not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

