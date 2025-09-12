# DtoInjectionV2CreateResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_count** | **int** |  | [optional] 
**created_items** | [**List[DtoInjectionV2Response]**](DtoInjectionV2Response.md) |  | [optional] 
**failed_count** | **int** |  | [optional] 
**failed_items** | [**List[DtoInjectionCreateError]**](DtoInjectionCreateError.md) |  | [optional] 
**message** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_v2_create_response import DtoInjectionV2CreateResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionV2CreateResponse from a JSON string
dto_injection_v2_create_response_instance = DtoInjectionV2CreateResponse.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionV2CreateResponse.to_json())

# convert the object into a dict
dto_injection_v2_create_response_dict = dto_injection_v2_create_response_instance.to_dict()
# create an instance of DtoInjectionV2CreateResponse from a dict
dto_injection_v2_create_response_from_dict = DtoInjectionV2CreateResponse.from_dict(dto_injection_v2_create_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


