# DtoGenericResponseDtoDatasetSearchResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**DtoDatasetSearchResponse**](DtoDatasetSearchResponse.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_dataset_search_response import DtoGenericResponseDtoDatasetSearchResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoDatasetSearchResponse from a JSON string
dto_generic_response_dto_dataset_search_response_instance = DtoGenericResponseDtoDatasetSearchResponse.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoDatasetSearchResponse.to_json())

# convert the object into a dict
dto_generic_response_dto_dataset_search_response_dict = dto_generic_response_dto_dataset_search_response_instance.to_dict()
# create an instance of DtoGenericResponseDtoDatasetSearchResponse from a dict
dto_generic_response_dto_dataset_search_response_from_dict = DtoGenericResponseDtoDatasetSearchResponse.from_dict(dto_generic_response_dto_dataset_search_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


