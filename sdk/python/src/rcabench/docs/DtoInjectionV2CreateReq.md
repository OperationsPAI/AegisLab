# DtoInjectionV2CreateReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**injections** | [**List[DtoInjectionV2CreateItem]**](DtoInjectionV2CreateItem.md) |  | 

## Example

```python
from rcabench.openapi.models.dto_injection_v2_create_req import DtoInjectionV2CreateReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionV2CreateReq from a JSON string
dto_injection_v2_create_req_instance = DtoInjectionV2CreateReq.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionV2CreateReq.to_json())

# convert the object into a dict
dto_injection_v2_create_req_dict = dto_injection_v2_create_req_instance.to_dict()
# create an instance of DtoInjectionV2CreateReq from a dict
dto_injection_v2_create_req_from_dict = DtoInjectionV2CreateReq.from_dict(dto_injection_v2_create_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


