# InjectionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**benchmark_id** | **int** |  | [optional] 
**benchmark_name** | **str** |  | [optional] 
**created_at** | **str** |  | [optional] 
**display_config** | **str** |  | [optional] 
**end_time** | **str** |  | [optional] 
**fault_type** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**labels** | [**List[LabelItem]**](LabelItem.md) |  | [optional] 
**name** | **str** |  | [optional] 
**pedestal_id** | **int** |  | [optional] 
**pedestal_name** | **str** |  | [optional] 
**pre_duration** | **int** |  | [optional] 
**start_time** | **str** |  | [optional] 
**state** | **str** |  | [optional] 
**status** | **str** |  | [optional] 
**task_id** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from openapi.models.injection_resp import InjectionResp

# TODO update the JSON string below
json = "{}"
# create an instance of InjectionResp from a JSON string
injection_resp_instance = InjectionResp.from_json(json)
# print the JSON string representation of the object
print(InjectionResp.to_json())

# convert the object into a dict
injection_resp_dict = injection_resp_instance.to_dict()
# create an instance of InjectionResp from a dict
injection_resp_from_dict = InjectionResp.from_dict(injection_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


