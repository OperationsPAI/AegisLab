# ListExecutionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[ExecutionResp]**](ExecutionResp.md) |  | [optional] 
**pagination** | [**PaginationInfo**](PaginationInfo.md) |  | [optional] 

## Example

```python
from openapi.models.list_execution_resp import ListExecutionResp

# TODO update the JSON string below
json = "{}"
# create an instance of ListExecutionResp from a JSON string
list_execution_resp_instance = ListExecutionResp.from_json(json)
# print the JSON string representation of the object
print(ListExecutionResp.to_json())

# convert the object into a dict
list_execution_resp_dict = list_execution_resp_instance.to_dict()
# create an instance of ListExecutionResp from a dict
list_execution_resp_from_dict = ListExecutionResp.from_dict(list_execution_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


