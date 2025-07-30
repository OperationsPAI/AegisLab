# DtoGenericResponseDtoAlgorithmResultUploadResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**DtoAlgorithmResultUploadResponse**](DtoAlgorithmResultUploadResponse.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_algorithm_result_upload_response import DtoGenericResponseDtoAlgorithmResultUploadResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoAlgorithmResultUploadResponse from a JSON string
dto_generic_response_dto_algorithm_result_upload_response_instance = DtoGenericResponseDtoAlgorithmResultUploadResponse.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoAlgorithmResultUploadResponse.to_json())

# convert the object into a dict
dto_generic_response_dto_algorithm_result_upload_response_dict = dto_generic_response_dto_algorithm_result_upload_response_instance.to_dict()
# create an instance of DtoGenericResponseDtoAlgorithmResultUploadResponse from a dict
dto_generic_response_dto_algorithm_result_upload_response_from_dict = DtoGenericResponseDtoAlgorithmResultUploadResponse.from_dict(dto_generic_response_dto_algorithm_result_upload_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


