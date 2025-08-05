# DtoInjectionV2CustomLabelManageReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**add_labels** | [**List[DtoLabelItem]**](DtoLabelItem.md) | List of labels to add | [optional] 
**remove_labels** | **List[str]** | List of label keys to remove | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_v2_custom_label_manage_req import DtoInjectionV2CustomLabelManageReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionV2CustomLabelManageReq from a JSON string
dto_injection_v2_custom_label_manage_req_instance = DtoInjectionV2CustomLabelManageReq.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionV2CustomLabelManageReq.to_json())

# convert the object into a dict
dto_injection_v2_custom_label_manage_req_dict = dto_injection_v2_custom_label_manage_req_instance.to_dict()
# create an instance of DtoInjectionV2CustomLabelManageReq from a dict
dto_injection_v2_custom_label_manage_req_from_dict = DtoInjectionV2CustomLabelManageReq.from_dict(dto_injection_v2_custom_label_manage_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


