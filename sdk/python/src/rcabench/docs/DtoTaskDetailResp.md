# DtoTaskDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**logs** | **List[str]** |  | [optional] 
**task** | [**DtoTaskItem**](DtoTaskItem.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_task_detail_resp import DtoTaskDetailResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoTaskDetailResp from a JSON string
dto_task_detail_resp_instance = DtoTaskDetailResp.from_json(json)
# print the JSON string representation of the object
print(DtoTaskDetailResp.to_json())

# convert the object into a dict
dto_task_detail_resp_dict = dto_task_detail_resp_instance.to_dict()
# create an instance of DtoTaskDetailResp from a dict
dto_task_detail_resp_from_dict = DtoTaskDetailResp.from_dict(dto_task_detail_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


