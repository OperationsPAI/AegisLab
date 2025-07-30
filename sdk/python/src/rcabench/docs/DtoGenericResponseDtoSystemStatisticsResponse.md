# DtoGenericResponseDtoSystemStatisticsResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**DtoSystemStatisticsResponse**](DtoSystemStatisticsResponse.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_system_statistics_response import DtoGenericResponseDtoSystemStatisticsResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoSystemStatisticsResponse from a JSON string
dto_generic_response_dto_system_statistics_response_instance = DtoGenericResponseDtoSystemStatisticsResponse.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoSystemStatisticsResponse.to_json())

# convert the object into a dict
dto_generic_response_dto_system_statistics_response_dict = dto_generic_response_dto_system_statistics_response_instance.to_dict()
# create an instance of DtoGenericResponseDtoSystemStatisticsResponse from a dict
dto_generic_response_dto_system_statistics_response_from_dict = DtoGenericResponseDtoSystemStatisticsResponse.from_dict(dto_generic_response_dto_system_statistics_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


