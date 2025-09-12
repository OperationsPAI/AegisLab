# DtoRelationListResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DtoRelationResponse]**](DtoRelationResponse.md) |  | [optional] 
**pagination** | [**DtoPaginationInfo**](DtoPaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_relation_list_response import DtoRelationListResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoRelationListResponse from a JSON string
dto_relation_list_response_instance = DtoRelationListResponse.from_json(json)
# print the JSON string representation of the object
print(DtoRelationListResponse.to_json())

# convert the object into a dict
dto_relation_list_response_dict = dto_relation_list_response_instance.to_dict()
# create an instance of DtoRelationListResponse from a dict
dto_relation_list_response_from_dict = DtoRelationListResponse.from_dict(dto_relation_list_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


