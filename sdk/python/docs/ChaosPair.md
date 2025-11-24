# ChaosPair


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**source** | **str** |  | [optional] 
**target** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.chaos_pair import ChaosPair

# TODO update the JSON string below
json = "{}"
# create an instance of ChaosPair from a JSON string
chaos_pair_instance = ChaosPair.from_json(json)
# print the JSON string representation of the object
print(ChaosPair.to_json())

# convert the object into a dict
chaos_pair_dict = chaos_pair_instance.to_dict()
# create an instance of ChaosPair from a dict
chaos_pair_from_dict = ChaosPair.from_dict(chaos_pair_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


