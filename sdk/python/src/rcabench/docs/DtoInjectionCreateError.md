# DtoInjectionCreateError


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**error** | **str** |  | [optional] 
**index** | **int** |  | [optional] 
**item** | [**DtoInjectionV2CreateItem**](DtoInjectionV2CreateItem.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_create_error import DtoInjectionCreateError

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionCreateError from a JSON string
dto_injection_create_error_instance = DtoInjectionCreateError.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionCreateError.to_json())

# convert the object into a dict
dto_injection_create_error_dict = dto_injection_create_error_instance.to_dict()
# create an instance of DtoInjectionCreateError from a dict
dto_injection_create_error_from_dict = DtoInjectionCreateError.from_dict(dto_injection_create_error_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


