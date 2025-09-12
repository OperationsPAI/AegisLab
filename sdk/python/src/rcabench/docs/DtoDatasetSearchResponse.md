# DtoDatasetSearchResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DtoDatasetV2Response]**](DtoDatasetV2Response.md) | Result list | [optional] 
**pagination** | [**DtoPaginationInfo**](DtoPaginationInfo.md) | Pagination info | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_search_response import DtoDatasetSearchResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetSearchResponse from a JSON string
dto_dataset_search_response_instance = DtoDatasetSearchResponse.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetSearchResponse.to_json())

# convert the object into a dict
dto_dataset_search_response_dict = dto_dataset_search_response_instance.to_dict()
# create an instance of DtoDatasetSearchResponse from a dict
dto_dataset_search_response_from_dict = DtoDatasetSearchResponse.from_dict(dto_dataset_search_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


