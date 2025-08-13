# DtoLabelCreateReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**category** | **str** |  | [optional] 
**color** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**key** | **str** |  | 
**value** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_label_create_req import DtoLabelCreateReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoLabelCreateReq from a JSON string
dto_label_create_req_instance = DtoLabelCreateReq.from_json(json)
# print the JSON string representation of the object
print(DtoLabelCreateReq.to_json())

# convert the object into a dict
dto_label_create_req_dict = dto_label_create_req_instance.to_dict()
# create an instance of DtoLabelCreateReq from a dict
dto_label_create_req_from_dict = DtoLabelCreateReq.from_dict(dto_label_create_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


