# DtoUserListResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DtoUserResponse]**](DtoUserResponse.md) |  | [optional] 
**pagination** | [**DtoPaginationInfo**](DtoPaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_user_list_response import DtoUserListResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoUserListResponse from a JSON string
dto_user_list_response_instance = DtoUserListResponse.from_json(json)
# print the JSON string representation of the object
print(DtoUserListResponse.to_json())

# convert the object into a dict
dto_user_list_response_dict = dto_user_list_response_instance.to_dict()
# create an instance of DtoUserListResponse from a dict
dto_user_list_response_from_dict = DtoUserListResponse.from_dict(dto_user_list_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


