# DtoLogoutRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**token** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_logout_request import DtoLogoutRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoLogoutRequest from a JSON string
dto_logout_request_instance = DtoLogoutRequest.from_json(json)
# print the JSON string representation of the object
print(DtoLogoutRequest.to_json())

# convert the object into a dict
dto_logout_request_dict = dto_logout_request_instance.to_dict()
# create an instance of DtoLogoutRequest from a dict
dto_logout_request_from_dict = DtoLogoutRequest.from_dict(dto_logout_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


