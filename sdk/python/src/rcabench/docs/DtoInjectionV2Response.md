# DtoInjectionV2Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**benchmark** | **str** |  | [optional] 
**created_at** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**display_config** | **str** |  | [optional] 
**end_time** | **str** |  | [optional] 
**engine_config** | **str** |  | [optional] 
**fault_type** | **int** |  | [optional] 
**id** | **int** |  | [optional] 
**injection_name** | **str** |  | [optional] 
**labels** | [**List[DatabaseLabel]**](DatabaseLabel.md) | Associated labels | [optional] 
**pre_duration** | **int** |  | [optional] 
**start_time** | **str** |  | [optional] 
**status** | **int** |  | [optional] 
**task** | [**DtoTaskV2Response**](DtoTaskV2Response.md) | Optional relations | [optional] 
**task_id** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_v2_response import DtoInjectionV2Response

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionV2Response from a JSON string
dto_injection_v2_response_instance = DtoInjectionV2Response.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionV2Response.to_json())

# convert the object into a dict
dto_injection_v2_response_dict = dto_injection_v2_response_instance.to_dict()
# create an instance of DtoInjectionV2Response from a dict
dto_injection_v2_response_from_dict = DtoInjectionV2Response.from_dict(dto_injection_v2_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


