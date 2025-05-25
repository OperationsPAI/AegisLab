# rcabench.openapi.AlgorithmApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_algorithms_build_post**](AlgorithmApi.md#api_v1_algorithms_build_post) | **POST** /api/v1/algorithms/build | 构建算法镜像
[**api_v1_algorithms_get**](AlgorithmApi.md#api_v1_algorithms_get) | **GET** /api/v1/algorithms | 获取算法列表
[**api_v1_algorithms_post**](AlgorithmApi.md#api_v1_algorithms_post) | **POST** /api/v1/algorithms | 执行算法


# **api_v1_algorithms_build_post**
> DtoGenericResponseDtoSubmitResp api_v1_algorithms_build_post(file=file, algo=algo)

构建算法镜像

通过上传文件或指定算法名称来构建算法镜像

### Example


```python
import rcabench.openapi
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
    file = None # bytearray | 算法文件 (zip/tar.gz) (optional)
    algo = 'algo_example' # str | 算法名称 (optional)

    try:
        # 构建算法镜像
        api_response = api_instance.api_v1_algorithms_build_post(file=file, algo=algo)
        print("The response of AlgorithmApi->api_v1_algorithms_build_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmApi->api_v1_algorithms_build_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **file** | **bytearray**| 算法文件 (zip/tar.gz) | [optional] 
 **algo** | **str**| 算法名称 | [optional] 

### Return type

[**DtoGenericResponseDtoSubmitResp**](DtoGenericResponseDtoSubmitResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**202** | Accepted |  -  |
**400** | Bad Request |  -  |
**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_algorithms_get**
> DtoGenericResponseDtoAlgorithmListResp api_v1_algorithms_get()

获取算法列表

获取算法列表

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_algorithm_list_resp import DtoGenericResponseDtoAlgorithmListResp
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

[**DtoGenericResponseDtoAlgorithmListResp**](DtoGenericResponseDtoAlgorithmListResp.md)

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

# **api_v1_algorithms_post**
> DtoGenericResponseDtoSubmitResp api_v1_algorithms_post(body)

执行算法

执行算法

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_algorithm_execution_payload import DtoAlgorithmExecutionPayload
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
    body = [rcabench.openapi.DtoAlgorithmExecutionPayload()] # List[DtoAlgorithmExecutionPayload] | 请求体

    try:
        # 执行算法
        api_response = api_instance.api_v1_algorithms_post(body)
        print("The response of AlgorithmApi->api_v1_algorithms_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmApi->api_v1_algorithms_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**List[DtoAlgorithmExecutionPayload]**](DtoAlgorithmExecutionPayload.md)| 请求体 | 

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

