# ExecutionsApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getExecutionById**](#getexecutionbyid) | **GET** /api/v2/executions/{id} | Get execution by ID|
|[**listExecutionLabels**](#listexecutionlabels) | **GET** /api/v2/executions/labels | List execution labels|
|[**listExecutions**](#listexecutions) | **GET** /api/v2/executions | List executions|
|[**runAlgorithm**](#runalgorithm) | **POST** /api/v2/executions/execute | Submit batch algorithm execution|
|[**uploadDetectionResults**](#uploaddetectionresults) | **POST** /api/v2/executions/{execution_id}/detector_results | Upload detector results|
|[**uploadLocalizationResults**](#uploadlocalizationresults) | **POST** /api/v2/executions/{execution_id}/granularity_results | Upload granularity results|

# **getExecutionById**
> GenericResponseExecutionDetailResp getExecutionById()

Get detailed information about a specific execution

### Example

```typescript
import {
    ExecutionsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ExecutionsApi(configuration);

let id: number; //Execution ID (default to undefined)

const { status, data } = await apiInstance.getExecutionById(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**number**] | Execution ID | defaults to undefined|


### Return type

**GenericResponseExecutionDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Execution retrieved successfully |  -  |
|**400** | Invalid execution ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Execution not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listExecutionLabels**
> GenericResponseArrayLabelItem listExecutionLabels()

List all available label keys for executions

### Example

```typescript
import {
    ExecutionsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ExecutionsApi(configuration);

const { status, data } = await apiInstance.listExecutionLabels();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**GenericResponseArrayLabelItem**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Available label keys |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listExecutions**
> GenericResponseListExecutionResp listExecutions()

Get a paginated list of executions with pagination and filtering

### Example

```typescript
import {
    ExecutionsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ExecutionsApi(configuration);

let page: number; //Page number (optional) (default to 1)
let size: number; //Page size (optional) (default to 20)
let state: ExecutionState; //Filter by execution state (optional) (default to undefined)
let status: StatusType; //Filter by status (optional) (default to undefined)
let labels: Array<string>; //Filter by labels (array of key:value strings, e.g., \'type:test\') (optional) (default to undefined)

const { status, data } = await apiInstance.listExecutions(
    page,
    size,
    state,
    status,
    labels
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **size** | [**number**] | Page size | (optional) defaults to 20|
| **state** | **ExecutionState** | Filter by execution state | (optional) defaults to undefined|
| **status** | **StatusType** | Filter by status | (optional) defaults to undefined|
| **labels** | **Array&lt;string&gt;** | Filter by labels (array of key:value strings, e.g., \&#39;type:test\&#39;) | (optional) defaults to undefined|


### Return type

**GenericResponseListExecutionResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Executions retrieved successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **runAlgorithm**
> GenericResponseSubmitExecutionResp runAlgorithm(request)

Submit multiple algorithm execution tasks in batch. Supports mixing datapack (v1 compatible) and dataset (v2 feature) executions.

### Example

```typescript
import {
    ExecutionsApi,
    Configuration,
    SubmitExecutionReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ExecutionsApi(configuration);

let request: SubmitExecutionReq; //Algorithm execution request

const { status, data } = await apiInstance.runAlgorithm(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **SubmitExecutionReq**| Algorithm execution request | |


### Return type

**GenericResponseSubmitExecutionResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Algorithm execution submitted successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Project, algorithm, datapack or dataset not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **uploadDetectionResults**
> GenericResponseUploadExecutionResultResp uploadDetectionResults(request)

Upload detection results for detector algorithm via API instead of file collection

### Example

```typescript
import {
    ExecutionsApi,
    Configuration,
    UploadDetectorResultReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ExecutionsApi(configuration);

let executionId: number; //Execution ID (default to undefined)
let request: UploadDetectorResultReq; //Detector results

const { status, data } = await apiInstance.uploadDetectionResults(
    executionId,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **UploadDetectorResultReq**| Detector results | |
| **executionId** | [**number**] | Execution ID | defaults to undefined|


### Return type

**GenericResponseUploadExecutionResultResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Results uploaded successfully |  -  |
|**400** | Invalid executionID or invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Execution not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **uploadLocalizationResults**
> GenericResponseUploadExecutionResultResp uploadLocalizationResults(request)

Upload granularity results for regular algorithms via API instead of file collection

### Example

```typescript
import {
    ExecutionsApi,
    Configuration,
    UploadGranularityResultReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new ExecutionsApi(configuration);

let executionId: number; //Execution ID (default to undefined)
let request: UploadGranularityResultReq; //Granularity results

const { status, data } = await apiInstance.uploadLocalizationResults(
    executionId,
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **UploadGranularityResultReq**| Granularity results | |
| **executionId** | [**number**] | Execution ID | defaults to undefined|


### Return type

**GenericResponseUploadExecutionResultResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Results uploaded successfully |  -  |
|**400** | Invalid exeuction ID or invalid request form or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Execution not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

