# DtoNumberRange


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**max** | **float** |  | [optional] 
**min** | **float** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_number_range import DtoNumberRange

# TODO update the JSON string below
json = "{}"
# create an instance of DtoNumberRange from a JSON string
dto_number_range_instance = DtoNumberRange.from_json(json)
# print the JSON string representation of the object
print(DtoNumberRange.to_json())

# convert the object into a dict
dto_number_range_dict = dto_number_range_instance.to_dict()
# create an instance of DtoNumberRange from a dict
dto_number_range_from_dict = DtoNumberRange.from_dict(dto_number_range_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


