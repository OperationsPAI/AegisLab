# DtoEvaluationListResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**results** | [**List[DtoEvaluationItem]**](DtoEvaluationItem.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_evaluation_list_resp import DtoEvaluationListResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoEvaluationListResp from a JSON string
dto_evaluation_list_resp_instance = DtoEvaluationListResp.from_json(json)
# print the JSON string representation of the object
print(DtoEvaluationListResp.to_json())

# convert the object into a dict
dto_evaluation_list_resp_dict = dto_evaluation_list_resp_instance.to_dict()
# create an instance of DtoEvaluationListResp from a dict
dto_evaluation_list_resp_from_dict = DtoEvaluationListResp.from_dict(dto_evaluation_list_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


