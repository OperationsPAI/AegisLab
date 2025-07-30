# DtoAlgorithmResultUploadResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm_id** | **int** |  | [optional] 
**execution_id** | **int** |  | [optional] 
**has_anomalies** | **bool** | Only included for detector results | [optional] 
**message** | **str** |  | [optional] 
**result_count** | **int** |  | [optional] 
**uploaded_at** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_algorithm_result_upload_response import DtoAlgorithmResultUploadResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAlgorithmResultUploadResponse from a JSON string
dto_algorithm_result_upload_response_instance = DtoAlgorithmResultUploadResponse.from_json(json)
# print the JSON string representation of the object
print(DtoAlgorithmResultUploadResponse.to_json())

# convert the object into a dict
dto_algorithm_result_upload_response_dict = dto_algorithm_result_upload_response_instance.to_dict()
# create an instance of DtoAlgorithmResultUploadResponse from a dict
dto_algorithm_result_upload_response_from_dict = DtoAlgorithmResultUploadResponse.from_dict(dto_algorithm_result_upload_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


