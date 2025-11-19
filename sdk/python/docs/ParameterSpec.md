# ParameterSpec


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**key** | **str** |  | [optional] 
**value** | **object** |  | [optional] 

## Example

```python
from openapi.models.parameter_spec import ParameterSpec

# TODO update the JSON string below
json = "{}"
# create an instance of ParameterSpec from a JSON string
parameter_spec_instance = ParameterSpec.from_json(json)
# print the JSON string representation of the object
print(ParameterSpec.to_json())

# convert the object into a dict
parameter_spec_dict = parameter_spec_instance.to_dict()
# create an instance of ParameterSpec from a dict
parameter_spec_from_dict = ParameterSpec.from_dict(parameter_spec_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


