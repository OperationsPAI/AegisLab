# BatchEvaluateDatapackResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**failed_count** | **int** |  | [optional] 
**failed_items** | **List[str]** |  | [optional] 
**success_count** | **int** |  | [optional] 
**success_items** | [**List[EvaluateDatapackItem]**](EvaluateDatapackItem.md) |  | [optional] 

## Example

```python
from openapi.models.batch_evaluate_datapack_resp import BatchEvaluateDatapackResp

# TODO update the JSON string below
json = "{}"
# create an instance of BatchEvaluateDatapackResp from a JSON string
batch_evaluate_datapack_resp_instance = BatchEvaluateDatapackResp.from_json(json)
# print the JSON string representation of the object
print(BatchEvaluateDatapackResp.to_json())

# convert the object into a dict
batch_evaluate_datapack_resp_dict = batch_evaluate_datapack_resp_instance.to_dict()
# create an instance of BatchEvaluateDatapackResp from a dict
batch_evaluate_datapack_resp_from_dict = BatchEvaluateDatapackResp.from_dict(batch_evaluate_datapack_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


