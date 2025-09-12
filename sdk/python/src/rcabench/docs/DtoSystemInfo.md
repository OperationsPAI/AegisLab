# DtoSystemInfo


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**cpu_usage** | **float** |  | [optional] 
**disk_usage** | **float** |  | [optional] 
**load_average** | **str** |  | [optional] 
**memory_usage** | **float** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_system_info import DtoSystemInfo

# TODO update the JSON string below
json = "{}"
# create an instance of DtoSystemInfo from a JSON string
dto_system_info_instance = DtoSystemInfo.from_json(json)
# print the JSON string representation of the object
print(DtoSystemInfo.to_json())

# convert the object into a dict
dto_system_info_dict = dto_system_info_instance.to_dict()
# create an instance of DtoSystemInfo from a dict
dto_system_info_from_dict = DtoSystemInfo.from_dict(dto_system_info_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


