# HandlerNode


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**children** | [**Dict[str, HandlerNode]**](HandlerNode.md) |  | [optional] 
**description** | **str** |  | [optional] 
**name** | **str** |  | [optional] 
**range** | **List[int]** |  | [optional] 
**value** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.handler_node import HandlerNode

# TODO update the JSON string below
json = "{}"
# create an instance of HandlerNode from a JSON string
handler_node_instance = HandlerNode.from_json(json)
# print the JSON string representation of the object
print(HandlerNode.to_json())

# convert the object into a dict
handler_node_dict = handler_node_instance.to_dict()
# create an instance of HandlerNode from a dict
handler_node_from_dict = HandlerNode.from_dict(handler_node_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


