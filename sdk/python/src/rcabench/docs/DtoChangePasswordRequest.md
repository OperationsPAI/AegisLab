# DtoChangePasswordRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**new_password** | **str** |  | 
**old_password** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_change_password_request import DtoChangePasswordRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoChangePasswordRequest from a JSON string
dto_change_password_request_instance = DtoChangePasswordRequest.from_json(json)
# print the JSON string representation of the object
print(DtoChangePasswordRequest.to_json())

# convert the object into a dict
dto_change_password_request_dict = dto_change_password_request_instance.to_dict()
# create an instance of DtoChangePasswordRequest from a dict
dto_change_password_request_from_dict = DtoChangePasswordRequest.from_dict(dto_change_password_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


