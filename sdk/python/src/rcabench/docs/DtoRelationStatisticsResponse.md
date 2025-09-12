# DtoRelationStatisticsResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**container_labels** | **int** |  | [optional] 
**dataset_labels** | **int** |  | [optional] 
**fault_injection_labels** | **int** |  | [optional] 
**project_labels** | **int** |  | [optional] 
**role_permissions** | **int** |  | [optional] 
**user_permissions** | **int** |  | [optional] 
**user_projects** | **int** |  | [optional] 
**user_roles** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_relation_statistics_response import DtoRelationStatisticsResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoRelationStatisticsResponse from a JSON string
dto_relation_statistics_response_instance = DtoRelationStatisticsResponse.from_json(json)
# print the JSON string representation of the object
print(DtoRelationStatisticsResponse.to_json())

# convert the object into a dict
dto_relation_statistics_response_dict = dto_relation_statistics_response_instance.to_dict()
# create an instance of DtoRelationStatisticsResponse from a dict
dto_relation_statistics_response_from_dict = DtoRelationStatisticsResponse.from_dict(dto_relation_statistics_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


