# DtoCreateUserRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**avatar** | **str** |  | [optional] 
**email** | **str** |  | 
**full_name** | **str** |  | 
**password** | **str** |  | 
**phone** | **str** |  | [optional] 
**username** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_create_user_request import DtoCreateUserRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoCreateUserRequest from a JSON string
dto_create_user_request_instance = DtoCreateUserRequest.from_json(json)
# print the JSON string representation of the object
print(DtoCreateUserRequest.to_json())

# convert the object into a dict
dto_create_user_request_dict = dto_create_user_request_instance.to_dict()
# create an instance of DtoCreateUserRequest from a dict
dto_create_user_request_from_dict = DtoCreateUserRequest.from_dict(dto_create_user_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


