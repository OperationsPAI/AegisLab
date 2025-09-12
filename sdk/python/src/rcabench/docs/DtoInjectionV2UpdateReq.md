# DtoInjectionV2UpdateReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**benchmark** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**display_config** | **str** |  | [optional] 
**end_time** | **str** |  | [optional] 
**engine_config** | **str** |  | [optional] 
**fault_type** | **int** |  | [optional] 
**injection_name** | **str** |  | [optional] 
**pre_duration** | **int** |  | [optional] 
**start_time** | **str** |  | [optional] 
**status** | **int** |  | [optional] 
**task_id** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_v2_update_req import DtoInjectionV2UpdateReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionV2UpdateReq from a JSON string
dto_injection_v2_update_req_instance = DtoInjectionV2UpdateReq.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionV2UpdateReq.to_json())

# convert the object into a dict
dto_injection_v2_update_req_dict = dto_injection_v2_update_req_instance.to_dict()
# create an instance of DtoInjectionV2UpdateReq from a dict
dto_injection_v2_update_req_from_dict = DtoInjectionV2UpdateReq.from_dict(dto_injection_v2_update_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


