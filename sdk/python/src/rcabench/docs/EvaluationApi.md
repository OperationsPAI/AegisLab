# rcabench.openapi.EvaluationApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_evaluations_groundtruth_get**](EvaluationApi.md#api_v1_evaluations_groundtruth_get) | **GET** /api/v1/evaluations/groundtruth | 获取数据集的 ground truth
[**api_v1_evaluations_raw_data_get**](EvaluationApi.md#api_v1_evaluations_raw_data_get) | **GET** /api/v1/evaluations/raw-data | 获取原始评估数据


# **api_v1_evaluations_groundtruth_get**
> DtoGenericResponseDtoGroundTruthResp api_v1_evaluations_groundtruth_get(datasets)

获取数据集的 ground truth

根据数据集数组获取对应的 ground truth 数据

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_ground_truth_resp import DtoGenericResponseDtoGroundTruthResp
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
    datasets = ['datasets_example'] # List[str] | 数据集数组

    try:
        # 获取数据集的 ground truth
        api_response = api_instance.api_v1_evaluations_groundtruth_get(datasets)
        print("The response of EvaluationApi->api_v1_evaluations_groundtruth_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EvaluationApi->api_v1_evaluations_groundtruth_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **datasets** | [**List[str]**](str.md)| 数据集数组 | 

### Return type

[**DtoGenericResponseDtoGroundTruthResp**](DtoGenericResponseDtoGroundTruthResp.md)

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

