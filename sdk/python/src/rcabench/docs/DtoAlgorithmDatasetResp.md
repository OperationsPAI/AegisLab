# DtoAlgorithmDatasetResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | **str** | Algorithm name | [optional] 
**dataset** | **str** | Dataset name | [optional] 
**dataset_version** | **str** | Dataset version | [optional] 
**executed_count** | **int** | Number of successfully executed datapacks | [optional] 
**items** | [**List[DtoDatapackEvaluationItem]**](DtoDatapackEvaluationItem.md) | Evaluation items for each datapack | [optional] 
**total_count** | **int** | Total number of datapacks in dataset | [optional] 

## Example

```python
from rcabench.openapi.models.dto_algorithm_dataset_resp import DtoAlgorithmDatasetResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAlgorithmDatasetResp from a JSON string
dto_algorithm_dataset_resp_instance = DtoAlgorithmDatasetResp.from_json(json)
# print the JSON string representation of the object
print(DtoAlgorithmDatasetResp.to_json())

# convert the object into a dict
dto_algorithm_dataset_resp_dict = dto_algorithm_dataset_resp_instance.to_dict()
# create an instance of DtoAlgorithmDatasetResp from a dict
dto_algorithm_dataset_resp_from_dict = DtoAlgorithmDatasetResp.from_dict(dto_algorithm_dataset_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


