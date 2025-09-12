# DtoTaskStatistics


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**completed** | **int** |  | [optional] 
**failed** | **int** |  | [optional] 
**pending** | **int** |  | [optional] 
**running** | **int** |  | [optional] 
**total** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_task_statistics import DtoTaskStatistics

# TODO update the JSON string below
json = "{}"
# create an instance of DtoTaskStatistics from a JSON string
dto_task_statistics_instance = DtoTaskStatistics.from_json(json)
# print the JSON string representation of the object
print(DtoTaskStatistics.to_json())

# convert the object into a dict
dto_task_statistics_dict = dto_task_statistics_instance.to_dict()
# create an instance of DtoTaskStatistics from a dict
dto_task_statistics_from_dict = DtoTaskStatistics.from_dict(dto_task_statistics_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


