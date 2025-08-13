# rcabench.openapi.LabelsApi

All URIs are relative to *http://localhost:8080/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**api_v2_labels_post**](LabelsApi.md#api_v2_labels_post) | **POST** /api/v2/labels | Create label


# **api_v2_labels_post**
> DtoGenericResponseDtoLabelResponse api_v2_labels_post(label)

Create label

Create a new label with key-value pair. If a deleted label with same key-value exists, it will be restored and updated.

### Example

* Api Key Authentication (BearerAuth):

```python
import rcabench.openapi
from rcabench.openapi.models.dto_generic_response_dto_label_response import DtoGenericResponseDtoLabelResponse
from rcabench.openapi.models.dto_label_create_req import DtoLabelCreateReq
from rcabench.openapi.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost:8080/api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = rcabench.openapi.Configuration(
    host = "http://localhost:8080/api/v1"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure API key authorization: BearerAuth
configuration.api_key['BearerAuth'] = os.environ["API_KEY"]

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['BearerAuth'] = 'Bearer'

# Enter a context with an instance of the API client
with rcabench.openapi.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = rcabench.openapi.LabelsApi(api_client)
    label = rcabench.openapi.DtoLabelCreateReq() # DtoLabelCreateReq | Label creation request

    try:
        # Create label
        api_response = api_instance.api_v2_labels_post(label)
        print("The response of LabelsApi->api_v2_labels_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling LabelsApi->api_v2_labels_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **label** | [**DtoLabelCreateReq**](DtoLabelCreateReq.md)| Label creation request | 

### Return type

[**DtoGenericResponseDtoLabelResponse**](DtoGenericResponseDtoLabelResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | Label created successfully |  -  |
**400** | Invalid request |  -  |
**409** | Label already exists |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

