# DtoTokenRefreshResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**expires_at** | **str** |  | [optional] 
**token** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_token_refresh_response import DtoTokenRefreshResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoTokenRefreshResponse from a JSON string
dto_token_refresh_response_instance = DtoTokenRefreshResponse.from_json(json)
# print the JSON string representation of the object
print(DtoTokenRefreshResponse.to_json())

# convert the object into a dict
dto_token_refresh_response_dict = dto_token_refresh_response_instance.to_dict()
# create an instance of DtoTokenRefreshResponse from a dict
dto_token_refresh_response_from_dict = DtoTokenRefreshResponse.from_dict(dto_token_refresh_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


