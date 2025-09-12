# rcabench.openapi.EvaluationApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_evaluations_executions_get**](EvaluationApi.md#api_v1_evaluations_executions_get) | **GET** /api/v1/evaluations/executions | Get successful algorithm execution records
[**api_v1_evaluations_groundtruth_post**](EvaluationApi.md#api_v1_evaluations_groundtruth_post) | **POST** /api/v1/evaluations/groundtruth | Get ground truth for datasets
[**api_v1_evaluations_raw_data_post**](EvaluationApi.md#api_v1_evaluations_raw_data_post) | **POST** /api/v1/evaluations/raw-data | Get raw evaluation data
[**api_v2_evaluations_datapacks_detector_post**](EvaluationApi.md#api_v2_evaluations_datapacks_detector_post) | **POST** /api/v2/evaluations/datapacks/detector | Get Datapack Detector Results
[**api_v2_evaluations_datapacks_post**](EvaluationApi.md#api_v2_evaluations_datapacks_post) | **POST** /api/v2/evaluations/datapacks | Get Batch Algorithm Datapack Evaluation
[**api_v2_evaluations_datasets_post**](EvaluationApi.md#api_v2_evaluations_datasets_post) | **POST** /api/v2/evaluations/datasets | Get Batch Algorithm Dataset Evaluation
[**api_v2_evaluations_label_keys_get**](EvaluationApi.md#api_v2_evaluations_label_keys_get) | **GET** /api/v2/evaluations/label-keys | Get Available Label Keys


# **api_v1_evaluations_executions_get**
> DtoGenericResponseDtoSuccessfulExecutionsResp api_v1_evaluations_executions_get(start_time=start_time, end_time=end_time, limit=limit, offset=offset)

Get successful algorithm execution records

Get all records in ExecutionResult with status ExecutionSuccess, supports time range filtering and quantity filtering

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_successful_executions_resp import DtoGenericResponseDtoSuccessfulExecutionsResp
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8080/api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8080/api/v1"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.EvaluationApi(api_client)
    start_time = 'start_time_example' # str | Start time, format: 2006-01-02T15:04:05Z07:00 (optional)
    end_time = 'end_time_example' # str | End time, format: 2006-01-02T15:04:05Z07:00 (optional)
    limit = 56 # int | Limit (optional)
    offset = 56 # int | Offset for pagination (optional)

    try:
        # Get successful algorithm execution records
        api_response = api_instance.api_v1_evaluations_executions_get(start_time=start_time, end_time=end_time, limit=limit, offset=offset)
        print("The response of EvaluationApi->api_v1_evaluations_executions_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EvaluationApi->api_v1_evaluations_executions_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **start_time** | **str**| Start time, format: 2006-01-02T15:04:05Z07:00 | [optional] 
 **end_time** | **str**| End time, format: 2006-01-02T15:04:05Z07:00 | [optional] 
 **limit** | **int**| Limit | [optional] 
 **offset** | **int**| Offset for pagination | [optional] 

### Return type

[**DtoGenericResponseDtoSuccessfulExecutionsResp**](DtoGenericResponseDtoSuccessfulExecutionsResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Successfully returns the list of successful algorithm execution records |  -  |
**400** | Request parameter error |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_evaluations_groundtruth_post**
> DtoGenericResponseDtoGroundTruthResp api_v1_evaluations_groundtruth_post(body)

Get ground truth for datasets

Get ground truth data for the given dataset array, used as benchmark data for algorithm evaluation. Supports batch query for ground truth information of multiple datasets

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_ground_truth_resp import DtoGenericResponseDtoGroundTruthResp
from rcabench.openapi.models.dto_ground_truth_req import DtoGroundTruthReq
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8080/api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8080/api/v1"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.EvaluationApi(api_client)
    body = rcabench.openapi.DtoGroundTruthReq() # DtoGroundTruthReq | Ground truth query request, contains dataset list

    try:
        # Get ground truth for datasets
        api_response = api_instance.api_v1_evaluations_groundtruth_post(body)
        print("The response of EvaluationApi->api_v1_evaluations_groundtruth_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EvaluationApi->api_v1_evaluations_groundtruth_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**DtoGroundTruthReq**](DtoGroundTruthReq.md)| Ground truth query request, contains dataset list | 

### Return type

[**DtoGenericResponseDtoGroundTruthResp**](DtoGenericResponseDtoGroundTruthResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Successfully returns ground truth information for datasets |  -  |
**400** | Request parameter error, such as incorrect JSON format or empty dataset array |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_evaluations_raw_data_post**
> DtoGenericResponseDtoRawDataResp api_v1_evaluations_raw_data_post(body)

Get raw evaluation data

Supports three query modes: 1) Directly pass an array of algorithm-dataset pairs for precise query; 2) Pass lists of algorithms and datasets for Cartesian product query; 3) Query by execution ID list. The three modes are mutually exclusive, only one can be selected

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_raw_data_resp import DtoGenericResponseDtoRawDataResp
from rcabench.openapi.models.dto_raw_data_req import DtoRawDataReq
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8080/api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8080/api/v1"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.EvaluationApi(api_client)
    body = rcabench.openapi.DtoRawDataReq() # DtoRawDataReq | Raw data query request, supports three modes: pairs array, (algorithms+datasets) Cartesian product, or execution_ids list

    try:
        # Get raw evaluation data
        api_response = api_instance.api_v1_evaluations_raw_data_post(body)
        print("The response of EvaluationApi->api_v1_evaluations_raw_data_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EvaluationApi->api_v1_evaluations_raw_data_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**DtoRawDataReq**](DtoRawDataReq.md)| Raw data query request, supports three modes: pairs array, (algorithms+datasets) Cartesian product, or execution_ids list | 

### Return type

[**DtoGenericResponseDtoRawDataResp**](DtoGenericResponseDtoRawDataResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Successfully returns the list of raw evaluation data |  -  |
**400** | Request parameter error, such as incorrect JSON format, query mode conflict or empty parameter |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_evaluations_datapacks_detector_post**
> DtoGenericResponseDtoDatapackDetectorResp api_v2_evaluations_datapacks_detector_post(request)

Get Datapack Detector Results

Get detector analysis results for multiple datapacks. If a datapack has multiple executions, returns the latest one.

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_datapack_detector_req import DtoDatapackDetectorReq
from rcabench.openapi.models.dto_generic_response_dto_datapack_detector_resp import DtoGenericResponseDtoDatapackDetectorResp
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8080/api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8080/api/v1"
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
    api_instance = rcabench.openapi.EvaluationApi(api_client)
    request = rcabench.openapi.DtoDatapackDetectorReq() # DtoDatapackDetectorReq | Datapack detector request

    try:
        # Get Datapack Detector Results
        api_response = api_instance.api_v2_evaluations_datapacks_detector_post(request)
        print("The response of EvaluationApi->api_v2_evaluations_datapacks_detector_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EvaluationApi->api_v2_evaluations_datapacks_detector_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoDatapackDetectorReq**](DtoDatapackDetectorReq.md)| Datapack detector request | 

### Return type

[**DtoGenericResponseDtoDatapackDetectorResp**](DtoGenericResponseDtoDatapackDetectorResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Datapack detector results |  -  |
**400** | Bad request |  -  |
**401** | Unauthorized |  -  |
**403** | Forbidden |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_evaluations_datapacks_post**
> DtoGenericResponseDtoDatapackEvaluationBatchResp api_v2_evaluations_datapacks_post(request)

Get Batch Algorithm Datapack Evaluation

Get execution results with predictions and ground truth for multiple algorithm-datapack pairs in a single request. Returns the latest execution for each pair if multiple executions exist.

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_datapack_evaluation_batch_req import DtoDatapackEvaluationBatchReq
from rcabench.openapi.models.dto_generic_response_dto_datapack_evaluation_batch_resp import DtoGenericResponseDtoDatapackEvaluationBatchResp
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8080/api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8080/api/v1"
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
    api_instance = rcabench.openapi.EvaluationApi(api_client)
    request = rcabench.openapi.DtoDatapackEvaluationBatchReq() # DtoDatapackEvaluationBatchReq | Batch evaluation request containing multiple algorithm-datapack pairs

    try:
        # Get Batch Algorithm Datapack Evaluation
        api_response = api_instance.api_v2_evaluations_datapacks_post(request)
        print("The response of EvaluationApi->api_v2_evaluations_datapacks_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EvaluationApi->api_v2_evaluations_datapacks_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoDatapackEvaluationBatchReq**](DtoDatapackEvaluationBatchReq.md)| Batch evaluation request containing multiple algorithm-datapack pairs | 

### Return type

[**DtoGenericResponseDtoDatapackEvaluationBatchResp**](DtoGenericResponseDtoDatapackEvaluationBatchResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Batch algorithm datapack evaluation data |  -  |
**400** | Bad request |  -  |
**401** | Unauthorized |  -  |
**403** | Forbidden |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_evaluations_datasets_post**
> DtoGenericResponseDtoDatasetEvaluationBatchResp api_v2_evaluations_datasets_post(request)

Get Batch Algorithm Dataset Evaluation

Get execution results with predictions and ground truth for multiple algorithm-dataset pairs in a single request

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_dataset_evaluation_batch_req import DtoDatasetEvaluationBatchReq
from rcabench.openapi.models.dto_generic_response_dto_dataset_evaluation_batch_resp import DtoGenericResponseDtoDatasetEvaluationBatchResp
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8080/api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8080/api/v1"
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
    api_instance = rcabench.openapi.EvaluationApi(api_client)
    request = rcabench.openapi.DtoDatasetEvaluationBatchReq() # DtoDatasetEvaluationBatchReq | Batch evaluation request containing multiple algorithm-dataset pairs

    try:
        # Get Batch Algorithm Dataset Evaluation
        api_response = api_instance.api_v2_evaluations_datasets_post(request)
        print("The response of EvaluationApi->api_v2_evaluations_datasets_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EvaluationApi->api_v2_evaluations_datasets_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoDatasetEvaluationBatchReq**](DtoDatasetEvaluationBatchReq.md)| Batch evaluation request containing multiple algorithm-dataset pairs | 

### Return type

[**DtoGenericResponseDtoDatasetEvaluationBatchResp**](DtoGenericResponseDtoDatasetEvaluationBatchResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Batch algorithm dataset evaluation data |  -  |
**400** | Bad request |  -  |
**401** | Unauthorized |  -  |
**403** | Forbidden |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_evaluations_label_keys_get**
> DtoGenericResponseArrayString api_v2_evaluations_label_keys_get()

Get Available Label Keys

Get the list of available label keys that can be used for filtering execution results

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_array_string import DtoGenericResponseArrayString
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8080/api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8080/api/v1"
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
    api_instance = rcabench.openapi.EvaluationApi(api_client)

    try:
        # Get Available Label Keys
        api_response = api_instance.api_v2_evaluations_label_keys_get()
        print("The response of EvaluationApi->api_v2_evaluations_label_keys_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EvaluationApi->api_v2_evaluations_label_keys_get: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**DtoGenericResponseArrayString**](DtoGenericResponseArrayString.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Available label keys |  -  |
**401** | Unauthorized |  -  |
**403** | Forbidden |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

