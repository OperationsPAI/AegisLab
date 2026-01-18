# PermissionsApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createPermission**](#createpermission) | **POST** /api/v2/permissions | Create a new permission|
|[**deletePermission**](#deletepermission) | **DELETE** /api/v2/permissions/{id} | Delete permission|
|[**getPermissionById**](#getpermissionbyid) | **GET** /api/v2/permissions/{id} | Get permission by ID|
|[**listPermissions**](#listpermissions) | **GET** /api/v2/permissions | List permissions|
|[**listRolesWithPermission**](#listroleswithpermission) | **GET** /api/v2/permissions/{permission_id}/roles | List roles from permission|
|[**updatePermission**](#updatepermission) | **PUT** /api/v2/permissions/{id} | Update permission|

# **createPermission**
> GenericResponsePermissionResp createPermission(request)

Create a new permission with specified resource and action

### Example

```typescript
import {
    PermissionsApi,
    Configuration,
    CreatePermissionReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new PermissionsApi(configuration);

let request: CreatePermissionReq; //Permission creation request

const { status, data } = await apiInstance.createPermission(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **CreatePermissionReq**| Permission creation request | |


### Return type

**GenericResponsePermissionResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Permission created successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**409** | Permission already exists |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deletePermission**
> GenericResponseAny deletePermission()

Delete a permission (soft delete by setting status to -1)

### Example

```typescript
import {
    PermissionsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new PermissionsApi(configuration);

let id: number; //Permission ID (default to undefined)

const { status, data } = await apiInstance.deletePermission(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**number**] | Permission ID | defaults to undefined|


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
|**204** | Permission deleted successfully |  -  |
|**400** | Invalid permission ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Cannot delete system permission |  -  |
|**404** | Permission not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getPermissionById**
> GenericResponsePermissionDetailResp getPermissionById()

Get detailed information about a specific permission

### Example

```typescript
import {
    PermissionsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new PermissionsApi(configuration);

let id: number; //Permission ID (default to undefined)

const { status, data } = await apiInstance.getPermissionById(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**number**] | Permission ID | defaults to undefined|


### Return type

**GenericResponsePermissionDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Permission retrieved successfully |  -  |
|**400** | Invalid permission ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Permission not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listPermissions**
> GenericResponsePermissionResp listPermissions()

Get paginated list of permissions with optional filtering

### Example

```typescript
import {
    PermissionsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new PermissionsApi(configuration);

let page: number; //Page number (optional) (default to 1)
let size: number; //Page size (optional) (default to 20)
let action: string; //Filter by action (optional) (default to undefined)
let isSystem: boolean; //Filter by system permission (optional) (default to undefined)
let status: StatusType; //Filter by status (optional) (default to undefined)

const { status, data } = await apiInstance.listPermissions(
    page,
    size,
    action,
    isSystem,
    status
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **size** | [**number**] | Page size | (optional) defaults to 20|
| **action** | [**string**] | Filter by action | (optional) defaults to undefined|
| **isSystem** | [**boolean**] | Filter by system permission | (optional) defaults to undefined|
| **status** | **StatusType** | Filter by status | (optional) defaults to undefined|


### Return type

**GenericResponsePermissionResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Permissions retrieved successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listRolesWithPermission**
> GenericResponseArrayRoleResp listRolesWithPermission()

Get list of roles assigned to a specific permission

### Example

```typescript
import {
    PermissionsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new PermissionsApi(configuration);

let permissionId: number; //Permission ID (default to undefined)

const { status, data } = await apiInstance.listRolesWithPermission(
    permissionId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **permissionId** | [**number**] | Permission ID | defaults to undefined|


### Return type

**GenericResponseArrayRoleResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Roles retrieved successfully |  -  |
|**400** | Invalid permission ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Permission not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updatePermission**
> GenericResponsePermissionResp updatePermission(request)

Update permission information (partial update supported)

### Example

```typescript
import {
    PermissionsApi,
    Configuration,
    UpdatePermissionReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new PermissionsApi(configuration);

let id: number; //Permission ID (default to undefined)
let request: UpdatePermissionReq; //Permission update request

const { status, data } = await apiInstance.updatePermission(
    id,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **UpdatePermissionReq**| Permission update request | |
| **id** | [**number**] | Permission ID | defaults to undefined|


### Return type

**GenericResponsePermissionResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**202** | Permission updated successfully |  -  |
|**400** | Invalid permission ID or request format/parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Permission not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

