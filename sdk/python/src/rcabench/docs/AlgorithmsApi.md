# rcabench.openapi.AlgorithmsApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_algorithms_algorithm_id_executions_execution_id_detectors_post**](AlgorithmsApi.md#api_v2_algorithms_algorithm_id_executions_execution_id_detectors_post) | **POST** /api/v2/algorithms/{algorithm_id}/executions/{execution_id}/detectors | Upload detector algorithm results
[**api_v2_algorithms_algorithm_id_executions_execution_id_results_post**](AlgorithmsApi.md#api_v2_algorithms_algorithm_id_executions_execution_id_results_post) | **POST** /api/v2/algorithms/{algorithm_id}/executions/{execution_id}/results | Upload granularity algorithm results
[**api_v2_algorithms_algorithm_id_results_post**](AlgorithmsApi.md#api_v2_algorithms_algorithm_id_results_post) | **POST** /api/v2/algorithms/{algorithm_id}/results | Upload granularity algorithm results
[**api_v2_algorithms_execute_batch_post**](AlgorithmsApi.md#api_v2_algorithms_execute_batch_post) | **POST** /api/v2/algorithms/execute/batch | Submit batch algorithm execution
[**api_v2_algorithms_execute_post**](AlgorithmsApi.md#api_v2_algorithms_execute_post) | **POST** /api/v2/algorithms/execute | Submit algorithm execution
[**api_v2_algorithms_get**](AlgorithmsApi.md#api_v2_algorithms_get) | **GET** /api/v2/algorithms | List algorithms
[**api_v2_algorithms_search_post**](AlgorithmsApi.md#api_v2_algorithms_search_post) | **POST** /api/v2/algorithms/search | Search algorithms


# **api_v2_algorithms_algorithm_id_executions_execution_id_detectors_post**
> DtoGenericResponseDtoAlgorithmResultUploadResponse api_v2_algorithms_algorithm_id_executions_execution_id_detectors_post(algorithm_id, execution_id, request)

Upload detector algorithm results

Upload detection results for detector algorithms via API instead of file collection

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_detector_result_request import DtoDetectorResultRequest
from rcabench.openapi.models.dto_generic_response_dto_algorithm_result_upload_response import DtoGenericResponseDtoAlgorithmResultUploadResponse
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
    api_instance = rcabench.openapi.AlgorithmsApi(api_client)
    algorithm_id = 56 # int | Algorithm ID
    execution_id = 56 # int | Execution ID
    request = rcabench.openapi.DtoDetectorResultRequest() # DtoDetectorResultRequest | Detector results

    try:
        # Upload detector algorithm results
        api_response = api_instance.api_v2_algorithms_algorithm_id_executions_execution_id_detectors_post(algorithm_id, execution_id, request)
        print("The response of AlgorithmsApi->api_v2_algorithms_algorithm_id_executions_execution_id_detectors_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmsApi->api_v2_algorithms_algorithm_id_executions_execution_id_detectors_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **algorithm_id** | **int**| Algorithm ID | 
 **execution_id** | **int**| Execution ID | 
 **request** | [**DtoDetectorResultRequest**](DtoDetectorResultRequest.md)| Detector results | 

### Return type

[**DtoGenericResponseDtoAlgorithmResultUploadResponse**](DtoGenericResponseDtoAlgorithmResultUploadResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Results uploaded successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**404** | Algorithm or execution not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_algorithms_algorithm_id_executions_execution_id_results_post**
> DtoGenericResponseDtoAlgorithmResultUploadResponse api_v2_algorithms_algorithm_id_executions_execution_id_results_post(algorithm_id, execution_id, request)

Upload granularity algorithm results

Upload granularity results for regular algorithms. Supports two modes: 1) Use existing execution_id, 2) Auto-create execution using algorithm_id and datapack_id

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_algorithm_result_upload_response import DtoGenericResponseDtoAlgorithmResultUploadResponse
from rcabench.openapi.models.dto_granularity_result_enhanced_request import DtoGranularityResultEnhancedRequest
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
    api_instance = rcabench.openapi.AlgorithmsApi(api_client)
    algorithm_id = 56 # int | Algorithm ID
    execution_id = 56 # int | Execution ID (optional - will create new if not provided)
    request = rcabench.openapi.DtoGranularityResultEnhancedRequest() # DtoGranularityResultEnhancedRequest | Granularity results with optional execution creation

    try:
        # Upload granularity algorithm results
        api_response = api_instance.api_v2_algorithms_algorithm_id_executions_execution_id_results_post(algorithm_id, execution_id, request)
        print("The response of AlgorithmsApi->api_v2_algorithms_algorithm_id_executions_execution_id_results_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmsApi->api_v2_algorithms_algorithm_id_executions_execution_id_results_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **algorithm_id** | **int**| Algorithm ID | 
 **execution_id** | **int**| Execution ID (optional - will create new if not provided) | 
 **request** | [**DtoGranularityResultEnhancedRequest**](DtoGranularityResultEnhancedRequest.md)| Granularity results with optional execution creation | 

### Return type

[**DtoGenericResponseDtoAlgorithmResultUploadResponse**](DtoGenericResponseDtoAlgorithmResultUploadResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Results uploaded successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**404** | Algorithm or datapack not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_algorithms_algorithm_id_results_post**
> DtoGenericResponseDtoAlgorithmResultUploadResponse api_v2_algorithms_algorithm_id_results_post(algorithm_id, request)

Upload granularity algorithm results

Upload granularity results for regular algorithms. Supports two modes: 1) Use existing execution_id, 2) Auto-create execution using algorithm_id and datapack_id

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_algorithm_result_upload_response import DtoGenericResponseDtoAlgorithmResultUploadResponse
from rcabench.openapi.models.dto_granularity_result_enhanced_request import DtoGranularityResultEnhancedRequest
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
    api_instance = rcabench.openapi.AlgorithmsApi(api_client)
    algorithm_id = 56 # int | Algorithm ID
    request = rcabench.openapi.DtoGranularityResultEnhancedRequest() # DtoGranularityResultEnhancedRequest | Granularity results with optional execution creation

    try:
        # Upload granularity algorithm results
        api_response = api_instance.api_v2_algorithms_algorithm_id_results_post(algorithm_id, request)
        print("The response of AlgorithmsApi->api_v2_algorithms_algorithm_id_results_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmsApi->api_v2_algorithms_algorithm_id_results_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **algorithm_id** | **int**| Algorithm ID | 
 **request** | [**DtoGranularityResultEnhancedRequest**](DtoGranularityResultEnhancedRequest.md)| Granularity results with optional execution creation | 

### Return type

[**DtoGenericResponseDtoAlgorithmResultUploadResponse**](DtoGenericResponseDtoAlgorithmResultUploadResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Results uploaded successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**404** | Algorithm or datapack not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_algorithms_execute_batch_post**
> DtoGenericResponseDtoBatchAlgorithmExecutionResponse api_v2_algorithms_execute_batch_post(request)

Submit batch algorithm execution

Submit multiple algorithm execution tasks in batch. Supports mixing datapack (v1 compatible) and dataset (v2 feature) executions.

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_batch_algorithm_execution_request import DtoBatchAlgorithmExecutionRequest
from rcabench.openapi.models.dto_generic_response_dto_batch_algorithm_execution_response import DtoGenericResponseDtoBatchAlgorithmExecutionResponse
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
    api_instance = rcabench.openapi.AlgorithmsApi(api_client)
    request = rcabench.openapi.DtoBatchAlgorithmExecutionRequest() # DtoBatchAlgorithmExecutionRequest | Batch algorithm execution request

    try:
        # Submit batch algorithm execution
        api_response = api_instance.api_v2_algorithms_execute_batch_post(request)
        print("The response of AlgorithmsApi->api_v2_algorithms_execute_batch_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmsApi->api_v2_algorithms_execute_batch_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoBatchAlgorithmExecutionRequest**](DtoBatchAlgorithmExecutionRequest.md)| Batch algorithm execution request | 

### Return type

[**DtoGenericResponseDtoBatchAlgorithmExecutionResponse**](DtoGenericResponseDtoBatchAlgorithmExecutionResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**202** | Batch algorithm execution submitted successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**404** | Project, algorithm, datapack or dataset not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_algorithms_execute_post**
> DtoGenericResponseDtoAlgorithmExecutionResponse api_v2_algorithms_execute_post(request)

Submit algorithm execution

Submit algorithm execution task for a single datapack (v1 compatible) or dataset (v2 feature). The system will create execution tasks and return tracking information.

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_algorithm_execution_request import DtoAlgorithmExecutionRequest
from rcabench.openapi.models.dto_generic_response_dto_algorithm_execution_response import DtoGenericResponseDtoAlgorithmExecutionResponse
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
    api_instance = rcabench.openapi.AlgorithmsApi(api_client)
    request = rcabench.openapi.DtoAlgorithmExecutionRequest() # DtoAlgorithmExecutionRequest | Algorithm execution request

    try:
        # Submit algorithm execution
        api_response = api_instance.api_v2_algorithms_execute_post(request)
        print("The response of AlgorithmsApi->api_v2_algorithms_execute_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmsApi->api_v2_algorithms_execute_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoAlgorithmExecutionRequest**](DtoAlgorithmExecutionRequest.md)| Algorithm execution request | 

### Return type

[**DtoGenericResponseDtoAlgorithmExecutionResponse**](DtoGenericResponseDtoAlgorithmExecutionResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**202** | Algorithm execution submitted successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**404** | Project, algorithm, datapack or dataset not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_algorithms_get**
> DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse api_v2_algorithms_get(page=page, size=size)

List algorithms

Get a simple list of all active algorithms without complex filtering

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_algorithm_response import DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse
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
    api_instance = rcabench.openapi.AlgorithmsApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)

    try:
        # List algorithms
        api_response = api_instance.api_v2_algorithms_get(page=page, size=size)
        print("The response of AlgorithmsApi->api_v2_algorithms_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmsApi->api_v2_algorithms_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]

### Return type

[**DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse**](DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Algorithms retrieved successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_algorithms_search_post**
> DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse api_v2_algorithms_search_post(request)

Search algorithms

Search algorithms with complex filtering, sorting and pagination. Algorithms are containers with type 'algorithm'

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_algorithm_search_request import DtoAlgorithmSearchRequest
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_algorithm_response import DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse
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
    api_instance = rcabench.openapi.AlgorithmsApi(api_client)
    request = rcabench.openapi.DtoAlgorithmSearchRequest() # DtoAlgorithmSearchRequest | Algorithm search request

    try:
        # Search algorithms
        api_response = api_instance.api_v2_algorithms_search_post(request)
        print("The response of AlgorithmsApi->api_v2_algorithms_search_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmsApi->api_v2_algorithms_search_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoAlgorithmSearchRequest**](DtoAlgorithmSearchRequest.md)| Algorithm search request | 

### Return type

[**DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse**](DtoGenericResponseDtoSearchResponseDtoAlgorithmResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Algorithms retrieved successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

