# ChaosNode


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**children** | [**Dict[str, ChaosNode]**](ChaosNode.md) |  | [optional] 
**description** | **str** |  | [optional] 
**name** | **str** |  | [optional] 
**range** | **List[int]** |  | [optional] 
**value** | **int** |  | [optional] 

## Example

```python
from openapi.models.chaos_node import ChaosNode

# TODO update the JSON string below
json = "{}"
# create an instance of ChaosNode from a JSON string
chaos_node_instance = ChaosNode.from_json(json)
# print the JSON string representation of the object
print(ChaosNode.to_json())

# convert the object into a dict
chaos_node_dict = chaos_node_instance.to_dict()
# create an instance of ChaosNode from a dict
chaos_node_from_dict = ChaosNode.from_dict(chaos_node_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


