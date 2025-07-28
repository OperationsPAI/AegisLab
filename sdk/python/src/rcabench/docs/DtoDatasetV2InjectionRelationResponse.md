# DtoDatasetV2InjectionRelationResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** | 创建时间 | [optional] 
**fault_injection** | [**DatabaseFaultInjectionSchedule**](DatabaseFaultInjectionSchedule.md) | 故障注入详情 | [optional] 
**fault_injection_id** | **int** | 故障注入ID | [optional] 
**id** | **int** | 关联ID | [optional] 
**updated_at** | **str** | 更新时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_v2_injection_relation_response import DtoDatasetV2InjectionRelationResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetV2InjectionRelationResponse from a JSON string
dto_dataset_v2_injection_relation_response_instance = DtoDatasetV2InjectionRelationResponse.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetV2InjectionRelationResponse.to_json())

# convert the object into a dict
dto_dataset_v2_injection_relation_response_dict = dto_dataset_v2_injection_relation_response_instance.to_dict()
# create an instance of DtoDatasetV2InjectionRelationResponse from a dict
dto_dataset_v2_injection_relation_response_from_dict = DtoDatasetV2InjectionRelationResponse.from_dict(dto_dataset_v2_injection_relation_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


