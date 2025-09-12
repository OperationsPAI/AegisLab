# DtoInjectionV2LabelManageReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**add_tags** | **List[str]** | List of tag values to add | [optional] 
**remove_tags** | **List[str]** | List of tag values to remove | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_v2_label_manage_req import DtoInjectionV2LabelManageReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionV2LabelManageReq from a JSON string
dto_injection_v2_label_manage_req_instance = DtoInjectionV2LabelManageReq.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionV2LabelManageReq.to_json())

# convert the object into a dict
dto_injection_v2_label_manage_req_dict = dto_injection_v2_label_manage_req_instance.to_dict()
# create an instance of DtoInjectionV2LabelManageReq from a dict
dto_injection_v2_label_manage_req_from_dict = DtoInjectionV2LabelManageReq.from_dict(dto_injection_v2_label_manage_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


