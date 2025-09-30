# rcabench.openapi.InjectionsApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_injections_batch_delete_post**](InjectionsApi.md#api_v2_injections_batch_delete_post) | **POST** /api/v2/injections/batch-delete | Batch delete injections
[**api_v2_injections_get**](InjectionsApi.md#api_v2_injections_get) | **GET** /api/v2/injections | List injections
[**api_v2_injections_id_delete**](InjectionsApi.md#api_v2_injections_id_delete) | **DELETE** /api/v2/injections/{id} | Delete injection
[**api_v2_injections_id_get**](InjectionsApi.md#api_v2_injections_id_get) | **GET** /api/v2/injections/{id} | Get injection by ID
[**api_v2_injections_id_put**](InjectionsApi.md#api_v2_injections_id_put) | **PUT** /api/v2/injections/{id} | Update injection
[**api_v2_injections_name_labels_patch**](InjectionsApi.md#api_v2_injections_name_labels_patch) | **PATCH** /api/v2/injections/{name}/labels | Manage injection custom labels
[**api_v2_injections_name_tags_patch**](InjectionsApi.md#api_v2_injections_name_tags_patch) | **PATCH** /api/v2/injections/{name}/tags | Manage injection tags
[**api_v2_injections_post**](InjectionsApi.md#api_v2_injections_post) | **POST** /api/v2/injections | Create injections
[**api_v2_injections_search_post**](InjectionsApi.md#api_v2_injections_search_post) | **POST** /api/v2/injections/search | Search injections


# **api_v2_injections_batch_delete_post**
> DtoGenericResponseDtoInjectionV2BatchDeleteResponse api_v2_injections_batch_delete_post(batch_delete)

Batch delete injections

Batch delete injections by IDs or labels with cascading deletion of related records

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_injection_v2_batch_delete_response import DtoGenericResponseDtoInjectionV2BatchDeleteResponse
from rcabench.openapi.models.dto_injection_v2_batch_delete_req import DtoInjectionV2BatchDeleteReq
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
    api_instance = rcabench.openapi.InjectionsApi(api_client)
    batch_delete = rcabench.openapi.DtoInjectionV2BatchDeleteReq() # DtoInjectionV2BatchDeleteReq | Batch delete request

    try:
        # Batch delete injections
        api_response = api_instance.api_v2_injections_batch_delete_post(batch_delete)
        print("The response of InjectionsApi->api_v2_injections_batch_delete_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->api_v2_injections_batch_delete_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **batch_delete** | [**DtoInjectionV2BatchDeleteReq**](DtoInjectionV2BatchDeleteReq.md)| Batch delete request | 

### Return type

[**DtoGenericResponseDtoInjectionV2BatchDeleteResponse**](DtoGenericResponseDtoInjectionV2BatchDeleteResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Injections deleted successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_injections_get**
> DtoGenericResponseDtoInjectionSearchResponse api_v2_injections_get(page=page, size=size, task_id=task_id, fault_type=fault_type, status=status, benchmark=benchmark, search=search, tags=tags, sort_by=sort_by, sort_order=sort_order, include=include)

List injections

Get a paginated list of injections with filtering and sorting

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_injection_search_response import DtoGenericResponseDtoInjectionSearchResponse
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
    api_instance = rcabench.openapi.InjectionsApi(api_client)
    page = 56 # int | Page number (default 1) (optional)
    size = 56 # int | Page size (default 20, max 100) (optional)
    task_id = 'task_id_example' # str | Filter by task ID (optional)
    fault_type = 56 # int | Filter by fault type (optional)
    status = 56 # int | Filter by status (optional)
    benchmark = 'benchmark_example' # str | Filter by benchmark (optional)
    search = 'search_example' # str | Search in injection name and description (optional)
    tags = ['tags_example'] # List[str] | Filter by tags (array of tag values) (optional)
    sort_by = 'sort_by_example' # str | Sort field (id,task_id,fault_type,status,benchmark,injection_name,created_at,updated_at) (optional)
    sort_order = 'sort_order_example' # str | Sort order (asc,desc) (optional)
    include = 'include_example' # str | Include related data (task) (optional)

    try:
        # List injections
        api_response = api_instance.api_v2_injections_get(page=page, size=size, task_id=task_id, fault_type=fault_type, status=status, benchmark=benchmark, search=search, tags=tags, sort_by=sort_by, sort_order=sort_order, include=include)
        print("The response of InjectionsApi->api_v2_injections_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->api_v2_injections_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number (default 1) | [optional] 
 **size** | **int**| Page size (default 20, max 100) | [optional] 
 **task_id** | **str**| Filter by task ID | [optional] 
 **fault_type** | **int**| Filter by fault type | [optional] 
 **status** | **int**| Filter by status | [optional] 
 **benchmark** | **str**| Filter by benchmark | [optional] 
 **search** | **str**| Search in injection name and description | [optional] 
 **tags** | [**List[str]**](str.md)| Filter by tags (array of tag values) | [optional] 
 **sort_by** | **str**| Sort field (id,task_id,fault_type,status,benchmark,injection_name,created_at,updated_at) | [optional] 
 **sort_order** | **str**| Sort order (asc,desc) | [optional] 
 **include** | **str**| Include related data (task) | [optional] 

### Return type

[**DtoGenericResponseDtoInjectionSearchResponse**](DtoGenericResponseDtoInjectionSearchResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Injections retrieved successfully |  -  |
**400** | Invalid request parameters |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_injections_id_delete**
> DtoGenericResponseAny api_v2_injections_id_delete(id)

Delete injection

Soft delete an injection (sets status to -1)

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny
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
    api_instance = rcabench.openapi.InjectionsApi(api_client)
    id = 56 # int | Injection ID

    try:
        # Delete injection
        api_response = api_instance.api_v2_injections_id_delete(id)
        print("The response of InjectionsApi->api_v2_injections_id_delete:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->api_v2_injections_id_delete: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Injection ID | 

### Return type

[**DtoGenericResponseAny**](DtoGenericResponseAny.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Injection deleted successfully |  -  |
**400** | Invalid injection ID |  -  |
**403** | Permission denied |  -  |
**404** | Injection not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_injections_id_get**
> DtoGenericResponseDtoInjectionV2Response api_v2_injections_id_get(id, include=include)

Get injection by ID

Get detailed information about a specific injection

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_injection_v2_response import DtoGenericResponseDtoInjectionV2Response
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
    api_instance = rcabench.openapi.InjectionsApi(api_client)
    id = 56 # int | Injection ID
    include = 'include_example' # str | Include related data (task) (optional)

    try:
        # Get injection by ID
        api_response = api_instance.api_v2_injections_id_get(id, include=include)
        print("The response of InjectionsApi->api_v2_injections_id_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->api_v2_injections_id_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Injection ID | 
 **include** | **str**| Include related data (task) | [optional] 

### Return type

[**DtoGenericResponseDtoInjectionV2Response**](DtoGenericResponseDtoInjectionV2Response.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Injection retrieved successfully |  -  |
**400** | Invalid injection ID |  -  |
**403** | Permission denied |  -  |
**404** | Injection not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_injections_id_put**
> DtoGenericResponseDtoInjectionV2Response api_v2_injections_id_put(id, injection)

Update injection

Update injection information

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_injection_v2_response import DtoGenericResponseDtoInjectionV2Response
from rcabench.openapi.models.dto_injection_v2_update_req import DtoInjectionV2UpdateReq
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
    api_instance = rcabench.openapi.InjectionsApi(api_client)
    id = 56 # int | Injection ID
    injection = rcabench.openapi.DtoInjectionV2UpdateReq() # DtoInjectionV2UpdateReq | Injection update request

    try:
        # Update injection
        api_response = api_instance.api_v2_injections_id_put(id, injection)
        print("The response of InjectionsApi->api_v2_injections_id_put:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->api_v2_injections_id_put: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Injection ID | 
 **injection** | [**DtoInjectionV2UpdateReq**](DtoInjectionV2UpdateReq.md)| Injection update request | 

### Return type

[**DtoGenericResponseDtoInjectionV2Response**](DtoGenericResponseDtoInjectionV2Response.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Injection updated successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**404** | Injection not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_injections_name_labels_patch**
> DtoGenericResponseDtoInjectionV2Response api_v2_injections_name_labels_patch(name, manage)

Manage injection custom labels

Add or remove custom labels (key-value pairs) for an injection

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_injection_v2_response import DtoGenericResponseDtoInjectionV2Response
from rcabench.openapi.models.dto_injection_v2_custom_label_manage_req import DtoInjectionV2CustomLabelManageReq
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
    api_instance = rcabench.openapi.InjectionsApi(api_client)
    name = 'name_example' # str | Injection Name
    manage = rcabench.openapi.DtoInjectionV2CustomLabelManageReq() # DtoInjectionV2CustomLabelManageReq | Custom label management request

    try:
        # Manage injection custom labels
        api_response = api_instance.api_v2_injections_name_labels_patch(name, manage)
        print("The response of InjectionsApi->api_v2_injections_name_labels_patch:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->api_v2_injections_name_labels_patch: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **str**| Injection Name | 
 **manage** | [**DtoInjectionV2CustomLabelManageReq**](DtoInjectionV2CustomLabelManageReq.md)| Custom label management request | 

### Return type

[**DtoGenericResponseDtoInjectionV2Response**](DtoGenericResponseDtoInjectionV2Response.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Custom labels managed successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**404** | Injection not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_injections_name_tags_patch**
> DtoGenericResponseDtoInjectionV2Response api_v2_injections_name_tags_patch(name, manage)

Manage injection tags

Add or remove tags for an injection

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_injection_v2_response import DtoGenericResponseDtoInjectionV2Response
from rcabench.openapi.models.dto_injection_v2_label_manage_req import DtoInjectionV2LabelManageReq
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
    api_instance = rcabench.openapi.InjectionsApi(api_client)
    name = 'name_example' # str | Injection Name
    manage = rcabench.openapi.DtoInjectionV2LabelManageReq() # DtoInjectionV2LabelManageReq | Tag management request

    try:
        # Manage injection tags
        api_response = api_instance.api_v2_injections_name_tags_patch(name, manage)
        print("The response of InjectionsApi->api_v2_injections_name_tags_patch:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->api_v2_injections_name_tags_patch: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **str**| Injection Name | 
 **manage** | [**DtoInjectionV2LabelManageReq**](DtoInjectionV2LabelManageReq.md)| Tag management request | 

### Return type

[**DtoGenericResponseDtoInjectionV2Response**](DtoGenericResponseDtoInjectionV2Response.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Tags managed successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**404** | Injection not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_injections_post**
> DtoGenericResponseDtoInjectionV2CreateResponse api_v2_injections_post(injections)

Create injections

Create one or multiple injection records with automatic labeling based on task_id

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_injection_v2_create_response import DtoGenericResponseDtoInjectionV2CreateResponse
from rcabench.openapi.models.dto_injection_v2_create_req import DtoInjectionV2CreateReq
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
    api_instance = rcabench.openapi.InjectionsApi(api_client)
    injections = rcabench.openapi.DtoInjectionV2CreateReq() # DtoInjectionV2CreateReq | Injection creation request

    try:
        # Create injections
        api_response = api_instance.api_v2_injections_post(injections)
        print("The response of InjectionsApi->api_v2_injections_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->api_v2_injections_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **injections** | [**DtoInjectionV2CreateReq**](DtoInjectionV2CreateReq.md)| Injection creation request | 

### Return type

[**DtoGenericResponseDtoInjectionV2CreateResponse**](DtoGenericResponseDtoInjectionV2CreateResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Injections created successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_injections_search_post**
> DtoGenericResponseDtoSearchResponseDtoInjectionV2Response api_v2_injections_search_post(search)

Search injections

Advanced search for injections with complex filtering including custom labels

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_injection_v2_response import DtoGenericResponseDtoSearchResponseDtoInjectionV2Response
from rcabench.openapi.models.dto_injection_v2_search_req import DtoInjectionV2SearchReq
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
    api_instance = rcabench.openapi.InjectionsApi(api_client)
    search = rcabench.openapi.DtoInjectionV2SearchReq() # DtoInjectionV2SearchReq | Search criteria

    try:
        # Search injections
        api_response = api_instance.api_v2_injections_search_post(search)
        print("The response of InjectionsApi->api_v2_injections_search_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->api_v2_injections_search_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **search** | [**DtoInjectionV2SearchReq**](DtoInjectionV2SearchReq.md)| Search criteria | 

### Return type

[**DtoGenericResponseDtoSearchResponseDtoInjectionV2Response**](DtoGenericResponseDtoSearchResponseDtoInjectionV2Response.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Search results |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

