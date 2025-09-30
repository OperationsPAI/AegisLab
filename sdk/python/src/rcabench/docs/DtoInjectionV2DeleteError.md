# DtoInjectionV2DeleteError


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**error** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**injection_name** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_v2_delete_error import DtoInjectionV2DeleteError

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionV2DeleteError from a JSON string
dto_injection_v2_delete_error_instance = DtoInjectionV2DeleteError.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionV2DeleteError.to_json())

# convert the object into a dict
dto_injection_v2_delete_error_dict = dto_injection_v2_delete_error_instance.to_dict()
# create an instance of DtoInjectionV2DeleteError from a dict
dto_injection_v2_delete_error_from_dict = DtoInjectionV2DeleteError.from_dict(dto_injection_v2_delete_error_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


