# DtoDatapackEvaluationBatchReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DtoAlgorithmDatapackReq]**](DtoAlgorithmDatapackReq.md) |  | 

## Example

```python
from rcabench.openapi.models.dto_datapack_evaluation_batch_req import DtoDatapackEvaluationBatchReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatapackEvaluationBatchReq from a JSON string
dto_datapack_evaluation_batch_req_instance = DtoDatapackEvaluationBatchReq.from_json(json)
# print the JSON string representation of the object
print(DtoDatapackEvaluationBatchReq.to_json())

# convert the object into a dict
dto_datapack_evaluation_batch_req_dict = dto_datapack_evaluation_batch_req_instance.to_dict()
# create an instance of DtoDatapackEvaluationBatchReq from a dict
dto_datapack_evaluation_batch_req_from_dict = DtoDatapackEvaluationBatchReq.from_dict(dto_datapack_evaluation_batch_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


