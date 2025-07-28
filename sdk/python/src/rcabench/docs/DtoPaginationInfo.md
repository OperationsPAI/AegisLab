# DtoPaginationInfo


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**page** | **int** |  | [optional] 
**size** | **int** |  | [optional] 
**total** | **int** |  | [optional] 
**total_pages** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_pagination_info import DtoPaginationInfo

# TODO update the JSON string below
json = "{}"
# create an instance of DtoPaginationInfo from a JSON string
dto_pagination_info_instance = DtoPaginationInfo.from_json(json)
# print the JSON string representation of the object
print(DtoPaginationInfo.to_json())

# convert the object into a dict
dto_pagination_info_dict = dto_pagination_info_instance.to_dict()
# create an instance of DtoPaginationInfo from a dict
dto_pagination_info_from_dict = DtoPaginationInfo.from_dict(dto_pagination_info_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


