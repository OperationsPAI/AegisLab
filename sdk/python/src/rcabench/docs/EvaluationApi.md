# rcabench.openapi.EvaluationApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_evaluations_get**](EvaluationApi.md#api_v1_evaluations_get) | **GET** /api/v1/evaluations | 获取每种算法的执行历史记录
[**api_v1_evaluations_raw_data_get**](EvaluationApi.md#api_v1_evaluations_raw_data_get) | **GET** /api/v1/evaluations/raw-data | 获取原始评估数据


# **api_v1_evaluations_get**
> DtoGenericResponseDtoEvaluationListResp api_v1_evaluations_get(execution_ids=execution_ids, algorithms=algorithms, levels=levels, metrics=metrics)

获取每种算法的执行历史记录

返回每种算法的执行历史记录

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_evaluation_list_resp import DtoGenericResponseDtoEvaluationListResp
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
    api_instance = rcabench.openapi.EvaluationApi(api_client)
    execution_ids = [56] # List[int] | 执行结果 ID 数组 (optional)
    algorithms = ['algorithms_example'] # List[str] | 算法名称数组 (optional)
    levels = ['levels_example'] # List[str] | 级别名称数组 (optional)
    metrics = ['metrics_example'] # List[str] | 指标名称数组 (optional)

    try:
        # 获取每种算法的执行历史记录
        api_response = api_instance.api_v1_evaluations_get(execution_ids=execution_ids, algorithms=algorithms, levels=levels, metrics=metrics)
        print("The response of EvaluationApi->api_v1_evaluations_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EvaluationApi->api_v1_evaluations_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **execution_ids** | [**List[int]**](int.md)| 执行结果 ID 数组 | [optional] 
 **algorithms** | [**List[str]**](str.md)| 算法名称数组 | [optional] 
 **levels** | [**List[str]**](str.md)| 级别名称数组 | [optional] 
 **metrics** | [**List[str]**](str.md)| 指标名称数组 | [optional] 

### Return type

[**DtoGenericResponseDtoEvaluationListResp**](DtoGenericResponseDtoEvaluationListResp.md)

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

# **api_v1_evaluations_raw_data_get**
> DtoGenericResponseArrayDtoRawDataItem api_v1_evaluations_raw_data_get(algorithms, datasets)

获取原始评估数据

根据算法和数据集的笛卡尔积获取对应的原始评估数据，包括粒度记录和真实值信息

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_array_dto_raw_data_item import DtoGenericResponseArrayDtoRawDataItem
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
    api_instance = rcabench.openapi.EvaluationApi(api_client)
    algorithms = ['algorithms_example'] # List[str] | 算法数组
    datasets = ['datasets_example'] # List[str] | 数据集数组

    try:
        # 获取原始评估数据
        api_response = api_instance.api_v1_evaluations_raw_data_get(algorithms, datasets)
        print("The response of EvaluationApi->api_v1_evaluations_raw_data_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EvaluationApi->api_v1_evaluations_raw_data_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **algorithms** | [**List[str]**](str.md)| 算法数组 | 
 **datasets** | [**List[str]**](str.md)| 数据集数组 | 

### Return type

[**DtoGenericResponseArrayDtoRawDataItem**](DtoGenericResponseArrayDtoRawDataItem.md)

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

