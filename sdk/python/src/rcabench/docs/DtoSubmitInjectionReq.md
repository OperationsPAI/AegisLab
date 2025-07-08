# DtoSubmitInjectionReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithms** | [**List[DtoAlgorithmItem]**](DtoAlgorithmItem.md) |  | [optional] 
**benchmark** | **str** |  | 
**interval** | **int** |  | 
**labels** | [**List[DtoLabelItem]**](DtoLabelItem.md) |  | [optional] 
**pre_duration** | **int** |  | 
**specs** | [**List[HandlerNode]**](HandlerNode.md) |  | 

## Example

```python
from rcabench.openapi.models.dto_submit_injection_req import DtoSubmitInjectionReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoSubmitInjectionReq from a JSON string
dto_submit_injection_req_instance = DtoSubmitInjectionReq.from_json(json)
# print the JSON string representation of the object
print(DtoSubmitInjectionReq.to_json())

# convert the object into a dict
dto_submit_injection_req_dict = dto_submit_injection_req_instance.to_dict()
# create an instance of DtoSubmitInjectionReq from a dict
dto_submit_injection_req_from_dict = DtoSubmitInjectionReq.from_dict(dto_submit_injection_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


