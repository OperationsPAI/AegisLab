# DtoGenericResponseDtoListInjectionsResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**DtoListInjectionsResp**](DtoListInjectionsResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_list_injections_resp import DtoGenericResponseDtoListInjectionsResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoListInjectionsResp from a JSON string
dto_generic_response_dto_list_injections_resp_instance = DtoGenericResponseDtoListInjectionsResp.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoListInjectionsResp.to_json())

# convert the object into a dict
dto_generic_response_dto_list_injections_resp_dict = dto_generic_response_dto_list_injections_resp_instance.to_dict()
# create an instance of DtoGenericResponseDtoListInjectionsResp from a dict
dto_generic_response_dto_list_injections_resp_from_dict = DtoGenericResponseDtoListInjectionsResp.from_dict(dto_generic_response_dto_list_injections_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


