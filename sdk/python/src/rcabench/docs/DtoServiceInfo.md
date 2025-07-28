# DtoServiceInfo


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**details** | **object** |  | [optional] 
**error** | **str** |  | [optional] 
**last_checked** | **str** |  | [optional] 
**response_time** | **str** |  | [optional] 
**status** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_service_info import DtoServiceInfo

# TODO update the JSON string below
json = "{}"
# create an instance of DtoServiceInfo from a JSON string
dto_service_info_instance = DtoServiceInfo.from_json(json)
# print the JSON string representation of the object
print(DtoServiceInfo.to_json())

# convert the object into a dict
dto_service_info_dict = dto_service_info_instance.to_dict()
# create an instance of DtoServiceInfo from a dict
dto_service_info_from_dict = DtoServiceInfo.from_dict(dto_service_info_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


