# SystemApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getSystemHealth**](#getsystemhealth) | **GET** /system/health | System health check|
|[**getSystemMetrics**](#getsystemmetrics) | **GET** /api/v2/system/metrics | Get current system metrics|
|[**getSystemMetricsHistory**](#getsystemmetricshistory) | **GET** /api/v2/system/metrics/history | Get historical system metrics|

# **getSystemHealth**
> GenericResponseHealthCheckResp getSystemHealth()

Get system health status and service information

### Example

```typescript
import {
    SystemApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new SystemApi(configuration);

const { status, data } = await apiInstance.getSystemHealth();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**GenericResponseHealthCheckResp**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Health check successful |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSystemMetrics**
> GenericResponseSystemMetricsResp getSystemMetrics()

Get current CPU, memory, and disk usage metrics

### Example

```typescript
import {
    SystemApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new SystemApi(configuration);

const { status, data } = await apiInstance.getSystemMetrics();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**GenericResponseSystemMetricsResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | System metrics retrieved successfully |  -  |
|**401** | Authentication required |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSystemMetricsHistory**
> GenericResponseSystemMetricsHistoryResp getSystemMetricsHistory()

Get 24-hour historical CPU and memory usage metrics

### Example

```typescript
import {
    SystemApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new SystemApi(configuration);

const { status, data } = await apiInstance.getSystemMetricsHistory();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**GenericResponseSystemMetricsHistoryResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | System metrics history retrieved successfully |  -  |
|**401** | Authentication required |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

