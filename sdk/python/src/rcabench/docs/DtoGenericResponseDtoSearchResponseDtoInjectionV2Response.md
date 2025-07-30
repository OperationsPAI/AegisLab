# DtoGenericResponseDtoSearchResponseDtoInjectionV2Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**DtoSearchResponseDtoInjectionV2Response**](DtoSearchResponseDtoInjectionV2Response.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_search_response_dto_injection_v2_response import DtoGenericResponseDtoSearchResponseDtoInjectionV2Response

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoSearchResponseDtoInjectionV2Response from a JSON string
dto_generic_response_dto_search_response_dto_injection_v2_response_instance = DtoGenericResponseDtoSearchResponseDtoInjectionV2Response.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoSearchResponseDtoInjectionV2Response.to_json())

# convert the object into a dict
dto_generic_response_dto_search_response_dto_injection_v2_response_dict = dto_generic_response_dto_search_response_dto_injection_v2_response_instance.to_dict()
# create an instance of DtoGenericResponseDtoSearchResponseDtoInjectionV2Response from a dict
dto_generic_response_dto_search_response_dto_injection_v2_response_from_dict = DtoGenericResponseDtoSearchResponseDtoInjectionV2Response.from_dict(dto_generic_response_dto_search_response_dto_injection_v2_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


