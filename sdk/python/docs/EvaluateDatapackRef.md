# EvaluateDatapackRef


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**datapack** | **str** |  | [optional] 
**detector_results** | [**List[DetectorResultItem]**](DetectorResultItem.md) | Detector results | [optional] 
**executed_at** | **str** | Execution time | [optional] 
**execution_duration** | **float** | Execution duration in seconds | [optional] 
**execution_id** | **int** | Execution ID | [optional] 
**predictions** | [**List[GranularityResultItem]**](GranularityResultItem.md) | Algorithm predictions | [optional] 

## Example

```python
from rcabench.openapi.models.evaluate_datapack_ref import EvaluateDatapackRef

# TODO update the JSON string below
json = "{}"
# create an instance of EvaluateDatapackRef from a JSON string
evaluate_datapack_ref_instance = EvaluateDatapackRef.from_json(json)
# print the JSON string representation of the object
print(EvaluateDatapackRef.to_json())

# convert the object into a dict
evaluate_datapack_ref_dict = evaluate_datapack_ref_instance.to_dict()
# create an instance of EvaluateDatapackRef from a dict
evaluate_datapack_ref_from_dict = EvaluateDatapackRef.from_dict(evaluate_datapack_ref_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


