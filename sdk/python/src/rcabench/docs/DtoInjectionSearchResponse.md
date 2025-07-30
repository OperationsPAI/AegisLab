# DtoInjectionSearchResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DtoInjectionV2Response]**](DtoInjectionV2Response.md) |  | [optional] 
**pagination** | [**DtoPaginationInfo**](DtoPaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_search_response import DtoInjectionSearchResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionSearchResponse from a JSON string
dto_injection_search_response_instance = DtoInjectionSearchResponse.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionSearchResponse.to_json())

# convert the object into a dict
dto_injection_search_response_dict = dto_injection_search_response_instance.to_dict()
# create an instance of DtoInjectionSearchResponse from a dict
dto_injection_search_response_from_dict = DtoInjectionSearchResponse.from_dict(dto_injection_search_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


