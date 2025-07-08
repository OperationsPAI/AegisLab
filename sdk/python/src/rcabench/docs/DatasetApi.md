# rcabench.openapi.DatasetApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_datasets_delete**](DatasetApi.md#api_v1_datasets_delete) | **DELETE** /api/v1/datasets | 删除数据集数据
[**api_v1_datasets_download_get**](DatasetApi.md#api_v1_datasets_download_get) | **GET** /api/v1/datasets/download | 下载数据集打包文件
[**api_v1_datasets_post**](DatasetApi.md#api_v1_datasets_post) | **POST** /api/v1/datasets | 批量构建数据集


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

将指定的多个数据集打包为 ZIP 文件下载，自动排除 result.csv 和检测器结论文件。支持按组ID或数据集名称进行下载，两种方式二选一。下载文件结构：按组ID下载时为 datasets/{groupId}/{datasetName}/...，按名称下载时为 datasets/{datasetName}/...

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
    group_ids = ['group_ids_example'] # List[str] | 任务组ID列表，格式：group1,group2,group3。与names参数二选一，优先使用group_ids (optional)
    names = ['names_example'] # List[str] | 数据集名称列表，格式：dataset1,dataset2,dataset3。与group_ids参数二选一 (optional)

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
 **group_ids** | [**List[str]**](str.md)| 任务组ID列表，格式：group1,group2,group3。与names参数二选一，优先使用group_ids | [optional] 
 **names** | [**List[str]**](str.md)| 数据集名称列表，格式：dataset1,dataset2,dataset3。与group_ids参数二选一 | [optional] 

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
**200** | ZIP 文件流，Content-Disposition 头中包含文件名 datasets.zip |  -  |
**400** | 请求参数错误：1) 参数绑定失败 2) 两个参数都为空 3) 同时提供两种参数 |  -  |
**403** | 权限错误：请求访问的数据集路径不在系统允许的范围内 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_datasets_post**
> DtoGenericResponseDtoSubmitResp api_v1_datasets_post(body)

批量构建数据集

根据指定的时间范围和基准测试容器批量构建数据集。

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
    body = [rcabench.openapi.DtoDatasetBuildPayload()] # List[DtoDatasetBuildPayload] | 数据集构建请求列表，每个请求包含数据集名称、时间范围、基准测试和环境变量配置

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
 **body** | [**List[DtoDatasetBuildPayload]**](DtoDatasetBuildPayload.md)| 数据集构建请求列表，每个请求包含数据集名称、时间范围、基准测试和环境变量配置 | 

### Return type

[**DtoGenericResponseDtoSubmitResp**](DtoGenericResponseDtoSubmitResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**202** | 成功提交数据集构建任务，返回任务组ID和跟踪信息列表 |  -  |
**400** | 请求参数错误：1) JSON格式不正确 2) 数据集名称为空 3) 时间范围无效 4) 基准测试不存在 5) 环境变量名称不支持 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

