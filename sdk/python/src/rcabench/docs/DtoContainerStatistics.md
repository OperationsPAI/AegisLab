# DtoContainerStatistics


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**active** | **int** |  | [optional] 
**deleted** | **int** |  | [optional] 
**total** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_container_statistics import DtoContainerStatistics

# TODO update the JSON string below
json = "{}"
# create an instance of DtoContainerStatistics from a JSON string
dto_container_statistics_instance = DtoContainerStatistics.from_json(json)
# print the JSON string representation of the object
print(DtoContainerStatistics.to_json())

# convert the object into a dict
dto_container_statistics_dict = dto_container_statistics_instance.to_dict()
# create an instance of DtoContainerStatistics from a dict
dto_container_statistics_from_dict = DtoContainerStatistics.from_dict(dto_container_statistics_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


