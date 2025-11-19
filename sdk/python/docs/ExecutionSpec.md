# ExecutionSpec


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | [**ContainerSpec**](ContainerSpec.md) |  | 
**datapack** | **str** |  | [optional] 
**dataset** | [**DatasetRef**](DatasetRef.md) |  | [optional] 

## Example

```python
from openapi.models.execution_spec import ExecutionSpec

# TODO update the JSON string below
json = "{}"
# create an instance of ExecutionSpec from a JSON string
execution_spec_instance = ExecutionSpec.from_json(json)
# print the JSON string representation of the object
print(ExecutionSpec.to_json())

# convert the object into a dict
execution_spec_dict = execution_spec_instance.to_dict()
# create an instance of ExecutionSpec from a dict
execution_spec_from_dict = ExecutionSpec.from_dict(execution_spec_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


