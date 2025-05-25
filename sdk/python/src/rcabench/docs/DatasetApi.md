# rcabench.openapi.DatasetApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_datasets_delete**](DatasetApi.md#api_v1_datasets_delete) | **DELETE** /api/v1/datasets | 删除数据集数据
[**api_v1_datasets_download_get**](DatasetApi.md#api_v1_datasets_download_get) | **GET** /api/v1/datasets/download | 下载数据集打包文件
[**api_v1_datasets_get**](DatasetApi.md#api_v1_datasets_get) | **GET** /api/v1/datasets | 分页查询数据集列表
[**api_v1_datasets_post**](DatasetApi.md#api_v1_datasets_post) | **POST** /api/v1/datasets | 批量构建数据集
[**api_v1_datasets_query_get**](DatasetApi.md#api_v1_datasets_query_get) | **GET** /api/v1/datasets/query | 查询单个数据集详情


# **api_v1_datasets_delete**
> DtoGenericResponseDtoDatasetDeleteResp api_v1_datasets_delete(names)

删除数据集数据

删除数据集数据

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_dataset_delete_resp import DtoGenericResponseDtoDatasetDeleteResp
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
    api_instance = rcabench.openapi.DatasetApi(api_client)
    names = ['names_example'] # List[str] | 数据集名称列表

    try:
        # 删除数据集数据
        api_response = api_instance.api_v1_datasets_delete(names)
        print("The response of DatasetApi->api_v1_datasets_delete:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetApi->api_v1_datasets_delete: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **names** | [**List[str]**](str.md)| 数据集名称列表 | 

### Return type

[**DtoGenericResponseDtoDatasetDeleteResp**](DtoGenericResponseDtoDatasetDeleteResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | OK |  -  |
**400** | Bad Request |  -  |
**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_datasets_download_get**
> str api_v1_datasets_download_get(group_ids=group_ids, names=names)

下载数据集打包文件

将指定路径的多个数据集打包为 ZIP 文件下载（自动排除 result.csv 文件）

### Example


```python
import rcabench.openapi
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
    api_instance = rcabench.openapi.DatasetApi(api_client)
    group_ids = ['group_ids_example'] # List[str] | 数据集组ID列表，与names参数二选一 (optional)
    names = ['names_example'] # List[str] | 数据集名称列表，与group_ids参数二选一 (optional)

    try:
        # 下载数据集打包文件
        api_response = api_instance.api_v1_datasets_download_get(group_ids=group_ids, names=names)
        print("The response of DatasetApi->api_v1_datasets_download_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetApi->api_v1_datasets_download_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **group_ids** | [**List[str]**](str.md)| 数据集组ID列表，与names参数二选一 | [optional] 
 **names** | [**List[str]**](str.md)| 数据集名称列表，与group_ids参数二选一 | [optional] 

### Return type

**str**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zip

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ZIP 文件流 |  -  |
**400** | 参数绑定错误 |  -  |
**403** | 非法路径访问 |  -  |
**500** | 文件打包失败 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_datasets_get**
> DtoGenericResponseDtoPaginationRespDtoDatasetItem api_v1_datasets_get(page_num, page_size)

分页查询数据集列表

获取状态为成功的注入数据集列表（支持分页参数）

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_pagination_resp_dto_dataset_item import DtoGenericResponseDtoPaginationRespDtoDatasetItem
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
    api_instance = rcabench.openapi.DatasetApi(api_client)
    page_num = 1 # int | 页码（从1开始） (default to 1)
    page_size = 10 # int | 每页数量 (default to 10)

    try:
        # 分页查询数据集列表
        api_response = api_instance.api_v1_datasets_get(page_num, page_size)
        print("The response of DatasetApi->api_v1_datasets_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetApi->api_v1_datasets_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page_num** | **int**| 页码（从1开始） | [default to 1]
 **page_size** | **int**| 每页数量 | [default to 10]

### Return type

[**DtoGenericResponseDtoPaginationRespDtoDatasetItem**](DtoGenericResponseDtoPaginationRespDtoDatasetItem.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | 成功响应 |  -  |
**400** | 参数校验失败 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_datasets_post**
> DtoGenericResponseDtoSubmitResp api_v1_datasets_post(body)

批量构建数据集

批量构建数据集

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_dataset_build_payload import DtoDatasetBuildPayload
from rcabench.openapi.models.dto_generic_response_dto_submit_resp import DtoGenericResponseDtoSubmitResp
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
    api_instance = rcabench.openapi.DatasetApi(api_client)
    body = [rcabench.openapi.DtoDatasetBuildPayload()] # List[DtoDatasetBuildPayload] | 请求体

    try:
        # 批量构建数据集
        api_response = api_instance.api_v1_datasets_post(body)
        print("The response of DatasetApi->api_v1_datasets_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetApi->api_v1_datasets_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**List[DtoDatasetBuildPayload]**](DtoDatasetBuildPayload.md)| 请求体 | 

### Return type

[**DtoGenericResponseDtoSubmitResp**](DtoGenericResponseDtoSubmitResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**202** | Accepted |  -  |
**400** | Bad Request |  -  |
**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_datasets_query_get**
> DtoGenericResponseDtoQueryDatasetResp api_v1_datasets_query_get(name, sort=sort)

查询单个数据集详情

根据数据集名称查询单个数据集的详细信息，包括检测器结果和执行记录

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_query_dataset_resp import DtoGenericResponseDtoQueryDatasetResp
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
    api_instance = rcabench.openapi.DatasetApi(api_client)
    name = 'name_example' # str | 数据集名称
    sort = 'sort_example' # str | 排序方式 (optional)

    try:
        # 查询单个数据集详情
        api_response = api_instance.api_v1_datasets_query_get(name, sort=sort)
        print("The response of DatasetApi->api_v1_datasets_query_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetApi->api_v1_datasets_query_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **str**| 数据集名称 | 
 **sort** | **str**| 排序方式 | [optional] 

### Return type

[**DtoGenericResponseDtoQueryDatasetResp**](DtoGenericResponseDtoQueryDatasetResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | OK |  -  |
**400** | Bad Request |  -  |
**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

