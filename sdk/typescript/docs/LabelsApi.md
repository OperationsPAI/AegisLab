# LabelsApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**batchDeleteLabels**](#batchdeletelabels) | **POST** /api/v2/labels/batch-delete | Batch delete labels|
|[**createLabel**](#createlabel) | **POST** /api/v2/labels | Create label|
|[**deleteLabel**](#deletelabel) | **DELETE** /api/v2/labels/{label_id} | Delete label|
|[**getLabelById**](#getlabelbyid) | **GET** /api/v2/labels/{label_id} | Get label by ID|
|[**listLabels**](#listlabels) | **GET** /api/v2/labels | List labels|
|[**updateLabel**](#updatelabel) | **PATCH** /api/v2/labels/{label_id} | Update label|

# **batchDeleteLabels**
> GenericResponseAny batchDeleteLabels(request)

Batch delete labels by IDs with cascading deletion of related records

### Example

```typescript
import {
    LabelsApi,
    Configuration,
    BatchDeleteLabelReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new LabelsApi(configuration);

let request: BatchDeleteLabelReq; //Batch delete request

const { status, data } = await apiInstance.batchDeleteLabels(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **BatchDeleteLabelReq**| Batch delete request | |


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
|**200** | Labels deleted successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createLabel**
> GenericResponseLabelResp createLabel(label)

Create a new label with key-value pair. If a deleted label with same key-value exists, it will be restored and updated.

### Example

```typescript
import {
    LabelsApi,
    Configuration,
    CreateLabelReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new LabelsApi(configuration);

let label: CreateLabelReq; //Label creation request

const { status, data } = await apiInstance.createLabel(
    label
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **label** | **CreateLabelReq**| Label creation request | |


### Return type

**GenericResponseLabelResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Label created successfully |  -  |
|**400** | Invalid request format/parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**409** | Label already exists |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteLabel**
> GenericResponseAny deleteLabel()

Delete a label and remove all its associations

### Example

```typescript
import {
    LabelsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new LabelsApi(configuration);

let labelId: number; //Label ID (default to undefined)

const { status, data } = await apiInstance.deleteLabel(
    labelId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **labelId** | [**number**] | Label ID | defaults to undefined|


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
|**204** | Label deleted successfully |  -  |
|**400** | Invalid label ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Label not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getLabelById**
> GenericResponseLabelDetailResp getLabelById()

Get detailed information about a specific label

### Example

```typescript
import {
    LabelsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new LabelsApi(configuration);

let labelId: number; //Label ID (default to undefined)

const { status, data } = await apiInstance.getLabelById(
    labelId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **labelId** | [**number**] | Label ID | defaults to undefined|


### Return type

**GenericResponseLabelDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Label retrieved successfully |  -  |
|**400** | Invalid label ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Label not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listLabels**
> GenericResponseListLabelResp listLabels()

Get paginated list of labels with filtering

### Example

```typescript
import {
    LabelsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new LabelsApi(configuration);

let page: number; //Page number (optional) (default to 1)
let size: number; //Page size (optional) (default to 20)
let key: string; //Filter by label key (optional) (default to undefined)
let value: string; //Filter by label value (optional) (default to undefined)
let category: LabelCategory; //Filter by category (optional) (default to undefined)
let isSystem: boolean; //Filter by system label (optional) (default to undefined)
let status: StatusType; //Filter by status (optional) (default to undefined)

const { status, data } = await apiInstance.listLabels(
    page,
    size,
    key,
    value,
    category,
    isSystem,
    status
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **size** | [**number**] | Page size | (optional) defaults to 20|
| **key** | [**string**] | Filter by label key | (optional) defaults to undefined|
| **value** | [**string**] | Filter by label value | (optional) defaults to undefined|
| **category** | **LabelCategory** | Filter by category | (optional) defaults to undefined|
| **isSystem** | [**boolean**] | Filter by system label | (optional) defaults to undefined|
| **status** | **StatusType** | Filter by status | (optional) defaults to undefined|


### Return type

**GenericResponseListLabelResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Labels retrieved successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateLabel**
> GenericResponseLabelResp updateLabel(request)

Update an existing label\'s information

### Example

```typescript
import {
    LabelsApi,
    Configuration,
    UpdateLabelReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new LabelsApi(configuration);

let labelId: number; //Label ID (default to undefined)
let request: UpdateLabelReq; //Label update request

const { status, data } = await apiInstance.updateLabel(
    labelId,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **UpdateLabelReq**| Label update request | |
| **labelId** | [**number**] | Label ID | defaults to undefined|


### Return type

**GenericResponseLabelResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**202** | Label updated successfully |  -  |
|**400** | Invalid label ID or invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Label not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

