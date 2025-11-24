# InjectionNoIssuesResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**datapack_id** | **int** |  | [optional] 
**datapack_name** | **str** |  | [optional] 
**engine_config** | [**ChaosNode**](ChaosNode.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.injection_no_issues_resp import InjectionNoIssuesResp

# TODO update the JSON string below
json = "{}"
# create an instance of InjectionNoIssuesResp from a JSON string
injection_no_issues_resp_instance = InjectionNoIssuesResp.from_json(json)
# print the JSON string representation of the object
print(InjectionNoIssuesResp.to_json())

# convert the object into a dict
injection_no_issues_resp_dict = injection_no_issues_resp_instance.to_dict()
# create an instance of InjectionNoIssuesResp from a dict
injection_no_issues_resp_from_dict = InjectionNoIssuesResp.from_dict(injection_no_issues_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


