# ResourcesApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getResourceById**](#getresourcebyid) | **GET** /api/v2/resources/{id} | Get resource by ID|
|[**listResourcePermissions**](#listresourcepermissions) | **GET** /api/v2/resources/{id}/permissions | List permissions from resource|
|[**listResources**](#listresources) | **GET** /api/v2/resources | List resources|

# **getResourceById**
> GenericResponseResourceResp getResourceById()

Get detailed information about a specific resource

### Example

```typescript
import {
    ResourcesApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ResourcesApi(configuration);

let id: number; //Resource ID (default to undefined)

const { status, data } = await apiInstance.getResourceById(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**number**] | Resource ID | defaults to undefined|


### Return type

**GenericResponseResourceResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Resource retrieved successfully |  -  |
|**400** | Invalid resource ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listResourcePermissions**
> GenericResponseArrayPermissionResp listResourcePermissions()

Get list of permissions assigned to a specific resource

### Example

```typescript
import {
    ResourcesApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ResourcesApi(configuration);

let id: number; //Resource ID (default to undefined)

const { status, data } = await apiInstance.listResourcePermissions(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**number**] | Resource ID | defaults to undefined|


### Return type

**GenericResponseArrayPermissionResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Permissions retrieved successfully |  -  |
|**400** | Invalid resource ID or request form |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listResources**
> GenericResponseListResourceResp listResources()

Get paginated list of resources with filtering

### Example

```typescript
import {
    ResourcesApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ResourcesApi(configuration);

let page: number; //Page number (optional) (default to 1)
let size: number; //Page size (optional) (default to 20)
let type: ResourceType; //Filter by resource type (optional) (default to undefined)
let category: ResourceCategory; //Filter by resource category (optional) (default to undefined)

const { status, data } = await apiInstance.listResources(
    page,
    size,
    type,
    category
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **size** | [**number**] | Page size | (optional) defaults to 20|
| **type** | **ResourceType** | Filter by resource type | (optional) defaults to undefined|
| **category** | **ResourceCategory** | Filter by resource category | (optional) defaults to undefined|


### Return type

**GenericResponseListResourceResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Resources retrieved successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

