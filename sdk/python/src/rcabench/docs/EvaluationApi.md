# rcabench.openapi.EvaluationApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_evaluations_executions_get**](EvaluationApi.md#api_v1_evaluations_executions_get) | **GET** /api/v1/evaluations/executions | Get successful algorithm execution records
[**api_v1_evaluations_groundtruth_post**](EvaluationApi.md#api_v1_evaluations_groundtruth_post) | **POST** /api/v1/evaluations/groundtruth | Get ground truth for datasets
[**api_v1_evaluations_raw_data_post**](EvaluationApi.md#api_v1_evaluations_raw_data_post) | **POST** /api/v1/evaluations/raw-data | Get raw evaluation data


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

