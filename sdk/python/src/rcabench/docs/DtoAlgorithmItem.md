# DtoAlgorithmItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**image** | **str** |  | [optional] 
**name** | **str** |  | 
**tag** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_algorithm_item import DtoAlgorithmItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAlgorithmItem from a JSON string
dto_algorithm_item_instance = DtoAlgorithmItem.from_json(json)
# print the JSON string representation of the object
print(DtoAlgorithmItem.to_json())

# convert the object into a dict
dto_algorithm_item_dict = dto_algorithm_item_instance.to_dict()
# create an instance of DtoAlgorithmItem from a dict
dto_algorithm_item_from_dict = DtoAlgorithmItem.from_dict(dto_algorithm_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


