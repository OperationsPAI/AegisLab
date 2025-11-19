# GenericResponseArrayInjectionWithIssuesResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**List[InjectionWithIssuesResp]**](InjectionWithIssuesResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from openapi.models.generic_response_array_injection_with_issues_resp import GenericResponseArrayInjectionWithIssuesResp

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseArrayInjectionWithIssuesResp from a JSON string
generic_response_array_injection_with_issues_resp_instance = GenericResponseArrayInjectionWithIssuesResp.from_json(json)
# print the JSON string representation of the object
print(GenericResponseArrayInjectionWithIssuesResp.to_json())

# convert the object into a dict
generic_response_array_injection_with_issues_resp_dict = generic_response_array_injection_with_issues_resp_instance.to_dict()
# create an instance of GenericResponseArrayInjectionWithIssuesResp from a dict
generic_response_array_injection_with_issues_resp_from_dict = GenericResponseArrayInjectionWithIssuesResp.from_dict(generic_response_array_injection_with_issues_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


