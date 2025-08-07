# rcabench.openapi.ProjectsApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_projects_id_get**](ProjectsApi.md#api_v2_projects_id_get) | **GET** /api/v2/projects/{id} | Get project by ID


# **api_v2_projects_id_get**
> DtoGenericResponseDtoProjectV2Response api_v2_projects_id_get(id, include_containers=include_containers, include_datasets=include_datasets, include_injections=include_injections, include_labels=include_labels)

Get project by ID

Get detailed information about a specific project

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_project_v2_response import DtoGenericResponseDtoProjectV2Response
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
    api_instance = rcabench.openapi.ProjectsApi(api_client)
    id = 56 # int | Project ID
    include_containers = True # bool | Include related containers (optional)
    include_datasets = True # bool | Include related datasets (optional)
    include_injections = True # bool | Include related fault injections (optional)
    include_labels = True # bool | Include related labels (optional)

    try:
        # Get project by ID
        api_response = api_instance.api_v2_projects_id_get(id, include_containers=include_containers, include_datasets=include_datasets, include_injections=include_injections, include_labels=include_labels)
        print("The response of ProjectsApi->api_v2_projects_id_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ProjectsApi->api_v2_projects_id_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Project ID | 
 **include_containers** | **bool**| Include related containers | [optional] 
 **include_datasets** | **bool**| Include related datasets | [optional] 
 **include_injections** | **bool**| Include related fault injections | [optional] 
 **include_labels** | **bool**| Include related labels | [optional] 

### Return type

[**DtoGenericResponseDtoProjectV2Response**](DtoGenericResponseDtoProjectV2Response.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Project retrieved successfully |  -  |
**400** | Invalid project ID |  -  |
**403** | Permission denied |  -  |
**404** | Project not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

