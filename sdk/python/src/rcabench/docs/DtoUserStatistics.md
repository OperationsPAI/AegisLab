# DtoUserStatistics


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**active** | **int** |  | [optional] 
**inactive** | **int** |  | [optional] 
**new_this_week** | **int** |  | [optional] 
**new_today** | **int** |  | [optional] 
**total** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_user_statistics import DtoUserStatistics

# TODO update the JSON string below
json = "{}"
# create an instance of DtoUserStatistics from a JSON string
dto_user_statistics_instance = DtoUserStatistics.from_json(json)
# print the JSON string representation of the object
print(DtoUserStatistics.to_json())

# convert the object into a dict
dto_user_statistics_dict = dto_user_statistics_instance.to_dict()
# create an instance of DtoUserStatistics from a dict
dto_user_statistics_from_dict = DtoUserStatistics.from_dict(dto_user_statistics_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


