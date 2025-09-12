# DtoGenericResponseDtoSubmitInjectionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**DtoSubmitInjectionResp**](DtoSubmitInjectionResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_submit_injection_resp import DtoGenericResponseDtoSubmitInjectionResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoSubmitInjectionResp from a JSON string
dto_generic_response_dto_submit_injection_resp_instance = DtoGenericResponseDtoSubmitInjectionResp.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoSubmitInjectionResp.to_json())

# convert the object into a dict
dto_generic_response_dto_submit_injection_resp_dict = dto_generic_response_dto_submit_injection_resp_instance.to_dict()
# create an instance of DtoGenericResponseDtoSubmitInjectionResp from a dict
dto_generic_response_dto_submit_injection_resp_from_dict = DtoGenericResponseDtoSubmitInjectionResp.from_dict(dto_generic_response_dto_submit_injection_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


