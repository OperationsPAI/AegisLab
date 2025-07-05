# HandlerResources


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**app_labels** | **List[str]** |  | [optional] 
**container_names** | **List[str]** |  | [optional] 
**database_app_names** | **List[str]** |  | [optional] 
**dns_app_names** | **List[str]** |  | [optional] 
**http_app_names** | **List[str]** |  | [optional] 
**jvm_app_names** | **List[str]** |  | [optional] 
**network_pairs** | [**List[HandlerPair]**](HandlerPair.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.handler_resources import HandlerResources

# TODO update the JSON string below
json = "{}"
# create an instance of HandlerResources from a JSON string
handler_resources_instance = HandlerResources.from_json(json)
# print the JSON string representation of the object
print(HandlerResources.to_json())

# convert the object into a dict
handler_resources_dict = handler_resources_instance.to_dict()
# create an instance of HandlerResources from a dict
handler_resources_from_dict = HandlerResources.from_dict(handler_resources_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


