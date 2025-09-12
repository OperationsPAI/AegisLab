# DtoDatasetStatistics


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**private** | **int** |  | [optional] 
**public** | **int** |  | [optional] 
**total** | **int** |  | [optional] 
**total_size** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_statistics import DtoDatasetStatistics

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetStatistics from a JSON string
dto_dataset_statistics_instance = DtoDatasetStatistics.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetStatistics.to_json())

# convert the object into a dict
dto_dataset_statistics_dict = dto_dataset_statistics_instance.to_dict()
# create an instance of DtoDatasetStatistics from a dict
dto_dataset_statistics_from_dict = DtoDatasetStatistics.from_dict(dto_dataset_statistics_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


