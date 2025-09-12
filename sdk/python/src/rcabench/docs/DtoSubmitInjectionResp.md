# DtoSubmitInjectionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**duplicated_count** | **int** |  | [optional] 
**group_id** | **str** |  | [optional] 
**original_count** | **int** |  | [optional] 
**traces** | [**List[DtoTrace]**](DtoTrace.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_submit_injection_resp import DtoSubmitInjectionResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoSubmitInjectionResp from a JSON string
dto_submit_injection_resp_instance = DtoSubmitInjectionResp.from_json(json)
# print the JSON string representation of the object
print(DtoSubmitInjectionResp.to_json())

# convert the object into a dict
dto_submit_injection_resp_dict = dto_submit_injection_resp_instance.to_dict()
# create an instance of DtoSubmitInjectionResp from a dict
dto_submit_injection_resp_from_dict = DtoSubmitInjectionResp.from_dict(dto_submit_injection_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


