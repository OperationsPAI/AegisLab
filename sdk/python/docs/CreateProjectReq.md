# CreateProjectReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**description** | **str** |  | [optional] 
**is_public** | **bool** |  | [optional] 
**name** | **str** |  | 

## Example

```python
from openapi.models.create_project_req import CreateProjectReq

# TODO update the JSON string below
json = "{}"
# create an instance of CreateProjectReq from a JSON string
create_project_req_instance = CreateProjectReq.from_json(json)
# print the JSON string representation of the object
print(CreateProjectReq.to_json())

# convert the object into a dict
create_project_req_dict = create_project_req_instance.to_dict()
# create an instance of CreateProjectReq from a dict
create_project_req_from_dict = CreateProjectReq.from_dict(create_project_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


