# rcabench.openapi.DatasetsApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_datasets_get**](DatasetsApi.md#api_v2_datasets_get) | **GET** /api/v2/datasets | List datasets
[**api_v2_datasets_id_delete**](DatasetsApi.md#api_v2_datasets_id_delete) | **DELETE** /api/v2/datasets/{id} | Delete dataset
[**api_v2_datasets_id_download_get**](DatasetsApi.md#api_v2_datasets_id_download_get) | **GET** /api/v2/datasets/{id}/download | Download dataset
[**api_v2_datasets_id_get**](DatasetsApi.md#api_v2_datasets_id_get) | **GET** /api/v2/datasets/{id} | Get dataset by ID
[**api_v2_datasets_id_injections_patch**](DatasetsApi.md#api_v2_datasets_id_injections_patch) | **PATCH** /api/v2/datasets/{id}/injections | Manage dataset injections
[**api_v2_datasets_id_labels_patch**](DatasetsApi.md#api_v2_datasets_id_labels_patch) | **PATCH** /api/v2/datasets/{id}/labels | Manage dataset labels
[**api_v2_datasets_id_put**](DatasetsApi.md#api_v2_datasets_id_put) | **PUT** /api/v2/datasets/{id} | Update dataset
[**api_v2_datasets_post**](DatasetsApi.md#api_v2_datasets_post) | **POST** /api/v2/datasets | Create dataset
[**api_v2_datasets_search_post**](DatasetsApi.md#api_v2_datasets_search_post) | **POST** /api/v2/datasets/search | Search datasets


# **api_v2_datasets_get**
> DtoGenericResponseDtoDatasetSearchResponse api_v2_datasets_get(page=page, size=size, type=type, status=status, is_public=is_public, search=search, sort_by=sort_by, sort_order=sort_order, include=include)

List datasets

Get a paginated list of datasets with filtering and sorting

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_dataset_search_response import DtoGenericResponseDtoDatasetSearchResponse
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
    api_instance = rcabench.openapi.DatasetsApi(api_client)
    page = 56 # int | Page number (default 1) (optional)
    size = 56 # int | Page size (default 20, max 100) (optional)
    type = 'type_example' # str | Filter by dataset type (optional)
    status = 56 # int | Filter by status (optional)
    is_public = True # bool | Filter by public status (optional)
    search = 'search_example' # str | Search in name and description (optional)
    sort_by = 'sort_by_example' # str | Sort field (id,name,created_at,updated_at) (optional)
    sort_order = 'sort_order_example' # str | Sort order (asc,desc) (optional)
    include = 'include_example' # str | Include related data (injections,labels) (optional)

    try:
        # List datasets
        api_response = api_instance.api_v2_datasets_get(page=page, size=size, type=type, status=status, is_public=is_public, search=search, sort_by=sort_by, sort_order=sort_order, include=include)
        print("The response of DatasetsApi->api_v2_datasets_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->api_v2_datasets_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number (default 1) | [optional] 
 **size** | **int**| Page size (default 20, max 100) | [optional] 
 **type** | **str**| Filter by dataset type | [optional] 
 **status** | **int**| Filter by status | [optional] 
 **is_public** | **bool**| Filter by public status | [optional] 
 **search** | **str**| Search in name and description | [optional] 
 **sort_by** | **str**| Sort field (id,name,created_at,updated_at) | [optional] 
 **sort_order** | **str**| Sort order (asc,desc) | [optional] 
 **include** | **str**| Include related data (injections,labels) | [optional] 

### Return type

[**DtoGenericResponseDtoDatasetSearchResponse**](DtoGenericResponseDtoDatasetSearchResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Datasets retrieved successfully |  -  |
**400** | Invalid request parameters |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_datasets_id_delete**
> DtoGenericResponseAny api_v2_datasets_id_delete(id)

Delete dataset

Soft delete a dataset (sets status to -1)

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny
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
    api_instance = rcabench.openapi.DatasetsApi(api_client)
    id = 56 # int | Dataset ID

    try:
        # Delete dataset
        api_response = api_instance.api_v2_datasets_id_delete(id)
        print("The response of DatasetsApi->api_v2_datasets_id_delete:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->api_v2_datasets_id_delete: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Dataset ID | 

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
**200** | Dataset deleted successfully |  -  |
**400** | Invalid dataset ID |  -  |
**403** | Permission denied |  -  |
**404** | Dataset not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_datasets_id_download_get**
> bytearray api_v2_datasets_id_download_get(id)

Download dataset

Download dataset file by ID

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
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
    api_instance = rcabench.openapi.DatasetsApi(api_client)
    id = 56 # int | Dataset ID

    try:
        # Download dataset
        api_response = api_instance.api_v2_datasets_id_download_get(id)
        print("The response of DatasetsApi->api_v2_datasets_id_download_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->api_v2_datasets_id_download_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Dataset ID | 

### Return type

**bytearray**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Dataset file |  -  |
**400** | Invalid dataset ID |  -  |
**403** | Permission denied |  -  |
**404** | Dataset not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_datasets_id_get**
> DtoGenericResponseDtoDatasetV2Response api_v2_datasets_id_get(id, include_injections=include_injections, include_labels=include_labels)

Get dataset by ID

Get detailed information about a specific dataset

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_dataset_v2_response import DtoGenericResponseDtoDatasetV2Response
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
    api_instance = rcabench.openapi.DatasetsApi(api_client)
    id = 56 # int | Dataset ID
    include_injections = True # bool | Include related fault injections (optional)
    include_labels = True # bool | Include related labels (optional)

    try:
        # Get dataset by ID
        api_response = api_instance.api_v2_datasets_id_get(id, include_injections=include_injections, include_labels=include_labels)
        print("The response of DatasetsApi->api_v2_datasets_id_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->api_v2_datasets_id_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Dataset ID | 
 **include_injections** | **bool**| Include related fault injections | [optional] 
 **include_labels** | **bool**| Include related labels | [optional] 

### Return type

[**DtoGenericResponseDtoDatasetV2Response**](DtoGenericResponseDtoDatasetV2Response.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Dataset retrieved successfully |  -  |
**400** | Invalid dataset ID |  -  |
**403** | Permission denied |  -  |
**404** | Dataset not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_datasets_id_injections_patch**
> DtoGenericResponseDtoDatasetV2Response api_v2_datasets_id_injections_patch(id, manage)

Manage dataset injections

Add or remove injection associations for a dataset

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_dataset_v2_injection_manage_req import DtoDatasetV2InjectionManageReq
from rcabench.openapi.models.dto_generic_response_dto_dataset_v2_response import DtoGenericResponseDtoDatasetV2Response
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
    api_instance = rcabench.openapi.DatasetsApi(api_client)
    id = 56 # int | Dataset ID
    manage = rcabench.openapi.DtoDatasetV2InjectionManageReq() # DtoDatasetV2InjectionManageReq | Injection management request

    try:
        # Manage dataset injections
        api_response = api_instance.api_v2_datasets_id_injections_patch(id, manage)
        print("The response of DatasetsApi->api_v2_datasets_id_injections_patch:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->api_v2_datasets_id_injections_patch: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Dataset ID | 
 **manage** | [**DtoDatasetV2InjectionManageReq**](DtoDatasetV2InjectionManageReq.md)| Injection management request | 

### Return type

[**DtoGenericResponseDtoDatasetV2Response**](DtoGenericResponseDtoDatasetV2Response.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Injections managed successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**404** | Dataset not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_datasets_id_labels_patch**
> DtoGenericResponseDtoDatasetV2Response api_v2_datasets_id_labels_patch(id, manage)

Manage dataset labels

Add, remove labels or create new labels for a dataset

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_dataset_v2_label_manage_req import DtoDatasetV2LabelManageReq
from rcabench.openapi.models.dto_generic_response_dto_dataset_v2_response import DtoGenericResponseDtoDatasetV2Response
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
    api_instance = rcabench.openapi.DatasetsApi(api_client)
    id = 56 # int | Dataset ID
    manage = rcabench.openapi.DtoDatasetV2LabelManageReq() # DtoDatasetV2LabelManageReq | Label management request

    try:
        # Manage dataset labels
        api_response = api_instance.api_v2_datasets_id_labels_patch(id, manage)
        print("The response of DatasetsApi->api_v2_datasets_id_labels_patch:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->api_v2_datasets_id_labels_patch: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Dataset ID | 
 **manage** | [**DtoDatasetV2LabelManageReq**](DtoDatasetV2LabelManageReq.md)| Label management request | 

### Return type

[**DtoGenericResponseDtoDatasetV2Response**](DtoGenericResponseDtoDatasetV2Response.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Labels managed successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**404** | Dataset not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_datasets_id_put**
> DtoGenericResponseDtoDatasetV2Response api_v2_datasets_id_put(id, dataset)

Update dataset

Update dataset information, injection and label associations

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_dataset_v2_update_req import DtoDatasetV2UpdateReq
from rcabench.openapi.models.dto_generic_response_dto_dataset_v2_response import DtoGenericResponseDtoDatasetV2Response
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
    api_instance = rcabench.openapi.DatasetsApi(api_client)
    id = 56 # int | Dataset ID
    dataset = rcabench.openapi.DtoDatasetV2UpdateReq() # DtoDatasetV2UpdateReq | Dataset update request

    try:
        # Update dataset
        api_response = api_instance.api_v2_datasets_id_put(id, dataset)
        print("The response of DatasetsApi->api_v2_datasets_id_put:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->api_v2_datasets_id_put: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Dataset ID | 
 **dataset** | [**DtoDatasetV2UpdateReq**](DtoDatasetV2UpdateReq.md)| Dataset update request | 

### Return type

[**DtoGenericResponseDtoDatasetV2Response**](DtoGenericResponseDtoDatasetV2Response.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Dataset updated successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**404** | Dataset not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_datasets_post**
> DtoGenericResponseDtoDatasetV2Response api_v2_datasets_post(dataset)

Create dataset

Create a new dataset with optional injection and label associations

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_dataset_v2_create_req import DtoDatasetV2CreateReq
from rcabench.openapi.models.dto_generic_response_dto_dataset_v2_response import DtoGenericResponseDtoDatasetV2Response
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
    api_instance = rcabench.openapi.DatasetsApi(api_client)
    dataset = rcabench.openapi.DtoDatasetV2CreateReq() # DtoDatasetV2CreateReq | Dataset creation request

    try:
        # Create dataset
        api_response = api_instance.api_v2_datasets_post(dataset)
        print("The response of DatasetsApi->api_v2_datasets_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->api_v2_datasets_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **dataset** | [**DtoDatasetV2CreateReq**](DtoDatasetV2CreateReq.md)| Dataset creation request | 

### Return type

[**DtoGenericResponseDtoDatasetV2Response**](DtoGenericResponseDtoDatasetV2Response.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | Dataset created successfully |  -  |
**400** | Invalid request |  -  |
**403** | Permission denied |  -  |
**409** | Dataset already exists |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v2_datasets_search_post**
> DtoGenericResponseDtoSearchResponseDtoDatasetV2Response api_v2_datasets_search_post(search)

Search datasets

Advanced search for datasets with complex filtering

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_dataset_v2_search_req import DtoDatasetV2SearchReq
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_dataset_v2_response import DtoGenericResponseDtoSearchResponseDtoDatasetV2Response
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
    api_instance = rcabench.openapi.DatasetsApi(api_client)
    search = rcabench.openapi.DtoDatasetV2SearchReq() # DtoDatasetV2SearchReq | Search criteria

    try:
        # Search datasets
        api_response = api_instance.api_v2_datasets_search_post(search)
        print("The response of DatasetsApi->api_v2_datasets_search_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->api_v2_datasets_search_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **search** | [**DtoDatasetV2SearchReq**](DtoDatasetV2SearchReq.md)| Search criteria | 

### Return type

[**DtoGenericResponseDtoSearchResponseDtoDatasetV2Response**](DtoGenericResponseDtoSearchResponseDtoDatasetV2Response.md)

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

