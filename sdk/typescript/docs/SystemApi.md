# SystemApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getSystemHealth**](#getsystemhealth) | **GET** /system/health | System health check|

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

