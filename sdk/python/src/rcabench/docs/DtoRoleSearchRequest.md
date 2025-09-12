# DtoRoleSearchRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | [**DtoDateRange**](DtoDateRange.md) | Common filters shortcuts | [optional] 
**description_pattern** | **str** | Description fuzzy match | [optional] 
**display_name_pattern** | **str** | Display name fuzzy match | [optional] 
**exclude_fields** | **List[str]** |  | [optional] 
**filters** | [**List[DtoSearchFilter]**](DtoSearchFilter.md) | Filters | [optional] 
**include** | **List[str]** | Include related entities | [optional] 
**include_fields** | **List[str]** | Include/Exclude fields | [optional] 
**is_active** | **bool** |  | [optional] 
**is_system** | **bool** | Whether system role | [optional] 
**keyword** | **str** | Search keyword (for general text search) | [optional] 
**name_pattern** | **str** | Role-specific filter shortcuts | [optional] 
**page** | **int** | Pagination | [optional] 
**permission_ids** | **List[int]** | Permission ID filter | [optional] 
**project_id** | **int** |  | [optional] 
**size** | **int** |  | [optional] 
**sort** | [**List[DtoSortOption]**](DtoSortOption.md) | Sort | [optional] 
**status** | **List[int]** |  | [optional] 
**types** | **List[str]** | Role type filter | [optional] 
**updated_at** | [**DtoDateRange**](DtoDateRange.md) |  | [optional] 
**user_count** | [**DtoNumberRange**](DtoNumberRange.md) | User count range | [optional] 
**user_id** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_role_search_request import DtoRoleSearchRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoRoleSearchRequest from a JSON string
dto_role_search_request_instance = DtoRoleSearchRequest.from_json(json)
# print the JSON string representation of the object
print(DtoRoleSearchRequest.to_json())

# convert the object into a dict
dto_role_search_request_dict = dto_role_search_request_instance.to_dict()
# create an instance of DtoRoleSearchRequest from a dict
dto_role_search_request_from_dict = DtoRoleSearchRequest.from_dict(dto_role_search_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


