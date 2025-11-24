# rcabench.openapi.DatasetsApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**create_dataset**](DatasetsApi.md#create_dataset) | **POST** /api/v2/datasets | Create dataset
[**create_dataset_version**](DatasetsApi.md#create_dataset_version) | **POST** /api/v2/datasets/{dataset_id}/versions | Create dataset version
[**download_dataset_version**](DatasetsApi.md#download_dataset_version) | **GET** /api/v2/datasets/{dataset_id}/versions/{version_id}/download | Download dataset version
[**get_dataset_by_id**](DatasetsApi.md#get_dataset_by_id) | **GET** /api/v2/datasets/{dataset_id} | Get dataset by ID
[**get_dataset_version_by_id**](DatasetsApi.md#get_dataset_version_by_id) | **GET** /api/v2/datasets/{dataset_id}/versions/{version_id} | Get dataset version by ID
[**list_dataset_versions**](DatasetsApi.md#list_dataset_versions) | **GET** /api/v2/datasets/{dataset_id}/versions | List dataset versions
[**list_datasets**](DatasetsApi.md#list_datasets) | **GET** /api/v2/datasets | List datasets


# **create_dataset**
> GenericResponseDatasetResp create_dataset(request)

Create dataset

Create a new dataset with an initial version

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.create_dataset_req import CreateDatasetReq
from rcabench.openapi.models.generic_response_dataset_resp import GenericResponseDatasetResp
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
    request = rcabench.openapi.CreateDatasetReq() # CreateDatasetReq | Dataset creation request

    try:
        # Create dataset
        api_response = api_instance.create_dataset(request)
        print("The response of DatasetsApi->create_dataset:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->create_dataset: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**CreateDatasetReq**](CreateDatasetReq.md)| Dataset creation request | 

### Return type

[**GenericResponseDatasetResp**](GenericResponseDatasetResp.md)

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
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**409** | Conflict error |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **create_dataset_version**
> GenericResponseDatasetVersionResp create_dataset_version(dataset_id, request)

Create dataset version

Create a new dataset version for an existing dataset.

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.create_dataset_version_req import CreateDatasetVersionReq
from rcabench.openapi.models.generic_response_dataset_version_resp import GenericResponseDatasetVersionResp
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
    dataset_id = 56 # int | Dataset ID
    request = rcabench.openapi.CreateDatasetVersionReq() # CreateDatasetVersionReq | Dataset version creation request

    try:
        # Create dataset version
        api_response = api_instance.create_dataset_version(dataset_id, request)
        print("The response of DatasetsApi->create_dataset_version:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->create_dataset_version: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **dataset_id** | **int**| Dataset ID | 
 **request** | [**CreateDatasetVersionReq**](CreateDatasetVersionReq.md)| Dataset version creation request | 

### Return type

[**GenericResponseDatasetVersionResp**](GenericResponseDatasetVersionResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | Dataset version created successfully |  -  |
**400** | Invalid request |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**409** | Conflict error |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **download_dataset_version**
> bytearray download_dataset_version(dataset_id, version_id)

Download dataset version

Download dataset file by version ID

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
    dataset_id = 56 # int | Dataset ID
    version_id = 56 # int | Dataset Version ID

    try:
        # Download dataset version
        api_response = api_instance.download_dataset_version(dataset_id, version_id)
        print("The response of DatasetsApi->download_dataset_version:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->download_dataset_version: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **dataset_id** | **int**| Dataset ID | 
 **version_id** | **int**| Dataset Version ID | 

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
**400** | Invalid dataset ID/dataset version ID |  -  |
**403** | Permission denied |  -  |
**404** | Dataset not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_dataset_by_id**
> GenericResponseDatasetDetailResp get_dataset_by_id(dataset_id)

Get dataset by ID

Get detailed information about a specific dataset

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.generic_response_dataset_detail_resp import GenericResponseDatasetDetailResp
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
    dataset_id = 56 # int | Dataset ID

    try:
        # Get dataset by ID
        api_response = api_instance.get_dataset_by_id(dataset_id)
        print("The response of DatasetsApi->get_dataset_by_id:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->get_dataset_by_id: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **dataset_id** | **int**| Dataset ID | 

### Return type

[**GenericResponseDatasetDetailResp**](GenericResponseDatasetDetailResp.md)

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
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Dataset not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_dataset_version_by_id**
> GenericResponseDatasetVersionDetailResp get_dataset_version_by_id(dataset_id, version_id)

Get dataset version by ID

Get detailed information about a specific dataset version

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.generic_response_dataset_version_detail_resp import GenericResponseDatasetVersionDetailResp
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
    dataset_id = 56 # int | Dataset ID
    version_id = 56 # int | Dataset Version ID

    try:
        # Get dataset version by ID
        api_response = api_instance.get_dataset_version_by_id(dataset_id, version_id)
        print("The response of DatasetsApi->get_dataset_version_by_id:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->get_dataset_version_by_id: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **dataset_id** | **int**| Dataset ID | 
 **version_id** | **int**| Dataset Version ID | 

### Return type

[**GenericResponseDatasetVersionDetailResp**](GenericResponseDatasetVersionDetailResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Dataset version retrieved successfully |  -  |
**400** | Invalid dataset ID/dataset version ID |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Dataset or version not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_dataset_versions**
> GenericResponseListDatasetVersionResp list_dataset_versions(dataset_id, page=page, size=size, status=status)

List dataset versions

Get paginated list of dataset versions for a specific dataset

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.generic_response_list_dataset_version_resp import GenericResponseListDatasetVersionResp
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
    api_instance = rcabench.openapi.DatasetsApi(api_client)
    dataset_id = 56 # int | Dataset ID
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)
    status = rcabench.openapi.StatusType() # StatusType | Dataset version status filter (optional)

    try:
        # List dataset versions
        api_response = api_instance.list_dataset_versions(dataset_id, page=page, size=size, status=status)
        print("The response of DatasetsApi->list_dataset_versions:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->list_dataset_versions: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **dataset_id** | **int**| Dataset ID | 
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]
 **status** | [**StatusType**](.md)| Dataset version status filter | [optional] 

### Return type

[**GenericResponseListDatasetVersionResp**](GenericResponseListDatasetVersionResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Dataset versions retrieved successfully |  -  |
**400** | Invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_datasets**
> GenericResponseListDatasetResp list_datasets(page=page, size=size, type=type, is_public=is_public, status=status)

List datasets

Get paginated list of datasets with pagination and filtering

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.generic_response_list_dataset_resp import GenericResponseListDatasetResp
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
    api_instance = rcabench.openapi.DatasetsApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)
    type = 'type_example' # str | Dataset type filter (optional)
    is_public = True # bool | Dataset public visibility filter (optional)
    status = rcabench.openapi.StatusType() # StatusType | Dataset status filter (optional)

    try:
        # List datasets
        api_response = api_instance.list_datasets(page=page, size=size, type=type, is_public=is_public, status=status)
        print("The response of DatasetsApi->list_datasets:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetsApi->list_datasets: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]
 **type** | **str**| Dataset type filter | [optional] 
 **is_public** | **bool**| Dataset public visibility filter | [optional] 
 **status** | [**StatusType**](.md)| Dataset status filter | [optional] 

### Return type

[**GenericResponseListDatasetResp**](GenericResponseListDatasetResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Datasets retrieved successfully |  -  |
**400** | Invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

