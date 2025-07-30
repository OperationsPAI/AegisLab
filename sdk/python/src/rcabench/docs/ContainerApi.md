# rcabench.openapi.ContainerApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_containers_post**](ContainerApi.md#api_v1_containers_post) | **POST** /api/v1/containers | Submit container build task


# **api_v1_containers_post**
> DtoGenericResponseDtoSubmitResp api_v1_containers_post(image, type=type, name=name, tag=tag, command=command, env_vars=env_vars, source_type=source_type, file=file, github_token=github_token, github_repo=github_repo, github_branch=github_branch, github_commit=github_commit, github_path=github_path, context_dir=context_dir, dockerfile_path=dockerfile_path, target=target, force_rebuild=force_rebuild)

Submit container build task

Build Docker images by uploading files, specifying GitHub repositories, or Harbor images. Supports zip and tar.gz file uploads, or automatically pulls code from GitHub for building, or directly updates the database from existing Harbor images. The system automatically validates required files (Dockerfile) and sets execution permissions.

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
    api_instance = rcabench.openapi.ContainerApi(api_client)
    image = 'image_example' # str | Docker image name. When source_type is harbor, specify the existing image name in Harbor; otherwise, supports the following formats: 1) image-name (automatically adds default Harbor address and namespace) 2) namespace/image-name (automatically adds default Harbor address)
    type = algorithm # str | Container type, specifies the purpose of the container (optional) (default to algorithm)
    name = 'name_example' # str | Container name, used to identify the container, will be used as the image build identifier, defaults to the name field in info.toml (optional)
    tag = 'latest' # str | Docker image tag. When source_type is harbor, specify the existing image tag in Harbor; otherwise, used for version control (optional) (default to 'latest')
    command = 'bash /entrypoint.sh' # str | Docker image startup command, defaults to bash /entrypoint.sh (optional) (default to 'bash /entrypoint.sh')
    env_vars = ['env_vars_example'] # List[str] | List of environment variable names, supports multiple variables (optional)
    source_type = file # str | Build source type, specifies the source of the code (optional) (default to file)
    file = None # bytearray | Source file (supports zip or tar.gz format), required when source_type is file, file size limit 5MB (optional)
    github_token = 'github_token_example' # str | GitHub access token, used for private repositories, not required for public repositories (optional)
    github_repo = 'github_repo_example' # str | GitHub repository address, format: owner/repo, required when source_type is github (optional)
    github_branch = 'main' # str | GitHub branch name, specifies the branch to build (optional) (default to 'main')
    github_commit = 'github_commit_example' # str | GitHub commit hash (supports short hash), if specified, branch parameter is ignored (optional)
    github_path = '.' # str | Subdirectory path in the repository, if the source code is not in the root directory (optional) (default to '.')
    context_dir = '.' # str | Docker build context path, relative to the source root directory (optional) (default to '.')
    dockerfile_path = 'Dockerfile' # str | Dockerfile path, relative to the source root directory (optional) (default to 'Dockerfile')
    target = 'target_example' # str | Dockerfile build target (used for multi-stage builds) (optional)
    force_rebuild = False # bool | Whether to force rebuild the image, ignore cache (optional) (default to False)

    try:
        # Submit container build task
        api_response = api_instance.api_v1_containers_post(image, type=type, name=name, tag=tag, command=command, env_vars=env_vars, source_type=source_type, file=file, github_token=github_token, github_repo=github_repo, github_branch=github_branch, github_commit=github_commit, github_path=github_path, context_dir=context_dir, dockerfile_path=dockerfile_path, target=target, force_rebuild=force_rebuild)
        print("The response of ContainerApi->api_v1_containers_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ContainerApi->api_v1_containers_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **image** | **str**| Docker image name. When source_type is harbor, specify the existing image name in Harbor; otherwise, supports the following formats: 1) image-name (automatically adds default Harbor address and namespace) 2) namespace/image-name (automatically adds default Harbor address) | 
 **type** | **str**| Container type, specifies the purpose of the container | [optional] [default to algorithm]
 **name** | **str**| Container name, used to identify the container, will be used as the image build identifier, defaults to the name field in info.toml | [optional] 
 **tag** | **str**| Docker image tag. When source_type is harbor, specify the existing image tag in Harbor; otherwise, used for version control | [optional] [default to &#39;latest&#39;]
 **command** | **str**| Docker image startup command, defaults to bash /entrypoint.sh | [optional] [default to &#39;bash /entrypoint.sh&#39;]
 **env_vars** | [**List[str]**](str.md)| List of environment variable names, supports multiple variables | [optional] 
 **source_type** | **str**| Build source type, specifies the source of the code | [optional] [default to file]
 **file** | **bytearray**| Source file (supports zip or tar.gz format), required when source_type is file, file size limit 5MB | [optional] 
 **github_token** | **str**| GitHub access token, used for private repositories, not required for public repositories | [optional] 
 **github_repo** | **str**| GitHub repository address, format: owner/repo, required when source_type is github | [optional] 
 **github_branch** | **str**| GitHub branch name, specifies the branch to build | [optional] [default to &#39;main&#39;]
 **github_commit** | **str**| GitHub commit hash (supports short hash), if specified, branch parameter is ignored | [optional] 
 **github_path** | **str**| Subdirectory path in the repository, if the source code is not in the root directory | [optional] [default to &#39;.&#39;]
 **context_dir** | **str**| Docker build context path, relative to the source root directory | [optional] [default to &#39;.&#39;]
 **dockerfile_path** | **str**| Dockerfile path, relative to the source root directory | [optional] [default to &#39;Dockerfile&#39;]
 **target** | **str**| Dockerfile build target (used for multi-stage builds) | [optional] 
 **force_rebuild** | **bool**| Whether to force rebuild the image, ignore cache | [optional] [default to False]

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
**202** | Successfully submitted container build task, returns task tracking information |  -  |
**400** | Request parameter error: unsupported file format (only zip, tar.gz), file size exceeds limit (5MB), parameter validation failed, invalid GitHub repository address, invalid Harbor image parameter, invalid force_rebuild value, etc. |  -  |
**404** | Resource not found: build context path does not exist, missing required files (Dockerfile, entrypoint.sh), image not found in Harbor |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

