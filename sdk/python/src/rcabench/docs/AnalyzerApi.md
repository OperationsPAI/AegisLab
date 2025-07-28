# rcabench.openapi.AnalyzerApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_analyzers_injections_get**](AnalyzerApi.md#api_v1_analyzers_injections_get) | **GET** /api/v1/analyzers/injections | 分析故障注入数据


# **api_v1_analyzers_injections_get**
> DtoGenericResponseDtoAnalyzeInjectionsResp api_v1_analyzers_injections_get(project_name=project_name, env=env, batch=batch, tag=tag, benchmark=benchmark, status=status, fault_type=fault_type, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)

分析故障注入数据

使用多种筛选条件分析故障注入数据，返回包括效率、多样性、种子之间的距离等统计信息

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_analyze_injections_resp import DtoGenericResponseDtoAnalyzeInjectionsResp
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
    api_instance = rcabench.openapi.AnalyzerApi(api_client)
    project_name = 'project_name_example' # str | 项目名称过滤 (optional)
    env = prod # str | 环境标签过滤 (optional) (default to prod)
    batch = 'batch_example' # str | 批次标签过滤 (optional)
    tag = train # str | 分类标签过滤 (optional) (default to train)
    benchmark = clickhouse # str | 基准测试类型过滤 (optional) (default to clickhouse)
    status = 0 # int | 状态过滤，具体值参考字段映射接口(/mapping) (optional) (default to 0)
    fault_type = 0 # int | 故障类型过滤，具体值参考字段映射接口(/mapping) (optional) (default to 0)
    lookback = 'lookback_example' # str | 时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置 (optional)
    custom_start_time = '2013-10-20T19:20:30+01:00' # datetime | 自定义开始时间，RFC3339格式，当lookback=custom时必需 (optional)
    custom_end_time = '2013-10-20T19:20:30+01:00' # datetime | 自定义结束时间，RFC3339格式，当lookback=custom时必需 (optional)

    try:
        # 分析故障注入数据
        api_response = api_instance.api_v1_analyzers_injections_get(project_name=project_name, env=env, batch=batch, tag=tag, benchmark=benchmark, status=status, fault_type=fault_type, lookback=lookback, custom_start_time=custom_start_time, custom_end_time=custom_end_time)
        print("The response of AnalyzerApi->api_v1_analyzers_injections_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AnalyzerApi->api_v1_analyzers_injections_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **project_name** | **str**| 项目名称过滤 | [optional] 
 **env** | **str**| 环境标签过滤 | [optional] [default to prod]
 **batch** | **str**| 批次标签过滤 | [optional] 
 **tag** | **str**| 分类标签过滤 | [optional] [default to train]
 **benchmark** | **str**| 基准测试类型过滤 | [optional] [default to clickhouse]
 **status** | **int**| 状态过滤，具体值参考字段映射接口(/mapping) | [optional] [default to 0]
 **fault_type** | **int**| 故障类型过滤，具体值参考字段映射接口(/mapping) | [optional] [default to 0]
 **lookback** | **str**| 时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置 | [optional] 
 **custom_start_time** | **datetime**| 自定义开始时间，RFC3339格式，当lookback&#x3D;custom时必需 | [optional] 
 **custom_end_time** | **datetime**| 自定义结束时间，RFC3339格式，当lookback&#x3D;custom时必需 | [optional] 

### Return type

[**DtoGenericResponseDtoAnalyzeInjectionsResp**](DtoGenericResponseDtoAnalyzeInjectionsResp.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | 返回故障注入分析统计信息 |  -  |
**400** | 请求参数错误，如参数格式不正确、验证失败等 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

