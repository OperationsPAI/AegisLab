# SubmitExecutionItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm_id** | **int** |  | [optional] 
**algorithm_version_id** | **int** |  | [optional] 
**datapack_id** | **int** |  | [optional] 
**dataset_id** | **int** |  | [optional] 
**index** | **int** |  | [optional] 
**task_id** | **str** |  | [optional] 
**trace_id** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.submit_execution_item import SubmitExecutionItem

# TODO update the JSON string below
json = "{}"
# create an instance of SubmitExecutionItem from a JSON string
submit_execution_item_instance = SubmitExecutionItem.from_json(json)
# print the JSON string representation of the object
print(SubmitExecutionItem.to_json())

# convert the object into a dict
submit_execution_item_dict = submit_execution_item_instance.to_dict()
# create an instance of SubmitExecutionItem from a dict
submit_execution_item_from_dict = SubmitExecutionItem.from_dict(submit_execution_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


