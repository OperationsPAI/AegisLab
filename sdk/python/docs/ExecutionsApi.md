# rcabench.openapi.ExecutionsApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**get_execution_by_id**](ExecutionsApi.md#get_execution_by_id) | **GET** /api/v2/executions/{id} | Get execution by ID
[**list_execution_labels**](ExecutionsApi.md#list_execution_labels) | **GET** /api/v2/executions/labels | List execution labels
[**list_executions**](ExecutionsApi.md#list_executions) | **GET** /api/v2/executions | List executions
[**run_algorithm**](ExecutionsApi.md#run_algorithm) | **POST** /api/v2/executions/execute | Submit batch algorithm execution
[**upload_detection_results**](ExecutionsApi.md#upload_detection_results) | **POST** /api/v2/executions/{execution_id}/detector_results | Upload detector results
[**upload_localization_results**](ExecutionsApi.md#upload_localization_results) | **POST** /api/v2/executions/{execution_id}/granularity_results | Upload granularity results


# **get_execution_by_id**
> GenericResponseExecutionDetailResp get_execution_by_id(id)

Get execution by ID

Get detailed information about a specific execution

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.generic_response_execution_detail_resp import GenericResponseExecutionDetailResp
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
    api_instance = rcabench.openapi.ExecutionsApi(api_client)
    id = 56 # int | Execution ID

    try:
        # Get execution by ID
        api_response = api_instance.get_execution_by_id(id)
        print("The response of ExecutionsApi->get_execution_by_id:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ExecutionsApi->get_execution_by_id: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Execution ID | 

### Return type

[**GenericResponseExecutionDetailResp**](GenericResponseExecutionDetailResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Execution retrieved successfully |  -  |
**400** | Invalid execution ID |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Execution not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_execution_labels**
> GenericResponseArrayLabelItem list_execution_labels()

List execution labels

List all available label keys for executions

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.generic_response_array_label_item import GenericResponseArrayLabelItem
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
    api_instance = rcabench.openapi.ExecutionsApi(api_client)

    try:
        # List execution labels
        api_response = api_instance.list_execution_labels()
        print("The response of ExecutionsApi->list_execution_labels:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ExecutionsApi->list_execution_labels: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**GenericResponseArrayLabelItem**](GenericResponseArrayLabelItem.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Available label keys |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_executions**
> GenericResponseListExecutionResp list_executions(page=page, size=size, state=state, status=status, labels=labels)

List executions

Get a paginated list of executions with pagination and filtering

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.execution_state import ExecutionState
from rcabench.openapi.models.generic_response_list_execution_resp import GenericResponseListExecutionResp
from rcabench.openapi.models.status_type import StatusType
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
    api_instance = rcabench.openapi.ExecutionsApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)
    state = rcabench.openapi.ExecutionState() # ExecutionState | Filter by execution state (optional)
    status = rcabench.openapi.StatusType() # StatusType | Filter by status (optional)
    labels = ['labels_example'] # List[str] | Filter by labels (array of key:value strings, e.g., 'type:test') (optional)

    try:
        # List executions
        api_response = api_instance.list_executions(page=page, size=size, state=state, status=status, labels=labels)
        print("The response of ExecutionsApi->list_executions:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ExecutionsApi->list_executions: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]
 **state** | [**ExecutionState**](.md)| Filter by execution state | [optional] 
 **status** | [**StatusType**](.md)| Filter by status | [optional] 
 **labels** | [**List[str]**](str.md)| Filter by labels (array of key:value strings, e.g., &#39;type:test&#39;) | [optional] 

### Return type

[**GenericResponseListExecutionResp**](GenericResponseListExecutionResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Executions retrieved successfully |  -  |
**400** | Invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **run_algorithm**
> GenericResponseSubmitExecutionResp run_algorithm(request)

Submit batch algorithm execution

Submit multiple algorithm execution tasks in batch. Supports mixing datapack (v1 compatible) and dataset (v2 feature) executions.

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.generic_response_submit_execution_resp import GenericResponseSubmitExecutionResp
from rcabench.openapi.models.submit_execution_req import SubmitExecutionReq
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
    api_instance = rcabench.openapi.ExecutionsApi(api_client)
    request = rcabench.openapi.SubmitExecutionReq() # SubmitExecutionReq | Algorithm execution request

    try:
        # Submit batch algorithm execution
        api_response = api_instance.run_algorithm(request)
        print("The response of ExecutionsApi->run_algorithm:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ExecutionsApi->run_algorithm: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**SubmitExecutionReq**](SubmitExecutionReq.md)| Algorithm execution request | 

### Return type

[**GenericResponseSubmitExecutionResp**](GenericResponseSubmitExecutionResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Algorithm execution submitted successfully |  -  |
**400** | Invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Project, algorithm, datapack or dataset not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **upload_detection_results**
> GenericResponseUploadExecutionResultResp upload_detection_results(execution_id, request)

Upload detector results

Upload detection results for detector algorithm via API instead of file collection

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.generic_response_upload_execution_result_resp import GenericResponseUploadExecutionResultResp
from rcabench.openapi.models.upload_detector_result_req import UploadDetectorResultReq
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
    api_instance = rcabench.openapi.ExecutionsApi(api_client)
    execution_id = 56 # int | Execution ID
    request = rcabench.openapi.UploadDetectorResultReq() # UploadDetectorResultReq | Detector results

    try:
        # Upload detector results
        api_response = api_instance.upload_detection_results(execution_id, request)
        print("The response of ExecutionsApi->upload_detection_results:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ExecutionsApi->upload_detection_results: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **execution_id** | **int**| Execution ID | 
 **request** | [**UploadDetectorResultReq**](UploadDetectorResultReq.md)| Detector results | 

### Return type

[**GenericResponseUploadExecutionResultResp**](GenericResponseUploadExecutionResultResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Results uploaded successfully |  -  |
**400** | Invalid executionID or invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Execution not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **upload_localization_results**
> GenericResponseUploadExecutionResultResp upload_localization_results(execution_id, request)

Upload granularity results

Upload granularity results for regular algorithms via API instead of file collection

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.generic_response_upload_execution_result_resp import GenericResponseUploadExecutionResultResp
from rcabench.openapi.models.upload_granularity_result_req import UploadGranularityResultReq
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
    api_instance = rcabench.openapi.ExecutionsApi(api_client)
    execution_id = 56 # int | Execution ID
    request = rcabench.openapi.UploadGranularityResultReq() # UploadGranularityResultReq | Granularity results

    try:
        # Upload granularity results
        api_response = api_instance.upload_localization_results(execution_id, request)
        print("The response of ExecutionsApi->upload_localization_results:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ExecutionsApi->upload_localization_results: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **execution_id** | **int**| Execution ID | 
 **request** | [**UploadGranularityResultReq**](UploadGranularityResultReq.md)| Granularity results | 

### Return type

[**GenericResponseUploadExecutionResultResp**](GenericResponseUploadExecutionResultResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Results uploaded successfully |  -  |
**400** | Invalid exeuction ID or invalid request form or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Execution not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

