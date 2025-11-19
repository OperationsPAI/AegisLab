# BatchEvaluateDatapackReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**specs** | [**List[EvaluateDatapackSpec]**](EvaluateDatapackSpec.md) |  | 

## Example

```python
from openapi.models.batch_evaluate_datapack_req import BatchEvaluateDatapackReq

# TODO update the JSON string below
json = "{}"
# create an instance of BatchEvaluateDatapackReq from a JSON string
batch_evaluate_datapack_req_instance = BatchEvaluateDatapackReq.from_json(json)
# print the JSON string representation of the object
print(BatchEvaluateDatapackReq.to_json())

# convert the object into a dict
batch_evaluate_datapack_req_dict = batch_evaluate_datapack_req_instance.to_dict()
# create an instance of BatchEvaluateDatapackReq from a dict
batch_evaluate_datapack_req_from_dict = BatchEvaluateDatapackReq.from_dict(batch_evaluate_datapack_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


