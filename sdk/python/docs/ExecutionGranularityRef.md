# ExecutionGranularityRef


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**detector_results** | [**List[DetectorResultItem]**](DetectorResultItem.md) | Detector results | [optional] 
**executed_at** | **str** | Execution time | [optional] 
**execution_duration** | **float** | Execution duration in seconds | [optional] 
**execution_id** | **int** | Execution ID | [optional] 
**predictions** | [**List[GranularityResultItem]**](GranularityResultItem.md) | Algorithm predictions | [optional] 

## Example

```python
from rcabench.openapi.models.execution_granularity_ref import ExecutionGranularityRef

# TODO update the JSON string below
json = "{}"
# create an instance of ExecutionGranularityRef from a JSON string
execution_granularity_ref_instance = ExecutionGranularityRef.from_json(json)
# print the JSON string representation of the object
print(ExecutionGranularityRef.to_json())

# convert the object into a dict
execution_granularity_ref_dict = execution_granularity_ref_instance.to_dict()
# create an instance of ExecutionGranularityRef from a dict
execution_granularity_ref_from_dict = ExecutionGranularityRef.from_dict(execution_granularity_ref_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


