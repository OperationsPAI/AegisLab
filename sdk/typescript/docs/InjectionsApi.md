# InjectionsApi

All URIs are relative to *http://http://localhost:8082*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**batchManageInjectionLabels**](#batchmanageinjectionlabels) | **PATCH** /api/v2/injections/labels/batch | Batch manage injection labels|
|[**buildDatapack**](#builddatapack) | **POST** /api/v2/injections/build | Submit batch datapack buildings|
|[**getInjectionById**](#getinjectionbyid) | **GET** /api/v2/injections/{id} | Get injection by ID|
|[**getInjectionMetadata**](#getinjectionmetadata) | **GET** /api/v2/injections/metadata | Get Injection Metadata|
|[**injectFault**](#injectfault) | **POST** /api/v2/injections/inject | Submit batch fault injections|
|[**listFailedInjections**](#listfailedinjections) | **GET** /api/v2/injections/analysis/no-issues | Query Fault Injection Records Without Issues|
|[**listInjections**](#listinjections) | **GET** /api/v2/injections | List injections|
|[**listSuccessfulInjections**](#listsuccessfulinjections) | **GET** /api/v2/injections/analysis/with-issues | Query Fault Injection Records With Issues|
|[**manageInjectionLabels**](#manageinjectionlabels) | **PATCH** /api/v2/injections/{id}/labels | Manage injection custom labels|
|[**searchInjections**](#searchinjections) | **POST** /api/v2/injections/search | Search injections|

# **batchManageInjectionLabels**
> GenericResponseBatchManageInjectionLabelResp batchManageInjectionLabels(batchManage)

Add or remove labels from multiple injections by IDs with success/failure tracking

### Example

```typescript
import {
    InjectionsApi,
    Configuration,
    BatchManageInjectionLabelReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new InjectionsApi(configuration);

let batchManage: BatchManageInjectionLabelReq; //Batch manage label request

const { status, data } = await apiInstance.batchManageInjectionLabels(
    batchManage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **batchManage** | **BatchManageInjectionLabelReq**| Batch manage label request | |


### Return type

**GenericResponseBatchManageInjectionLabelResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Injection labels managed successfully |  -  |
|**400** | Invalid request |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **buildDatapack**
> GenericResponseSubmitDatapackBuildingResp buildDatapack(body)


### Example

```typescript
import {
    InjectionsApi,
    Configuration,
    SubmitDatapackBuildingReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new InjectionsApi(configuration);

let body: SubmitDatapackBuildingReq; //Datapack building request body

const { status, data } = await apiInstance.buildDatapack(
    body
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **body** | **SubmitDatapackBuildingReq**| Datapack building request body | |


### Return type

**GenericResponseSubmitDatapackBuildingResp**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**202** | Datapack building submitted successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getInjectionById**
> GenericResponseInjectionDetailResp getInjectionById()

Get detailed information about a specific injection

### Example

```typescript
import {
    InjectionsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new InjectionsApi(configuration);

let id: number; //Injection ID (default to undefined)

const { status, data } = await apiInstance.getInjectionById(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**number**] | Injection ID | defaults to undefined|


### Return type

**GenericResponseInjectionDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Injection retrieved successfully |  -  |
|**400** | Invalid injection ID |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Injection not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getInjectionMetadata**
> GenericResponseInjectionMetadataResp getInjectionMetadata()

Get injection-related metadata including configuration, field mappings, and system resources

### Example

```typescript
import {
    InjectionsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new InjectionsApi(configuration);

let system: 'ts' | 'otel-demo' | 'media' | 'hs' | 'sn' | 'ob' | 'ts' | 'otel-demo' | 'media' | 'hs' | 'sn' | 'ob'; //System for config and resources metadata (default to undefined)

const { status, data } = await apiInstance.getInjectionMetadata(
    system
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **system** | [**&#39;ts&#39; | &#39;otel-demo&#39; | &#39;media&#39; | &#39;hs&#39; | &#39;sn&#39; | &#39;ob&#39; | &#39;ts&#39; | &#39;otel-demo&#39; | &#39;media&#39; | &#39;hs&#39; | &#39;sn&#39; | &#39;ob&#39;**]**Array<&#39;ts&#39; &#124; &#39;otel-demo&#39; &#124; &#39;media&#39; &#124; &#39;hs&#39; &#124; &#39;sn&#39; &#124; &#39;ob&#39; &#124; &#39;ts&#39; &#124; &#39;otel-demo&#39; &#124; &#39;media&#39; &#124; &#39;hs&#39; &#124; &#39;sn&#39; &#124; &#39;ob&#39;>** | System for config and resources metadata | defaults to undefined|


### Return type

**GenericResponseInjectionMetadataResp**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successfully returned metadata |  -  |
|**400** | Invalid system |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **injectFault**
> GenericResponseSubmitInjectionResp injectFault(body)

Submit multiple fault injection tasks in batch

### Example

```typescript
import {
    InjectionsApi,
    Configuration,
    SubmitInjectionReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new InjectionsApi(configuration);

let body: SubmitInjectionReq; //Fault injection request body

const { status, data } = await apiInstance.injectFault(
    body
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **body** | **SubmitInjectionReq**| Fault injection request body | |


### Return type

**GenericResponseSubmitInjectionResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Fault injection submitted successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Resource not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listFailedInjections**
> GenericResponseArrayInjectionNoIssuesResp listFailedInjections()

Query all fault injection records without issues based on time range, returning detailed records including configuration information

### Example

```typescript
import {
    InjectionsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new InjectionsApi(configuration);

let labels: Array<string>; //Filter by labels (array of key:value strings, e.g., \'type:chaos\') (optional) (default to undefined)
let lookback: string; //Time range query, supports custom relative time (1h/24h/7d) or custom, default not set (optional) (default to undefined)
let customStartTime: string; //Custom start time, RFC3339 format, required when lookback=custom (optional) (default to undefined)
let customEndTime: string; //Custom end time, RFC3339 format, required when lookback=custom (optional) (default to undefined)

const { status, data } = await apiInstance.listFailedInjections(
    labels,
    lookback,
    customStartTime,
    customEndTime
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **labels** | **Array&lt;string&gt;** | Filter by labels (array of key:value strings, e.g., \&#39;type:chaos\&#39;) | (optional) defaults to undefined|
| **lookback** | [**string**] | Time range query, supports custom relative time (1h/24h/7d) or custom, default not set | (optional) defaults to undefined|
| **customStartTime** | [**string**] | Custom start time, RFC3339 format, required when lookback&#x3D;custom | (optional) defaults to undefined|
| **customEndTime** | [**string**] | Custom end time, RFC3339 format, required when lookback&#x3D;custom | (optional) defaults to undefined|


### Return type

**GenericResponseArrayInjectionNoIssuesResp**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successfully returned fault injection records without issues |  -  |
|**400** | Request parameter error, such as incorrect time format or parameter validation failure, etc. |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listInjections**
> GenericResponseListInjectionResp listInjections()

Get a paginated list of injections with pagination and filtering

### Example

```typescript
import {
    InjectionsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new InjectionsApi(configuration);

let page: number; //Page number (optional) (default to 1)
let size: number; //Page size (optional) (default to 20)
let type: 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 | 13 | 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21 | 22 | 23 | 24 | 25 | 26 | 27 | 28 | 29 | 30; //Filter by fault type (optional) (default to undefined)
let benchmark: string; //Filter by benchmark (optional) (default to undefined)
let state: DatapackState; //Filter by injection state (optional) (default to undefined)
let status: number; //Filter by status (optional) (default to undefined)
let labels: Array<string>; //Filter by labels (array of key:value strings, e.g., \'type:chaos\') (optional) (default to undefined)

const { status, data } = await apiInstance.listInjections(
    page,
    size,
    type,
    benchmark,
    state,
    status,
    labels
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **size** | [**number**] | Page size | (optional) defaults to 20|
| **type** | [**0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 | 13 | 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21 | 22 | 23 | 24 | 25 | 26 | 27 | 28 | 29 | 30**]**Array<0 &#124; 1 &#124; 2 &#124; 3 &#124; 4 &#124; 5 &#124; 6 &#124; 7 &#124; 8 &#124; 9 &#124; 10 &#124; 11 &#124; 12 &#124; 13 &#124; 14 &#124; 15 &#124; 16 &#124; 17 &#124; 18 &#124; 19 &#124; 20 &#124; 21 &#124; 22 &#124; 23 &#124; 24 &#124; 25 &#124; 26 &#124; 27 &#124; 28 &#124; 29 &#124; 30>** | Filter by fault type | (optional) defaults to undefined|
| **benchmark** | [**string**] | Filter by benchmark | (optional) defaults to undefined|
| **state** | **DatapackState** | Filter by injection state | (optional) defaults to undefined|
| **status** | [**number**] | Filter by status | (optional) defaults to undefined|
| **labels** | **Array&lt;string&gt;** | Filter by labels (array of key:value strings, e.g., \&#39;type:chaos\&#39;) | (optional) defaults to undefined|


### Return type

**GenericResponseListInjectionResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Injections retrieved successfully |  -  |
|**400** | Invalid request format or parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listSuccessfulInjections**
> GenericResponseArrayInjectionWithIssuesResp listSuccessfulInjections()

Query all fault injection records with issues based on time range

### Example

```typescript
import {
    InjectionsApi,
    Configuration
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new InjectionsApi(configuration);

let labels: Array<string>; //Filter by labels (array of key:value strings, e.g., \'type:chaos\') (optional) (default to undefined)
let lookback: string; //Time range query, supports custom relative time (1h/24h/7d) or custom, default not set (optional) (default to undefined)
let customStartTime: string; //Custom start time, RFC3339 format, required when lookback=custom (optional) (default to undefined)
let customEndTime: string; //Custom end time, RFC3339 format, required when lookback=custom (optional) (default to undefined)

const { status, data } = await apiInstance.listSuccessfulInjections(
    labels,
    lookback,
    customStartTime,
    customEndTime
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **labels** | **Array&lt;string&gt;** | Filter by labels (array of key:value strings, e.g., \&#39;type:chaos\&#39;) | (optional) defaults to undefined|
| **lookback** | [**string**] | Time range query, supports custom relative time (1h/24h/7d) or custom, default not set | (optional) defaults to undefined|
| **customStartTime** | [**string**] | Custom start time, RFC3339 format, required when lookback&#x3D;custom | (optional) defaults to undefined|
| **customEndTime** | [**string**] | Custom end time, RFC3339 format, required when lookback&#x3D;custom | (optional) defaults to undefined|


### Return type

**GenericResponseArrayInjectionWithIssuesResp**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Request parameter error, such as incorrect time format or parameter validation failure, etc. |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **manageInjectionLabels**
> GenericResponseInjectionResp manageInjectionLabels(manage)

Add or remove custom labels (key-value pairs) for an injection

### Example

```typescript
import {
    InjectionsApi,
    Configuration,
    ManageInjectionLabelReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new InjectionsApi(configuration);

let id: number; //Injection ID (default to undefined)
let manage: ManageInjectionLabelReq; //Custom label management request

const { status, data } = await apiInstance.manageInjectionLabels(
    id,
    manage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **manage** | **ManageInjectionLabelReq**| Custom label management request | |
| **id** | [**number**] | Injection ID | defaults to undefined|


### Return type

**GenericResponseInjectionResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Custom labels managed successfully |  -  |
|**400** | Invalid injection ID or request format/parameters |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**404** | Injection not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **searchInjections**
> GenericResponseSearchRespInjectionDetailResp searchInjections(search)

Advanced search for injections with complex filtering including name search, custom labels, tags, and time ranges

### Example

```typescript
import {
    InjectionsApi,
    Configuration,
    SearchInjectionReq
} from 'rcabench-client';

const configuration = new Configuration();
const apiInstance = new InjectionsApi(configuration);

let search: SearchInjectionReq; //Search criteria

const { status, data } = await apiInstance.searchInjections(
    search
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **search** | **SearchInjectionReq**| Search criteria | |


### Return type

**GenericResponseSearchRespInjectionDetailResp**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Search results |  -  |
|**400** | Invalid request |  -  |
|**401** | Authentication required |  -  |
|**403** | Permission denied |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

