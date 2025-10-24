# DtoInfoPayloadTemplate


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**msg** | **str** |  | [optional] 
**status** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_info_payload_template import DtoInfoPayloadTemplate

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInfoPayloadTemplate from a JSON string
dto_info_payload_template_instance = DtoInfoPayloadTemplate.from_json(json)
# print the JSON string representation of the object
print(DtoInfoPayloadTemplate.to_json())

# convert the object into a dict
dto_info_payload_template_dict = dto_info_payload_template_instance.to_dict()
# create an instance of DtoInfoPayloadTemplate from a dict
dto_info_payload_template_from_dict = DtoInfoPayloadTemplate.from_dict(dto_info_payload_template_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


