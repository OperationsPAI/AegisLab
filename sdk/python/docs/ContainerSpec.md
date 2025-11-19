# ContainerSpec


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**env_vars** | [**List[ParameterSpec]**](ParameterSpec.md) |  | [optional] 
**name** | **str** |  | 
**version** | **str** |  | [optional] 

## Example

```python
from openapi.models.container_spec import ContainerSpec

# TODO update the JSON string below
json = "{}"
# create an instance of ContainerSpec from a JSON string
container_spec_instance = ContainerSpec.from_json(json)
# print the JSON string representation of the object
print(ContainerSpec.to_json())

# convert the object into a dict
container_spec_dict = container_spec_instance.to_dict()
# create an instance of ContainerSpec from a dict
container_spec_from_dict = ContainerSpec.from_dict(container_spec_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


