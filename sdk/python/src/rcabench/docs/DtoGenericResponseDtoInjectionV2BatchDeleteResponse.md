# DtoGenericResponseDtoInjectionV2BatchDeleteResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**DtoInjectionV2BatchDeleteResponse**](DtoInjectionV2BatchDeleteResponse.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_injection_v2_batch_delete_response import DtoGenericResponseDtoInjectionV2BatchDeleteResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoInjectionV2BatchDeleteResponse from a JSON string
dto_generic_response_dto_injection_v2_batch_delete_response_instance = DtoGenericResponseDtoInjectionV2BatchDeleteResponse.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoInjectionV2BatchDeleteResponse.to_json())

# convert the object into a dict
dto_generic_response_dto_injection_v2_batch_delete_response_dict = dto_generic_response_dto_injection_v2_batch_delete_response_instance.to_dict()
# create an instance of DtoGenericResponseDtoInjectionV2BatchDeleteResponse from a dict
dto_generic_response_dto_injection_v2_batch_delete_response_from_dict = DtoGenericResponseDtoInjectionV2BatchDeleteResponse.from_dict(dto_generic_response_dto_injection_v2_batch_delete_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


