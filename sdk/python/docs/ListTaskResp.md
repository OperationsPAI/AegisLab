# ListTaskResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[TaskResp]**](TaskResp.md) |  | [optional] 
**pagination** | [**PaginationInfo**](PaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.list_task_resp import ListTaskResp

# TODO update the JSON string below
json = "{}"
# create an instance of ListTaskResp from a JSON string
list_task_resp_instance = ListTaskResp.from_json(json)
# print the JSON string representation of the object
print(ListTaskResp.to_json())

# convert the object into a dict
list_task_resp_dict = list_task_resp_instance.to_dict()
# create an instance of ListTaskResp from a dict
list_task_resp_from_dict = ListTaskResp.from_dict(list_task_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


