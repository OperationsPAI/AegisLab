# EvaluateDatasetItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | **str** | Algorithm name | [optional] 
**algorithm_version** | **str** | Algorithm version | [optional] 
**dataset** | **str** | Dataset name | [optional] 
**dataset_version** | **str** | Dataset version | [optional] 
**evalaute_refs** | [**List[EvaluateDatapackRef]**](EvaluateDatapackRef.md) | Evaluation refs for each dataset | [optional] 
**executed_count** | **int** | Number of successfully executed datapacks | [optional] 
**total_count** | **int** | Total number of datapacks in dataset | [optional] 

## Example

```python
from openapi.models.evaluate_dataset_item import EvaluateDatasetItem

# TODO update the JSON string below
json = "{}"
# create an instance of EvaluateDatasetItem from a JSON string
evaluate_dataset_item_instance = EvaluateDatasetItem.from_json(json)
# print the JSON string representation of the object
print(EvaluateDatasetItem.to_json())

# convert the object into a dict
evaluate_dataset_item_dict = evaluate_dataset_item_instance.to_dict()
# create an instance of EvaluateDatasetItem from a dict
evaluate_dataset_item_from_dict = EvaluateDatasetItem.from_dict(evaluate_dataset_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


