# DtoLoginResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**expires_at** | **str** |  | [optional] 
**token** | **str** |  | [optional] 
**user** | [**DtoUserInfo**](DtoUserInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_login_response import DtoLoginResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoLoginResponse from a JSON string
dto_login_response_instance = DtoLoginResponse.from_json(json)
# print the JSON string representation of the object
print(DtoLoginResponse.to_json())

# convert the object into a dict
dto_login_response_dict = dto_login_response_instance.to_dict()
# create an instance of DtoLoginResponse from a dict
dto_login_response_from_dict = DtoLoginResponse.from_dict(dto_login_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


