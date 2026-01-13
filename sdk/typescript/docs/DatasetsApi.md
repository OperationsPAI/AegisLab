# DatasetsApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createDataset**](#createdataset) | **POST** /api/v2/datasets | Create dataset|
|[**createDatasetVersion**](#createdatasetversion) | **POST** /api/v2/datasets/{dataset_id}/versions | Create dataset version|
|[**downloadDatasetVersion**](#downloaddatasetversion) | **GET** /api/v2/datasets/{dataset_id}/versions/{version_id}/download | Download dataset version|
|[**getDatasetById**](#getdatasetbyid) | **GET** /api/v2/datasets/{dataset_id} | Get dataset by ID|
|[**getDatasetVersionById**](#getdatasetversionbyid) | **GET** /api/v2/datasets/{dataset_id}/versions/{version_id} | Get dataset version by ID|
|[**listDatasetVersions**](#listdatasetversions) | **GET** /api/v2/datasets/{dataset_id}/versions | List dataset versions|
|[**listDatasets**](#listdatasets) | **GET** /api/v2/datasets | List datasets|
|[**manageDatasetVersionInjections**](#managedatasetversioninjections) | **PATCH** /api/v2/datasets/{dataset_id}/version/{version_id}/injections | Manage dataset injections|
|[**searchDatasets**](#searchdatasets) | **POST** /api/v2/datasets/search | Search datasets|

# **createDataset**
> GenericResponseDatasetResp createDataset(request)

Create a new dataset with an initial version

### Example

```typescript
import {
    DatasetsApi,
    Configuration,
    CreateDatasetReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new DatasetsApi(configuration);

let request: CreateDatasetReq; //Dataset creation request

const { status, data } = await apiInstance.createDataset(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **CreateDatasetReq**| Dataset creation request | |


### Return type

**GenericResponseDatasetResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Dataset created successfully |  -  |
|**400** | Invalid request |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**409** | Conflict error |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createDatasetVersion**
> GenericResponseDatasetVersionResp createDatasetVersion(request)

Create a new dataset version for an existing dataset.

### Example

```typescript
import {
    DatasetsApi,
    Configuration,
    CreateDatasetVersionReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new DatasetsApi(configuration);

let datasetId: number; //Dataset ID (default to undefined)
let request: CreateDatasetVersionReq; //Dataset version creation request

const { status, data } = await apiInstance.createDatasetVersion(
    datasetId,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **CreateDatasetVersionReq**| Dataset version creation request | |
| **datasetId** | [**number**] | Dataset ID | defaults to undefined|


### Return type

**GenericResponseDatasetVersionResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Dataset version created successfully |  -  |
|**400** | Invalid request |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**409** | Conflict error |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **downloadDatasetVersion**
> File downloadDatasetVersion()

Download dataset file by version ID

### Example

```typescript
import {
    DatasetsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new DatasetsApi(configuration);

let datasetId: number; //Dataset ID (default to undefined)
let versionId: number; //Dataset Version ID (default to undefined)

const { status, data } = await apiInstance.downloadDatasetVersion(
    datasetId,
    versionId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **datasetId** | [**number**] | Dataset ID | defaults to undefined|
| **versionId** | [**number**] | Dataset Version ID | defaults to undefined|


### Return type

**File**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Dataset file |  -  |
|**400** | Invalid dataset ID/dataset version ID |  -  |
|**403** | Permission denied |  -  |
|**404** | Dataset not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getDatasetById**
> GenericResponseDatasetDetailResp getDatasetById()

Get detailed information about a specific dataset

### Example

```typescript
import {
    DatasetsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new DatasetsApi(configuration);

let datasetId: number; //Dataset ID (default to undefined)

const { status, data } = await apiInstance.getDatasetById(
    datasetId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **datasetId** | [**number**] | Dataset ID | defaults to undefined|


### Return type

**GenericResponseDatasetDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Dataset retrieved successfully |  -  |
|**400** | Invalid dataset ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Dataset not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getDatasetVersionById**
> GenericResponseDatasetVersionDetailResp getDatasetVersionById()

Get detailed information about a specific dataset version

### Example

```typescript
import {
    DatasetsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new DatasetsApi(configuration);

let datasetId: number; //Dataset ID (default to undefined)
let versionId: number; //Dataset Version ID (default to undefined)

const { status, data } = await apiInstance.getDatasetVersionById(
    datasetId,
    versionId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **datasetId** | [**number**] | Dataset ID | defaults to undefined|
| **versionId** | [**number**] | Dataset Version ID | defaults to undefined|


### Return type

**GenericResponseDatasetVersionDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Dataset version retrieved successfully |  -  |
|**400** | Invalid dataset ID/dataset version ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Dataset or version not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listDatasetVersions**
> GenericResponseListDatasetVersionResp listDatasetVersions()

Get paginated list of dataset versions for a specific dataset

### Example

```typescript
import {
    DatasetsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new DatasetsApi(configuration);

let datasetId: number; //Dataset ID (default to undefined)
let page: number; //Page number (optional) (default to 1)
let size: number; //Page size (optional) (default to 20)
let status: StatusType; //Dataset version status filter (optional) (default to undefined)

const { status, data } = await apiInstance.listDatasetVersions(
    datasetId,
    page,
    size,
    status
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **datasetId** | [**number**] | Dataset ID | defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **size** | [**number**] | Page size | (optional) defaults to 20|
| **status** | **StatusType** | Dataset version status filter | (optional) defaults to undefined|


### Return type

**GenericResponseListDatasetVersionResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Dataset versions retrieved successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listDatasets**
> GenericResponseListDatasetResp listDatasets()

Get paginated list of datasets with pagination and filtering

### Example

```typescript
import {
    DatasetsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new DatasetsApi(configuration);

let page: number; //Page number (optional) (default to 1)
let size: number; //Page size (optional) (default to 20)
let type: string; //Dataset type filter (optional) (default to undefined)
let isPublic: boolean; //Dataset public visibility filter (optional) (default to undefined)
let status: StatusType; //Dataset status filter (optional) (default to undefined)

const { status, data } = await apiInstance.listDatasets(
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
| **size** | [**number**] | Page size | (optional) defaults to 20|
| **type** | [**string**] | Dataset type filter | (optional) defaults to undefined|
| **isPublic** | [**boolean**] | Dataset public visibility filter | (optional) defaults to undefined|
| **status** | **StatusType** | Dataset status filter | (optional) defaults to undefined|


### Return type

**GenericResponseListDatasetResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Datasets retrieved successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **manageDatasetVersionInjections**
> GenericResponseDatasetVersionDetailResp manageDatasetVersionInjections(manage)

Add or remove injections for a dataset

### Example

```typescript
import {
    DatasetsApi,
    Configuration,
    ManageDatasetVersionInjectionReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new DatasetsApi(configuration);

let datasetId: number; //Dataset ID (default to undefined)
let versionId: number; //Dataset Version ID (default to undefined)
let manage: ManageDatasetVersionInjectionReq; //Injection management request

const { status, data } = await apiInstance.manageDatasetVersionInjections(
    datasetId,
    versionId,
    manage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **manage** | **ManageDatasetVersionInjectionReq**| Injection management request | |
| **datasetId** | [**number**] | Dataset ID | defaults to undefined|
| **versionId** | [**number**] | Dataset Version ID | defaults to undefined|


### Return type

**GenericResponseDatasetVersionDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Injections managed successfully |  -  |
|**400** | Invalid dataset ID or invalid request format/parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Dataset not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **searchDatasets**
> GenericResponseListDatasetDetailResp searchDatasets(request)

Search datasets with advanced filtering options

### Example

```typescript
import {
    DatasetsApi,
    Configuration,
    SearchDatasetReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new DatasetsApi(configuration);

let request: SearchDatasetReq; //Dataset search request

const { status, data } = await apiInstance.searchDatasets(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **SearchDatasetReq**| Dataset search request | |


### Return type

**GenericResponseListDatasetDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Datasets retrieved successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

