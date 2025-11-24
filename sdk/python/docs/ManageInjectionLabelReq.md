# ManageInjectionLabelReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**add_labels** | [**List[LabelItem]**](LabelItem.md) | List of labels to add | [optional] 
**remove_labels** | **List[str]** | List of label keys to remove | [optional] 

## Example

```python
from rcabench.openapi.models.manage_injection_label_req import ManageInjectionLabelReq

# TODO update the JSON string below
json = "{}"
# create an instance of ManageInjectionLabelReq from a JSON string
manage_injection_label_req_instance = ManageInjectionLabelReq.from_json(json)
# print the JSON string representation of the object
print(ManageInjectionLabelReq.to_json())

# convert the object into a dict
manage_injection_label_req_dict = manage_injection_label_req_instance.to_dict()
# create an instance of ManageInjectionLabelReq from a dict
manage_injection_label_req_from_dict = ManageInjectionLabelReq.from_dict(manage_injection_label_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


