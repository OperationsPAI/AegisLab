# DtoGenericResponseHandlerNode


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**HandlerNode**](HandlerNode.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_handler_node import DtoGenericResponseHandlerNode

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseHandlerNode from a JSON string
dto_generic_response_handler_node_instance = DtoGenericResponseHandlerNode.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseHandlerNode.to_json())

# convert the object into a dict
dto_generic_response_handler_node_dict = dto_generic_response_handler_node_instance.to_dict()
# create an instance of DtoGenericResponseHandlerNode from a dict
dto_generic_response_handler_node_from_dict = DtoGenericResponseHandlerNode.from_dict(dto_generic_response_handler_node_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


