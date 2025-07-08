# rcabench.openapi.AlgorithmApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_algorithms_build_post**](AlgorithmApi.md#api_v1_algorithms_build_post) | **POST** /api/v1/algorithms/build | 提交算法构建任务
[**api_v1_algorithms_get**](AlgorithmApi.md#api_v1_algorithms_get) | **GET** /api/v1/algorithms | 获取算法列表
[**api_v1_algorithms_post**](AlgorithmApi.md#api_v1_algorithms_post) | **POST** /api/v1/algorithms | 提交算法执行任务


# **api_v1_algorithms_build_post**
> DtoGenericResponseDtoSubmitResp api_v1_algorithms_build_post(algorithm, image, tag=tag, command=command, source_type=source_type, file=file, github_token=github_token, github_repo=github_repo, github_branch=github_branch, github_commit=github_commit, github_path=github_path, context_dir=context_dir, dockerfile_path=dockerfile_path, target=target, force_rebuild=force_rebuild)

提交算法构建任务

通过上传文件或指定GitHub仓库来构建算法Docker镜像。支持zip和tar.gz格式的文件上传，或从GitHub仓库自动拉取代码进行构建。系统会自动验证必需文件（Dockerfile）并设置执行权限

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
    algorithm = 'algorithm_example' # str | 算法名称，用于标识算法，将作为镜像构建的标识符
    image = 'image_example' # str | Docker镜像名称。支持以下格式：1) image-name（自动添加默认Harbor地址和命名空间）2) namespace/image-name（自动添加默认Harbor地址）
    tag = 'latest' # str | Docker镜像标签，用于版本控制 (optional) (default to 'latest')
    command = 'bash /entrypoint.sh' # str | Docker镜像启动命令，默认为bash /entrypoint.sh (optional) (default to 'bash /entrypoint.sh')
    source_type = file # str | 构建源类型，指定算法源码来源 (optional) (default to file)
    file = None # bytearray | 算法源码文件（支持zip或tar.gz格式），当source_type为file时必需，文件大小限制5MB (optional)
    github_token = 'github_token_example' # str | GitHub访问令牌，用于访问私有仓库，公开仓库可不提供 (optional)
    github_repo = 'github_repo_example' # str | GitHub仓库地址，格式：owner/repo，当source_type为github时必需 (optional)
    github_branch = 'main' # str | GitHub分支名，指定要构建的分支 (optional) (default to 'main')
    github_commit = 'github_commit_example' # str | GitHub commit哈希值（支持短hash），如果指定commit则忽略branch参数 (optional)
    github_path = '.' # str | 仓库内的子目录路径，如果算法源码不在根目录 (optional) (default to '.')
    context_dir = '.' # str | Docker构建上下文路径，相对于源码根目录 (optional) (default to '.')
    dockerfile_path = 'Dockerfile' # str | Dockerfile路径，相对于源码根目录 (optional) (default to 'Dockerfile')
    target = 'target_example' # str | Dockerfile构建目标（multi-stage build时使用） (optional)
    force_rebuild = False # bool | 是否强制重新构建镜像，忽略缓存 (optional) (default to False)

    try:
        # 提交算法构建任务
        api_response = api_instance.api_v1_algorithms_build_post(algorithm, image, tag=tag, command=command, source_type=source_type, file=file, github_token=github_token, github_repo=github_repo, github_branch=github_branch, github_commit=github_commit, github_path=github_path, context_dir=context_dir, dockerfile_path=dockerfile_path, target=target, force_rebuild=force_rebuild)
        print("The response of AlgorithmApi->api_v1_algorithms_build_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AlgorithmApi->api_v1_algorithms_build_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **algorithm** | **str**| 算法名称，用于标识算法，将作为镜像构建的标识符 | 
 **image** | **str**| Docker镜像名称。支持以下格式：1) image-name（自动添加默认Harbor地址和命名空间）2) namespace/image-name（自动添加默认Harbor地址） | 
 **tag** | **str**| Docker镜像标签，用于版本控制 | [optional] [default to &#39;latest&#39;]
 **command** | **str**| Docker镜像启动命令，默认为bash /entrypoint.sh | [optional] [default to &#39;bash /entrypoint.sh&#39;]
 **source_type** | **str**| 构建源类型，指定算法源码来源 | [optional] [default to file]
 **file** | **bytearray**| 算法源码文件（支持zip或tar.gz格式），当source_type为file时必需，文件大小限制5MB | [optional] 
 **github_token** | **str**| GitHub访问令牌，用于访问私有仓库，公开仓库可不提供 | [optional] 
 **github_repo** | **str**| GitHub仓库地址，格式：owner/repo，当source_type为github时必需 | [optional] 
 **github_branch** | **str**| GitHub分支名，指定要构建的分支 | [optional] [default to &#39;main&#39;]
 **github_commit** | **str**| GitHub commit哈希值（支持短hash），如果指定commit则忽略branch参数 | [optional] 
 **github_path** | **str**| 仓库内的子目录路径，如果算法源码不在根目录 | [optional] [default to &#39;.&#39;]
 **context_dir** | **str**| Docker构建上下文路径，相对于源码根目录 | [optional] [default to &#39;.&#39;]
 **dockerfile_path** | **str**| Dockerfile路径，相对于源码根目录 | [optional] [default to &#39;Dockerfile&#39;]
 **target** | **str**| Dockerfile构建目标（multi-stage build时使用） | [optional] 
 **force_rebuild** | **bool**| 是否强制重新构建镜像，忽略缓存 | [optional] [default to False]

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
**202** | 成功提交算法构建任务，返回任务跟踪信息 |  -  |
**400** | 请求参数错误：文件格式不支持（仅支持zip、tar.gz）、文件大小超限（5MB）、参数验证失败、GitHub仓库地址无效、force_rebuild值格式错误等 |  -  |
**404** | 资源不存在：构建上下文路径不存在、缺少必需文件（Dockerfile、entrypoint.sh） |  -  |
**500** | 服务器内部错误 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

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

