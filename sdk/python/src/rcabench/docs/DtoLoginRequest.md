# DtoLoginRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**password** | **str** |  | 
**username** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_login_request import DtoLoginRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoLoginRequest from a JSON string
dto_login_request_instance = DtoLoginRequest.from_json(json)
# print the JSON string representation of the object
print(DtoLoginRequest.to_json())

# convert the object into a dict
dto_login_request_dict = dto_login_request_instance.to_dict()
# create an instance of DtoLoginRequest from a dict
dto_login_request_from_dict = DtoLoginRequest.from_dict(dto_login_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


