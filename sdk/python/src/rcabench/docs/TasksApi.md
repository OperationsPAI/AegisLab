# rcabench.openapi.TasksApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_tasks_get**](TasksApi.md#api_v2_tasks_get) | **GET** /api/v2/tasks | List tasks
[**api_v2_tasks_id_get**](TasksApi.md#api_v2_tasks_id_get) | **GET** /api/v2/tasks/{id} | Get task by ID
[**api_v2_tasks_queue_post**](TasksApi.md#api_v2_tasks_queue_post) | **POST** /api/v2/tasks/queue | Get queued tasks
[**api_v2_tasks_search_post**](TasksApi.md#api_v2_tasks_search_post) | **POST** /api/v2/tasks/search | Search tasks


# **api_v2_tasks_get**
> DtoGenericResponseDtoSearchResponseDtoTaskResponse api_v2_tasks_get(page=page, size=size, task_id=task_id, trace_id=trace_id, group_id=group_id, task_type=task_type, status=status, immediate=immediate)

List tasks

Get a simple list of tasks with basic filtering via query parameters

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_task_response import DtoGenericResponseDtoSearchResponseDtoTaskResponse
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
    api_instance = rcabench.openapi.TasksApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)
    task_id = 'task_id_example' # str | Filter by task ID (optional)
    trace_id = 'trace_id_example' # str | Filter by trace ID (optional)
    group_id = 'group_id_example' # str | Filter by group ID (optional)
    task_type = 'task_type_example' # str | Filter by task type (optional)
    status = 'status_example' # str | Filter by status (optional)
    immediate = True # bool | Filter by immediate execution (optional)

    try:
        # List tasks
        api_response = api_instance.api_v2_tasks_get(page=page, size=size, task_id=task_id, trace_id=trace_id, group_id=group_id, task_type=task_type, status=status, immediate=immediate)
        print("The response of TasksApi->api_v2_tasks_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TasksApi->api_v2_tasks_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]
 **task_id** | **str**| Filter by task ID | [optional] 
 **trace_id** | **str**| Filter by trace ID | [optional] 
 **group_id** | **str**| Filter by group ID | [optional] 
 **task_type** | **str**| Filter by task type | [optional] 
 **status** | **str**| Filter by status | [optional] 
 **immediate** | **bool**| Filter by immediate execution | [optional] 

### Return type

[**DtoGenericResponseDtoSearchResponseDtoTaskResponse**](DtoGenericResponseDtoSearchResponseDtoTaskResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Tasks retrieved successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_tasks_id_get**
> DtoGenericResponseDtoTaskDetailResponse api_v2_tasks_id_get(id, include=include)

Get task by ID

Get detailed information about a specific task including logs

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_task_detail_response import DtoGenericResponseDtoTaskDetailResponse
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
    api_instance = rcabench.openapi.TasksApi(api_client)
    id = 'id_example' # str | Task ID
    include = ['include_example'] # List[str] | Include additional data (logs) (optional)

    try:
        # Get task by ID
        api_response = api_instance.api_v2_tasks_id_get(id, include=include)
        print("The response of TasksApi->api_v2_tasks_id_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TasksApi->api_v2_tasks_id_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **str**| Task ID | 
 **include** | [**List[str]**](str.md)| Include additional data (logs) | [optional] 

### Return type

[**DtoGenericResponseDtoTaskDetailResponse**](DtoGenericResponseDtoTaskDetailResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Task retrieved successfully |  -  |
**400** | Invalid task ID |  -  |
**403** | Permission denied |  -  |
**404** | Task not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_tasks_queue_post**
> DtoGenericResponseDtoSearchResponseDtoTaskResponse api_v2_tasks_queue_post(request)

Get queued tasks

Get tasks in queue (ready and delayed) with pagination and filtering

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_advanced_search_request import DtoAdvancedSearchRequest
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_task_response import DtoGenericResponseDtoSearchResponseDtoTaskResponse
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
    api_instance = rcabench.openapi.TasksApi(api_client)
    request = rcabench.openapi.DtoAdvancedSearchRequest() # DtoAdvancedSearchRequest | Search request with pagination

    try:
        # Get queued tasks
        api_response = api_instance.api_v2_tasks_queue_post(request)
        print("The response of TasksApi->api_v2_tasks_queue_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TasksApi->api_v2_tasks_queue_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoAdvancedSearchRequest**](DtoAdvancedSearchRequest.md)| Search request with pagination | 

### Return type

[**DtoGenericResponseDtoSearchResponseDtoTaskResponse**](DtoGenericResponseDtoSearchResponseDtoTaskResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Queued tasks retrieved successfully |  -  |
**400** | Invalid request |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_tasks_search_post**
> DtoGenericResponseDtoSearchResponseDtoTaskResponse api_v2_tasks_search_post(request)

Search tasks

Search tasks with complex filtering, sorting and pagination

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_task_response import DtoGenericResponseDtoSearchResponseDtoTaskResponse
from rcabench.openapi.models.dto_task_search_request import DtoTaskSearchRequest
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
    api_instance = rcabench.openapi.TasksApi(api_client)
    request = rcabench.openapi.DtoTaskSearchRequest() # DtoTaskSearchRequest | Task search request

    try:
        # Search tasks
        api_response = api_instance.api_v2_tasks_search_post(request)
        print("The response of TasksApi->api_v2_tasks_search_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling TasksApi->api_v2_tasks_search_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoTaskSearchRequest**](DtoTaskSearchRequest.md)| Task search request | 

### Return type

[**DtoGenericResponseDtoSearchResponseDtoTaskResponse**](DtoGenericResponseDtoSearchResponseDtoTaskResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Tasks retrieved successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

