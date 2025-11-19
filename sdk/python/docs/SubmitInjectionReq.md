# SubmitInjectionReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithms** | [**List[ContainerSpec]**](ContainerSpec.md) |  | [optional] 
**benchmark** | [**ContainerSpec**](ContainerSpec.md) |  | 
**interval** | **int** |  | 
**labels** | [**List[LabelItem]**](LabelItem.md) |  | [optional] 
**pedestal** | [**ContainerSpec**](ContainerSpec.md) |  | 
**pre_duration** | **int** |  | 
**project_name** | **str** |  | 
**specs** | [**List[ChaosNode]**](ChaosNode.md) |  | 

## Example

```python
from openapi.models.submit_injection_req import SubmitInjectionReq

# TODO update the JSON string below
json = "{}"
# create an instance of SubmitInjectionReq from a JSON string
submit_injection_req_instance = SubmitInjectionReq.from_json(json)
# print the JSON string representation of the object
print(SubmitInjectionReq.to_json())

# convert the object into a dict
submit_injection_req_dict = submit_injection_req_instance.to_dict()
# create an instance of SubmitInjectionReq from a dict
submit_injection_req_from_dict = SubmitInjectionReq.from_dict(submit_injection_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


