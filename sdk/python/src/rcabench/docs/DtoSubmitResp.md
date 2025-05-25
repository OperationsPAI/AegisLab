# DtoSubmitResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**group_id** | **str** |  | [optional] 
**traces** | [**List[DtoTrace]**](DtoTrace.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_submit_resp import DtoSubmitResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoSubmitResp from a JSON string
dto_submit_resp_instance = DtoSubmitResp.from_json(json)
# print the JSON string representation of the object
print(DtoSubmitResp.to_json())

# convert the object into a dict
dto_submit_resp_dict = dto_submit_resp_instance.to_dict()
# create an instance of DtoSubmitResp from a dict
dto_submit_resp_from_dict = DtoSubmitResp.from_dict(dto_submit_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


