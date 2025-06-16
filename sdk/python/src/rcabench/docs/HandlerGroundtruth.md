# HandlerGroundtruth


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**container** | **List[str]** |  | [optional] 
**function** | **List[str]** |  | [optional] 
**metric** | **List[str]** |  | [optional] 
**pod** | **List[str]** |  | [optional] 
**service** | **List[str]** |  | [optional] 
**span** | **List[str]** |  | [optional] 

## Example

```python
from rcabench.openapi.models.handler_groundtruth import HandlerGroundtruth

# TODO update the JSON string below
json = "{}"
# create an instance of HandlerGroundtruth from a JSON string
handler_groundtruth_instance = HandlerGroundtruth.from_json(json)
# print the JSON string representation of the object
print(HandlerGroundtruth.to_json())

# convert the object into a dict
handler_groundtruth_dict = handler_groundtruth_instance.to_dict()
# create an instance of HandlerGroundtruth from a dict
handler_groundtruth_from_dict = HandlerGroundtruth.from_dict(handler_groundtruth_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


