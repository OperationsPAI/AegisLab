# DtoInjectionRef


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **int** | Injection ID | [optional] 
**name** | **str** | Injection name | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_ref import DtoInjectionRef

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionRef from a JSON string
dto_injection_ref_instance = DtoInjectionRef.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionRef.to_json())

# convert the object into a dict
dto_injection_ref_dict = dto_injection_ref_instance.to_dict()
# create an instance of DtoInjectionRef from a dict
dto_injection_ref_from_dict = DtoInjectionRef.from_dict(dto_injection_ref_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


