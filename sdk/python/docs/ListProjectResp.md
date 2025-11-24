# ListProjectResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[ProjectResp]**](ProjectResp.md) |  | [optional] 
**pagination** | [**PaginationInfo**](PaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.list_project_resp import ListProjectResp

# TODO update the JSON string below
json = "{}"
# create an instance of ListProjectResp from a JSON string
list_project_resp_instance = ListProjectResp.from_json(json)
# print the JSON string representation of the object
print(ListProjectResp.to_json())

# convert the object into a dict
list_project_resp_dict = list_project_resp_instance.to_dict()
# create an instance of ListProjectResp from a dict
list_project_resp_from_dict = ListProjectResp.from_dict(list_project_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


