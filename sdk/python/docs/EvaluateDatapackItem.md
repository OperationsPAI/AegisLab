# EvaluateDatapackItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | **str** |  | [optional] 
**algorithm_version** | **str** |  | [optional] 
**datapack** | **str** |  | [optional] 
**execution_refs** | [**List[ExecutionGranularityRef]**](ExecutionGranularityRef.md) |  | [optional] 
**groundtruth** | [**ChaosGroundtruth**](ChaosGroundtruth.md) |  | [optional] 

## Example

```python
from openapi.models.evaluate_datapack_item import EvaluateDatapackItem

# TODO update the JSON string below
json = "{}"
# create an instance of EvaluateDatapackItem from a JSON string
evaluate_datapack_item_instance = EvaluateDatapackItem.from_json(json)
# print the JSON string representation of the object
print(EvaluateDatapackItem.to_json())

# convert the object into a dict
evaluate_datapack_item_dict = evaluate_datapack_item_instance.to_dict()
# create an instance of EvaluateDatapackItem from a dict
evaluate_datapack_item_from_dict = EvaluateDatapackItem.from_dict(evaluate_datapack_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


