# RolesApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createRole**](#createrole) | **POST** /api/v2/roles | Create a new role|
|[**deleteRole**](#deleterole) | **DELETE** /api/v2/roles/{id} | Delete role|
|[**getRoleById**](#getrolebyid) | **GET** /api/v2/roles/{id} | Get role by ID|
|[**grantPermissionsToRole**](#grantpermissionstorole) | **POST** /api/v2/roles/{role_id}/permissions/assign | Assign permissions to role|
|[**listRoles**](#listroles) | **GET** /api/v2/roles | List roles|
|[**revokePermissionsFromRole**](#revokepermissionsfromrole) | **POST** /api/v2/roles/{role_id}/permissions/remove | Remove permissions from role|
|[**updateRole**](#updaterole) | **PATCH** /api/v2/roles/{id} | Update role|

# **createRole**
> GenericResponseRoleResp createRole(request)

Create a new role with specified permissions

### Example

```typescript
import {
    RolesApi,
    Configuration,
    CreateRoleReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new RolesApi(configuration);

let request: CreateRoleReq; //Role creation request

const { status, data } = await apiInstance.createRole(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **CreateRoleReq**| Role creation request | |


### Return type

**GenericResponseRoleResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Role created successfully |  -  |
|**400** | Invalid request format |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**409** | Role already exists |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteRole**
> GenericResponseAny deleteRole()

Delete a role (soft delete by setting status to -1)

### Example

```typescript
import {
    RolesApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new RolesApi(configuration);

let id: number; //Role ID (default to undefined)

const { status, data } = await apiInstance.deleteRole(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**number**] | Role ID | defaults to undefined|


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
|**200** | Role deleted successfully |  -  |
|**400** | Invalid role ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied or cannot delete system role |  -  |
|**404** | Role not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getRoleById**
> GenericResponseRoleDetailResp getRoleById()

Get detailed information about a specific role

### Example

```typescript
import {
    RolesApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new RolesApi(configuration);

let id: number; //Role ID (default to undefined)

const { status, data } = await apiInstance.getRoleById(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**number**] | Role ID | defaults to undefined|


### Return type

**GenericResponseRoleDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Role retrieved successfully |  -  |
|**400** | Invalid role ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Role not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **grantPermissionsToRole**
> GenericResponseAny grantPermissionsToRole(request)

Assign multiple permissions to a role

### Example

```typescript
import {
    RolesApi,
    Configuration,
    AssignRolePermissionReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new RolesApi(configuration);

let roleId: number; //Role ID (default to undefined)
let request: AssignRolePermissionReq; //Permission assignment request

const { status, data } = await apiInstance.grantPermissionsToRole(
    roleId,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **AssignRolePermissionReq**| Permission assignment request | |
| **roleId** | [**number**] | Role ID | defaults to undefined|


### Return type

**GenericResponseAny**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Permissions assigned successfully |  -  |
|**400** | Invalid role ID or request format |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Role not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listRoles**
> GenericResponseListRoleResp listRoles()

Get paginated list of roles with optional filtering

### Example

```typescript
import {
    RolesApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new RolesApi(configuration);

let page: number; //Page number (optional) (default to 1)
let size: number; //Page size (optional) (default to 20)
let isSystem: boolean; //Filter by system role (optional) (default to undefined)
let status: StatusType; //Filter by status (optional) (default to undefined)

const { status, data } = await apiInstance.listRoles(
    page,
    size,
    isSystem,
    status
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **size** | [**number**] | Page size | (optional) defaults to 20|
| **isSystem** | [**boolean**] | Filter by system role | (optional) defaults to undefined|
| **status** | **StatusType** | Filter by status | (optional) defaults to undefined|


### Return type

**GenericResponseListRoleResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Roles retrieved successfully |  -  |
|**400** | Invalid request parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **revokePermissionsFromRole**
> GenericResponseAny revokePermissionsFromRole(request)

Remove multiple permissions from a role

### Example

```typescript
import {
    RolesApi,
    Configuration,
    RemoveRolePermissionReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new RolesApi(configuration);

let roleId: number; //Role ID (default to undefined)
let request: RemoveRolePermissionReq; //Permission removal request

const { status, data } = await apiInstance.revokePermissionsFromRole(
    roleId,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **RemoveRolePermissionReq**| Permission removal request | |
| **roleId** | [**number**] | Role ID | defaults to undefined|


### Return type

**GenericResponseAny**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Permissions removed successfully |  -  |
|**400** | Invalid role ID or request format |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Role not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateRole**
> GenericResponseRoleResp updateRole(request)

Update role information (partial update supported)

### Example

```typescript
import {
    RolesApi,
    Configuration,
    UpdateRoleReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new RolesApi(configuration);

let id: number; //Role ID (default to undefined)
let request: UpdateRoleReq; //Role update request

const { status, data } = await apiInstance.updateRole(
    id,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **UpdateRoleReq**| Role update request | |
| **id** | [**number**] | Role ID | defaults to undefined|


### Return type

**GenericResponseRoleResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**202** | Role updated successfully |  -  |
|**400** | Invalid request |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Role not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

