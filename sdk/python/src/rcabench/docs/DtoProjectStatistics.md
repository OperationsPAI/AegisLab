# DtoProjectStatistics


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**active** | **int** |  | [optional] 
**inactive** | **int** |  | [optional] 
**new_today** | **int** |  | [optional] 
**total** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_project_statistics import DtoProjectStatistics

# TODO update the JSON string below
json = "{}"
# create an instance of DtoProjectStatistics from a JSON string
dto_project_statistics_instance = DtoProjectStatistics.from_json(json)
# print the JSON string representation of the object
print(DtoProjectStatistics.to_json())

# convert the object into a dict
dto_project_statistics_dict = dto_project_statistics_instance.to_dict()
# create an instance of DtoProjectStatistics from a dict
dto_project_statistics_from_dict = DtoProjectStatistics.from_dict(dto_project_statistics_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


