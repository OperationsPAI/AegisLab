# HandlerPair


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**source** | **str** |  | [optional] 
**target** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.handler_pair import HandlerPair

# TODO update the JSON string below
json = "{}"
# create an instance of HandlerPair from a JSON string
handler_pair_instance = HandlerPair.from_json(json)
# print the JSON string representation of the object
print(HandlerPair.to_json())

# convert the object into a dict
handler_pair_dict = handler_pair_instance.to_dict()
# create an instance of HandlerPair from a dict
handler_pair_from_dict = HandlerPair.from_dict(handler_pair_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


