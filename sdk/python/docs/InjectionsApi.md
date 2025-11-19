# openapi.InjectionsApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**build_datapack**](InjectionsApi.md#build_datapack) | **POST** /api/v2/injections/build | Submit batch datapack buildings
[**get_injection_by_id**](InjectionsApi.md#get_injection_by_id) | **GET** /api/v2/injections/{id} | Get injection by ID
[**get_injection_metadata**](InjectionsApi.md#get_injection_metadata) | **GET** /api/v2/injections/metadata | Get Injection Metadata
[**inject_fault**](InjectionsApi.md#inject_fault) | **POST** /api/v2/injections/inject | Submit batch fault injections
[**list_failed_injections**](InjectionsApi.md#list_failed_injections) | **GET** /api/v2/injections/analysis/with-issues | Query Fault Injection Records With Issues
[**list_injections**](InjectionsApi.md#list_injections) | **GET** /api/v2/injections | List injections
[**list_successful_injections**](InjectionsApi.md#list_successful_injections) | **GET** /api/v2/injections/analysis/no-issues | Query Fault Injection Records Without Issues
[**search_injections**](InjectionsApi.md#search_injections) | **POST** /api/v2/injections/search | Search injections


# **build_datapack**
> GenericResponseSubmitDatapackBuildingResp build_datapack(body)

Submit batch datapack buildings

### Example


```python
import openapi
from openapi.models.generic_response_submit_datapack_building_resp import GenericResponseSubmitDatapackBuildingResp
from openapi.models.submit_datapack_building_req import SubmitDatapackBuildingReq
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.InjectionsApi(api_client)
    body = openapi.SubmitDatapackBuildingReq() # SubmitDatapackBuildingReq | Datapack building request body

    try:
        # Submit batch datapack buildings
        api_response = api_instance.build_datapack(body)
        print("The response of InjectionsApi->build_datapack:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->build_datapack: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**SubmitDatapackBuildingReq**](SubmitDatapackBuildingReq.md)| Datapack building request body | 

### Return type

[**GenericResponseSubmitDatapackBuildingResp**](GenericResponseSubmitDatapackBuildingResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**202** | Datapack building submitted successfully |  -  |
**400** | Invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Resource not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_injection_by_id**
> GenericResponseInjectionDetailResp get_injection_by_id(id)

Get injection by ID

Get detailed information about a specific injection

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.generic_response_injection_detail_resp import GenericResponseInjectionDetailResp
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
    api_instance = openapi.InjectionsApi(api_client)
    id = 56 # int | Injection ID

    try:
        # Get injection by ID
        api_response = api_instance.get_injection_by_id(id)
        print("The response of InjectionsApi->get_injection_by_id:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->get_injection_by_id: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Injection ID | 

### Return type

[**GenericResponseInjectionDetailResp**](GenericResponseInjectionDetailResp.md)

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
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Injection not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_injection_metadata**
> GenericResponseInjectionMetadataResp get_injection_metadata(namespace)

Get Injection Metadata

Get injection-related metadata including configuration, field mappings, and namespace resources

### Example


```python
import openapi
from openapi.models.generic_response_injection_metadata_resp import GenericResponseInjectionMetadataResp
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.InjectionsApi(api_client)
    namespace = 'namespace_example' # str | Namespace prefix for config and resources metadata

    try:
        # Get Injection Metadata
        api_response = api_instance.get_injection_metadata(namespace)
        print("The response of InjectionsApi->get_injection_metadata:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->get_injection_metadata: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **namespace** | **str**| Namespace prefix for config and resources metadata | 

### Return type

[**GenericResponseInjectionMetadataResp**](GenericResponseInjectionMetadataResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Successfully returned metadata |  -  |
**400** | Invalid namespace prefix |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Resource not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **inject_fault**
> GenericResponseSubmitInjectionResp inject_fault(body)

Submit batch fault injections

Submit multiple fault injection tasks in batch

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.generic_response_submit_injection_resp import GenericResponseSubmitInjectionResp
from openapi.models.submit_injection_req import SubmitInjectionReq
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
    api_instance = openapi.InjectionsApi(api_client)
    body = openapi.SubmitInjectionReq() # SubmitInjectionReq | Fault injection request body

    try:
        # Submit batch fault injections
        api_response = api_instance.inject_fault(body)
        print("The response of InjectionsApi->inject_fault:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->inject_fault: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**SubmitInjectionReq**](SubmitInjectionReq.md)| Fault injection request body | 

### Return type

[**GenericResponseSubmitInjectionResp**](GenericResponseSubmitInjectionResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Fault injection submitted successfully |  -  |
**400** | Invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**404** | Resource not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_failed_injections**
> GenericResponseArrayInjectionWithIssuesResp list_failed_injections(labels=labels, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

Query Fault Injection Records With Issues

Query all fault injection records with issues based on time range

### Example


```python
import openapi
from openapi.models.generic_response_array_injection_with_issues_resp import GenericResponseArrayInjectionWithIssuesResp
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.InjectionsApi(api_client)
    labels = ['labels_example'] # List[str] | Filter by labels (array of key:value strings, e.g., 'type:chaos') (optional)
    lookback = 'lookback_example' # str | Time range query, supports custom relative time (1h/24h/7d) or custom, default not set (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | Custom start time, RFC3339 format, required when lookback=custom (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | Custom end time, RFC3339 format, required when lookback=custom (optional)

    try:
        # Query Fault Injection Records With Issues
        api_response = api_instance.list_failed_injections(labels=labels, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of InjectionsApi->list_failed_injections:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->list_failed_injections: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **labels** | [**List[str]**](str.md)| Filter by labels (array of key:value strings, e.g., &#39;type:chaos&#39;) | [optional] 
 **lookback** | **str**| Time range query, supports custom relative time (1h/24h/7d) or custom, default not set | [optional] 
 **custom_start_time** | **datetime**| Custom start time, RFC3339 format, required when lookback&#x3D;custom | [optional] 
 **custom_end_time** | **datetime**| Custom end time, RFC3339 format, required when lookback&#x3D;custom | [optional] 

### Return type

[**GenericResponseArrayInjectionWithIssuesResp**](GenericResponseArrayInjectionWithIssuesResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | OK |  -  |
**400** | Request parameter error, such as incorrect time format or parameter validation failure, etc. |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_injections**
> GenericResponseListInjectionResp list_injections(page=page, size=size, type=type, benchmark=benchmark, state=state, status=status, labels=labels)

List injections

Get a paginated list of injections with pagination and filtering

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.datapack_state import DatapackState
from openapi.models.generic_response_list_injection_resp import GenericResponseListInjectionResp
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
    api_instance = openapi.InjectionsApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)
    type = 56 # int | Filter by fault type (optional)
    benchmark = 'benchmark_example' # str | Filter by benchmark (optional)
    state = openapi.DatapackState() # DatapackState | Filter by injection state (optional)
    status = 56 # int | Filter by status (optional)
    labels = ['labels_example'] # List[str] | Filter by labels (array of key:value strings, e.g., 'type:chaos') (optional)

    try:
        # List injections
        api_response = api_instance.list_injections(page=page, size=size, type=type, benchmark=benchmark, state=state, status=status, labels=labels)
        print("The response of InjectionsApi->list_injections:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->list_injections: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]
 **type** | **int**| Filter by fault type | [optional] 
 **benchmark** | **str**| Filter by benchmark | [optional] 
 **state** | [**DatapackState**](.md)| Filter by injection state | [optional] 
 **status** | **int**| Filter by status | [optional] 
 **labels** | [**List[str]**](str.md)| Filter by labels (array of key:value strings, e.g., &#39;type:chaos&#39;) | [optional] 

### Return type

[**GenericResponseListInjectionResp**](GenericResponseListInjectionResp.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Injections retrieved successfully |  -  |
**400** | Invalid request format or parameters |  -  |
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_successful_injections**
> GenericResponseArrayInjectionNoIssuesResp list_successful_injections(labels=labels, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

Query Fault Injection Records Without Issues

Query all fault injection records without issues based on time range, returning detailed records including configuration information

### Example


```python
import openapi
from openapi.models.generic_response_array_injection_no_issues_resp import GenericResponseArrayInjectionNoIssuesResp
from openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi.InjectionsApi(api_client)
    labels = ['labels_example'] # List[str] | Filter by labels (array of key:value strings, e.g., 'type:chaos') (optional)
    lookback = 'lookback_example' # str | Time range query, supports custom relative time (1h/24h/7d) or custom, default not set (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | Custom start time, RFC3339 format, required when lookback=custom (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | Custom end time, RFC3339 format, required when lookback=custom (optional)

    try:
        # Query Fault Injection Records Without Issues
        api_response = api_instance.list_successful_injections(labels=labels, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of InjectionsApi->list_successful_injections:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->list_successful_injections: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **labels** | [**List[str]**](str.md)| Filter by labels (array of key:value strings, e.g., &#39;type:chaos&#39;) | [optional] 
 **lookback** | **str**| Time range query, supports custom relative time (1h/24h/7d) or custom, default not set | [optional] 
 **custom_start_time** | **datetime**| Custom start time, RFC3339 format, required when lookback&#x3D;custom | [optional] 
 **custom_end_time** | **datetime**| Custom end time, RFC3339 format, required when lookback&#x3D;custom | [optional] 

### Return type

[**GenericResponseArrayInjectionNoIssuesResp**](GenericResponseArrayInjectionNoIssuesResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Successfully returned fault injection records without issues |  -  |
**400** | Request parameter error, such as incorrect time format or parameter validation failure, etc. |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **search_injections**
> GenericResponseSearchRespInjectionResp search_injections(search)

Search injections

Advanced search for injections with complex filtering including custom labels

### Example

* Api Key Authentication (BearerAuth):

```python
import openapi
from openapi.models.generic_response_search_resp_injection_resp import GenericResponseSearchRespInjectionResp
from openapi.models.search_injection_req import SearchInjectionReq
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
    api_instance = openapi.InjectionsApi(api_client)
    search = openapi.SearchInjectionReq() # SearchInjectionReq | Search criteria

    try:
        # Search injections
        api_response = api_instance.search_injections(search)
        print("The response of InjectionsApi->search_injections:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling InjectionsApi->search_injections: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **search** | [**SearchInjectionReq**](SearchInjectionReq.md)| Search criteria | 

### Return type

[**GenericResponseSearchRespInjectionResp**](GenericResponseSearchRespInjectionResp.md)

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
**401** | Authentication required |  -  |
**403** | Permission denied |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

