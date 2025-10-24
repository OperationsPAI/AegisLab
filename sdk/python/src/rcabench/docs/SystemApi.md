# rcabench.openapi.SystemApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**system_audit_get**](SystemApi.md#system_audit_get) | **GET** /system/audit | List audit logs
[**system_audit_id_get**](SystemApi.md#system_audit_id_get) | **GET** /system/audit/{id} | Get audit log by ID
[**system_audit_post**](SystemApi.md#system_audit_post) | **POST** /system/audit | Create audit log
[**system_health_get**](SystemApi.md#system_health_get) | **GET** /system/health | System health check
[**system_monitor_info_get**](SystemApi.md#system_monitor_info_get) | **GET** /system/monitor/info | Get system information
[**system_monitor_metrics_post**](SystemApi.md#system_monitor_metrics_post) | **POST** /system/monitor/metrics | Get monitoring metrics
[**system_statistics_get**](SystemApi.md#system_statistics_get) | **GET** /system/statistics | Get system statistics


# **system_audit_get**
> DtoGenericResponseDtoAuditLogListResponse system_audit_get(page=page, size=size, user_id=user_id, action=action, resource=resource, success=success, start_date=start_date, end_date=end_date)

List audit logs

Get paginated list of audit logs with optional filtering

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_audit_log_list_response import DtoGenericResponseDtoAuditLogListResponse
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.SystemApi(api_client)
    page = 1 # int | Page number (optional) (default to 1)
    size = 20 # int | Page size (optional) (default to 20)
    user_id = 56 # int | Filter by user ID (optional)
    action = 'action_example' # str | Filter by action (optional)
    resource = 'resource_example' # str | Filter by resource (optional)
    success = True # bool | Filter by success status (optional)
    start_date = 'start_date_example' # str | Filter from date (YYYY-MM-DD) (optional)
    end_date = 'end_date_example' # str | Filter to date (YYYY-MM-DD) (optional)

    try:
        # List audit logs
        api_response = api_instance.system_audit_get(page=page, size=size, user_id=user_id, action=action, resource=resource, success=success, start_date=start_date, end_date=end_date)
        print("The response of SystemApi->system_audit_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling SystemApi->system_audit_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int**| Page number | [optional] [default to 1]
 **size** | **int**| Page size | [optional] [default to 20]
 **user_id** | **int**| Filter by user ID | [optional] 
 **action** | **str**| Filter by action | [optional] 
 **resource** | **str**| Filter by resource | [optional] 
 **success** | **bool**| Filter by success status | [optional] 
 **start_date** | **str**| Filter from date (YYYY-MM-DD) | [optional] 
 **end_date** | **str**| Filter to date (YYYY-MM-DD) | [optional] 

### Return type

[**DtoGenericResponseDtoAuditLogListResponse**](DtoGenericResponseDtoAuditLogListResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Audit logs retrieved successfully |  -  |
**400** | Invalid request parameters |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **system_audit_id_get**
> DtoGenericResponseDtoAuditLogResponse system_audit_id_get(id)

Get audit log by ID

Get a specific audit log entry by ID

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_audit_log_response import DtoGenericResponseDtoAuditLogResponse
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.SystemApi(api_client)
    id = 56 # int | Audit log ID

    try:
        # Get audit log by ID
        api_response = api_instance.system_audit_id_get(id)
        print("The response of SystemApi->system_audit_id_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling SystemApi->system_audit_id_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| Audit log ID | 

### Return type

[**DtoGenericResponseDtoAuditLogResponse**](DtoGenericResponseDtoAuditLogResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Audit log retrieved successfully |  -  |
**400** | Invalid ID |  -  |
**404** | Audit log not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **system_audit_post**
> DtoGenericResponseDtoAuditLogResponse system_audit_post(audit_log)

Create audit log

Create a new audit log entry

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_audit_log_request import DtoAuditLogRequest
from rcabench.openapi.models.dto_generic_response_dto_audit_log_response import DtoGenericResponseDtoAuditLogResponse
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
    api_instance = rcabench.openapi.SystemApi(api_client)
    audit_log = rcabench.openapi.DtoAuditLogRequest() # DtoAuditLogRequest | Audit log data

    try:
        # Create audit log
        api_response = api_instance.system_audit_post(audit_log)
        print("The response of SystemApi->system_audit_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling SystemApi->system_audit_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **audit_log** | [**DtoAuditLogRequest**](DtoAuditLogRequest.md)| Audit log data | 

### Return type

[**DtoGenericResponseDtoAuditLogResponse**](DtoGenericResponseDtoAuditLogResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | Audit log created successfully |  -  |
**400** | Invalid request |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **system_health_get**
> DtoGenericResponseDtoHealthCheckResponse system_health_get()

System health check

Get system health status and service information

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_health_check_response import DtoGenericResponseDtoHealthCheckResponse
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.SystemApi(api_client)

    try:
        # System health check
        api_response = api_instance.system_health_get()
        print("The response of SystemApi->system_health_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling SystemApi->system_health_get: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**DtoGenericResponseDtoHealthCheckResponse**](DtoGenericResponseDtoHealthCheckResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Health check successful |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **system_monitor_info_get**
> DtoGenericResponseDtoSystemInfo system_monitor_info_get()

Get system information

Get basic system information and status

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_system_info import DtoGenericResponseDtoSystemInfo
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.SystemApi(api_client)

    try:
        # Get system information
        api_response = api_instance.system_monitor_info_get()
        print("The response of SystemApi->system_monitor_info_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling SystemApi->system_monitor_info_get: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**DtoGenericResponseDtoSystemInfo**](DtoGenericResponseDtoSystemInfo.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | System info retrieved successfully |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **system_monitor_metrics_post**
> DtoGenericResponseDtoMonitoringMetricsResponse system_monitor_metrics_post(request)

Get monitoring metrics

Query monitoring metrics for system performance

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_monitoring_metrics_response import DtoGenericResponseDtoMonitoringMetricsResponse
from rcabench.openapi.models.dto_monitoring_query_request import DtoMonitoringQueryRequest
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.SystemApi(api_client)
    request = rcabench.openapi.DtoMonitoringQueryRequest() # DtoMonitoringQueryRequest | Metrics query request

    try:
        # Get monitoring metrics
        api_response = api_instance.system_monitor_metrics_post(request)
        print("The response of SystemApi->system_monitor_metrics_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling SystemApi->system_monitor_metrics_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DtoMonitoringQueryRequest**](DtoMonitoringQueryRequest.md)| Metrics query request | 

### Return type

[**DtoGenericResponseDtoMonitoringMetricsResponse**](DtoGenericResponseDtoMonitoringMetricsResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Metrics retrieved successfully |  -  |
**400** | Invalid request |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **system_statistics_get**
> DtoGenericResponseDtoSystemStatisticsResponse system_statistics_get()

Get system statistics

Get comprehensive system statistics and metrics

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_system_statistics_response import DtoGenericResponseDtoSystemStatisticsResponse
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8082
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8082"
)


# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.SystemApi(api_client)

    try:
        # Get system statistics
        api_response = api_instance.system_statistics_get()
        print("The response of SystemApi->system_statistics_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling SystemApi->system_statistics_get: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**DtoGenericResponseDtoSystemStatisticsResponse**](DtoGenericResponseDtoSystemStatisticsResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Statistics retrieved successfully |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

