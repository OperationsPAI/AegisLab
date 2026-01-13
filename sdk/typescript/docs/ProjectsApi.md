# ProjectsApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createProject**](#createproject) | **POST** /api/v2/projects | Create a new project|
|[**getProjectById**](#getprojectbyid) | **GET** /api/v2/projects/{project_id} | Get project by ID|
|[**listProjects**](#listprojects) | **GET** /api/v2/projects | List projects|

# **createProject**
> GenericResponseProjectResp createProject(request)

Create a new project with specified details

### Example

```typescript
import {
    ProjectsApi,
    Configuration,
    CreateProjectReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ProjectsApi(configuration);

let request: CreateProjectReq; //Project creation request

const { status, data } = await apiInstance.createProject(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **CreateProjectReq**| Project creation request | |


### Return type

**GenericResponseProjectResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Project created successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**409** | Project already exists |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getProjectById**
> GenericResponseProjectDetailResp getProjectById()

Get detailed information about a specific project

### Example

```typescript
import {
    ProjectsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ProjectsApi(configuration);

let projectId: number; //Project ID (default to undefined)

const { status, data } = await apiInstance.getProjectById(
    projectId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **projectId** | [**number**] | Project ID | defaults to undefined|


### Return type

**GenericResponseProjectDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Project retrieved successfully |  -  |
|**400** | Invalid project ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Project not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listProjects**
> GenericResponseListProjectResp listProjects()

Get paginated list of projects with filtering

### Example

```typescript
import {
    ProjectsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ProjectsApi(configuration);

let page: number; //Page number (optional) (default to 1)
let size: number; //Page size (optional) (default to 20)
let isPublic: boolean; //Filter by public status (optional) (default to undefined)
let status: StatusType; //Filter by status (optional) (default to undefined)

const { status, data } = await apiInstance.listProjects(
    page,
    size,
    isPublic,
    status
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **size** | [**number**] | Page size | (optional) defaults to 20|
| **isPublic** | [**boolean**] | Filter by public status | (optional) defaults to undefined|
| **status** | **StatusType** | Filter by status | (optional) defaults to undefined|


### Return type

**GenericResponseListProjectResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Projects retrieved successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

