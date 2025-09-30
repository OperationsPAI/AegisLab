# DtoInjectionV2BatchDeleteResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**cascade_deleted** | [**DtoInjectionV2CascadeDeleteStats**](DtoInjectionV2CascadeDeleteStats.md) |  | [optional] 
**failed_count** | **int** |  | [optional] 
**failed_items** | [**List[DtoInjectionV2DeleteError]**](DtoInjectionV2DeleteError.md) |  | [optional] 
**message** | **str** |  | [optional] 
**success_count** | **int** |  | [optional] 
**success_items** | [**List[DtoInjectionV2DeletedItem]**](DtoInjectionV2DeletedItem.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_v2_batch_delete_response import DtoInjectionV2BatchDeleteResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionV2BatchDeleteResponse from a JSON string
dto_injection_v2_batch_delete_response_instance = DtoInjectionV2BatchDeleteResponse.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionV2BatchDeleteResponse.to_json())

# convert the object into a dict
dto_injection_v2_batch_delete_response_dict = dto_injection_v2_batch_delete_response_instance.to_dict()
# create an instance of DtoInjectionV2BatchDeleteResponse from a dict
dto_injection_v2_batch_delete_response_from_dict = DtoInjectionV2BatchDeleteResponse.from_dict(dto_injection_v2_batch_delete_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


