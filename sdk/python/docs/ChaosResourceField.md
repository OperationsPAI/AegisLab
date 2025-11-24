# ChaosResourceField


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**index_name** | **str** |  | [optional] 
**name** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.chaos_resource_field import ChaosResourceField

# TODO update the JSON string below
json = "{}"
# create an instance of ChaosResourceField from a JSON string
chaos_resource_field_instance = ChaosResourceField.from_json(json)
# print the JSON string representation of the object
print(ChaosResourceField.to_json())

# convert the object into a dict
chaos_resource_field_dict = chaos_resource_field_instance.to_dict()
# create an instance of ChaosResourceField from a dict
chaos_resource_field_from_dict = ChaosResourceField.from_dict(chaos_resource_field_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


