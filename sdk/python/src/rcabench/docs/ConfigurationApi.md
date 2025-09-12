# rcabench.openapi.ConfigurationApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_configurations_ns_index_get**](ConfigurationApi.md#api_v1_configurations_ns_index_get) | **GET** /api/v1/configurations/ns/{index} | 根据索引获取命名空间前缀
[**api_v1_configurations_ns_services_get**](ConfigurationApi.md#api_v1_configurations_ns_services_get) | **GET** /api/v1/configurations/ns-services | 获取命名空间服务映射


# **api_v1_configurations_ns_index_get**
> DtoGenericResponseString api_v1_configurations_ns_index_get(index)

根据索引获取命名空间前缀

根据索引获取对应的命名空间前缀。索引从0开始，按配置文件中的顺序返回对应的命名空间前缀

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_string import DtoGenericResponseString
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
    api_instance = rcabench.openapi.ConfigurationApi(api_client)
    index = 56 # int | 命名空间前缀索引，从0开始

    try:
        # 根据索引获取命名空间前缀
        api_response = api_instance.api_v1_configurations_ns_index_get(index)
        print("The response of ConfigurationApi->api_v1_configurations_ns_index_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ConfigurationApi->api_v1_configurations_ns_index_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **index** | **int**| 命名空间前缀索引，从0开始 | 

### Return type

[**DtoGenericResponseString**](DtoGenericResponseString.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | 成功返回命名空间前缀 |  -  |
**400** | 请求参数错误，索引参数无效或为负数 |  -  |
**404** | 索引超出范围，不存在对应的命名空间前缀 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_configurations_ns_services_get**
> DtoGenericResponseAny api_v1_configurations_ns_services_get(namespace=namespace)

获取命名空间服务映射

获取命名空间及其对应的服务列表映射关系。不指定namespace时返回所有命名空间的映射，指定namespace时返回该命名空间的服务列表

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny
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
    api_instance = rcabench.openapi.ConfigurationApi(api_client)
    namespace = 'namespace_example' # str | 命名空间名称，不指定时返回所有命名空间映射 (optional)

    try:
        # 获取命名空间服务映射
        api_response = api_instance.api_v1_configurations_ns_services_get(namespace=namespace)
        print("The response of ConfigurationApi->api_v1_configurations_ns_services_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ConfigurationApi->api_v1_configurations_ns_services_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **namespace** | **str**| 命名空间名称，不指定时返回所有命名空间映射 | [optional] 

### Return type

[**DtoGenericResponseAny**](DtoGenericResponseAny.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | 成功返回命名空间服务映射表或指定命名空间的服务列表 |  -  |
**400** | 请求参数错误 |  -  |
**404** | 指定的命名空间不存在 |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

