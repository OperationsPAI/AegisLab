# DtoInjectionFieldMappingResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**fault_resource** | **object** |  | [optional] 
**fault_type** | **object** |  | [optional] 
**status** | **object** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_field_mapping_resp import DtoInjectionFieldMappingResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionFieldMappingResp from a JSON string
dto_injection_field_mapping_resp_instance = DtoInjectionFieldMappingResp.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionFieldMappingResp.to_json())

# convert the object into a dict
dto_injection_field_mapping_resp_dict = dto_injection_field_mapping_resp_instance.to_dict()
# create an instance of DtoInjectionFieldMappingResp from a dict
dto_injection_field_mapping_resp_from_dict = DtoInjectionFieldMappingResp.from_dict(dto_injection_field_mapping_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


