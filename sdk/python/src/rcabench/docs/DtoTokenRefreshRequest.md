# DtoTokenRefreshRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**token** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_token_refresh_request import DtoTokenRefreshRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoTokenRefreshRequest from a JSON string
dto_token_refresh_request_instance = DtoTokenRefreshRequest.from_json(json)
# print the JSON string representation of the object
print(DtoTokenRefreshRequest.to_json())

# convert the object into a dict
dto_token_refresh_request_dict = dto_token_refresh_request_instance.to_dict()
# create an instance of DtoTokenRefreshRequest from a dict
dto_token_refresh_request_from_dict = DtoTokenRefreshRequest.from_dict(dto_token_refresh_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


