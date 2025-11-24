# SubmitInjectionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**duplicated_count** | **int** |  | [optional] 
**group_id** | **str** |  | [optional] 
**items** | [**List[SubmitInjectionItem]**](SubmitInjectionItem.md) |  | [optional] 
**original_count** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.submit_injection_resp import SubmitInjectionResp

# TODO update the JSON string below
json = "{}"
# create an instance of SubmitInjectionResp from a JSON string
submit_injection_resp_instance = SubmitInjectionResp.from_json(json)
# print the JSON string representation of the object
print(SubmitInjectionResp.to_json())

# convert the object into a dict
submit_injection_resp_dict = submit_injection_resp_instance.to_dict()
# create an instance of SubmitInjectionResp from a dict
submit_injection_resp_from_dict = SubmitInjectionResp.from_dict(submit_injection_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


