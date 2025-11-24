# InjectionWithIssuesResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**abnormal_avg_duration** | **float** |  | [optional] 
**abnormal_p99** | **float** |  | [optional] 
**abnormal_succ_rate** | **float** |  | [optional] 
**datapack_id** | **int** |  | [optional] 
**datapack_name** | **str** |  | [optional] 
**engine_config** | [**ChaosNode**](ChaosNode.md) |  | [optional] 
**issues** | **str** |  | [optional] 
**normal_avg_duration** | **float** |  | [optional] 
**normal_p99** | **float** |  | [optional] 
**normal_succ_rate** | **float** |  | [optional] 

## Example

```python
from rcabench.openapi.models.injection_with_issues_resp import InjectionWithIssuesResp

# TODO update the JSON string below
json = "{}"
# create an instance of InjectionWithIssuesResp from a JSON string
injection_with_issues_resp_instance = InjectionWithIssuesResp.from_json(json)
# print the JSON string representation of the object
print(InjectionWithIssuesResp.to_json())

# convert the object into a dict
injection_with_issues_resp_dict = injection_with_issues_resp_instance.to_dict()
# create an instance of InjectionWithIssuesResp from a dict
injection_with_issues_resp_from_dict = InjectionWithIssuesResp.from_dict(injection_with_issues_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


