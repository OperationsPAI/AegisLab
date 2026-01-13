# TracesApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getGroupStats**](#getgroupstats) | **GET** /api/v2/traces/group/stats | Get statistics for a group of traces|
|[**getTraceEvents**](#gettraceevents) | **GET** /api/v2/traces/{trace_id}/stream | Stream trace events in real-time|

# **getGroupStats**
> GenericResponseGroupStats getGroupStats()

Retrieves statistics such as total traces, average duration, and state distribution for a specified group of traces.

### Example

```typescript
import {
    TracesApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new TracesApi(configuration);

let groupId: string; //Group ID to query (default to undefined)

const { status, data } = await apiInstance.getGroupStats(
    groupId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **groupId** | [**string**] | Group ID to query | defaults to undefined|


### Return type

**GenericResponseGroupStats**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Group trace statistics |  -  |
|**400** | Invalid request format/parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getTraceEvents**
> string getTraceEvents()

Establishes a Server-Sent Events (SSE) connection to stream trace logs and task execution events in real-time. Returns historical events first, then switches to live monitoring.

### Example

```typescript
import {
    TracesApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new TracesApi(configuration);

let traceId: string; //Trace ID (default to undefined)
let lastId: string; //Last event ID received (optional) (default to '\"0\"')

const { status, data } = await apiInstance.getTraceEvents(
    traceId,
    lastId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **traceId** | [**string**] | Trace ID | defaults to undefined|
| **lastId** | [**string**] | Last event ID received | (optional) defaults to '\"0\"'|


### Return type

**string**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | A stream of event messages (e.g., log entries, task status updates). |  -  |
|**400** | Invalid trace ID or invalid request format/parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

