# DtoInjectionV2BatchDeleteReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ids** | **List[int]** | List of injection IDs to delete | [optional] 
**labels** | **List[str]** | List of label keys to match for deletion (key1:value1,key2:value2) | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_v2_batch_delete_req import DtoInjectionV2BatchDeleteReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionV2BatchDeleteReq from a JSON string
dto_injection_v2_batch_delete_req_instance = DtoInjectionV2BatchDeleteReq.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionV2BatchDeleteReq.to_json())

# convert the object into a dict
dto_injection_v2_batch_delete_req_dict = dto_injection_v2_batch_delete_req_instance.to_dict()
# create an instance of DtoInjectionV2BatchDeleteReq from a dict
dto_injection_v2_batch_delete_req_from_dict = DtoInjectionV2BatchDeleteReq.from_dict(dto_injection_v2_batch_delete_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


