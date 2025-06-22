# DtoInjectionSubmitReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithms** | **List[str]** |  | [optional] 
**benchmark** | **str** |  | [optional] 
**direct** | **bool** |  | [optional] 
**interval** | **int** |  | [optional] 
**labels** | [**List[DtoLabelItem]**](DtoLabelItem.md) |  | [optional] 
**pre_duration** | **int** |  | [optional] 
**specs** | [**List[HandlerNode]**](HandlerNode.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_submit_req import DtoInjectionSubmitReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionSubmitReq from a JSON string
dto_injection_submit_req_instance = DtoInjectionSubmitReq.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionSubmitReq.to_json())

# convert the object into a dict
dto_injection_submit_req_dict = dto_injection_submit_req_instance.to_dict()
# create an instance of DtoInjectionSubmitReq from a dict
dto_injection_submit_req_from_dict = DtoInjectionSubmitReq.from_dict(dto_injection_submit_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


