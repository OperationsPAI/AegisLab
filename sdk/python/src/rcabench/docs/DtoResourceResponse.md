# DtoResourceResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**category** | **str** |  | [optional] 
**display_name** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**name** | **str** |  | [optional] 
**type** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_resource_response import DtoResourceResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoResourceResponse from a JSON string
dto_resource_response_instance = DtoResourceResponse.from_json(json)
# print the JSON string representation of the object
print(DtoResourceResponse.to_json())

# convert the object into a dict
dto_resource_response_dict = dto_resource_response_instance.to_dict()
# create an instance of DtoResourceResponse from a dict
dto_resource_response_from_dict = DtoResourceResponse.from_dict(dto_resource_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


