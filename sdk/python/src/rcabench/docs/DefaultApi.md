# rcabench.openapi.DefaultApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_datasets_get**](DefaultApi.md#api_v2_datasets_get) | **GET** /api/v2/datasets | 


# **api_v2_datasets_get**
> DtoGenericResponseDtoDatasetSearchResponse api_v2_datasets_get(type=type, status=status, is_public=is_public, search=search, sort_by=sort_by, sort_order=sort_order, include=include)

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_dataset_search_response import DtoGenericResponseDtoDatasetSearchResponse
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
    api_instance = rcabench.openapi.DefaultApi(api_client)
    type = 'type_example' # str | Filter by dataset type (optional)
    status = 56 # int | Filter by status (optional)
    is_public = True # bool | Filter by public status (optional)
    search = 'search_example' # str | Search in name and description (optional)
    sort_by = 'sort_by_example' # str | Sort field (id,name,created_at,updated_at) (optional)
    sort_order = 'sort_order_example' # str | Sort order (asc,desc) (optional)
    include = 'include_example' # str | Include related data (injections,labels) (optional)

    try:
        api_response = api_instance.api_v2_datasets_get(type=type, status=status, is_public=is_public, search=search, sort_by=sort_by, sort_order=sort_order, include=include)
        print("The response of DefaultApi->api_v2_datasets_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DefaultApi->api_v2_datasets_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
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

No authorization required

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

