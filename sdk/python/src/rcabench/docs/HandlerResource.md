# HandlerResource


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**app_labels** | **List[str]** |  | [optional] 
**container_app_names** | **List[str]** |  | [optional] 
**database_app_names** | **List[str]** |  | [optional] 
**dns_app_names** | **List[str]** |  | [optional] 
**http_app_names** | **List[str]** |  | [optional] 
**jvm_app_names** | **List[str]** |  | [optional] 
**network_pairs** | [**List[HandlerPair]**](HandlerPair.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.handler_resource import HandlerResource

# TODO update the JSON string below
json = "{}"
# create an instance of HandlerResource from a JSON string
handler_resource_instance = HandlerResource.from_json(json)
# print the JSON string representation of the object
print(HandlerResource.to_json())

# convert the object into a dict
handler_resource_dict = handler_resource_instance.to_dict()
# create an instance of HandlerResource from a dict
handler_resource_from_dict = HandlerResource.from_dict(handler_resource_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


