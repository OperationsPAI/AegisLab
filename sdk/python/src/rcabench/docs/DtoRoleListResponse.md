# DtoRoleListResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DtoRoleResponse]**](DtoRoleResponse.md) |  | [optional] 
**pagination** | [**DtoPaginationInfo**](DtoPaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_role_list_response import DtoRoleListResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoRoleListResponse from a JSON string
dto_role_list_response_instance = DtoRoleListResponse.from_json(json)
# print the JSON string representation of the object
print(DtoRoleListResponse.to_json())

# convert the object into a dict
dto_role_list_response_dict = dto_role_list_response_instance.to_dict()
# create an instance of DtoRoleListResponse from a dict
dto_role_list_response_from_dict = DtoRoleListResponse.from_dict(dto_role_list_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


