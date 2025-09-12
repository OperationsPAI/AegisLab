# DtoSubmitExecutionReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**payloads** | [**List[DtoExecutionPayload]**](DtoExecutionPayload.md) |  | 
**project_name** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_submit_execution_req import DtoSubmitExecutionReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoSubmitExecutionReq from a JSON string
dto_submit_execution_req_instance = DtoSubmitExecutionReq.from_json(json)
# print the JSON string representation of the object
print(DtoSubmitExecutionReq.to_json())

# convert the object into a dict
dto_submit_execution_req_dict = dto_submit_execution_req_instance.to_dict()
# create an instance of DtoSubmitExecutionReq from a dict
dto_submit_execution_req_from_dict = DtoSubmitExecutionReq.from_dict(dto_submit_execution_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


