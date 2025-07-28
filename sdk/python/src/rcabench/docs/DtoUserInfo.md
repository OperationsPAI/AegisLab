# DtoUserInfo


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**avatar** | **str** |  | [optional] 
**created_at** | **str** |  | [optional] 
**email** | **str** |  | [optional] 
**full_name** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**is_active** | **bool** |  | [optional] 
**phone** | **str** |  | [optional] 
**status** | **int** |  | [optional] 
**updated_at** | **str** |  | [optional] 
**username** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_user_info import DtoUserInfo

# TODO update the JSON string below
json = "{}"
# create an instance of DtoUserInfo from a JSON string
dto_user_info_instance = DtoUserInfo.from_json(json)
# print the JSON string representation of the object
print(DtoUserInfo.to_json())

# convert the object into a dict
dto_user_info_dict = dto_user_info_instance.to_dict()
# create an instance of DtoUserInfo from a dict
dto_user_info_from_dict = DtoUserInfo.from_dict(dto_user_info_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


