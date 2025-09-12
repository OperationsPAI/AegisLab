# DtoGenericResponseDtoLabelResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**DtoLabelResponse**](DtoLabelResponse.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_label_response import DtoGenericResponseDtoLabelResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoLabelResponse from a JSON string
dto_generic_response_dto_label_response_instance = DtoGenericResponseDtoLabelResponse.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoLabelResponse.to_json())

# convert the object into a dict
dto_generic_response_dto_label_response_dict = dto_generic_response_dto_label_response_instance.to_dict()
# create an instance of DtoGenericResponseDtoLabelResponse from a dict
dto_generic_response_dto_label_response_from_dict = DtoGenericResponseDtoLabelResponse.from_dict(dto_generic_response_dto_label_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


