# DtoGenericResponseDtoListAlgorithmsResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**List[DatabaseContainer]**](DatabaseContainer.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_list_algorithms_resp import DtoGenericResponseDtoListAlgorithmsResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoListAlgorithmsResp from a JSON string
dto_generic_response_dto_list_algorithms_resp_instance = DtoGenericResponseDtoListAlgorithmsResp.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoListAlgorithmsResp.to_json())

# convert the object into a dict
dto_generic_response_dto_list_algorithms_resp_dict = dto_generic_response_dto_list_algorithms_resp_instance.to_dict()
# create an instance of DtoGenericResponseDtoListAlgorithmsResp from a dict
dto_generic_response_dto_list_algorithms_resp_from_dict = DtoGenericResponseDtoListAlgorithmsResp.from_dict(dto_generic_response_dto_list_algorithms_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


