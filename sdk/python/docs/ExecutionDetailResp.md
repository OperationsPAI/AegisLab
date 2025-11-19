# ExecutionDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm_id** | **int** |  | [optional] 
**algorithm_name** | **str** |  | [optional] 
**algorithm_version** | **str** |  | [optional] 
**algorithm_version_id** | **int** |  | [optional] 
**created_at** | **str** |  | [optional] 
**datapack_id** | **int** |  | [optional] 
**datapack_name** | **str** |  | [optional] 
**detector_results** | [**List[DetectorResultItem]**](DetectorResultItem.md) |  | [optional] 
**duration** | **float** |  | [optional] 
**granularity_results** | [**List[GranularityResultItem]**](GranularityResultItem.md) |  | [optional] 
**id** | **int** |  | [optional] 
**labels** | [**List[LabelItem]**](LabelItem.md) |  | [optional] 
**state** | **str** |  | [optional] 
**status** | **str** |  | [optional] 
**task_id** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from openapi.models.execution_detail_resp import ExecutionDetailResp

# TODO update the JSON string below
json = "{}"
# create an instance of ExecutionDetailResp from a JSON string
execution_detail_resp_instance = ExecutionDetailResp.from_json(json)
# print the JSON string representation of the object
print(ExecutionDetailResp.to_json())

# convert the object into a dict
execution_detail_resp_dict = execution_detail_resp_instance.to_dict()
# create an instance of ExecutionDetailResp from a dict
execution_detail_resp_from_dict = ExecutionDetailResp.from_dict(execution_detail_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


