# GenericResponseBatchEvaluateDatapackResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**BatchEvaluateDatapackResp**](BatchEvaluateDatapackResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.generic_response_batch_evaluate_datapack_resp import GenericResponseBatchEvaluateDatapackResp

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseBatchEvaluateDatapackResp from a JSON string
generic_response_batch_evaluate_datapack_resp_instance = GenericResponseBatchEvaluateDatapackResp.from_json(json)
# print the JSON string representation of the object
print(GenericResponseBatchEvaluateDatapackResp.to_json())

# convert the object into a dict
generic_response_batch_evaluate_datapack_resp_dict = generic_response_batch_evaluate_datapack_resp_instance.to_dict()
# create an instance of GenericResponseBatchEvaluateDatapackResp from a dict
generic_response_batch_evaluate_datapack_resp_from_dict = GenericResponseBatchEvaluateDatapackResp.from_dict(generic_response_batch_evaluate_datapack_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


