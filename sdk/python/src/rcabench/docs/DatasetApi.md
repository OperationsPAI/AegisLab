# rcabench.openapi.DatasetApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v1_datasets_delete**](DatasetApi.md#api_v1_datasets_delete) | **DELETE** /api/v1/datasets | Delete dataset data
[**api_v1_datasets_download_get**](DatasetApi.md#api_v1_datasets_download_get) | **GET** /api/v1/datasets/download | Download dataset archive file
[**api_v1_datasets_post**](DatasetApi.md#api_v1_datasets_post) | **POST** /api/v1/datasets | Batch build datasets


# **api_v1_datasets_delete**
> DtoGenericResponseDtoDatasetDeleteResp api_v1_datasets_delete(names)

Delete dataset data

Delete dataset data

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
    names = ['names_example'] # List[str] | Dataset name list

    try:
        # Delete dataset data
        api_response = api_instance.api_v1_datasets_delete(names)
        print("The response of DatasetApi->api_v1_datasets_delete:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetApi->api_v1_datasets_delete: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **names** | [**List[str]**](str.md)| Dataset name list | 

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

Download dataset archive file

Package specified datasets into a ZIP file for download, automatically excluding result.csv and detector conclusion files. Supports downloading by group ID or dataset name (mutually exclusive). Directory structure: when downloading by group ID: datasets/{groupId}/{datasetName}/...; when by name: datasets/{datasetName}/...

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
    group_ids = ['group_ids_example'] # List[str] | List of task group IDs, format: group1,group2,group3. Mutually exclusive with names parameter; group_ids takes precedence (optional)
    names = ['names_example'] # List[str] | List of dataset names, format: dataset1,dataset2,dataset3. Mutually exclusive with group_ids parameter (optional)

    try:
        # Download dataset archive file
        api_response = api_instance.api_v1_datasets_download_get(group_ids=group_ids, names=names)
        print("The response of DatasetApi->api_v1_datasets_download_get:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetApi->api_v1_datasets_download_get: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **group_ids** | [**List[str]**](str.md)| List of task group IDs, format: group1,group2,group3. Mutually exclusive with names parameter; group_ids takes precedence | [optional] 
 **names** | [**List[str]**](str.md)| List of dataset names, format: dataset1,dataset2,dataset3. Mutually exclusive with group_ids parameter | [optional] 

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
**200** | ZIP file stream; the Content-Disposition header contains filename datasets.zip |  -  |
**400** | Bad request parameters: 1) Parameter binding failed 2) Both parameters are empty 3) Both parameters provided |  -  |
**403** | Permission error: requested dataset path is not within allowed scope |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **api_v1_datasets_post**
> DtoGenericResponseDtoSubmitResp api_v1_datasets_post(body)

Batch build datasets

Batch build datasets based on specified time range and benchmark container

### Example


```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_submit_resp import DtoGenericResponseDtoSubmitResp
from rcabench.openapi.models.dto_submit_dataset_building_req import DtoSubmitDatasetBuildingReq
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
    body = rcabench.openapi.DtoSubmitDatasetBuildingReq() # DtoSubmitDatasetBuildingReq | List of dataset build requests; each request includes dataset name, time range, benchmark, and environment variable configuration

    try:
        # Batch build datasets
        api_response = api_instance.api_v1_datasets_post(body)
        print("The response of DatasetApi->api_v1_datasets_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DatasetApi->api_v1_datasets_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**DtoSubmitDatasetBuildingReq**](DtoSubmitDatasetBuildingReq.md)| List of dataset build requests; each request includes dataset name, time range, benchmark, and environment variable configuration | 

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
**202** | Successfully submitted dataset building tasks; returns group ID and trace information list |  -  |
**400** | Bad request parameters: 1) Invalid JSON format 2) Empty dataset name 3) Invalid time range 4) Benchmark does not exist 5) Unsupported environment variable name |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

