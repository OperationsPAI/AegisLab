# DtoRelationResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**source** | [**DtoRelationEntity**](DtoRelationEntity.md) |  | [optional] 
**target** | [**DtoRelationEntity**](DtoRelationEntity.md) |  | [optional] 
**type** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_relation_response import DtoRelationResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoRelationResponse from a JSON string
dto_relation_response_instance = DtoRelationResponse.from_json(json)
# print the JSON string representation of the object
print(DtoRelationResponse.to_json())

# convert the object into a dict
dto_relation_response_dict = dto_relation_response_instance.to_dict()
# create an instance of DtoRelationResponse from a dict
dto_relation_response_from_dict = DtoRelationResponse.from_dict(dto_relation_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


