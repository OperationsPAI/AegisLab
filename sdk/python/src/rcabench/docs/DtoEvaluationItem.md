# DtoEvaluationItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | **str** |  | [optional] 
**conclusions** | [**List[DtoConclusion]**](DtoConclusion.md) |  | [optional] 
**executions** | [**List[DtoExecution]**](DtoExecution.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_evaluation_item import DtoEvaluationItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoEvaluationItem from a JSON string
dto_evaluation_item_instance = DtoEvaluationItem.from_json(json)
# print the JSON string representation of the object
print(DtoEvaluationItem.to_json())

# convert the object into a dict
dto_evaluation_item_dict = dto_evaluation_item_instance.to_dict()
# create an instance of DtoEvaluationItem from a dict
dto_evaluation_item_from_dict = DtoEvaluationItem.from_dict(dto_evaluation_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


