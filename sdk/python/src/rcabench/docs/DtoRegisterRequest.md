# DtoRegisterRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**email** | **str** |  | 
**full_name** | **str** |  | 
**password** | **str** |  | 
**phone** | **str** |  | [optional] 
**username** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_register_request import DtoRegisterRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoRegisterRequest from a JSON string
dto_register_request_instance = DtoRegisterRequest.from_json(json)
# print the JSON string representation of the object
print(DtoRegisterRequest.to_json())

# convert the object into a dict
dto_register_request_dict = dto_register_request_instance.to_dict()
# create an instance of DtoRegisterRequest from a dict
dto_register_request_from_dict = DtoRegisterRequest.from_dict(dto_register_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


