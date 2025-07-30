# DtoGenericResponseHandlerResources


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**HandlerResources**](HandlerResources.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_handler_resources import DtoGenericResponseHandlerResources

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseHandlerResources from a JSON string
dto_generic_response_handler_resources_instance = DtoGenericResponseHandlerResources.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseHandlerResources.to_json())

# convert the object into a dict
dto_generic_response_handler_resources_dict = dto_generic_response_handler_resources_instance.to_dict()
# create an instance of DtoGenericResponseHandlerResources from a dict
dto_generic_response_handler_resources_from_dict = DtoGenericResponseHandlerResources.from_dict(dto_generic_response_handler_resources_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


