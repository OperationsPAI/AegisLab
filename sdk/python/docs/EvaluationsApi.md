# rcabench.openapi.EvaluationsApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**evaluate_algorithm_on_datapacks**](EvaluationsApi.md#evaluate_algorithm_on_datapacks) | **POST** /api/v2/evaluations/datapacks | List Datapack Evaluation Results
[**evaluate_algorithm_on_datasets**](EvaluationsApi.md#evaluate_algorithm_on_datasets) | **POST** /api/v2/evaluations/datasets | List Dataset Evaluation Results


# **evaluate_algorithm_on_datapacks**
> GenericResponseBatchEvaluateDatapackResp evaluate_algorithm_on_datapacks(request)

List Datapack Evaluation Results

Retrieve evaluation data for multiple algorithm-datapack pairs.

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.batch_evaluate_datapack_req import BatchEvaluateDatapackReq
from rcabench.openapi.models.generic_response_batch_evaluate_datapack_resp import GenericResponseBatchEvaluateDatapackResp
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure API key authorization: BearerAuth
configuration.api_key['BearerAuth'] = os.environ["API_KEY"]

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['BearerAuth'] = 'Bearer'

# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.EvaluationsApi(api_client)
    request = rcabench.openapi.BatchEvaluateDatapackReq() # BatchEvaluateDatapackReq | Batch evaluation request containing multiple algorithm-datapack pairs

    try:
        # List Datapack Evaluation Results
        api_response = api_instance.evaluate_algorithm_on_datapacks(request)
        print("The response of EvaluationsApi->evaluate_algorithm_on_datapacks:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EvaluationsApi->evaluate_algorithm_on_datapacks: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**BatchEvaluateDatapackReq**](BatchEvaluateDatapackReq.md)| Batch evaluation request containing multiple algorithm-datapack pairs | 

### Return type

[**GenericResponseBatchEvaluateDatapackResp**](GenericResponseBatchEvaluateDatapackResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Batch algorithm datapack evaluation data retrieved successfully |  -  |
**400** | Invalid request format/parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **evaluate_algorithm_on_datasets**
> GenericResponseBatchEvaluateDatasetResp evaluate_algorithm_on_datasets(request)

List Dataset Evaluation Results

Retrieve evaluation data for multiple algorithm-dataset pairs.

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.batch_evaluate_datapack_req import BatchEvaluateDatapackReq
from rcabench.openapi.models.generic_response_batch_evaluate_dataset_resp import GenericResponseBatchEvaluateDatasetResp
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure API key authorization: BearerAuth
configuration.api_key['BearerAuth'] = os.environ["API_KEY"]

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['BearerAuth'] = 'Bearer'

# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.EvaluationsApi(api_client)
    request = rcabench.openapi.BatchEvaluateDatapackReq() # BatchEvaluateDatapackReq | Batch evaluation request containing multiple algorithm-dataset pairs

    try:
        # List Dataset Evaluation Results
        api_response = api_instance.evaluate_algorithm_on_datasets(request)
        print("The response of EvaluationsApi->evaluate_algorithm_on_datasets:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EvaluationsApi->evaluate_algorithm_on_datasets: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**BatchEvaluateDatapackReq**](BatchEvaluateDatapackReq.md)| Batch evaluation request containing multiple algorithm-dataset pairs | 

### Return type

[**GenericResponseBatchEvaluateDatasetResp**](GenericResponseBatchEvaluateDatasetResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Batch algorithm dataset evaluation data retrieved successfully |  -  |
**400** | Invalid request format/parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

