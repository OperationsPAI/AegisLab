# DtoDatasetEvaluationBatchReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DtoAlgorithmDatasetReq]**](DtoAlgorithmDatasetReq.md) |  | 

## Example

```python
from rcabench.openapi.models.dto_dataset_evaluation_batch_req import DtoDatasetEvaluationBatchReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetEvaluationBatchReq from a JSON string
dto_dataset_evaluation_batch_req_instance = DtoDatasetEvaluationBatchReq.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetEvaluationBatchReq.to_json())

# convert the object into a dict
dto_dataset_evaluation_batch_req_dict = dto_dataset_evaluation_batch_req_instance.to_dict()
# create an instance of DtoDatasetEvaluationBatchReq from a dict
dto_dataset_evaluation_batch_req_from_dict = DtoDatasetEvaluationBatchReq.from_dict(dto_dataset_evaluation_batch_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


