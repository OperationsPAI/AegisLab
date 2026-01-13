# EvaluationsApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**evaluateAlgorithmOnDatapacks**](#evaluatealgorithmondatapacks) | **POST** /api/v2/evaluations/datapacks | List Datapack Evaluation Results|
|[**evaluateAlgorithmOnDatasets**](#evaluatealgorithmondatasets) | **POST** /api/v2/evaluations/datasets | List Dataset Evaluation Results|

# **evaluateAlgorithmOnDatapacks**
> GenericResponseBatchEvaluateDatapackResp evaluateAlgorithmOnDatapacks(request)

Retrieve evaluation data for multiple algorithm-datapack pairs.

### Example

```typescript
import {
    EvaluationsApi,
    Configuration,
    BatchEvaluateDatapackReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new EvaluationsApi(configuration);

let request: BatchEvaluateDatapackReq; //Batch evaluation request containing multiple algorithm-datapack pairs

const { status, data } = await apiInstance.evaluateAlgorithmOnDatapacks(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **BatchEvaluateDatapackReq**| Batch evaluation request containing multiple algorithm-datapack pairs | |


### Return type

**GenericResponseBatchEvaluateDatapackResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Batch algorithm datapack evaluation data retrieved successfully |  -  |
|**400** | Invalid request format/parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **evaluateAlgorithmOnDatasets**
> GenericResponseBatchEvaluateDatasetResp evaluateAlgorithmOnDatasets(request)

Retrieve evaluation data for multiple algorithm-dataset pairs.

### Example

```typescript
import {
    EvaluationsApi,
    Configuration,
    BatchEvaluateDatasetReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new EvaluationsApi(configuration);

let request: BatchEvaluateDatasetReq; //Batch evaluation request containing multiple algorithm-dataset pairs

const { status, data } = await apiInstance.evaluateAlgorithmOnDatasets(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **BatchEvaluateDatasetReq**| Batch evaluation request containing multiple algorithm-dataset pairs | |


### Return type

**GenericResponseBatchEvaluateDatasetResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Batch algorithm dataset evaluation data retrieved successfully |  -  |
|**400** | Invalid request format/parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

