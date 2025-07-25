# DtoFaultInjectionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**batch** | **str** |  | [optional] 
**benchmark** | **str** |  | [optional] 
**created_at** | **str** |  | [optional] 
**display_config** | **str** |  | [optional] 
**end_time** | **str** |  | [optional] 
**engine_config** | **str** |  | [optional] 
**env** | **str** |  | [optional] 
**fault_type** | **int** |  | [optional] 
**id** | **int** |  | [optional] 
**injection_name** | **str** |  | [optional] 
**pre_duration** | **int** |  | [optional] 
**start_time** | **str** |  | [optional] 
**status** | **int** |  | [optional] 
**tag** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_fault_injection_resp import DtoFaultInjectionResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoFaultInjectionResp from a JSON string
dto_fault_injection_resp_instance = DtoFaultInjectionResp.from_json(json)
# print the JSON string representation of the object
print(DtoFaultInjectionResp.to_json())

# convert the object into a dict
dto_fault_injection_resp_dict = dto_fault_injection_resp_instance.to_dict()
# create an instance of DtoFaultInjectionResp from a dict
dto_fault_injection_resp_from_dict = DtoFaultInjectionResp.from_dict(dto_fault_injection_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


