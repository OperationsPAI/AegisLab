# GenericResponseUserInfo


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**UserInfo**](UserInfo.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.generic_response_user_info import GenericResponseUserInfo

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseUserInfo from a JSON string
generic_response_user_info_instance = GenericResponseUserInfo.from_json(json)
# print the JSON string representation of the object
print(GenericResponseUserInfo.to_json())

# convert the object into a dict
generic_response_user_info_dict = generic_response_user_info_instance.to_dict()
# create an instance of GenericResponseUserInfo from a dict
generic_response_user_info_from_dict = GenericResponseUserInfo.from_dict(generic_response_user_info_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


