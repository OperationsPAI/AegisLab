# ContainersApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**buildContainerImage**](#buildcontainerimage) | **POST** /api/v2/containers/build | Submit container building|
|[**createContainer**](#createcontainer) | **POST** /api/v2/containers | Create container|
|[**createContainerVersion**](#createcontainerversion) | **POST** /api/v2/containers/{container_id}/versions | Create container version|
|[**getContainerById**](#getcontainerbyid) | **GET** /api/v2/containers/{container_id} | Get container by ID|
|[**getContainerVersionById**](#getcontainerversionbyid) | **GET** /api/v2/containers/{container_id}/versions/{version_id} | Get container version by ID|
|[**listContainerVersions**](#listcontainerversions) | **GET** /api/v2/containers/{container_id}/versions | List container versions|
|[**listContainers**](#listcontainers) | **GET** /api/v2/containers | List containers|

# **buildContainerImage**
> GenericResponseSubmitContainerBuildResp buildContainerImage(request)

Submit a container build task to build a container image from provided source files.

### Example

```typescript
import {
    ContainersApi,
    Configuration,
    SubmitBuildContainerReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ContainersApi(configuration);

let request: SubmitBuildContainerReq; //Container build request

const { status, data } = await apiInstance.buildContainerImage(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **SubmitBuildContainerReq**| Container build request | |


### Return type

**GenericResponseSubmitContainerBuildResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Container build task submitted successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Required files not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createContainer**
> GenericResponseContainerResp createContainer(request)

Create a new container without build configuration.

### Example

```typescript
import {
    ContainersApi,
    Configuration,
    CreateContainerReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ContainersApi(configuration);

let request: CreateContainerReq; //Container creation request

const { status, data } = await apiInstance.createContainer(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **CreateContainerReq**| Container creation request | |


### Return type

**GenericResponseContainerResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Container created successfully |  -  |
|**400** | Invalid request |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**409** | Conflict error |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createContainerVersion**
> GenericResponseContainerVersionResp createContainerVersion(request)

Create a new container version for an existing container.

### Example

```typescript
import {
    ContainersApi,
    Configuration,
    CreateContainerVersionReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ContainersApi(configuration);

let containerId: number; //Container ID (default to undefined)
let request: CreateContainerVersionReq; //Container version creation request

const { status, data } = await apiInstance.createContainerVersion(
    containerId,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **CreateContainerVersionReq**| Container version creation request | |
| **containerId** | [**number**] | Container ID | defaults to undefined|


### Return type

**GenericResponseContainerVersionResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Container version created successfully |  -  |
|**400** | Invalid container ID or invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**409** | Conflict error |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getContainerById**
> GenericResponseContainerDetailResp getContainerById()

Get detailed information about a specific container

### Example

```typescript
import {
    ContainersApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ContainersApi(configuration);

let containerId: number; //Container ID (default to undefined)

const { status, data } = await apiInstance.getContainerById(
    containerId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **containerId** | [**number**] | Container ID | defaults to undefined|


### Return type

**GenericResponseContainerDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Container retrieved successfully |  -  |
|**400** | Invalid container ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Container not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getContainerVersionById**
> GenericResponseContainerVersionDetailResp getContainerVersionById()

Get detailed information about a specific container version

### Example

```typescript
import {
    ContainersApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ContainersApi(configuration);

let containerId: number; //Container ID (default to undefined)
let versionId: number; //Container Version ID (default to undefined)

const { status, data } = await apiInstance.getContainerVersionById(
    containerId,
    versionId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **containerId** | [**number**] | Container ID | defaults to undefined|
| **versionId** | [**number**] | Container Version ID | defaults to undefined|


### Return type

**GenericResponseContainerVersionDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Container version retrieved successfully |  -  |
|**400** | Invalid container ID/container version ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Container or version not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listContainerVersions**
> GenericResponseListContainerVersionResp listContainerVersions()

Get paginated list of container versions for a specific container

### Example

```typescript
import {
    ContainersApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ContainersApi(configuration);

let containerId: number; //Container ID (default to undefined)
let page: number; //Page number (optional) (default to 1)
let size: number; //Page size (optional) (default to 20)
let status: StatusType; //Container version status filter (optional) (default to undefined)

const { status, data } = await apiInstance.listContainerVersions(
    containerId,
    page,
    size,
    status
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **containerId** | [**number**] | Container ID | defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **size** | [**number**] | Page size | (optional) defaults to 20|
| **status** | **StatusType** | Container version status filter | (optional) defaults to undefined|


### Return type

**GenericResponseListContainerVersionResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Container versions retrieved successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listContainers**
> GenericResponseListContainerResp listContainers()

Get paginated list of containers with pagination and filtering

### Example

```typescript
import {
    ContainersApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ContainersApi(configuration);

let page: number; //Page number (optional) (default to 1)
let size: PageSize; //Page size (optional) (default to undefined)
let type: ContainerType; //Container type filter (optional) (default to undefined)
let isPublic: boolean; //Container public visibility filter (optional) (default to undefined)
let status: StatusType; //Container status filter (optional) (default to undefined)

const { status, data } = await apiInstance.listContainers(
    page,
    size,
    type,
    isPublic,
    status
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **size** | **PageSize** | Page size | (optional) defaults to undefined|
| **type** | **ContainerType** | Container type filter | (optional) defaults to undefined|
| **isPublic** | [**boolean**] | Container public visibility filter | (optional) defaults to undefined|
| **status** | **StatusType** | Container status filter | (optional) defaults to undefined|


### Return type

**GenericResponseListContainerResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Containers retrieved successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

