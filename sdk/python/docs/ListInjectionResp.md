# ListInjectionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[InjectionResp]**](InjectionResp.md) |  | [optional] 
**pagination** | [**PaginationInfo**](PaginationInfo.md) |  | [optional] 

## Example

```python
from openapi.models.list_injection_resp import ListInjectionResp

# TODO update the JSON string below
json = "{}"
# create an instance of ListInjectionResp from a JSON string
list_injection_resp_instance = ListInjectionResp.from_json(json)
# print the JSON string representation of the object
print(ListInjectionResp.to_json())

# convert the object into a dict
list_injection_resp_dict = list_injection_resp_instance.to_dict()
# create an instance of ListInjectionResp from a dict
list_injection_resp_from_dict = ListInjectionResp.from_dict(list_injection_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


