# BuildingSpec


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**benchmark** | [**ContainerSpec**](ContainerSpec.md) |  | 
**datapack** | **str** |  | [optional] 
**dataset** | [**DatasetRef**](DatasetRef.md) |  | [optional] 
**pre_duration** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.building_spec import BuildingSpec

# TODO update the JSON string below
json = "{}"
# create an instance of BuildingSpec from a JSON string
building_spec_instance = BuildingSpec.from_json(json)
# print the JSON string representation of the object
print(BuildingSpec.to_json())

# convert the object into a dict
building_spec_dict = building_spec_instance.to_dict()
# create an instance of BuildingSpec from a dict
building_spec_from_dict = BuildingSpec.from_dict(building_spec_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


