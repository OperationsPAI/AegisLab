# BatchEvaluateDatasetResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**failed_count** | **int** |  | [optional] 
**failed_items** | **List[str]** |  | [optional] 
**success_count** | **int** |  | [optional] 
**success_items** | [**List[EvaluateDatasetItem]**](EvaluateDatasetItem.md) |  | [optional] 

## Example

```python
from openapi.models.batch_evaluate_dataset_resp import BatchEvaluateDatasetResp

# TODO update the JSON string below
json = "{}"
# create an instance of BatchEvaluateDatasetResp from a JSON string
batch_evaluate_dataset_resp_instance = BatchEvaluateDatasetResp.from_json(json)
# print the JSON string representation of the object
print(BatchEvaluateDatasetResp.to_json())

# convert the object into a dict
batch_evaluate_dataset_resp_dict = batch_evaluate_dataset_resp_instance.to_dict()
# create an instance of BatchEvaluateDatasetResp from a dict
batch_evaluate_dataset_resp_from_dict = BatchEvaluateDatasetResp.from_dict(batch_evaluate_dataset_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


