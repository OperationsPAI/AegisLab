# SubmitExecutionReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**labels** | [**List[LabelItem]**](LabelItem.md) |  | [optional] 
**project_name** | **str** |  | 
**specs** | [**List[ExecutionSpec]**](ExecutionSpec.md) |  | 

## Example

```python
from rcabench.openapi.models.submit_execution_req import SubmitExecutionReq

# TODO update the JSON string below
json = "{}"
# create an instance of SubmitExecutionReq from a JSON string
submit_execution_req_instance = SubmitExecutionReq.from_json(json)
# print the JSON string representation of the object
print(SubmitExecutionReq.to_json())

# convert the object into a dict
submit_execution_req_dict = submit_execution_req_instance.to_dict()
# create an instance of SubmitExecutionReq from a dict
submit_execution_req_from_dict = SubmitExecutionReq.from_dict(submit_execution_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


