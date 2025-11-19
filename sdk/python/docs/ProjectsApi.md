# openapi.ProjectsApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**create_project**](ProjectsApi.md#create_project) | **POST** /api/v2/projects | Create a new project
[**get_project_by_id**](ProjectsApi.md#get_project_by_id) | **GET** /api/v2/projects/{project_id} | Get project by ID
[**list_projects**](ProjectsApi.md#list_projects) | **GET** /api/v2/projects | List projects


# **create_project**
> GenericResponseProjectResp create_project(request)

Create a new project

Create a new project with specified details

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.create_project_req import CreateProjectReq
from openapi.models.generic_response_project_resp import GenericResponseProjectResp
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
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
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.ProjectsApi(api_client)
    request = openapi.CreateProjectReq() # CreateProjectReq | Project creation request

    try:
        # Create a new project
        api_response = api_instance.create_project(request)
        print("The response of ProjectsApi->create_project:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ProjectsApi->create_project: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**CreateProjectReq**](CreateProjectReq.md)| Project creation request | 

### Return type

[**GenericResponseProjectResp**](GenericResponseProjectResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | Project created successfully |  -  |
**400** | Invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**409** | Project already exists |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_project_by_id**
> GenericResponseProjectDetailResp get_project_by_id(project_id)

Get project by ID

Get detailed information about a specific project

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.generic_response_project_detail_resp import GenericResponseProjectDetailResp
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
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
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.ProjectsApi(api_client)
    project_id = 56 # int | Project ID

    try:
        # Get project by ID
        api_response = api_instance.get_project_by_id(project_id)
        print("The response of ProjectsApi->get_project_by_id:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ProjectsApi->get_project_by_id: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **project_id** | **int**| Project ID | 

### Return type

[**GenericResponseProjectDetailResp**](GenericResponseProjectDetailResp.md)

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
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Project not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_projects**
> GenericResponseListProjectResp list_projects(page=page, size=size, is_public=is_public, status=status)

List projects

Get paginated list of projects with filtering

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.generic_response_list_project_resp import GenericResponseListProjectResp
from openapi.models.status_type import StatusType
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
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
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.ProjectsApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)
    is_public = True # bool | Filter by public status (optional)
    status = openapi.StatusType() # StatusType | Filter by status (optional)

    try:
        # List projects
        api_response = api_instance.list_projects(page=page, size=size, is_public=is_public, status=status)
        print("The response of ProjectsApi->list_projects:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ProjectsApi->list_projects: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]
 **is_public** | **bool**| Filter by public status | [optional] 
 **status** | [**StatusType**](.md)| Filter by status | [optional] 

### Return type

[**GenericResponseListProjectResp**](GenericResponseListProjectResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Projects retrieved successfully |  -  |
**400** | Invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

