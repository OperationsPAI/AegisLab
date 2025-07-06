# rcabench.openapi.EvaluationApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_evaluations_groundtruth_post**](EvaluationApi.md#api_v1_evaluations_groundtruth_post) | **POST** /api/v1/evaluations/groundtruth | 获取数据集的 ground truth
[**api_v1_evaluations_raw_data_post**](EvaluationApi.md#api_v1_evaluations_raw_data_post) | **POST** /api/v1/evaluations/raw-data | 获取原始评估数据


# **api_v1_evaluations_groundtruth_post**
> DtoGenericResponseDtoGroundTruthResp api_v1_evaluations_groundtruth_post(body)

获取数据集的 ground truth

根据数据集数组获取对应的 ground truth 数据，用于算法评估的基准数据。支持批量查询多个数据集的真实标签信息

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_ground_truth_resp import DtoGenericResponseDtoGroundTruthResp
from rcabench.openapi.models.dto_ground_truth_req import DtoGroundTruthReq
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
    body = rcabench.openapi.DtoGroundTruthReq() # DtoGroundTruthReq | Ground truth查询请求，包含数据集列表

    try:
        # 获取数据集的 ground truth
        api_response = api_instance.api_v1_evaluations_groundtruth_post(body)
        print("The response of EvaluationApi->api_v1_evaluations_groundtruth_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EvaluationApi->api_v1_evaluations_groundtruth_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**DtoGroundTruthReq**](DtoGroundTruthReq.md)| Ground truth查询请求，包含数据集列表 | 

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
**200** | 成功返回数据集的ground truth信息 |  -  |
**400** | 请求参数错误，如JSON格式不正确、数据集数组为空 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_evaluations_raw_data_post**
> DtoGenericResponseDtoRawDataResp api_v1_evaluations_raw_data_post(body)

获取原始评估数据

根据算法和数据集的笛卡尔积获取对应的原始评估数据，包括粒度记录和真实值信息。支持批量查询多个算法在多个数据集上的执行结果

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_raw_data_resp import DtoGenericResponseDtoRawDataResp
from rcabench.openapi.models.dto_raw_data_req import DtoRawDataReq
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
    body = rcabench.openapi.DtoRawDataReq() # DtoRawDataReq | 原始数据查询请求，包含算法列表和数据集列表

    try:
        # 获取原始评估数据
        api_response = api_instance.api_v1_evaluations_raw_data_post(body)
        print("The response of EvaluationApi->api_v1_evaluations_raw_data_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EvaluationApi->api_v1_evaluations_raw_data_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**DtoRawDataReq**](DtoRawDataReq.md)| 原始数据查询请求，包含算法列表和数据集列表 | 

### Return type

[**DtoGenericResponseDtoRawDataResp**](DtoGenericResponseDtoRawDataResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | 成功返回原始评估数据列表 |  -  |
**400** | 请求参数错误，如JSON格式不正确、算法或数据集数组为空 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

