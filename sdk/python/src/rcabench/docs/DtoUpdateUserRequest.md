# DtoUpdateUserRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**avatar** | **str** |  | [optional] 
**email** | **str** |  | [optional] 
**full_name** | **str** |  | [optional] 
**is_active** | **bool** |  | [optional] 
**phone** | **str** |  | [optional] 
**status** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_update_user_request import DtoUpdateUserRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoUpdateUserRequest from a JSON string
dto_update_user_request_instance = DtoUpdateUserRequest.from_json(json)
# print the JSON string representation of the object
print(DtoUpdateUserRequest.to_json())

# convert the object into a dict
dto_update_user_request_dict = dto_update_user_request_instance.to_dict()
# create an instance of DtoUpdateUserRequest from a dict
dto_update_user_request_from_dict = DtoUpdateUserRequest.from_dict(dto_update_user_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


