# DtoRelationEntity


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **int** |  | [optional] 
**name** | **str** |  | [optional] 
**type** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_relation_entity import DtoRelationEntity

# TODO update the JSON string below
json = "{}"
# create an instance of DtoRelationEntity from a JSON string
dto_relation_entity_instance = DtoRelationEntity.from_json(json)
# print the JSON string representation of the object
print(DtoRelationEntity.to_json())

# convert the object into a dict
dto_relation_entity_dict = dto_relation_entity_instance.to_dict()
# create an instance of DtoRelationEntity from a dict
dto_relation_entity_from_dict = DtoRelationEntity.from_dict(dto_relation_entity_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


