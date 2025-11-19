# SubmitExecutionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**group_id** | **str** |  | [optional] 
**items** | [**List[SubmitExecutionItem]**](SubmitExecutionItem.md) |  | [optional] 

## Example

```python
from openapi.models.submit_execution_resp import SubmitExecutionResp

# TODO update the JSON string below
json = "{}"
# create an instance of SubmitExecutionResp from a JSON string
submit_execution_resp_instance = SubmitExecutionResp.from_json(json)
# print the JSON string representation of the object
print(SubmitExecutionResp.to_json())

# convert the object into a dict
submit_execution_resp_dict = submit_execution_resp_instance.to_dict()
# create an instance of SubmitExecutionResp from a dict
submit_execution_resp_from_dict = SubmitExecutionResp.from_dict(submit_execution_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


