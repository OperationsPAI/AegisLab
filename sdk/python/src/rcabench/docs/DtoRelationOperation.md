# DtoRelationOperation


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**action** | **str** |  | 
**source_id** | **int** |  | 
**target_id** | **int** |  | 
**type** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_relation_operation import DtoRelationOperation

# TODO update the JSON string below
json = "{}"
# create an instance of DtoRelationOperation from a JSON string
dto_relation_operation_instance = DtoRelationOperation.from_json(json)
# print the JSON string representation of the object
print(DtoRelationOperation.to_json())

# convert the object into a dict
dto_relation_operation_dict = dto_relation_operation_instance.to_dict()
# create an instance of DtoRelationOperation from a dict
dto_relation_operation_from_dict = DtoRelationOperation.from_dict(dto_relation_operation_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


