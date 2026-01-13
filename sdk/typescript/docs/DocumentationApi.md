# DocumentationApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**apiDocsModelsGet**](#apidocsmodelsget) | **GET** /api/_docs/models | API Model Definitions|

# **apiDocsModelsGet**
> SSEEventName apiDocsModelsGet()

Virtual endpoint for including all DTO type definitions in Swagger documentation. DO NOT USE in production.

### Example

```typescript
import {
    DocumentationApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new DocumentationApi(configuration);

const { status, data } = await apiInstance.apiDocsModelsGet();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**SSEEventName**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | SSE event name constants |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

