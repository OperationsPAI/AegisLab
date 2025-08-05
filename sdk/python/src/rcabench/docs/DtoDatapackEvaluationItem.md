# DtoDatapackEvaluationItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**datapack_name** | **str** | Datapack name (from FaultInjectionSchedule) | [optional] 
**executed_at** | **str** | Execution time | [optional] 
**execution_id** | **int** | Execution ID | [optional] 
**groundtruth** | [**HandlerGroundtruth**](HandlerGroundtruth.md) | Ground truth for this datapack | [optional] 
**predictions** | [**List[DtoGranularityRecord]**](DtoGranularityRecord.md) | Algorithm predictions | [optional] 

## Example

```python
from rcabench.openapi.models.dto_datapack_evaluation_item import DtoDatapackEvaluationItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatapackEvaluationItem from a JSON string
dto_datapack_evaluation_item_instance = DtoDatapackEvaluationItem.from_json(json)
# print the JSON string representation of the object
print(DtoDatapackEvaluationItem.to_json())

# convert the object into a dict
dto_datapack_evaluation_item_dict = dto_datapack_evaluation_item_instance.to_dict()
# create an instance of DtoDatapackEvaluationItem from a dict
dto_datapack_evaluation_item_from_dict = DtoDatapackEvaluationItem.from_dict(dto_datapack_evaluation_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


