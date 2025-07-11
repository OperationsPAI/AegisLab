# rcabench.openapi.AlgorithmApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_algorithms_get**](AlgorithmApi.md#api_v1_algorithms_get) | **GET** /api/v1/algorithms | 获取算法列表
[**api_v1_algorithms_post**](AlgorithmApi.md#api_v1_algorithms_post) | **POST** /api/v1/algorithms | 提交算法执行任务


# **api_v1_algorithms_get**
> DtoGenericResponseDtoListAlgorithmsResp api_v1_algorithms_get()

获取算法列表

获取系统中所有可用的算法列表，包括算法的镜像信息、标签和更新时间。只返回状态为激活的算法容器

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_list_algorithms_resp import DtoGenericResponseDtoListAlgorithmsResp
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
    api_instance = rcabench.openapi.AlgorithmApi(api_client)

    try:
        # 获取算法列表
        api_response = api_instance.api_v1_algorithms_get()
        print("The response of AlgorithmApi->api_v1_algorithms_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmApi->api_v1_algorithms_get: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**DtoGenericResponseDtoListAlgorithmsResp**](DtoGenericResponseDtoListAlgorithmsResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | 成功返回算法列表 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_algorithms_post**
> DtoGenericResponseDtoSubmitResp api_v1_algorithms_post(body)

提交算法执行任务

批量提交算法执行任务，支持多个算法和数据集的组合执行。系统将为每个执行任务分配唯一的 TraceID 用于跟踪任务状态和结果

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_execution_payload import DtoExecutionPayload
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
    api_instance = rcabench.openapi.AlgorithmApi(api_client)
    body = [rcabench.openapi.DtoExecutionPayload()] # List[DtoExecutionPayload] | 算法执行请求列表，包含算法名称、数据集和环境变量

    try:
        # 提交算法执行任务
        api_response = api_instance.api_v1_algorithms_post(body)
        print("The response of AlgorithmApi->api_v1_algorithms_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmApi->api_v1_algorithms_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**List[DtoExecutionPayload]**](DtoExecutionPayload.md)| 算法执行请求列表，包含算法名称、数据集和环境变量 | 

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
**202** | 成功提交算法执行任务，返回任务跟踪信息 |  -  |
**400** | 请求参数错误，如JSON格式不正确、算法名称或数据集名称无效、环境变量名称不支持等 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

