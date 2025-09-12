# DtoLabelResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**category** | **str** |  | [optional] 
**color** | **str** |  | [optional] 
**created_at** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**is_system** | **bool** |  | [optional] 
**key** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 
**usage** | **int** |  | [optional] 
**value** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_label_response import DtoLabelResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoLabelResponse from a JSON string
dto_label_response_instance = DtoLabelResponse.from_json(json)
# print the JSON string representation of the object
print(DtoLabelResponse.to_json())

# convert the object into a dict
dto_label_response_dict = dto_label_response_instance.to_dict()
# create an instance of DtoLabelResponse from a dict
dto_label_response_from_dict = DtoLabelResponse.from_dict(dto_label_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


